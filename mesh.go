package fauxgl

import (
	"math"
)

// Mesh f
type Mesh struct {
	Triangles []*Triangle
	Lines     []*Line
	box       *Box
}

// NewEmptyMesh returns an empty mesh
func NewEmptyMesh() *Mesh {
	return &Mesh{}
}

// NewMesh returns a mesh with given data
func NewMesh(triangles []*Triangle, lines []*Line) *Mesh {
	return &Mesh{triangles, lines, nil}
}

// NewTriangleMesh returns a mesh with given data
func NewTriangleMesh(triangles []*Triangle) *Mesh {
	return &Mesh{triangles, nil, nil}
}

// NewLineMesh returns a mesh with given data
func NewLineMesh(lines []*Line) *Mesh {
	return &Mesh{nil, lines, nil}
}

func (m *Mesh) dirty() {
	m.box = nil
}

// Copy f
func (m *Mesh) Copy() *Mesh {
	triangles := make([]*Triangle, len(m.Triangles))
	lines := make([]*Line, len(m.Lines))
	for i, t := range m.Triangles {
		a := *t
		triangles[i] = &a
	}
	for i, l := range m.Lines {
		a := *l
		lines[i] = &a
	}
	return NewMesh(triangles, lines)
}

// Add f
func (m *Mesh) Add(b *Mesh) {
	m.Triangles = append(m.Triangles, b.Triangles...)
	m.Lines = append(m.Lines, b.Lines...)
	m.dirty()
}

// Volume f
func (m *Mesh) Volume() float64 {
	var v float64
	for _, t := range m.Triangles {
		p1 := t.V1.Position
		p2 := t.V2.Position
		p3 := t.V3.Position
		v += p1.X*(p2.Y*p3.Z-p3.Y*p2.Z) - p2.X*(p1.Y*p3.Z-p3.Y*p1.Z) + p3.X*(p1.Y*p2.Z-p2.Y*p1.Z)
	}
	return math.Abs(v / 6)
}

// SurfaceArea f
func (m *Mesh) SurfaceArea() float64 {
	var a float64
	for _, t := range m.Triangles {
		a += t.Area()
	}
	return a
}

func smoothNormalsThreshold(normal Vector, normals []Vector, threshold float64) Vector {
	result := Vector{}
	for _, x := range normals {
		if x.Dot(normal) >= threshold {
			result = result.Add(x)
		}
	}
	return result.Normalize()
}

// SmoothNormalsThreshold f
func (m *Mesh) SmoothNormalsThreshold(radians float64) {
	threshold := math.Cos(radians)
	lookup := make(map[Vector][]Vector)
	for _, t := range m.Triangles {
		lookup[t.V1.Position] = append(lookup[t.V1.Position], t.V1.Normal)
		lookup[t.V2.Position] = append(lookup[t.V2.Position], t.V2.Normal)
		lookup[t.V3.Position] = append(lookup[t.V3.Position], t.V3.Normal)
	}
	for _, t := range m.Triangles {
		t.V1.Normal = smoothNormalsThreshold(t.V1.Normal, lookup[t.V1.Position], threshold)
		t.V2.Normal = smoothNormalsThreshold(t.V2.Normal, lookup[t.V2.Position], threshold)
		t.V3.Normal = smoothNormalsThreshold(t.V3.Normal, lookup[t.V3.Position], threshold)
	}
}

// SmoothNormals f
func (m *Mesh) SmoothNormals() {
	lookup := make(map[Vector]Vector)
	for _, t := range m.Triangles {
		lookup[t.V1.Position] = lookup[t.V1.Position].Add(t.V1.Normal)
		lookup[t.V2.Position] = lookup[t.V2.Position].Add(t.V2.Normal)
		lookup[t.V3.Position] = lookup[t.V3.Position].Add(t.V3.Normal)
	}
	for k, v := range lookup {
		lookup[k] = v.Normalize()
	}
	for _, t := range m.Triangles {
		t.V1.Normal = lookup[t.V1.Position]
		t.V2.Normal = lookup[t.V2.Position]
		t.V3.Normal = lookup[t.V3.Position]
	}
}

// UnitCube f
func (m *Mesh) UnitCube() Matrix {
	const r = 0.5
	return m.FitInside(Box{Vector{-r, -r, -r}, Vector{r, r, r}}, Vector{0.5, 0.5, 0.5})
}

// BiUnitCube f
func (m *Mesh) BiUnitCube() Matrix {
	const r = 1
	return m.FitInside(Box{Vector{-r, -r, -r}, Vector{r, r, r}}, Vector{0.5, 0.5, 0.5})
}

// MoveTo f
func (m *Mesh) MoveTo(position, anchor Vector) Matrix {
	matrix := Translate(position.Sub(m.BoundingBox().Anchor(anchor)))
	m.Transform(matrix)
	return matrix
}

// Center f
func (m *Mesh) Center() Matrix {
	return m.MoveTo(Vector{}, Vector{0.5, 0.5, 0.5})
}

// FitInside f
func (m *Mesh) FitInside(box Box, anchor Vector) Matrix {
	scale := box.Size().Div(m.BoundingBox().Size()).MinComponent()
	extra := box.Size().Sub(m.BoundingBox().Size().MulScalar(scale))
	matrix := Identity()
	matrix = matrix.Translate(m.BoundingBox().Min.Negate())
	matrix = matrix.Scale(Vector{scale, scale, scale})
	matrix = matrix.Translate(box.Min.Add(extra.Mul(anchor)))
	m.Transform(matrix)
	return matrix
}

// BoundingBox f
func (m *Mesh) BoundingBox() Box {
	if m.box == nil {
		box := EmptyBox
		for _, t := range m.Triangles {
			box = box.Extend(t.BoundingBox())
		}
		for _, l := range m.Lines {
			box = box.Extend(l.BoundingBox())
		}
		m.box = &box
	}
	return *m.box
}

// Transform f
func (m *Mesh) Transform(matrix Matrix) {
	// 使用SIMD优化的批量变换
	if len(m.Triangles) > 1000 {
		// 对于大网格，使用SIMD优化
		m.transformWithSIMD(matrix)
	} else {
		// 对于小网格，使用传统方法
		for _, t := range m.Triangles {
			t.Transform(matrix)
		}
		for _, l := range m.Lines {
			l.Transform(matrix)
		}
	}
	m.dirty()
}

// transformWithSIMD 使用SIMD优化的变换
func (m *Mesh) transformWithSIMD(matrix Matrix) {
	// 将矩阵转换为SIMD格式
	simdMatrix := NewSIMDMat4(
		matrix.X00, matrix.X01, matrix.X02, matrix.X03,
		matrix.X10, matrix.X11, matrix.X12, matrix.X13,
		matrix.X20, matrix.X21, matrix.X22, matrix.X23,
		matrix.X30, matrix.X31, matrix.X32, matrix.X33,
	)

	// 计算法线变换矩阵（逆转置）
	normalMatrix := matrix.Transpose().Inverse()

	// 批量处理三角形顶点
	for _, t := range m.Triangles {
		// 转换顶点为SIMD格式
		sv1 := NewSIMDVertex(t.V1.Position, t.V1.Normal, t.V1.Texture, t.V1.Color)
		sv2 := NewSIMDVertex(t.V2.Position, t.V2.Normal, t.V2.Texture, t.V2.Color)
		sv3 := NewSIMDVertex(t.V3.Position, t.V3.Normal, t.V3.Texture, t.V3.Color)

		// 使用SIMD矩阵变换顶点
		tv1 := simdMatrix.MulPositionSIMD(sv1.Position)
		tv2 := simdMatrix.MulPositionSIMD(sv2.Position)
		tv3 := simdMatrix.MulPositionSIMD(sv3.Position)

		// 转换回普通顶点
		t.V1.Position = tv1.ToVector()
		t.V2.Position = tv2.ToVector()
		t.V3.Position = tv3.ToVector()

		// 变换法线（使用矩阵的逆转置）
		if !math.IsNaN(normalMatrix.X00) && !math.IsInf(normalMatrix.X00, 0) {
			t.V1.Normal = normalMatrix.MulDirection(t.V1.Normal)
			t.V2.Normal = normalMatrix.MulDirection(t.V2.Normal)
			t.V3.Normal = normalMatrix.MulDirection(t.V3.Normal)
		}
	}

	// 批量处理线条顶点
	for _, l := range m.Lines {
		sv1 := NewSIMDVertex(l.V1.Position, l.V1.Normal, l.V1.Texture, l.V1.Color)
		sv2 := NewSIMDVertex(l.V2.Position, l.V2.Normal, l.V2.Texture, l.V2.Color)

		tv1 := simdMatrix.MulPositionSIMD(sv1.Position)
		tv2 := simdMatrix.MulPositionSIMD(sv2.Position)

		l.V1.Position = tv1.ToVector()
		l.V2.Position = tv2.ToVector()

		if !math.IsNaN(normalMatrix.X00) && !math.IsInf(normalMatrix.X00, 0) {
			l.V1.Normal = normalMatrix.MulDirection(l.V1.Normal)
			l.V2.Normal = normalMatrix.MulDirection(l.V2.Normal)
		}
	}
}

// ReverseWinding  f
func (m *Mesh) ReverseWinding() {
	for _, t := range m.Triangles {
		t.ReverseWinding()
	}
}

// Simplify f
func (m *Mesh) Simplify(factor float64) {
	// 直接使用优化的网格简化
	m.simplifyInternally(factor)
	m.dirty()
}

// simplifyInternally 内部实现的网格简化
func (m *Mesh) simplifyInternally(factor float64) {
	// 如果简化因子为1.0或更高，则不进行简化
	if factor >= 1.0 {
		return
	}

	// 如果简化因子为0或更低，则清空网格
	if factor <= 0 {
		m.Triangles = make([]*Triangle, 0)
		return
	}

	// 计算要保留的三角形数量
	targetCount := int(float64(len(m.Triangles)) * factor)
	if targetCount <= 0 {
		targetCount = 1
	}

	// 简单的简化算法：均匀采样保留三角形
	// 在实际应用中，可以实现更复杂的算法如边折叠等
	step := len(m.Triangles) / targetCount
	if step <= 0 {
		step = 1
	}

	newTriangles := make([]*Triangle, 0, targetCount)
	for i := 0; i < len(m.Triangles); i += step {
		newTriangles = append(newTriangles, m.Triangles[i])
		if len(newTriangles) >= targetCount {
			break
		}
	}

	m.Triangles = newTriangles
}

func (m *Mesh) SplitTriangles(maxEdgeLength float64) {
	var triangles []*Triangle

	var split func(t *Triangle)

	split = func(t *Triangle) {
		v1 := t.V1
		v2 := t.V2
		v3 := t.V3
		p1 := v1.Position
		p2 := v2.Position
		p3 := v3.Position
		d12 := p1.Distance(p2)
		d23 := p2.Distance(p3)
		d31 := p3.Distance(p1)
		max := math.Max(d12, math.Max(d23, d31))
		if max <= maxEdgeLength {
			triangles = append(triangles, t)
		} else if d12 == max {
			v := InterpolateVertexes(v1, v2, v3, VectorW{0.5, 0.5, 0, 1})
			t1 := NewTriangle(v3, v1, v)
			t2 := NewTriangle(v2, v3, v)
			split(t1)
			split(t2)
		} else if d23 == max {
			v := InterpolateVertexes(v1, v2, v3, VectorW{0, 0.5, 0.5, 1})
			t1 := NewTriangle(v1, v2, v)
			t2 := NewTriangle(v3, v1, v)
			split(t1)
			split(t2)
		} else {
			v := InterpolateVertexes(v1, v2, v3, VectorW{0.5, 0, 0.5, 1})
			t1 := NewTriangle(v2, v3, v)
			t2 := NewTriangle(v1, v2, v)
			split(t1)
			split(t2)
		}
	}

	for _, t := range m.Triangles {
		split(t)
	}

	m.Triangles = triangles
	m.dirty()
}

func (m *Mesh) SharpEdges(angleThreshold float64) *Mesh {
	type Edge struct {
		A, B Vector
	}

	makeEdge := func(a, b Vector) Edge {
		if a.Less(b) {
			return Edge{a, b}
		}
		return Edge{b, a}
	}

	var lines []*Line
	other := make(map[Edge]*Triangle)
	for _, t := range m.Triangles {
		p1 := t.V1.Position
		p2 := t.V2.Position
		p3 := t.V3.Position
		e1 := makeEdge(p1, p2)
		e2 := makeEdge(p2, p3)
		e3 := makeEdge(p3, p1)
		for _, e := range []Edge{e1, e2, e3} {
			if u, ok := other[e]; ok {
				a := math.Acos(t.Normal().Dot(u.Normal()))
				if a > angleThreshold {
					lines = append(lines, NewLineForPoints(e.A, e.B))
				}
			}
		}
		other[e1] = t
		other[e2] = t
		other[e3] = t
	}
	return NewLineMesh(lines)
}
