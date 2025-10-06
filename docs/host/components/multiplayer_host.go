package components

import (
	"math"
	"math/rand"
	"time"

	"github.com/rfwlab/rfw/v1/host"
	"github.com/rfwlab/rfw/v1/netcode"
)

const (
	multiplayerChannel   = "MultiplayerArena"
	multiplayerTick      = 50 * time.Millisecond
	arenaWidth           = 800.0
	arenaHeight          = 520.0
	playerRadius         = 18.0
	bulletRadius         = 6.0
	playerSpeed          = 220.0
	bulletSpeed          = 420.0
	bulletLifetime       = 2.5
	maxLives             = 3
	shootCooldownSeconds = 0.4
)

var colorPalette = []string{
	"#ef4444", // red
	"#22c55e", // green
	"#3b82f6", // blue
	"#f97316", // orange
	"#a855f7", // purple
	"#14b8a6", // teal
	"#facc15", // yellow
	"#ec4899", // pink
}

type multiplayerState struct {
	Players map[string]multiplayerPlayer `json:"players"`
	Bullets []multiplayerBullet          `json:"bullets"`
	Winner  string                       `json:"winner"`
}

type multiplayerPlayer struct {
	ID       string  `json:"id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	VX       float64 `json:"vx"`
	VY       float64 `json:"vy"`
	AimX     float64 `json:"aimX"`
	AimY     float64 `json:"aimY"`
	Color    string  `json:"color"`
	Lives    int     `json:"lives"`
	Alive    bool    `json:"alive"`
	Cooldown float64 `json:"-"`
}

type multiplayerBullet struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	VX    float64 `json:"vx"`
	VY    float64 `json:"vy"`
	Owner string  `json:"owner"`
	Life  float64 `json:"-"`
}

// RegisterMultiplayerHost exposes the multiplayer arena netcode server.
func RegisterMultiplayerHost() {
	initial := multiplayerState{Players: make(map[string]multiplayerPlayer)}
	srv := netcode.NewServer(multiplayerChannel, initial, applyMultiplayerCommand)
	host.Register(srv.HostComponent())

	ticker := time.NewTicker(multiplayerTick)
	go func() {
		defer ticker.Stop()
		var tick int64
		for range ticker.C {
			tick += int64(multiplayerTick / time.Millisecond)
			srv.Update(func(state *multiplayerState) {
				stepMultiplayer(state, multiplayerTick.Seconds())
			})
			srv.Broadcast(tick)
		}
	}()
}

func applyMultiplayerCommand(state *multiplayerState, cmd any) {
	payload, ok := cmd.(map[string]any)
	if !ok {
		return
	}
	session, _ := payload["session"].(string)
	if session == "" {
		return
	}
	if state.Players == nil {
		state.Players = make(map[string]multiplayerPlayer)
	}

	action, _ := payload["type"].(string)
	switch action {
	case "join":
		ensurePlayer(state, session)
		state.Winner = ""
		return
	}

	player := ensurePlayer(state, session)
	if !player.Alive {
		state.Players[session] = player
		return
	}

	dx := floatFrom(payload["dx"])
	dy := floatFrom(payload["dy"])
	mag := math.Hypot(dx, dy)
	if mag > 1 {
		dx /= mag
		dy /= mag
		mag = 1
	}
	player.VX = dx * playerSpeed
	player.VY = dy * playerSpeed

	aimX := floatFrom(payload["aimX"])
	aimY := floatFrom(payload["aimY"])
	if aimLen := math.Hypot(aimX, aimY); aimLen > 0 {
		player.AimX = aimX / aimLen
		player.AimY = aimY / aimLen
	} else if mag > 0 {
		norm := math.Hypot(dx, dy)
		if norm > 0 {
			player.AimX = dx / norm
			player.AimY = dy / norm
		}
	}

	shoot := boolFrom(payload["shoot"])
	if shoot && player.Cooldown <= 0 && player.Alive {
		ax, ay := player.AimX, player.AimY
		if math.Hypot(ax, ay) == 0 {
			ax, ay = 0, -1
			player.AimX, player.AimY = ax, ay
		}
		bullet := multiplayerBullet{
			X:     player.X + ax*(playerRadius+bulletRadius),
			Y:     player.Y + ay*(playerRadius+bulletRadius),
			VX:    ax * bulletSpeed,
			VY:    ay * bulletSpeed,
			Owner: session,
		}
		state.Bullets = append(state.Bullets, bullet)
		player.Cooldown = shootCooldownSeconds
	}

	state.Players[session] = player
}

func stepMultiplayer(state *multiplayerState, dt float64) {
	if state.Players == nil {
		state.Players = make(map[string]multiplayerPlayer)
	}

	for id, player := range state.Players {
		if player.Cooldown > 0 {
			player.Cooldown -= dt
			if player.Cooldown < 0 {
				player.Cooldown = 0
			}
		}
		if player.Alive {
			player.X += player.VX * dt
			player.Y += player.VY * dt
			if player.X < playerRadius {
				player.X = playerRadius
			}
			if player.X > arenaWidth-playerRadius {
				player.X = arenaWidth - playerRadius
			}
			if player.Y < playerRadius {
				player.Y = playerRadius
			}
			if player.Y > arenaHeight-playerRadius {
				player.Y = arenaHeight - playerRadius
			}
		}
		state.Players[id] = player
	}

	newBullets := make([]multiplayerBullet, 0, len(state.Bullets))
	for _, bullet := range state.Bullets {
		bullet.X += bullet.VX * dt
		bullet.Y += bullet.VY * dt
		bullet.Life += dt
		if bullet.Life > bulletLifetime {
			continue
		}
		if bullet.X < 0 || bullet.X > arenaWidth || bullet.Y < 0 || bullet.Y > arenaHeight {
			continue
		}

		hit := false
		for id, player := range state.Players {
			if !player.Alive || id == bullet.Owner {
				continue
			}
			if overlaps(bullet.X, bullet.Y, player.X, player.Y, playerRadius+bulletRadius) {
				player.Lives--
				if player.Lives < 0 {
					player.Lives = 0
				}
				if player.Lives <= 0 {
					player.Alive = false
					player.VX, player.VY = 0, 0
				}
				state.Players[id] = player
				hit = true
				break
			}
		}
		if !hit {
			newBullets = append(newBullets, bullet)
		}
	}
	state.Bullets = newBullets

	aliveCount := 0
	lastAlive := ""
	for id, player := range state.Players {
		if player.Alive {
			aliveCount++
			lastAlive = id
		}
	}
	if aliveCount == 1 {
		state.Winner = lastAlive
	} else if aliveCount == 0 {
		state.Winner = ""
	} else {
		state.Winner = ""
	}
}

func ensurePlayer(state *multiplayerState, session string) multiplayerPlayer {
	if player, ok := state.Players[session]; ok {
		return player
	}
	spawnX := playerRadius + rand.Float64()*(arenaWidth-2*playerRadius)
	spawnY := playerRadius + rand.Float64()*(arenaHeight-2*playerRadius)
	player := multiplayerPlayer{
		ID:    session,
		X:     spawnX,
		Y:     spawnY,
		AimX:  0,
		AimY:  -1,
		Color: pickColor(state),
		Lives: maxLives,
		Alive: true,
	}
	state.Players[session] = player
	return player
}

func pickColor(state *multiplayerState) string {
	used := make(map[string]bool, len(state.Players))
	for _, player := range state.Players {
		used[player.Color] = true
	}
	for _, color := range colorPalette {
		if !used[color] {
			return color
		}
	}
	return colorPalette[rand.Intn(len(colorPalette))]
}

func overlaps(ax, ay, bx, by, radius float64) bool {
	dx := ax - bx
	dy := ay - by
	return dx*dx+dy*dy <= radius*radius
}

func floatFrom(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	}
	return 0
}

func boolFrom(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case float64:
		return val != 0
	case int:
		return val != 0
	}
	return false
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
