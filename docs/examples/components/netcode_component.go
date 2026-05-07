//go:build js && wasm

package components

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	core "github.com/rfwlab/rfw/v2/core"
	dom "github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/netcode"
)

//go:embed templates/netcode_component.rtml
var netcodeComponentTpl []byte

type ncState struct {
	X float64 `json:"x"`
}

type NetcodeComponent struct {
	*core.HTMLComponent
	client *netcode.Client[ncState]
	stop   chan struct{}
	tick   int64
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
	c := &NetcodeComponent{}
	c.HTMLComponent = core.NewComponentWith("NetcodeComponent", netcodeComponentTpl, nil, c)
	dom.RegisterHandlerFunc("move", c.move)
	return c.HTMLComponent
}

func (c *NetcodeComponent) OnMount() {
	c.client = netcode.NewClient[ncState]("NetcodeHost", decodeState, lerp)
	c.stop = make(chan struct{})
	c.tick = 0
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-c.stop:
				return
			case <-ticker.C:
				c.tick += 50
				c.client.Flush(c.tick)
				s := c.client.State(c.tick)
				dom.Doc().ByID("pos").SetText(fmt.Sprintf("x: %.1f", s.X))
			}
		}
	}()
}

func (c *NetcodeComponent) OnUnmount() {
	if c.stop != nil {
		close(c.stop)
		c.stop = nil
	}
	c.client = nil
}

func (c *NetcodeComponent) move() {
	if c.client == nil {
		return
	}
	c.client.Enqueue(map[string]any{"dx": 1.0})
}
