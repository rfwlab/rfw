//go:build js && wasm

package components

import (
	_ "embed"
	"fmt"
	"strings"
	jst "syscall/js"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	events "github.com/rfwlab/rfw/v1/events"
	js "github.com/rfwlab/rfw/v1/js"
	webgl "github.com/rfwlab/rfw/v1/webgl"
)

//go:embed templates/webgl_component.rtml
var webglComponentTpl []byte

type point struct{ x, y int }

const gridSize = 20

var cellSize = 2.0 / float32(gridSize)

var (
	snake []point
	dir   point
	food  point
	score int

	ctx      webgl.Context
	colorLoc jst.Value
	transLoc jst.Value
	scaleLoc jst.Value
	timeLoc  jst.Value
	keyState = map[string]bool{}
	render   jst.Func
	running  bool
	lastMove float64
)

// NewWebGLComponent returns a component demonstrating WebGL via a snake game.
func NewWebGLComponent() *core.HTMLComponent {
	c := core.NewComponent("WebGLComponent", webglComponentTpl, nil)
	dom.RegisterHandlerFunc("webglStart", func() { startGame() })
	dom.RegisterHandlerFunc("webglFullscreen", fullscreen)
	return c
}

func init() {
	events.OnKeyDown(func(v jst.Value) {
		key := strings.ToLower(v.Get("key").String())
		keyState[key] = true
	})
	events.OnKeyUp(func(v jst.Value) {
		key := strings.ToLower(v.Get("key").String())
		keyState[key] = false
	})
}

func fullscreen() {
	doc := js.Document()
	fs := doc.Get("fullscreenElement")
	if !fs.IsUndefined() && !fs.IsNull() {
		doc.Call("exitFullscreen")
	} else {
		dom.ByID("game-container").Call("requestFullscreen")
	}
}

func startGame() {
	if ctx.Value().IsUndefined() || ctx.Value().IsNull() {
		ctx = webgl.NewContext("glcanvas")
		ctx.ClearColor(0, 0, 0, 1)
		canvas := dom.ByID("glcanvas")
		ctx.Viewport(0, 0, canvas.Get("width").Int(), canvas.Get("height").Int())
		ctx.Enable(webgl.DEPTH_TEST)
		ctx.DepthFunc(webgl.LEQUAL)
		ctx.Enable(webgl.BLEND)
		ctx.BlendFunc(webgl.SRC_ALPHA, webgl.ONE)

		vertexSrc := `attribute vec2 a_position;
uniform vec2 u_translation;
uniform vec2 u_scale;
void main(){
    vec2 pos = a_position * u_scale + u_translation;
    gl_Position = vec4(pos,0.0,1.0);
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
		inds := js.Get("Uint16Array").New(6)
		for i, v := range []uint16{0, 1, 2, 2, 1, 3} {
			inds.SetIndex(i, v)
		}
		ctx.BufferData(webgl.ELEMENT_ARRAY_BUFFER, inds, webgl.STATIC_DRAW)

		posLoc := ctx.GetAttribLocation(prog, "a_position")
		ctx.EnableVertexAttribArray(posLoc)
		ctx.VertexAttribPointer(posLoc, 2, webgl.FLOAT, false, 0, 0)

		colorLoc = ctx.GetUniformLocation(prog, "u_color")
		transLoc = ctx.GetUniformLocation(prog, "u_translation")
		scaleLoc = ctx.GetUniformLocation(prog, "u_scale")
		timeLoc = ctx.GetUniformLocation(prog, "u_time")

		render = js.FuncOf(renderLoop)
		js.RequestAnimationFrame(render)
	}

	snake = []point{{gridSize / 2, gridSize / 2}, {gridSize/2 - 1, gridSize / 2}, {gridSize/2 - 2, gridSize / 2}}
	dir = point{1, 0}
	score = 0
	running = true
	lastMove = 0
	updateScore()
	newFood()
	dom.AddClass(dom.ByID("menu"), "hidden")
}

func renderLoop(this jst.Value, args []jst.Value) any {
	t := args[0].Float() / 1000
	ctx.Clear(webgl.COLOR_BUFFER_BIT)
	ctx.Uniform1f(timeLoc, float32(t))

	if running {
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
		if t-lastMove > 0.15 {
			moveSnake()
			lastMove = t
		}
	}

	drawSquare(food, [4]float32{1, 0, 0, 1})
	for _, p := range snake {
		drawSquare(p, [4]float32{0, 1, 0, 1})
	}

	js.RequestAnimationFrame(render)
	return nil
}

func moveSnake() {
	head := point{snake[0].x + dir.x, snake[0].y + dir.y}
	if head.x < 0 {
		head.x = gridSize - 1
	}
	if head.x >= gridSize {
		head.x = 0
	}
	if head.y < 0 {
		head.y = gridSize - 1
	}
	if head.y >= gridSize {
		head.y = 0
	}
	for _, s := range snake {
		if s == head {
			running = false
			dom.RemoveClass(dom.ByID("menu"), "hidden")
			return
		}
	}
	snake = append([]point{head}, snake...)
	if head == food {
		score++
		updateScore()
		newFood()
	} else {
		snake = snake[:len(snake)-1]
	}
}

func drawSquare(p point, color [4]float32) {
	x := -1 + cellSize*float32(p.x) + cellSize/2
	y := -1 + cellSize*float32(p.y) + cellSize/2
	ctx.Uniform2f(transLoc, x, y)

	ctx.Uniform2f(scaleLoc, cellSize*1.4, cellSize*1.4)
	ctx.Uniform4f(colorLoc, color[0], color[1], color[2], color[3]*0.3)
	ctx.DrawElements(webgl.TRIANGLES, 6, webgl.UNSIGNED_SHORT, 0)

	ctx.Uniform2f(scaleLoc, cellSize, cellSize)
	ctx.Uniform4f(colorLoc, color[0], color[1], color[2], color[3])
	ctx.DrawElements(webgl.TRIANGLES, 6, webgl.UNSIGNED_SHORT, 0)
}

func newFood() {
	for {
		fx := js.Get("Math").Call("floor", js.Get("Math").Call("random").Float()*float64(gridSize)).Int()
		fy := js.Get("Math").Call("floor", js.Get("Math").Call("random").Float()*float64(gridSize)).Int()
		p := point{fx, fy}
		collision := false
		for _, s := range snake {
			if s == p {
				collision = true
				break
			}
		}
		if !collision {
			food = p
			return
		}
	}
}

func updateScore() {
	dom.SetText(dom.ByID("score"), fmt.Sprintf("Score: %d", score))
}
