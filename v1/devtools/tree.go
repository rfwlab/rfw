package devtools

import (
	"encoding/json"
	"sync"
)

type node struct {
	ID       int     `json:"id"`
	Kind     string  `json:"kind"`
	Name     string  `json:"name"`
	Time     float64 `json:"time"`
	Path     string  `json:"path"`
	Children []*node `json:"children,omitempty"`
}

var (
	mu     sync.RWMutex
	roots  []*node
	nodes  = map[string]*node{}
	nextID int
)

func addComponent(id, kind, name, parentID string) {
	mu.Lock()
	defer mu.Unlock()
	n := &node{ID: nextID, Kind: kind, Name: name, Path: "/" + name}
	nextID++
	nodes[id] = n
	if parentID != "" {
		if p, ok := nodes[parentID]; ok {
			n.Path = p.Path + "/" + name
			p.Children = append(p.Children, n)
			return
		}
	}
	roots = append(roots, n)
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
