package fauxgl

import "math"

// Matrix f
type Matrix struct {
	X00, X01, X02, X03 float64
	X10, X11, X12, X13 float64
	X20, X21, X22, X23 float64
	X30, X31, X32, X33 float64
}

// Identity f
func Identity() Matrix {
	return Matrix{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1}
}

// Translate f
func Translate(v Vector) Matrix {
	return Matrix{
		1, 0, 0, v.X,
		0, 1, 0, v.Y,
		0, 0, 1, v.Z,
		0, 0, 0, 1}
}

// Scale f
func Scale(v Vector) Matrix {
	return Matrix{
		v.X, 0, 0, 0,
		0, v.Y, 0, 0,
		0, 0, v.Z, 0,
		0, 0, 0, 1}
}

// Rotate f
func Rotate(v Vector, a float64) Matrix {
	v = v.Normalize()
	s := math.Sin(a)
	c := math.Cos(a)
	m := 1 - c
	return Matrix{
		m*v.X*v.X + c, m*v.X*v.Y + v.Z*s, m*v.Z*v.X - v.Y*s, 0,
		m*v.X*v.Y - v.Z*s, m*v.Y*v.Y + c, m*v.Y*v.Z + v.X*s, 0,
		m*v.Z*v.X + v.Y*s, m*v.Y*v.Z - v.X*s, m*v.Z*v.Z + c, 0,
		0, 0, 0, 1}
}

// RotateTo f
func RotateTo(a, b Vector) Matrix {
	dot := b.Dot(a)
	if dot == 1 {
		return Identity()
	} else if dot == -1 {
		return Rotate(a.Perpendicular(), math.Pi)
	} else {
		angle := math.Acos(dot)
		v := b.Cross(a).Normalize()
		return Rotate(v, angle)
	}
}

// Orient f
func Orient(position, size, up Vector, rotation float64) Matrix {
	m := Rotate(Vector{0, 0, 1}, rotation)
	m = m.Scale(size)
	m = m.RotateTo(Vector{0, 0, 1}, up)
	m = m.Translate(position)
	return m
}

// Frustum f
func Frustum(l, r, b, t, n, f float64) Matrix {
	return Matrix{
		2 * n / (r - l), 0, (r + l) / (r - l), 0,
		0, 2 * n / (t - b), (t + b) / (t - b), 0,
		0, 0, -(f + n) / (f - n), -2 * f * n / (f - n),
		0, 0, -1, 0}
}

// Orthographic f
func Orthographic(l, r, b, t, n, f float64) Matrix {
	return Matrix{
		2 / (r - l), 0, 0, -(r + l) / (r - l),
		0, 2 / (t - b), 0, -(t + b) / (t - b),
		0, 0, -2 / (f - n), -(f + n) / (f - n),
		0, 0, 0, 1}
}

// Perspective f
func Perspective(fovy, aspect, near, far float64) Matrix {
	ymax := near * math.Tan(fovy*math.Pi/360)
	xmax := ymax * aspect
	return Frustum(-xmax, xmax, -ymax, ymax, near, far)
}

// LookAt f
func LookAt(eye, center, up Vector) Matrix {
	z := eye.Sub(center).Normalize()
	x := up.Cross(z).Normalize()
	y := z.Cross(x)
	return Matrix{
		x.X, x.Y, x.Z, -x.Dot(eye),
		y.X, y.Y, y.Z, -y.Dot(eye),
		z.X, z.Y, z.Z, -z.Dot(eye),
		0, 0, 0, 1,
	}
}

// LookAtDirection f
func LookAtDirection(forward, up Vector) Matrix {
	z := forward.Normalize()
	x := up.Cross(z).Normalize()
	y := z.Cross(x)
	return Matrix{
		x.X, x.Y, x.Z, 0,
		y.X, y.Y, y.Z, 0,
		z.X, z.Y, z.Z, 0,
		0, 0, 0, 1,
	}
}

// Screen f
func Screen(w, h int) Matrix {
	w2 := float64(w) / 2
	h2 := float64(h) / 2
	return Matrix{
		w2, 0, 0, w2,
		0, -h2, 0, h2,
		0, 0, 0.5, 0.5,
		0, 0, 0, 1,
	}
}

// Viewport f
func Viewport(x, y, w, h float64) Matrix {
	l := x
	b := y
	r := x + w
	t := y + h
	return Matrix{
		(r - l) / 2, 0, 0, (r + l) / 2,
		0, (t - b) / 2, 0, (t + b) / 2,
		0, 0, 0.5, 0.5,
		0, 0, 0, 1,
	}
}

// Translate f
func (a Matrix) Translate(v Vector) Matrix {
	return Translate(v).Mul(a)
}

// Scale f
func (a Matrix) Scale(v Vector) Matrix {
	return Scale(v).Mul(a)
}

// Rotate f
func (a Matrix) Rotate(v Vector, f float64) Matrix {
	return Rotate(v, f).Mul(a)
}

// RotateTo f
func (a Matrix) RotateTo(b, c Vector) Matrix {
	return RotateTo(b, c).Mul(a)
}

// Frustum f
func (a Matrix) Frustum(l, r, b, t, n, f float64) Matrix {
	return Frustum(l, r, b, t, n, f).Mul(a)
}

// Orthographic f
func (a Matrix) Orthographic(l, r, b, t, n, f float64) Matrix {
	return Orthographic(l, r, b, t, n, f).Mul(a)
}

// Perspective f
func (a Matrix) Perspective(fovy, aspect, near, far float64) Matrix {
	return Perspective(fovy, aspect, near, far).Mul(a)
}

// LookAt f
func (a Matrix) LookAt(eye, center, up Vector) Matrix {
	return LookAt(eye, center, up).Mul(a)
}

// Mul f
func (a Matrix) Mul(b Matrix) Matrix {
	return Matrix{
		a.X00*b.X00 + a.X01*b.X10 + a.X02*b.X20 + a.X03*b.X30,
		a.X00*b.X01 + a.X01*b.X11 + a.X02*b.X21 + a.X03*b.X31,
		a.X00*b.X02 + a.X01*b.X12 + a.X02*b.X22 + a.X03*b.X32,
		a.X00*b.X03 + a.X01*b.X13 + a.X02*b.X23 + a.X03*b.X33,
		a.X10*b.X00 + a.X11*b.X10 + a.X12*b.X20 + a.X13*b.X30,
		a.X10*b.X01 + a.X11*b.X11 + a.X12*b.X21 + a.X13*b.X31,
		a.X10*b.X02 + a.X11*b.X12 + a.X12*b.X22 + a.X13*b.X32,
		a.X10*b.X03 + a.X11*b.X13 + a.X12*b.X23 + a.X13*b.X33,
		a.X20*b.X00 + a.X21*b.X10 + a.X22*b.X20 + a.X23*b.X30,
		a.X20*b.X01 + a.X21*b.X11 + a.X22*b.X21 + a.X23*b.X31,
		a.X20*b.X02 + a.X21*b.X12 + a.X22*b.X22 + a.X23*b.X32,
		a.X20*b.X03 + a.X21*b.X13 + a.X22*b.X23 + a.X23*b.X33,
		a.X30*b.X00 + a.X31*b.X10 + a.X32*b.X20 + a.X33*b.X30,
		a.X30*b.X01 + a.X31*b.X11 + a.X32*b.X21 + a.X33*b.X31,
		a.X30*b.X02 + a.X31*b.X12 + a.X32*b.X22 + a.X33*b.X32,
		a.X30*b.X03 + a.X31*b.X13 + a.X32*b.X23 + a.X33*b.X33,
	}
}

// MulPosition f
func (a Matrix) MulPosition(b Vector) Vector {
	x := a.X00*b.X + a.X01*b.Y + a.X02*b.Z + a.X03
	y := a.X10*b.X + a.X11*b.Y + a.X12*b.Z + a.X13
	z := a.X20*b.X + a.X21*b.Y + a.X22*b.Z + a.X23
	return Vector{x, y, z}
}

// MulPositionW f
func (a Matrix) MulPositionW(b Vector) VectorW {
	// 使用SIMD优化的位置变换
	simdMatrix := NewSIMDMat4(
		a.X00, a.X01, a.X02, a.X03,
		a.X10, a.X11, a.X12, a.X13,
		a.X20, a.X21, a.X22, a.X23,
		a.X30, a.X31, a.X32, a.X33,
	)
	simdVector := NewSIMDVector4FromVector(b)
	result := simdMatrix.MulPositionSIMD(simdVector)
	return VectorW{result[0], result[1], result[2], result[3]}
}

// MulDirection f
func (a Matrix) MulDirection(b Vector) Vector {
	// 使用SIMD优化的方向向量变换
	simdMatrix := NewSIMDMat4(
		a.X00, a.X01, a.X02, 0,
		a.X10, a.X11, a.X12, 0,
		a.X20, a.X21, a.X22, 0,
		0, 0, 0, 1,
	)
	simdVector := NewSIMDVector4FromVector(b)
	result := simdMatrix.MulPositionSIMD(simdVector)
	return result.ToVector().Normalize()
}

// MulBox f
func (a Matrix) MulBox(box Box) Box {
	// http://dev.theomader.com/transform-bounding-boxes/
	r := Vector{a.X00, a.X10, a.X20}
	u := Vector{a.X01, a.X11, a.X21}
	b := Vector{a.X02, a.X12, a.X22}
	t := Vector{a.X03, a.X13, a.X23}
	xa := r.MulScalar(box.Min.X)
	xb := r.MulScalar(box.Max.X)
	ya := u.MulScalar(box.Min.Y)
	yb := u.MulScalar(box.Max.Y)
	za := b.MulScalar(box.Min.Z)
	zb := b.MulScalar(box.Max.Z)
	xa, xb = xa.Min(xb), xa.Max(xb)
	ya, yb = ya.Min(yb), ya.Max(yb)
	za, zb = za.Min(zb), za.Max(zb)
	min := xa.Add(ya).Add(za).Add(t)
	max := xb.Add(yb).Add(zb).Add(t)
	return Box{min, max}
}

// Transpose f
func (a Matrix) Transpose() Matrix {
	return Matrix{
		a.X00, a.X10, a.X20, a.X30,
		a.X01, a.X11, a.X21, a.X31,
		a.X02, a.X12, a.X22, a.X32,
		a.X03, a.X13, a.X23, a.X33}
}

// Determinant f
func (a Matrix) Determinant() float64 {
	return (a.X00*a.X11*a.X22*a.X33 - a.X00*a.X11*a.X23*a.X32 +
		a.X00*a.X12*a.X23*a.X31 - a.X00*a.X12*a.X21*a.X33 +
		a.X00*a.X13*a.X21*a.X32 - a.X00*a.X13*a.X22*a.X31 -
		a.X01*a.X12*a.X23*a.X30 + a.X01*a.X12*a.X20*a.X33 -
		a.X01*a.X13*a.X20*a.X32 + a.X01*a.X13*a.X22*a.X30 -
		a.X01*a.X10*a.X22*a.X33 + a.X01*a.X10*a.X23*a.X32 +
		a.X02*a.X13*a.X20*a.X31 - a.X02*a.X13*a.X21*a.X30 +
		a.X02*a.X10*a.X21*a.X33 - a.X02*a.X10*a.X23*a.X31 +
		a.X02*a.X11*a.X23*a.X30 - a.X02*a.X11*a.X20*a.X33 -
		a.X03*a.X10*a.X21*a.X32 + a.X03*a.X10*a.X22*a.X31 -
		a.X03*a.X11*a.X22*a.X30 + a.X03*a.X11*a.X20*a.X32 -
		a.X03*a.X12*a.X20*a.X31 + a.X03*a.X12*a.X21*a.X30)
}

// Inverse f
func (a Matrix) Inverse() Matrix {
	d := a.Determinant()
	if d == 0 {
		return Identity()
	}
	m := Matrix{}
	m.X00 = (a.X12*a.X23*a.X31 - a.X13*a.X22*a.X31 + a.X13*a.X21*a.X32 - a.X11*a.X23*a.X32 - a.X12*a.X21*a.X33 + a.X11*a.X22*a.X33) / d
	m.X01 = (a.X03*a.X22*a.X31 - a.X02*a.X23*a.X31 - a.X03*a.X21*a.X32 + a.X01*a.X23*a.X32 + a.X02*a.X21*a.X33 - a.X01*a.X22*a.X33) / d
	m.X02 = (a.X02*a.X13*a.X31 - a.X03*a.X12*a.X31 + a.X03*a.X11*a.X32 - a.X01*a.X13*a.X32 - a.X02*a.X11*a.X33 + a.X01*a.X12*a.X33) / d
	m.X03 = (a.X03*a.X12*a.X21 - a.X02*a.X13*a.X21 - a.X03*a.X11*a.X22 + a.X01*a.X13*a.X22 + a.X02*a.X11*a.X23 - a.X01*a.X12*a.X23) / d
	m.X10 = (a.X13*a.X22*a.X30 - a.X12*a.X23*a.X30 - a.X13*a.X20*a.X32 + a.X10*a.X23*a.X32 + a.X12*a.X20*a.X33 - a.X10*a.X22*a.X33) / d
	m.X11 = (a.X02*a.X23*a.X30 - a.X03*a.X22*a.X30 + a.X03*a.X20*a.X32 - a.X00*a.X23*a.X32 - a.X02*a.X20*a.X33 + a.X00*a.X22*a.X33) / d
	m.X12 = (a.X03*a.X12*a.X30 - a.X02*a.X13*a.X30 - a.X03*a.X10*a.X32 + a.X00*a.X13*a.X32 + a.X02*a.X10*a.X33 - a.X00*a.X12*a.X33) / d
	m.X13 = (a.X02*a.X13*a.X20 - a.X03*a.X12*a.X20 + a.X03*a.X10*a.X22 - a.X00*a.X13*a.X22 - a.X02*a.X10*a.X23 + a.X00*a.X12*a.X23) / d
	m.X20 = (a.X11*a.X23*a.X30 - a.X13*a.X21*a.X30 + a.X13*a.X20*a.X31 - a.X10*a.X23*a.X31 - a.X11*a.X20*a.X33 + a.X10*a.X21*a.X33) / d
	m.X21 = (a.X03*a.X21*a.X30 - a.X01*a.X23*a.X30 - a.X03*a.X20*a.X31 + a.X00*a.X23*a.X31 + a.X01*a.X20*a.X33 - a.X00*a.X21*a.X33) / d
	m.X22 = (a.X01*a.X13*a.X30 - a.X03*a.X11*a.X30 + a.X03*a.X10*a.X31 - a.X00*a.X13*a.X31 - a.X01*a.X10*a.X33 + a.X00*a.X11*a.X33) / d
	m.X23 = (a.X03*a.X11*a.X20 - a.X01*a.X13*a.X20 - a.X03*a.X10*a.X21 + a.X00*a.X13*a.X21 + a.X01*a.X10*a.X23 - a.X00*a.X11*a.X23) / d
	m.X30 = (a.X12*a.X21*a.X30 - a.X11*a.X22*a.X30 - a.X12*a.X20*a.X31 + a.X10*a.X22*a.X31 + a.X11*a.X20*a.X32 - a.X10*a.X21*a.X32) / d
	m.X31 = (a.X01*a.X22*a.X30 - a.X02*a.X21*a.X30 + a.X02*a.X20*a.X31 - a.X00*a.X22*a.X31 - a.X01*a.X20*a.X32 + a.X00*a.X21*a.X32) / d
	m.X32 = (a.X02*a.X11*a.X30 - a.X01*a.X12*a.X30 - a.X02*a.X10*a.X31 + a.X00*a.X12*a.X31 + a.X01*a.X10*a.X32 - a.X00*a.X11*a.X32) / d
	m.X33 = (a.X01*a.X12*a.X20 - a.X02*a.X11*a.X20 + a.X02*a.X10*a.X21 - a.X00*a.X12*a.X21 - a.X01*a.X10*a.X22 + a.X00*a.X11*a.X22) / d
	return m
}
