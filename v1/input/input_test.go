package input

import (
	"testing"

	"github.com/rfwlab/rfw/v1/math"
)

func TestKeyBinding(t *testing.T) {
	m := New()
	m.BindKey("jump", "Space")
	m.handleKeyDown("Space")
	if !m.IsActive("jump") {
		t.Fatalf("action not active on key down")
	}
	m.handleKeyUp("Space")
	if m.IsActive("jump") {
		t.Fatalf("action still active after key up")
	}
}

func TestDragRect(t *testing.T) {
	m := New()
	m.handleMouseDown(0, 10, 20)
	m.handleMouseMove(30, 40)
	start, end, dragging := m.DragRect()
	if !dragging {
		t.Fatalf("expected dragging state")
	}
	if start != (math.Vec2{10, 20}) || end != (math.Vec2{30, 40}) {
		t.Fatalf("unexpected drag rect %v %v", start, end)
	}
	m.handleMouseUp(0, 30, 40)
	_, _, dragging = m.DragRect()
	if dragging {
		t.Fatalf("drag state not cleared")
	}
}

func TestCameraControls(t *testing.T) {
	m := New()
	m.BindMouse("pan", 1)
	m.handleMouseDown(1, 0, 0)
	m.handleMouseMove(5, -5)
	cam := m.Camera()
	if cam.Position != (math.Vec2{5, -5}) {
		t.Fatalf("pan not applied, got %v", cam.Position)
	}
	m.handleWheel(-120)
	cam = m.Camera()
	if cam.Zoom <= 1 {
		t.Fatalf("zoom not applied, got %v", cam.Zoom)
	}
	m.BindMouse("rotate", 2)
	m.handleMouseDown(2, 0, 0)
	m.handleMouseMove(10, 0)
	cam = m.Camera()
	if cam.Rotation == 0 {
		t.Fatalf("rotation not applied")
	}
}
