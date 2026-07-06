package draw

// Canvas represents a 2D drawing surface built on top of a canvas element.
type Canvas struct {
	impl canvasImpl
}

type canvasImpl interface {
	valid() bool
	setSize(width, height float64)
	drawRect(Rect)
	drawCircle(Circle)
	drawLine(Line)
}

// Valid reports whether the canvas is ready for drawing commands.
func (c Canvas) Valid() bool {
	if c.impl == nil {
		return false
	}
	return c.impl.valid()
}

// SetSize updates the backing canvas dimensions.
func (c Canvas) SetSize(width, height float64) {
	if c.impl == nil {
		return
	}
	c.impl.setSize(width, height)
}

// Draw executes the provided commands in order when the canvas is valid.
func (c Canvas) Draw(cmds ...Command) {
	if !c.Valid() {
		return
	}
	for _, cmd := range cmds {
		if cmd != nil {
			cmd.draw(c)
		}
	}
}

func (c Canvas) drawRect(r Rect) {
	if c.impl == nil {
		return
	}
	c.impl.drawRect(r)
}

func (c Canvas) drawCircle(circle Circle) {
	if c.impl == nil {
		return
	}
	c.impl.drawCircle(circle)
}

func (c Canvas) drawLine(line Line) {
	if c.impl == nil {
		return
	}
	c.impl.drawLine(line)
}

// Command represents a draw instruction that can be executed on a Canvas.
type Command interface {
	draw(Canvas)
}

// Rect describes a rectangle drawing command.
type Rect struct {
	X, Y          float64
	Width, Height float64
	fillColor     string
	strokeColor   string
	strokeWidth   float64
}

// Rectangle returns a rectangle command builder for the given bounds.
func Rectangle(x, y, width, height float64) *Rect {
	return &Rect{X: x, Y: y, Width: width, Height: height}
}

// Fill configures the fill color for the rectangle.
func (r *Rect) Fill(color string) *Rect {
	r.fillColor = color
	return r
}

// Stroke configures the stroke color and width for the rectangle outline.
func (r *Rect) Stroke(color string, width float64) *Rect {
	r.strokeColor = color
	r.strokeWidth = width
	return r
}

func (r *Rect) draw(c Canvas) {
	c.drawRect(*r)
}

// Circle describes a circle drawing command.
type Circle struct {
	X, Y        float64
	Radius      float64
	fillColor   string
	strokeColor string
	strokeWidth float64
}

// Disc returns a circle command builder centered at the provided coordinates.
func Disc(x, y, radius float64) *Circle {
	return &Circle{X: x, Y: y, Radius: radius}
}

// Fill configures the fill color for the circle.
func (c *Circle) Fill(color string) *Circle {
	c.fillColor = color
	return c
}

// Stroke configures the stroke color and width for the circle outline.
func (c *Circle) Stroke(color string, width float64) *Circle {
	c.strokeColor = color
	c.strokeWidth = width
	return c
}

func (c *Circle) draw(canvas Canvas) {
	canvas.drawCircle(*c)
}

// Line describes a line segment drawing command.
type Line struct {
	X1, Y1      float64
	X2, Y2      float64
	strokeColor string
	strokeWidth float64
}

// Segment returns a line command builder for the provided endpoints.
func Segment(x1, y1, x2, y2 float64) *Line {
	return &Line{X1: x1, Y1: y1, X2: x2, Y2: y2}
}

// Stroke configures the stroke color and width for the line.
func (l *Line) Stroke(color string, width float64) *Line {
	l.strokeColor = color
	l.strokeWidth = width
	return l
}

func (l *Line) draw(c Canvas) {
	c.drawLine(*l)
}
