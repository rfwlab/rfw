package draw

import "testing"

type recorder struct {
	validState bool
	width      float64
	height     float64
	rects      []Rect
	circles    []Circle
	lines      []Line
}

func (r *recorder) valid() bool { return r.validState }

func (r *recorder) setSize(width, height float64) {
	r.width = width
	r.height = height
}

func (r *recorder) drawRect(rect Rect) { r.rects = append(r.rects, rect) }

func (r *recorder) drawCircle(circle Circle) { r.circles = append(r.circles, circle) }

func (r *recorder) drawLine(line Line) { r.lines = append(r.lines, line) }

func TestCanvasDrawDelegates(t *testing.T) {
	impl := &recorder{validState: true}
	canvas := Canvas{impl: impl}

	rect := Rectangle(0, 0, 10, 20).Fill("#111111").Stroke("#ffffff", 2)
	circle := Disc(1, 2, 3).Fill("#eeeeee").Stroke("#000000", 1.5)
	line := Segment(-1, -1, 1, 1).Stroke("#ff00ff", 3)

	canvas.Draw(rect, nil, circle, line)

	if got := len(impl.rects); got != 1 {
		t.Fatalf("expected 1 rect, got %d", got)
	}
	if got := len(impl.circles); got != 1 {
		t.Fatalf("expected 1 circle, got %d", got)
	}
	if got := len(impl.lines); got != 1 {
		t.Fatalf("expected 1 line, got %d", got)
	}

	if impl.rects[0].fillColor != "#111111" || impl.rects[0].strokeColor != "#ffffff" || impl.rects[0].strokeWidth != 2 {
		t.Fatalf("unexpected rect payload: %#v", impl.rects[0])
	}
	if impl.circles[0].fillColor != "#eeeeee" || impl.circles[0].strokeColor != "#000000" || impl.circles[0].strokeWidth != 1.5 {
		t.Fatalf("unexpected circle payload: %#v", impl.circles[0])
	}
	if impl.lines[0].strokeColor != "#ff00ff" || impl.lines[0].strokeWidth != 3 {
		t.Fatalf("unexpected line payload: %#v", impl.lines[0])
	}
}

func TestCanvasInvalid(t *testing.T) {
	canvas := Canvas{}
	if canvas.Valid() {
		t.Fatal("expected zero-value canvas to be invalid")
	}
	canvas.Draw(Rectangle(0, 0, 1, 1))
	canvas.SetSize(10, 20)
}

func TestCanvasSetSizeDelegates(t *testing.T) {
	impl := &recorder{validState: true}
	canvas := Canvas{impl: impl}
	canvas.SetSize(640, 480)
	if impl.width != 640 || impl.height != 480 {
		t.Fatalf("expected setSize delegation, got width=%f height=%f", impl.width, impl.height)
	}
}

func TestCanvasValidUsesImpl(t *testing.T) {
	impl := &recorder{validState: false}
	canvas := Canvas{impl: impl}
	if canvas.Valid() {
		t.Fatal("expected canvas to report invalid when impl invalid")
	}
	impl.validState = true
	if !canvas.Valid() {
		t.Fatal("expected canvas to report valid after impl change")
	}
}
