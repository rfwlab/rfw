package math

import (
	stdmath "math"
	"testing"
)

func almostEqual(a, b float32) bool { return stdmath.Abs(float64(a-b)) < 1e-5 }

func TestVecOperations(t *testing.T) {
	v := Vec2{1, 2}.Add(Vec2{3, 4})
	if v.X != 4 || v.Y != 6 {
		t.Fatalf("unexpected sum: %+v", v)
	}
	n := Vec3{3, 4, 0}.Normalize()
	if !almostEqual(n.X, 0.6) || !almostEqual(n.Y, 0.8) || !almostEqual(n.Z, 0) {
		t.Fatalf("unexpected normalization: %+v", n)
	}
}

func TestMat4Multiply(t *testing.T) {
	a := Translation(Vec3{1, 2, 3})
	b := Scale(Vec3{2, 2, 2})
	r := a.Mul(b)
	exp := Mat4{
		2, 0, 0, 0,
		0, 2, 0, 0,
		0, 0, 2, 0,
		1, 2, 3, 1,
	}
	for i := 0; i < 16; i++ {
		if !almostEqual(r[i], exp[i]) {
			t.Fatalf("matrix mismatch at %d: got %v want %v", i, r[i], exp[i])
		}
	}
}
