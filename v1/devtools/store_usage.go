package devtools

import (
	"sort"
	"sync"
)

type storeBinding struct {
	Module string   `json:"module"`
	Name   string   `json:"name"`
	Keys   []string `json:"keys,omitempty"`
}

var (
	storeUsageMu sync.RWMutex
	storeUsage   = map[string]map[string]map[string]map[string]struct{}{}
)

func recordStoreBinding(componentID, module, store, key string) {
	if componentID == "" || module == "" || store == "" || key == "" {
		return
	}
	storeUsageMu.Lock()
	defer storeUsageMu.Unlock()
	moduleMap, ok := storeUsage[componentID]
	if !ok {
		moduleMap = make(map[string]map[string]map[string]struct{})
		storeUsage[componentID] = moduleMap
	}
	storeMap, ok := moduleMap[module]
	if !ok {
		storeMap = make(map[string]map[string]struct{})
		moduleMap[module] = storeMap
	}
	keySet, ok := storeMap[store]
	if !ok {
		keySet = make(map[string]struct{})
		storeMap[store] = keySet
	}
	keySet[key] = struct{}{}
}

func dropStoreBindings(componentID string) {
	if componentID == "" {
		return
	}
	storeUsageMu.Lock()
	delete(storeUsage, componentID)
	storeUsageMu.Unlock()
}

func snapshotStoreBindings(componentID string) []storeBinding {
	storeUsageMu.RLock()
	moduleMap := storeUsage[componentID]
	storeUsageMu.RUnlock()
	if len(moduleMap) == 0 {
		return nil
	}
	modules := make([]string, 0, len(moduleMap))
	for module := range moduleMap {
		modules = append(modules, module)
	}
	sort.Strings(modules)
	bindings := make([]storeBinding, 0)
	for _, module := range modules {
		storeMap := moduleMap[module]
		storeNames := make([]string, 0, len(storeMap))
		for storeName := range storeMap {
			storeNames = append(storeNames, storeName)
		}
		sort.Strings(storeNames)
		for _, storeName := range storeNames {
			keySet := storeMap[storeName]
			keys := make([]string, 0, len(keySet))
			for key := range keySet {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			bindings = append(bindings, storeBinding{Module: module, Name: storeName, Keys: keys})
		}
	}
	return bindings
}

func resetStoreUsage() {
	storeUsageMu.Lock()
	storeUsage = map[string]map[string]map[string]map[string]struct{}{}
	storeUsageMu.Unlock()
}
