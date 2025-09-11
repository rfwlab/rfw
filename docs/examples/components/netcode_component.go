//go:build js && wasm

package components

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	hostclient "github.com/rfwlab/rfw/v1/hostclient"
	"github.com/rfwlab/rfw/v1/netcode"
)

//go:embed templates/netcode_component.rtml
var netcodeComponentTpl []byte

type ncState struct {
	X float64 `json:"x"`
}

func decodeState(m map[string]any) ncState {
	b, _ := json.Marshal(m)
	var s ncState
	_ = json.Unmarshal(b, &s)
	return s
}

func lerp(a, b ncState, alpha float64) ncState {
	return ncState{X: a.X + (b.X-a.X)*alpha}
}

// NewNetcodeComponent renders a simple netcode demo.
func NewNetcodeComponent() *core.HTMLComponent {
	hostclient.EnableDebug()
	c := core.NewComponent("NetcodeComponent", netcodeComponentTpl, nil)
	client := netcode.NewClient[ncState]("NetcodeHost", decodeState, lerp)
	var tick int64
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		for range ticker.C {
			tick += 50
			client.Flush(tick)
			s := client.State(tick)
			dom.SetText(dom.ByID("pos"), fmt.Sprintf("x: %.1f", s.X))
		}
	}()
	dom.RegisterHandlerFunc("move", func() {
		client.Enqueue(map[string]any{"dx": 1})
	})
	return c
}
