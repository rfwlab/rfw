//go:build js && wasm

package components

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	events "github.com/rfwlab/rfw/v1/events"
	game "github.com/rfwlab/rfw/v1/game/loop"
	scene "github.com/rfwlab/rfw/v1/game/scene"
	js "github.com/rfwlab/rfw/v1/js"
	m "github.com/rfwlab/rfw/v1/math"
	webgl "github.com/rfwlab/rfw/v1/webgl"
)

//go:embed templates/webgl_component.rtml
var webglComponentTpl []byte

type point struct{ x, y int }

const gridSize = 20

var cellSize = 2.0 / float32(gridSize)

var (
	root      *scene.Node
	snakeRoot *scene.Node
	foodNode  *scene.Node
	snake     []*scene.Node
	dir       point
	score     int

	ctx      webgl.Context
	colorLoc js.Value
	mvpLoc   js.Value
	timeLoc  js.Value
	proj     m.Mat4
	keyState = map[string]bool{}
	running  bool
	lastMove time.Duration
	elapsed  float64
)

// NewWebGLComponent returns a component demonstrating WebGL via a snake game.
func NewWebGLComponent() *core.HTMLComponent {
	c := core.NewComponent("WebGLComponent", webglComponentTpl, nil)
	dom.RegisterHandlerFunc("webglStart", func() { startGame() })
	dom.RegisterHandlerFunc("webglFullscreen", fullscreen)
	return c
}

func init() {
	events.OnKeyDown(func(v js.Value) {
		key := strings.ToLower(v.Get("key").String())
		keyState[key] = true
	})
	events.OnKeyUp(func(v js.Value) {
		key := strings.ToLower(v.Get("key").String())
		keyState[key] = false
	})
	game.OnUpdate(updateLoop)
	game.OnRender(renderLoop)
}

func fullscreen() {
	doc := dom.Doc()
	fs := doc.Get("fullscreenElement")
	if !fs.IsUndefined() && !fs.IsNull() {
		doc.Call("exitFullscreen")
	} else {
		doc.ByID("game-container").Call("requestFullscreen")
	}
}

func startGame() {
	doc := dom.Doc()
	if ctx.Value().IsUndefined() || ctx.Value().IsNull() {
		ctx = webgl.NewContext("glcanvas")
		ctx.ClearColor(0, 0, 0, 1)
		canvas := doc.ByID("glcanvas")
		ctx.Viewport(0, 0, canvas.Get("width").Int(), canvas.Get("height").Int())
		ctx.Enable(webgl.DEPTH_TEST)
		ctx.DepthFunc(webgl.LEQUAL)
		ctx.Enable(webgl.BLEND)
		ctx.BlendFunc(webgl.SRC_ALPHA, webgl.ONE)

		vertexSrc := `attribute vec2 a_position;
uniform mat4 u_mvp;
void main(){
    vec4 pos = vec4(a_position,0.0,1.0);
    gl_Position = u_mvp * pos;
}`
		fragmentSrc := `precision mediump float;
uniform vec4 u_color;
uniform float u_time;
void main(){
    float glow = 0.5 + 0.5 * sin(u_time);
    gl_FragColor = vec4(u_color.rgb * glow, u_color.a);
}`

		vs := ctx.CreateShader(webgl.VERTEX_SHADER)
		ctx.ShaderSource(vs, vertexSrc)
		ctx.CompileShader(vs)
		fs := ctx.CreateShader(webgl.FRAGMENT_SHADER)
		ctx.ShaderSource(fs, fragmentSrc)
		ctx.CompileShader(fs)

		prog := ctx.CreateProgram()
		ctx.AttachShader(prog, vs)
		ctx.AttachShader(prog, fs)
		ctx.LinkProgram(prog)
		ctx.UseProgram(prog)

		if !ctx.Get("createVertexArray").IsUndefined() {
			vao := ctx.CreateVertexArray()
			ctx.BindVertexArray(vao)
		}
		vbuf := ctx.CreateBuffer()
		ctx.BindBuffer(webgl.ARRAY_BUFFER, vbuf)
		vertices := []float32{
			-0.5, -0.5,
			0.5, -0.5,
			-0.5, 0.5,
			0.5, 0.5,
		}
		ctx.BufferDataFloat32(webgl.ARRAY_BUFFER, vertices, webgl.STATIC_DRAW)
		ibuf := ctx.CreateBuffer()
		ctx.BindBuffer(webgl.ELEMENT_ARRAY_BUFFER, ibuf)
		inds := js.Uint16Array().New(6)
		for i, v := range []uint16{0, 1, 2, 2, 1, 3} {
			inds.SetIndex(i, v)
		}
		ctx.BufferData(webgl.ELEMENT_ARRAY_BUFFER, inds, webgl.STATIC_DRAW)

		posLoc := ctx.GetAttribLocation(prog, "a_position")
		ctx.EnableVertexAttribArray(posLoc)
		ctx.VertexAttribPointer(posLoc, 2, webgl.FLOAT, false, 0, 0)

		colorLoc = ctx.GetUniformLocation(prog, "u_color")
		mvpLoc = ctx.GetUniformLocation(prog, "u_mvp")
		timeLoc = ctx.GetUniformLocation(prog, "u_time")
		proj = m.Orthographic(-1, 1, -1, 1, -1, 1)

		game.Start()
	}

	root = scene.NewNode()
	snakeRoot = scene.NewNode()
	root.AddChild(snakeRoot)
	snakeRoot.AddEntity(&snakeEntity{comps: []scene.Component{&snakeComponent{}}})
	snake = nil
	for _, p := range []point{{gridSize / 2, gridSize / 2}, {gridSize/2 - 1, gridSize / 2}, {gridSize/2 - 2, gridSize / 2}} {
		seg := scene.NewNode()
		seg.Transform = scene.Transform{X: float64(p.x), Y: float64(p.y)}
		seg.AddEntity(&renderEntity{color: [4]float32{0, 1, 0, 1}})
		snakeRoot.AddChild(seg)
		snake = append(snake, seg)
	}
	foodNode = scene.NewNode()
	foodNode.AddEntity(&renderEntity{color: [4]float32{1, 0, 0, 1}})
	root.AddChild(foodNode)
	dir = point{1, 0}
	score = 0
	running = true
	lastMove = 0
	elapsed = 0
	updateScore()
	newFood()
	doc.ByID("menu").AddClass("hidden")
}

func updateLoop(t game.Ticker) {
	scene.Update(root, scene.Ticker{Delta: t.Delta, FPS: t.FPS})
}

func renderLoop(t game.Ticker) {
	if ctx.Value().IsUndefined() || ctx.Value().IsNull() {
		return
	}
	elapsed += t.Delta.Seconds()
	ctx.Clear(webgl.COLOR_BUFFER_BIT)
	ctx.Uniform1f(timeLoc, float32(elapsed))
	scene.Traverse(root, func(n *scene.Node) {
		for _, e := range n.Entities {
			if r, ok := e.(*renderEntity); ok {
				drawSquare(n, r.color)
			}
		}
	})
}

type snakeComponent struct{}

func (c *snakeComponent) Update(n *scene.Node, t scene.Ticker) {
	if !running {
		return
	}
	if keyState["a"] && dir.x != 1 {
		dir = point{-1, 0}
	}
	if keyState["d"] && dir.x != -1 {
		dir = point{1, 0}
	}
	if keyState["w"] && dir.y != -1 {
		dir = point{0, 1}
	}
	if keyState["s"] && dir.y != 1 {
		dir = point{0, -1}
	}
	lastMove += t.Delta
	if lastMove > 150*time.Millisecond {
		moveSnake()
		lastMove = 0
	}
}

type snakeEntity struct{ comps []scene.Component }

func (e *snakeEntity) Components() []scene.Component { return e.comps }

type renderEntity struct{ color [4]float32 }

func (e *renderEntity) Components() []scene.Component { return nil }

func moveSnake() {
	doc := dom.Doc()
	head := snake[0]
	nx := int(head.Transform.X) + dir.x
	ny := int(head.Transform.Y) + dir.y
	if nx < 0 {
		nx = gridSize - 1
	}
	if nx >= gridSize {
		nx = 0
	}
	if ny < 0 {
		ny = gridSize - 1
	}
	if ny >= gridSize {
		ny = 0
	}
	for _, s := range snake {
		if int(s.Transform.X) == nx && int(s.Transform.Y) == ny {
			running = false
			doc.ByID("menu").RemoveClass("hidden")
			return
		}
	}
	newHead := scene.NewNode()
	newHead.Transform = scene.Transform{X: float64(nx), Y: float64(ny)}
	newHead.AddEntity(&renderEntity{color: [4]float32{0, 1, 0, 1}})
	snakeRoot.AddChild(newHead)
	snake = append([]*scene.Node{newHead}, snake...)
	if nx == int(foodNode.Transform.X) && ny == int(foodNode.Transform.Y) {
		score++
		updateScore()
		newFood()
	} else {
		tail := snake[len(snake)-1]
		snake = snake[:len(snake)-1]
		for i, c := range snakeRoot.Children {
			if c == tail {
				snakeRoot.Children = append(snakeRoot.Children[:i], snakeRoot.Children[i+1:]...)
				break
			}
		}
	}
}

func drawSquare(n *scene.Node, color [4]float32) {
	x := -1 + cellSize*float32(n.Transform.X) + cellSize/2
	y := -1 + cellSize*float32(n.Transform.Y) + cellSize/2

	model := m.Translation(m.Vec3{x, y, 0}).Mul(m.Scale(m.Vec3{cellSize * 1.4, cellSize * 1.4, 1}))
	mvp := proj.Mul(model)
	arr := js.Float32Array().New(len(mvp))
	for i, v := range mvp {
		arr.SetIndex(i, v)
	}
	ctx.Call("uniformMatrix4fv", mvpLoc, false, arr)
	ctx.Uniform4f(colorLoc, color[0], color[1], color[2], color[3]*0.3)
	ctx.DrawElements(webgl.TRIANGLES, 6, webgl.UNSIGNED_SHORT, 0)

	model = m.Translation(m.Vec3{x, y, 0}).Mul(m.Scale(m.Vec3{cellSize, cellSize, 1}))
	mvp = proj.Mul(model)
	arr = js.Float32Array().New(len(mvp))
	for i, v := range mvp {
		arr.SetIndex(i, v)
	}
	ctx.Call("uniformMatrix4fv", mvpLoc, false, arr)
	ctx.Uniform4f(colorLoc, color[0], color[1], color[2], color[3])
	ctx.DrawElements(webgl.TRIANGLES, 6, webgl.UNSIGNED_SHORT, 0)
}

func newFood() {
	for {
		m := js.Math()
		fx := m.Call("floor", m.Call("random").Float()*float64(gridSize)).Int()
		fy := m.Call("floor", m.Call("random").Float()*float64(gridSize)).Int()
		collision := false
		for _, s := range snake {
			if int(s.Transform.X) == fx && int(s.Transform.Y) == fy {
				collision = true
				break
			}
		}
		if !collision {
			foodNode.Transform = scene.Transform{X: float64(fx), Y: float64(fy)}
			return
		}
	}
}

func updateScore() {
	dom.Doc().ByID("score").SetText(fmt.Sprintf("Score: %d", score))
}
