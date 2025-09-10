package math

import stdmath "math"

// Vec2 represents a 2D vector.
type Vec2 struct{ X, Y float32 }

// Add adds two vectors component-wise.
func (v Vec2) Add(o Vec2) Vec2 { return Vec2{v.X + o.X, v.Y + o.Y} }

// Normalize returns the unit vector in the same direction.
func (v Vec2) Normalize() Vec2 {
	l := float32(stdmath.Sqrt(float64(v.X*v.X + v.Y*v.Y)))
	if l == 0 {
		return Vec2{}
	}
	return Vec2{v.X / l, v.Y / l}
}

// Vec3 represents a 3D vector.
type Vec3 struct{ X, Y, Z float32 }

// Add adds two Vec3 values component-wise.
func (v Vec3) Add(o Vec3) Vec3 { return Vec3{v.X + o.X, v.Y + o.Y, v.Z + o.Z} }

// Normalize returns the unit vector of v.
func (v Vec3) Normalize() Vec3 {
	l := float32(stdmath.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z)))
	if l == 0 {
		return Vec3{}
	}
	return Vec3{v.X / l, v.Y / l, v.Z / l}
}

// Mat4 represents a 4x4 matrix in column-major order.
type Mat4 [16]float32

// Identity returns an identity matrix.
func Identity() Mat4 {
	return Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// Translation returns a matrix translating by v.
func Translation(v Vec3) Mat4 {
	m := Identity()
	m[12], m[13], m[14] = v.X, v.Y, v.Z
	return m
}

// Scale returns a scaling matrix for v.
func Scale(v Vec3) Mat4 {
	m := Identity()
	m[0], m[5], m[10] = v.X, v.Y, v.Z
	return m
}

// Mul multiplies m by n and returns the result.
func (m Mat4) Mul(n Mat4) Mat4 {
	var r Mat4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			r[j*4+i] = m[i]*n[j*4] + m[4+i]*n[j*4+1] + m[8+i]*n[j*4+2] + m[12+i]*n[j*4+3]
		}
	}
	return r
}

// Perspective returns a perspective projection matrix.
func Perspective(fovY, aspect, near, far float32) Mat4 {
	f := float32(1 / stdmath.Tan(float64(fovY)/2))
	var m Mat4
	m[0] = f / aspect
	m[5] = f
	m[10] = (far + near) / (near - far)
	m[11] = -1
	m[14] = (2 * far * near) / (near - far)
	return m
}

// Orthographic returns an orthographic projection matrix.
func Orthographic(left, right, bottom, top, near, far float32) Mat4 {
	var m Mat4
	m[0] = 2 / (right - left)
	m[5] = 2 / (top - bottom)
	m[10] = -2 / (far - near)
	m[12] = -(right + left) / (right - left)
	m[13] = -(top + bottom) / (top - bottom)
	m[14] = -(far + near) / (far - near)
	m[15] = 1
	return m
}
