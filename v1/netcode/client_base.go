package netcode

import "sync"

type sendFunc func(string, any)
type registerFunc func(string, func(map[string]any))

type snapshot[T any] struct {
	tick  int64
	state T
}

// Client maintains a command queue and interpolates server snapshots.
type Client[T any] struct {
	name   string
	send   sendFunc
	decode func(map[string]any) T
	interp func(T, T, float64) T
	snaps  []snapshot[T]
	cmds   []any
	mu     sync.Mutex
}

func newClient[T any](name string, decode func(map[string]any) T, interp func(T, T, float64) T, send sendFunc, register registerFunc) *Client[T] {
	c := &Client[T]{name: name, send: send, decode: decode, interp: interp}
	register(name, c.handle)
	return c
}

func (c *Client[T]) handle(payload map[string]any) {
	var tick int64
	switch v := payload["tick"].(type) {
	case float64:
		tick = int64(v)
	case int64:
		tick = v
	case int:
		tick = int64(v)
	}
	m, _ := payload["state"].(map[string]any)
	s := snapshot[T]{tick: tick, state: c.decode(m)}
	c.mu.Lock()
	c.snaps = append(c.snaps, s)
	if len(c.snaps) > 2 {
		c.snaps = c.snaps[len(c.snaps)-2:]
	}
	c.mu.Unlock()
}

// Enqueue adds a command to be sent on the next flush.
func (c *Client[T]) Enqueue(cmd any) {
	c.mu.Lock()
	c.cmds = append(c.cmds, cmd)
	c.mu.Unlock()
}

// Flush sends queued commands with the associated tick.
func (c *Client[T]) Flush(tick int64) {
	c.mu.Lock()
	cmds := c.cmds
	c.cmds = nil
	c.mu.Unlock()
	if len(cmds) == 0 {
		return
	}
	c.send(c.name, map[string]any{"tick": tick, "commands": cmds})
}

// State returns the interpolated snapshot for the given tick.
func (c *Client[T]) State(now int64) (out T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.snaps) == 0 {
		return
	}
	if len(c.snaps) == 1 || now >= c.snaps[1].tick {
		return c.snaps[len(c.snaps)-1].state
	}
	a, b := c.snaps[0], c.snaps[1]
	if now <= a.tick {
		return a.state
	}
	alpha := float64(now-a.tick) / float64(b.tick-a.tick)
	return c.interp(a.state, b.state, alpha)
}
