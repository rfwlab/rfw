package devtools

import (
	"encoding/json"
	"sync"
)

type node struct {
	ID            int             `json:"id"`
	Kind          string          `json:"kind"`
	Name          string          `json:"name"`
	Time          float64         `json:"time"`
	Average       float64         `json:"average,omitempty"`
	Total         float64         `json:"total,omitempty"`
	Path          string          `json:"path"`
	Owner         string          `json:"owner,omitempty"`
	Host          string          `json:"hostComponent,omitempty"`
	Updates       int             `json:"updates,omitempty"`
	Props         map[string]any  `json:"props,omitempty"`
	Slots         map[string]any  `json:"slots,omitempty"`
	Signals       map[string]any  `json:"signals,omitempty"`
	StoreBindings []storeBinding  `json:"storeBindings,omitempty"`
	Store         *storeSnapshot  `json:"store,omitempty"`
	Children      []*node         `json:"children,omitempty"`
	Timeline      []timelineEntry `json:"timeline,omitempty"`
}

type timelineEntry struct {
	Kind     string  `json:"kind"`
	At       int64   `json:"at"`
	Duration float64 `json:"duration,omitempty"`
}

type storeSnapshot struct {
	Module string         `json:"module"`
	Name   string         `json:"name"`
	State  map[string]any `json:"state,omitempty"`
}

var (
	mu     sync.RWMutex
	roots  []*node
	nodes  = map[string]*node{}
	nextID int
)

func addComponent(id, kind, name, parentID string) *node {
	mu.Lock()
	defer mu.Unlock()
	n := &node{ID: nextID, Kind: kind, Name: name, Path: "/" + name}
	nextID++
	nodes[id] = n
	if parentID != "" {
		if p, ok := nodes[parentID]; ok {
			n.Path = p.Path + "/" + name
			n.Owner = p.Name
			p.Children = append(p.Children, n)
			return n
		}
	}
	roots = append(roots, n)
	return n
}

func removeComponent(id string) {
	mu.Lock()
	defer mu.Unlock()
	n, ok := nodes[id]
	if !ok {
		return
	}
	for _, p := range nodes {
		for i, ch := range p.Children {
			if ch == n {
				p.Children = append(p.Children[:i], p.Children[i+1:]...)
				delete(nodes, id)
				dropLifecycle(id)
				return
			}
		}
	}
	for i, r := range roots {
		if r == n {
			roots = append(roots[:i], roots[i+1:]...)
			break
		}
	}
	delete(nodes, id)
	dropLifecycle(id)
}

func resetTree() {
	mu.Lock()
	roots = nil
	nodes = map[string]*node{}
	nextID = 0
	mu.Unlock()
}

func treeJSON() string {
	mu.RLock()
	defer mu.RUnlock()
	b, _ := json.Marshal(roots)
	return string(b)
}
