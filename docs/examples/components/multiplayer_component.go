//go:build js && wasm

package components

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	events "github.com/rfwlab/rfw/v1/events"
	draw "github.com/rfwlab/rfw/v1/game/draw"
	hostclient "github.com/rfwlab/rfw/v1/hostclient"
	js "github.com/rfwlab/rfw/v1/js"
	"github.com/rfwlab/rfw/v1/netcode"
)

//go:embed templates/multiplayer_component.rtml
var multiplayerComponentTpl []byte

const (
	multiplayerRoute = "MultiplayerArena"
	frameInterval    = 50 * time.Millisecond
)

type mpSnapshot struct {
	Players map[string]mpPlayer `json:"players"`
	Bullets []mpBullet          `json:"bullets"`
	Winner  string              `json:"winner"`
}

type mpPlayer struct {
	ID    string  `json:"id"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	VX    float64 `json:"vx"`
	VY    float64 `json:"vy"`
	AimX  float64 `json:"aimX"`
	AimY  float64 `json:"aimY"`
	Color string  `json:"color"`
	Lives int     `json:"lives"`
	Alive bool    `json:"alive"`
}

type mpBullet struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type multiplayerComponent struct {
	*core.HTMLComponent

	client       *netcode.Client[mpSnapshot]
	keyState     map[string]bool
	shootPending bool
	lastAimX     float64
	lastAimY     float64
	cancelFuncs  []func()
	stop         chan struct{}
	sessionID    string
	surface      draw.Canvas
}

// NewMultiplayerComponent renders the multiplayer arena example page.
func NewMultiplayerComponent() *core.HTMLComponent {
	cmp := &multiplayerComponent{keyState: make(map[string]bool), lastAimY: -1}
	cmp.HTMLComponent = core.NewComponentWith("MultiplayerComponent", multiplayerComponentTpl, nil, cmp)
	return cmp.HTMLComponent
}

func (c *multiplayerComponent) OnMount() {
	c.keyState = make(map[string]bool)
	c.stop = make(chan struct{})
	c.lastAimX, c.lastAimY = 0, -1
	c.client = netcode.NewClient[mpSnapshot](multiplayerRoute, decodeSnapshot, passSnapshot)

	doc := dom.Doc()
	canvas := doc.ByID("mp-canvas")
	if !canvas.IsNull() && !canvas.IsUndefined() {
		if surface, ok := draw.NewCanvas(canvas); ok {
			surface.SetSize(arenaWidth, arenaHeight)
			c.surface = surface
		} else {
			c.surface = draw.Canvas{}
		}
	}

	c.cancelFuncs = []func(){
		events.OnKeyDown(c.onKeyDown),
		events.OnKeyUp(c.onKeyUp),
	}

	go c.loop()
}

func (c *multiplayerComponent) OnUnmount() {
	if c.stop != nil {
		close(c.stop)
		c.stop = nil
	}
	for _, cancel := range c.cancelFuncs {
		if cancel != nil {
			cancel()
		}
	}
	c.cancelFuncs = nil
	c.client = nil
	c.surface = draw.Canvas{}
}

func (c *multiplayerComponent) loop() {
	ticker := time.NewTicker(frameInterval)
	defer ticker.Stop()
	var tick int64
	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			tick += int64(frameInterval / time.Millisecond)
			c.ensureSession()
			if c.sessionID == "" {
				continue
			}
			dx, dy := c.inputVector()
			aimX, aimY := c.aimVector()
			shoot := c.consumeShoot()
			payload := map[string]any{
				"type":    "input",
				"session": c.sessionID,
				"dx":      dx,
				"dy":      dy,
				"aimX":    aimX,
				"aimY":    aimY,
			}
			if shoot {
				payload["shoot"] = true
			}
			c.client.Enqueue(payload)
			c.client.Flush(tick)
			snap := c.client.State(tick)
			c.render(snap)
		}
	}
}

func (c *multiplayerComponent) ensureSession() {
	if c.sessionID != "" {
		return
	}
	id := hostclient.SessionID()
	if id == "" {
		return
	}
	c.sessionID = id
	c.client.Enqueue(map[string]any{"type": "join", "session": id})
	c.client.Flush(0)
}

func (c *multiplayerComponent) onKeyDown(evt js.Value) {
	key := strings.ToLower(evt.Get("key").String())
	c.keyState[key] = true
	if key == " " || key == "space" {
		if !evt.Get("repeat").Bool() {
			c.shootPending = true
		}
	}
}

func (c *multiplayerComponent) onKeyUp(evt js.Value) {
	key := strings.ToLower(evt.Get("key").String())
	c.keyState[key] = false
	if key == " " || key == "space" {
		c.shootPending = false
	}
}

func (c *multiplayerComponent) inputVector() (float64, float64) {
	dx := 0.0
	dy := 0.0
	if c.keyState["arrowleft"] || c.keyState["a"] {
		dx -= 1
	}
	if c.keyState["arrowright"] || c.keyState["d"] {
		dx += 1
	}
	if c.keyState["arrowup"] || c.keyState["w"] {
		dy -= 1
	}
	if c.keyState["arrowdown"] || c.keyState["s"] {
		dy += 1
	}
	if dx != 0 || dy != 0 {
		length := math.Hypot(dx, dy)
		if length > 0 {
			dx /= length
			dy /= length
			c.lastAimX = dx
			c.lastAimY = dy
		}
	}
	return dx, dy
}

func (c *multiplayerComponent) aimVector() (float64, float64) {
	ax, ay := c.lastAimX, c.lastAimY
	if ax == 0 && ay == 0 {
		ax, ay = 0, -1
	}
	return ax, ay
}

func (c *multiplayerComponent) consumeShoot() bool {
	if !c.shootPending {
		return false
	}
	c.shootPending = false
	return true
}

func (c *multiplayerComponent) render(state mpSnapshot) {
	if !c.surface.Valid() {
		return
	}

	commands := make([]draw.Command, 0, 1+len(state.Bullets)+len(state.Players)*3)
	commands = append(commands, draw.Rectangle(0, 0, arenaWidth, arenaHeight).Fill(arenaBackground))

	for _, bullet := range state.Bullets {
		commands = append(commands, draw.Disc(bullet.X, bullet.Y, bulletRadius).Fill(bulletColor))
	}

	for id, player := range state.Players {
		marker := draw.Disc(player.X, player.Y, playerRadius).Fill(player.Color)
		if id == c.sessionID {
			marker.Stroke(activePlayerStroke, outlineWidth)
		}
		commands = append(commands, marker)
		if !player.Alive {
			commands = append(commands,
				draw.Segment(player.X-playerRadius, player.Y-playerRadius, player.X+playerRadius, player.Y+playerRadius).Stroke(eliminatedStroke, outlineWidth),
				draw.Segment(player.X-playerRadius, player.Y+playerRadius, player.X+playerRadius, player.Y-playerRadius).Stroke(eliminatedStroke, outlineWidth),
			)
		}
	}

	c.surface.Draw(commands...)

	c.updateHUD(state)
}

func (c *multiplayerComponent) updateHUD(state mpSnapshot) {
	doc := dom.Doc()
	livesEl := doc.ByID("mp-lives")
	if !livesEl.IsNull() && !livesEl.IsUndefined() {
		keys := make([]string, 0, len(state.Players))
		for id := range state.Players {
			keys = append(keys, id)
		}
		sort.Strings(keys)
		var b strings.Builder
		b.WriteString("<ul class=\"space-y-1\">")
		for _, id := range keys {
			player := state.Players[id]
			name := fmt.Sprintf("Player %s", shortID(id))
			if id == c.sessionID {
				name = "You"
			}
			hearts := strings.Repeat("❤", player.Lives)
			if hearts == "" {
				hearts = "✖"
			}
			stateClass := ""
			if !player.Alive {
				stateClass = "opacity-60"
			}
			fmt.Fprintf(&b,
				"<li class=\"flex items-center justify-between text-sm %s\"><span class=\"flex items-center gap-2\">"+
					"<span class=\"inline-block h-3 w-3 rounded-full\" style=\"background:%s\"></span><span>%s</span></span>"+
					"<span class=\"font-mono\">%s</span></li>",
				stateClass, player.Color, name, hearts,
			)
		}
		b.WriteString("</ul>")
		livesEl.SetHTML(b.String())
	}

	statusEl := doc.ByID("mp-status")
	if statusEl.IsNull() || statusEl.IsUndefined() {
		return
	}
	message := ""
	if state.Winner != "" {
		if state.Winner == c.sessionID {
			message = "You won! Last player standing."
		} else if winner, ok := state.Players[state.Winner]; ok {
			message = fmt.Sprintf("Player %s wins", shortID(winner.ID))
		} else {
			message = "Match finished."
		}
	} else if player, ok := state.Players[c.sessionID]; ok && !player.Alive {
		message = "Game over! Press WASD to move when the next round starts."
	}
	statusEl.SetText(message)
}

func decodeSnapshot(m map[string]any) mpSnapshot {
	b, _ := json.Marshal(m)
	var snap mpSnapshot
	_ = json.Unmarshal(b, &snap)
	return snap
}

func passSnapshot(_ mpSnapshot, next mpSnapshot, _ float64) mpSnapshot { return next }

func shortID(id string) string {
	if len(id) <= 6 {
		return id
	}
	return id[:6]
}

const (
	bulletRadius       = 6.0
	playerRadius       = 18.0
	arenaWidth         = 800.0
	arenaHeight        = 520.0
	outlineWidth       = 2.0
	arenaBackground    = "#0f172a"
	bulletColor        = "#f8fafc"
	activePlayerStroke = "#ffffff"
	eliminatedStroke   = "rgba(15, 23, 42, 0.7)"
)
