package input

import "github.com/rfwlab/rfw/v1/math"

// Camera represents a simple 2D camera state.
type Camera struct {
	Position math.Vec2
	Zoom     float32
	Rotation float32
}

// Manager handles input bindings and state.
type Manager struct {
	keyBindings   map[string]string
	mouseBindings map[int]string
	active        map[string]bool

	dragStart math.Vec2
	dragEnd   math.Vec2
	dragging  bool
	lastPos   math.Vec2

	camera Camera
}

func newManager() *Manager {
	return &Manager{
		keyBindings:   make(map[string]string),
		mouseBindings: make(map[int]string),
		active:        make(map[string]bool),
		camera:        Camera{Zoom: 1},
	}
}

// BindKey associates a keyboard key with an action.
func (m *Manager) BindKey(action, key string) { m.keyBindings[key] = action }

// RebindKey changes the key mapped to an action.
func (m *Manager) RebindKey(action, key string) {
	for k, a := range m.keyBindings {
		if a == action {
			delete(m.keyBindings, k)
			break
		}
	}
	m.keyBindings[key] = action
}

// BindMouse associates a mouse button with an action.
func (m *Manager) BindMouse(action string, button int) { m.mouseBindings[button] = action }

// RebindMouse changes the button mapped to an action.
func (m *Manager) RebindMouse(action string, button int) {
	for b, a := range m.mouseBindings {
		if a == action {
			delete(m.mouseBindings, b)
			break
		}
	}
	m.mouseBindings[button] = action
}

// IsActive reports whether an action is currently engaged.
func (m *Manager) IsActive(action string) bool { return m.active[action] }

// DragRect returns the start and end of the current drag along with its state.
func (m *Manager) DragRect() (start, end math.Vec2, dragging bool) {
	return m.dragStart, m.dragEnd, m.dragging
}

// Camera returns a copy of the current camera state.
func (m *Manager) Camera() Camera { return m.camera }

// Pan translates the camera by dx, dy.
func (m *Manager) Pan(dx, dy float32) {
	m.camera.Position = m.camera.Position.Add(math.Vec2{dx, dy})
}

// Zoom adjusts the camera zoom level.
func (m *Manager) Zoom(delta float32) { m.camera.Zoom += delta }

// Rotate adjusts the camera rotation.
func (m *Manager) Rotate(delta float32) { m.camera.Rotation += delta }

func (m *Manager) handleKeyDown(key string) {
	if action, ok := m.keyBindings[key]; ok {
		m.active[action] = true
	}
}

func (m *Manager) handleKeyUp(key string) {
	if action, ok := m.keyBindings[key]; ok {
		delete(m.active, action)
	}
}

func (m *Manager) handleMouseDown(button int, x, y float32) {
	if action, ok := m.mouseBindings[button]; ok {
		m.active[action] = true
	}
	if button == 0 {
		m.dragging = true
		m.dragStart = math.Vec2{x, y}
		m.dragEnd = m.dragStart
	}
	m.lastPos = math.Vec2{x, y}
}

func (m *Manager) handleMouseUp(button int, x, y float32) {
	if action, ok := m.mouseBindings[button]; ok {
		delete(m.active, action)
	}
	if button == 0 {
		m.dragging = false
	}
	m.lastPos = math.Vec2{x, y}
}

func (m *Manager) handleMouseMove(x, y float32) {
	pos := math.Vec2{x, y}
	if m.dragging {
		m.dragEnd = pos
	}
	dx := pos.X - m.lastPos.X
	dy := pos.Y - m.lastPos.Y
	if m.IsActive("pan") {
		m.Pan(dx, dy)
	}
	if m.IsActive("rotate") {
		m.Rotate(dx)
	}
	m.lastPos = pos
}

func (m *Manager) handleWheel(delta float32) { m.Zoom(-delta / 100) }
