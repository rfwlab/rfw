//go:build devtools && js && wasm

package devtools

import (
	"encoding/json"
	jst "syscall/js"
)

type node struct {
	ID       int     `json:"id"`
	Kind     string  `json:"kind"`
	Name     string  `json:"name"`
	Time     float64 `json:"time"`
	Path     string  `json:"path"`
	Children []node  `json:"children,omitempty"`
}

var mock = []node{
	{ID: 1, Kind: "Layout", Name: "AppLayout", Time: 2.3, Path: "/app/Layout", Children: []node{
		{ID: 2, Kind: "Header", Name: "Header", Time: 1.2, Path: "/app/Layout/Header", Children: []node{
			{ID: 5, Kind: "Button", Name: "ThemeToggle", Time: 0.4, Path: "/app/common/ThemeToggle"},
		}},
		{ID: 3, Kind: "Route", Name: "Dashboard", Time: 3.0, Path: "/routes/dashboard", Children: []node{
			{ID: 6, Kind: "Card", Name: "StatsCard", Time: 0.9, Path: "/routes/dashboard/StatsCard"},
			{ID: 7, Kind: "Chart", Name: "UsageChart", Time: 2.8, Path: "/routes/dashboard/UsageChart"},
		}},
		{ID: 4, Kind: "Footer", Name: "Footer", Time: 0.7, Path: "/app/Layout/Footer"},
	}},
}

func init() {
	jst.Global().Set("RFW_DEVTOOLS_TREE", jst.FuncOf(func(this jst.Value, args []jst.Value) any {
		b, _ := json.Marshal(mock)
		return string(b)
	}))
}
