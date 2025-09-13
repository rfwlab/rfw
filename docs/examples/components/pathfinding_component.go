//go:build js && wasm

package components

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	pathfinding "github.com/rfwlab/rfw/v1/ai/pathfinding"
	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	m "github.com/rfwlab/rfw/v1/math"
)

//go:embed templates/pathfinding_component.rtml
var pathfindingComponentTpl []byte

// Global-ish test state for demo purposes.
// In a real app you'd encapsulate this in your component state.
var (
	pf     = pathfinding.New()
	lastID int
)

// NewPathfindingComponent returns a demo component for pathfinding.
func NewPathfindingComponent() *core.HTMLComponent {
	c := core.NewComponent("PathfindingComponent", pathfindingComponentTpl, nil)

	// Run a grid A* and render the result as an ASCII map.
	dom.RegisterHandlerFunc("runPathfinding", func() {
		go func() {
			// Demo grid (0 walkable, 1 blocked).
			grid := pathfinding.Grid{
				{0, 0, 0, 0, 0},
				{1, 1, 0, 1, 0},
				{0, 0, 0, 1, 0},
				{0, 1, 0, 0, 0},
				{0, 1, 0, 1, 0},
			}
			start := m.Vec2{0, 0}
			goal := m.Vec2{4, 4}

			ctx := context.Background()
			id, ch := pf.RequestGridPath(ctx, grid, start, goal)
			lastID = id

			res := <-ch // GridPathResult{Path, Found, Err}

			doc := dom.Doc()
			switch {
			case res.Err != nil:
				doc.ByID("path-output").SetText(fmt.Sprintf("Error: %v", res.Err))
			case !res.Found:
				doc.ByID("path-output").SetText("No path found.")
			default:
				// Render ASCII map with path overlay.
				txt := renderGridASCII(grid, start, goal, res.Path)
				info := fmt.Sprintf("Found: %v\nPath: %v\n\n%s", res.Found, res.Path, txt)
				doc.ByID("path-output").SetText(info)
			}
		}()
	})

	// Cancel the last pending request (if any).
	dom.RegisterHandlerFunc("cancelPathfinding", func() {
		pf.Cancel(lastID)
	})

	return c
}

// renderGridASCII draws an ASCII map.
// Legend: '■' wall, '·' empty, 'S' start, 'G' goal, '*' path (excluding S/G).
func renderGridASCII(grid pathfinding.Grid, start, goal m.Vec2, path []m.Vec2) string {
	h := len(grid)
	if h == 0 {
		return "(empty grid)"
	}
	w := len(grid[0])

	// Build a quick lookup for path positions.
	pathSet := make(map[[2]int]bool, len(path))
	for _, p := range path {
		pathSet[[2]int{int(p.X), int(p.Y)}] = true
	}
	sx, sy := int(start.X), int(start.Y)
	gx, gy := int(goal.X), int(goal.Y)

	var b strings.Builder
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			switch {
			case x == sx && y == sy:
				b.WriteString("S ")
			case x == gx && y == gy:
				b.WriteString("G ")
			case grid[y][x] == 1:
				b.WriteString("■ ")
			case pathSet[[2]int{x, y}]:
				// Avoid overwriting S/G markers (already handled above).
				b.WriteString("* ")
			default:
				b.WriteString("· ")
			}
		}
		if y < h-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
