package state

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrResourcePending is returned by Read while a resource is loading.
	ErrResourcePending = errors.New("state: resource pending")
	// ErrNilResource is returned when Read is called on a nil resource.
	ErrNilResource = errors.New("state: nil resource")
)

// ResourceStatus describes the current resource state.
type ResourceStatus string

const (
	// ResourceIdle indicates that loading has not started.
	ResourceIdle ResourceStatus = "idle"
	// ResourceLoading indicates that a load is in progress.
	ResourceLoading ResourceStatus = "loading"
	// ResourceReady indicates that a value is available.
	ResourceReady ResourceStatus = "ready"
	// ResourceError indicates that loading failed.
	ResourceError ResourceStatus = "error"
)

type resourceConfig struct {
	key       string
	ttl       time.Duration
	immediate bool
}

// ResourceOption configures a Resource.
type ResourceOption func(*resourceConfig)

// WithResourceKey enables request deduplication and caching for key.
func WithResourceKey(key string) ResourceOption {
	return func(config *resourceConfig) { config.key = key }
}

// WithResourceTTL expires a keyed cache entry after ttl.
func WithResourceTTL(ttl time.Duration) ResourceOption {
	return func(config *resourceConfig) { config.ttl = ttl }
}

// WithoutImmediateLoad leaves a resource idle until Load is called.
func WithoutImmediateLoad() ResourceOption {
	return func(config *resourceConfig) { config.immediate = false }
}

type resourceCacheEntry struct {
	value   any
	expires time.Time
}

type resourceFlight struct {
	done    chan struct{}
	cancel  context.CancelFunc
	value   any
	err     error
	waiters int
	closed  bool
}

var resourceShared = struct {
	sync.Mutex
	cache   map[string]resourceCacheEntry
	flights map[string]*resourceFlight
}{
	cache:   make(map[string]resourceCacheEntry),
	flights: make(map[string]*resourceFlight),
}

// ClearResourceCache removes a shared resource cache entry.
func ClearResourceCache(key string) {
	resourceShared.Lock()
	delete(resourceShared.cache, key)
	resourceShared.Unlock()
}

// Resource wraps cancellable asynchronous data in reactive signals.
type Resource[T any] struct {
	mu         sync.Mutex
	loader     func(context.Context) (T, error)
	key        string
	ttl        time.Duration
	generation uint64
	cancel     context.CancelFunc
	closed     bool

	value  *Signal[T]
	status *Signal[ResourceStatus]
	err    *Signal[error]
}

// NewResource creates a resource and starts loading by default.
func NewResource[T any](loader func(context.Context) (T, error), opts ...ResourceOption) *Resource[T] {
	config := resourceConfig{immediate: true}
	for _, opt := range opts {
		opt(&config)
	}
	resource := &Resource[T]{
		loader: loader,
		key:    config.key,
		ttl:    config.ttl,
		value:  NewSignal(*new(T)),
		status: NewSignal(ResourceIdle),
		err:    NewSignal[error](nil),
	}
	if config.immediate {
		resource.Load(context.Background())
	}
	return resource
}

// Load starts or joins the current keyed request.
func (r *Resource[T]) Load(ctx context.Context) {
	if r == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}

	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	r.generation++
	generation := r.generation
	if r.cancel != nil {
		r.cancel()
	}
	loadCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	loader := r.loader
	key := r.key
	r.mu.Unlock()

	if cached, ok := loadResourceCache[T](key); ok {
		r.finish(generation, cached, nil)
		return
	}
	if loader == nil {
		var zero T
		r.finish(generation, zero, errors.New("state: nil resource loader"))
		return
	}

	r.err.Set(nil)
	r.status.Set(ResourceLoading)
	flight := acquireResourceFlight(loadCtx, key, loader)
	go func() {
		value, err := waitResourceFlight[T](loadCtx, flight)
		r.finish(generation, value, err)
	}()
}

func loadResourceCache[T any](key string) (T, bool) {
	var zero T
	if key == "" {
		return zero, false
	}
	resourceShared.Lock()
	defer resourceShared.Unlock()
	entry, ok := resourceShared.cache[key]
	if !ok {
		return zero, false
	}
	if !entry.expires.IsZero() && time.Now().After(entry.expires) {
		delete(resourceShared.cache, key)
		return zero, false
	}
	value, ok := entry.value.(T)
	return value, ok
}

func acquireResourceFlight[T any](ctx context.Context, key string, loader func(context.Context) (T, error)) *resourceFlight {
	resourceShared.Lock()
	if key != "" {
		if flight := resourceShared.flights[key]; flight != nil {
			flight.waiters++
			resourceShared.Unlock()
			return flight
		}
	}
	base := context.WithoutCancel(ctx)
	flightCtx, cancel := context.WithCancel(base)
	flight := &resourceFlight{done: make(chan struct{}), cancel: cancel, waiters: 1}
	if key != "" {
		resourceShared.flights[key] = flight
	}
	resourceShared.Unlock()

	go func() {
		value, err := runResourceLoader(flightCtx, loader)
		resourceShared.Lock()
		flight.value = value
		flight.err = err
		flight.closed = true
		if key != "" && resourceShared.flights[key] == flight {
			delete(resourceShared.flights, key)
		}
		close(flight.done)
		resourceShared.Unlock()
	}()
	return flight
}

func runResourceLoader[T any](ctx context.Context, loader func(context.Context) (T, error)) (value T, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("state: resource loader panic: %v", recovered)
		}
	}()
	return loader(ctx)
}

func waitResourceFlight[T any](ctx context.Context, flight *resourceFlight) (T, error) {
	var zero T
	select {
	case <-flight.done:
		resourceShared.Lock()
		value := flight.value
		err := flight.err
		flight.waiters--
		resourceShared.Unlock()
		typed, ok := value.(T)
		if !ok && err == nil {
			return zero, fmt.Errorf("state: resource result type mismatch")
		}
		return typed, err
	case <-ctx.Done():
		resourceShared.Lock()
		flight.waiters--
		if flight.waiters == 0 && !flight.closed {
			flight.cancel()
		}
		resourceShared.Unlock()
		return zero, ctx.Err()
	}
}

func (r *Resource[T]) finish(generation uint64, value T, err error) {
	r.mu.Lock()
	if r.closed || generation != r.generation {
		r.mu.Unlock()
		return
	}
	key := r.key
	ttl := r.ttl
	r.cancel = nil
	r.mu.Unlock()

	if err != nil {
		r.err.Set(err)
		r.status.Set(ResourceError)
		return
	}
	if key != "" {
		entry := resourceCacheEntry{value: value}
		if ttl > 0 {
			entry.expires = time.Now().Add(ttl)
		}
		resourceShared.Lock()
		resourceShared.cache[key] = entry
		resourceShared.Unlock()
	}
	Batch(func() {
		r.value.Set(value)
		r.err.Set(nil)
		r.status.Set(ResourceReady)
	})
}

// Reload clears the keyed cache and starts a new load.
func (r *Resource[T]) Reload(ctx context.Context) {
	if r == nil {
		return
	}
	ClearResourceCache(r.key)
	r.Load(ctx)
}

// Mutate replaces the value without loading.
func (r *Resource[T]) Mutate(value T) {
	if r == nil {
		return
	}
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	r.generation++
	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}
	key := r.key
	ttl := r.ttl
	r.mu.Unlock()

	if key != "" {
		entry := resourceCacheEntry{value: value}
		if ttl > 0 {
			entry.expires = time.Now().Add(ttl)
		}
		resourceShared.Lock()
		resourceShared.cache[key] = entry
		resourceShared.Unlock()
	}
	Batch(func() {
		r.value.Set(value)
		r.err.Set(nil)
		r.status.Set(ResourceReady)
	})
}

// Invalidate clears the cache and returns the resource to idle.
func (r *Resource[T]) Invalidate() {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.generation++
	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}
	key := r.key
	r.mu.Unlock()
	ClearResourceCache(key)
	r.err.Set(nil)
	r.status.Set(ResourceIdle)
}

// Read returns the value, the load error, or ErrResourcePending.
func (r *Resource[T]) Read() (T, error) {
	if r == nil {
		var zero T
		return zero, ErrNilResource
	}
	switch r.status.Get() {
	case ResourceReady:
		return r.value.Get(), nil
	case ResourceError:
		return r.value.Get(), r.err.Get()
	default:
		var zero T
		return zero, ErrResourcePending
	}
}

// Value returns the latest successful value.
func (r *Resource[T]) Value() T {
	if r == nil {
		var zero T
		return zero
	}
	return r.value.Get()
}

// Status returns the current reactive status.
func (r *Resource[T]) Status() ResourceStatus {
	if r == nil {
		return ResourceIdle
	}
	return r.status.Get()
}

// Error returns the current load error.
func (r *Resource[T]) Error() error {
	if r == nil {
		return ErrNilResource
	}
	return r.err.Get()
}

// Loading reports whether a load is in progress.
func (r *Resource[T]) Loading() bool { return r.Status() == ResourceLoading }

// Close cancels pending work and prevents future loads.
func (r *Resource[T]) Close() {
	if r == nil {
		return
	}
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	r.closed = true
	r.generation++
	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}
	r.mu.Unlock()
	r.status.Set(ResourceIdle)
}
