package routeranalytics

import (
	"encoding/json"
	"sort"
	"strings"
	"sync"

	"github.com/rfwlab/rfw/v1/core"
)

type Options struct {
	// Normalize maps a navigation path to a bucket used for analytics.
	// Defaults to stripping query/hash fragments and ensuring a leading slash.
	Normalize func(string) string
	// PrefetchLimit limits how many predicted routes are sent per navigation.
	// Set to a negative value to disable prefetch hints (default 3).
	PrefetchLimit int
	// PrefetchThreshold drops predictions below this probability when
	// enqueuing prefetch hints. Values <= 0 fall back to 0.2.
	PrefetchThreshold float64
	// Channel customizes the netcode channel used for hints.
	Channel string
}

type TransitionProbability struct {
	From        string
	To          string
	Count       int
	Probability float64
}

type Plugin struct {
	opts     Options
	tracker  *transitionTracker
	prefetch prefetcher
}

var (
	defaultMu       sync.RWMutex
	defaultInstance *Plugin
)

func New(opts Options) *Plugin {
	if opts.Normalize == nil {
		opts.Normalize = defaultNormalize
	}
	if opts.PrefetchLimit < 0 {
		opts.PrefetchLimit = 0
	} else if opts.PrefetchLimit == 0 {
		opts.PrefetchLimit = 3
	}
	if opts.PrefetchThreshold <= 0 {
		opts.PrefetchThreshold = 0.2
	}
	if opts.Channel == "" {
		opts.Channel = "RouterPrefetch"
	}
	plugin := &Plugin{
		opts:    opts,
		tracker: newTransitionTracker(),
	}
	if opts.PrefetchLimit > 0 {
		plugin.prefetch = newPrefetcher(opts.Channel)
	} else {
		plugin.prefetch = noopPrefetcher{}
	}
	return plugin
}

func (p *Plugin) Name() string {
	return "routeranalytics"
}

func (p *Plugin) Build(json.RawMessage) error { return nil }

func (p *Plugin) Install(a *core.App) {
	if a == nil {
		return
	}
	setDefaultInstance(p)
	a.RegisterRouter(func(path string) {
		p.handleNavigation(path)
	})
}

func (p *Plugin) TransitionProbabilities(from string) []TransitionProbability {
	if p == nil {
		return nil
	}
	probs := p.tracker.probabilities(from)
	if len(probs) == 0 {
		return nil
	}
	out := make([]TransitionProbability, len(probs))
	copy(out, probs)
	return out
}

func (p *Plugin) MostLikelyNext(from string, limit int) []TransitionProbability {
	if p == nil || limit <= 0 {
		return nil
	}
	probs := p.tracker.mostLikelyNext(from, limit)
	if len(probs) == 0 {
		return nil
	}
	out := make([]TransitionProbability, len(probs))
	copy(out, probs)
	return out
}

func (p *Plugin) Reset() {
	if p == nil {
		return
	}
	p.tracker.reset()
	if p.prefetch != nil {
		p.prefetch.reset()
	}
}

func TransitionProbabilities(from string) []TransitionProbability {
	if inst := getDefaultInstance(); inst != nil {
		return inst.TransitionProbabilities(from)
	}
	return nil
}

func MostLikelyNext(from string, limit int) []TransitionProbability {
	if inst := getDefaultInstance(); inst != nil {
		return inst.MostLikelyNext(from, limit)
	}
	return nil
}

func Reset() {
	if inst := getDefaultInstance(); inst != nil {
		inst.Reset()
	}
}

type transitionTracker struct {
	mu          sync.RWMutex
	transitions map[string]map[string]int
	totals      map[string]int
	last        string
}

func newTransitionTracker() *transitionTracker {
	return &transitionTracker{
		transitions: make(map[string]map[string]int),
		totals:      make(map[string]int),
	}
}

func (t *transitionTracker) visit(routeID string) {
	if routeID == "" {
		return
	}
	t.mu.Lock()
	if t.last != "" {
		t.recordLocked(t.last, routeID)
	}
	t.last = routeID
	t.mu.Unlock()
}

func (t *transitionTracker) recordLocked(from, to string) {
	if from == "" || to == "" {
		return
	}
	if t.transitions[from] == nil {
		t.transitions[from] = make(map[string]int)
	}
	t.transitions[from][to]++
	t.totals[from]++
}

func (t *transitionTracker) reset() {
	t.mu.Lock()
	t.transitions = make(map[string]map[string]int)
	t.totals = make(map[string]int)
	t.last = ""
	t.mu.Unlock()
}

func (t *transitionTracker) probabilities(from string) []TransitionProbability {
	t.mu.RLock()
	total := t.totals[from]
	if total == 0 {
		t.mu.RUnlock()
		return nil
	}
	counts := t.transitions[from]
	result := make([]TransitionProbability, 0, len(counts))
	for to, count := range counts {
		result = append(result, TransitionProbability{
			From:        from,
			To:          to,
			Count:       count,
			Probability: float64(count) / float64(total),
		})
	}
	t.mu.RUnlock()
	sort.Slice(result, func(i, j int) bool {
		if result[i].Probability == result[j].Probability {
			return result[i].To < result[j].To
		}
		return result[i].Probability > result[j].Probability
	})
	return result
}

func (t *transitionTracker) mostLikelyNext(from string, limit int) []TransitionProbability {
	if limit <= 0 {
		return nil
	}
	probs := t.probabilities(from)
	if len(probs) == 0 {
		return nil
	}
	if limit >= len(probs) {
		out := make([]TransitionProbability, len(probs))
		copy(out, probs)
		return out
	}
	out := make([]TransitionProbability, limit)
	copy(out, probs[:limit])
	return out
}

func (p *Plugin) handleNavigation(path string) {
	normalized := strings.TrimSpace(path)
	if p.opts.Normalize != nil {
		normalized = p.opts.Normalize(path)
	}
	if normalized == "" {
		return
	}
	p.tracker.visit(normalized)
	p.enqueuePrefetch(normalized)
}

func (p *Plugin) enqueuePrefetch(current string) {
	if p.opts.PrefetchLimit <= 0 || p.prefetch == nil {
		return
	}
	predictions := p.tracker.mostLikelyNext(current, p.opts.PrefetchLimit)
	if len(predictions) == 0 {
		return
	}
	threshold := p.opts.PrefetchThreshold
	if threshold < 0 {
		threshold = 0
	}
	resources := make([]string, 0, len(predictions))
	for _, prob := range predictions {
		if prob.Probability < threshold {
			continue
		}
		resources = append(resources, prob.To)
	}
	if len(resources) == 0 {
		return
	}
	p.prefetch.request(resources)
}

func defaultNormalize(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	if idx := strings.IndexAny(trimmed, "?#"); idx >= 0 {
		trimmed = trimmed[:idx]
	}
	if trimmed == "" {
		return ""
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return trimmed
}

// NormalizePath applica la normalizzazione predefinita (trim, rimozione di query/hash e slash iniziale).
func NormalizePath(path string) string {
	return defaultNormalize(path)
}

type prefetcher interface {
	request(resources []string)
	reset()
}

type noopPrefetcher struct{}

func (noopPrefetcher) request([]string) {}

func (noopPrefetcher) reset() {}

func setDefaultInstance(p *Plugin) {
	defaultMu.Lock()
	defaultInstance = p
	defaultMu.Unlock()
}

func getDefaultInstance() *Plugin {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultInstance
}
