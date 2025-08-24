package fauxgl

import "math"

// NewCube f
func NewCube() *Mesh {
	v := []Vector{
		{-1, -1, -1}, {-1, -1, 1}, {-1, 1, -1}, {-1, 1, 1},
		{1, -1, -1}, {1, -1, 1}, {1, 1, -1}, {1, 1, 1},
	}
	mesh := NewTriangleMesh([]*Triangle{
		NewTriangleForPoints(v[3], v[5], v[7]),
		NewTriangleForPoints(v[5], v[3], v[1]),
		NewTriangleForPoints(v[0], v[6], v[4]),
		NewTriangleForPoints(v[6], v[0], v[2]),
		NewTriangleForPoints(v[0], v[5], v[1]),
		NewTriangleForPoints(v[5], v[0], v[4]),
		NewTriangleForPoints(v[5], v[6], v[7]),
		NewTriangleForPoints(v[6], v[5], v[4]),
		NewTriangleForPoints(v[6], v[3], v[7]),
		NewTriangleForPoints(v[3], v[6], v[2]),
		NewTriangleForPoints(v[0], v[3], v[2]),
		NewTriangleForPoints(v[3], v[0], v[1]),
	})
	mesh.Transform(Scale(Vector{0.5, 0.5, 0.5}))
	return mesh
}

// NewCubeForBox f
func NewCubeForBox(box Box) *Mesh {
	m := Translate(Vector{0.5, 0.5, 0.5})
	m = m.Scale(box.Size())
	m = m.Translate(box.Min)
	cube := NewCube()
	cube.Transform(m)
	return cube
}

// NewCubeOutlineForBox f
func NewCubeOutlineForBox(box Box) *Mesh {
	x0 := box.Min.X
	y0 := box.Min.Y
	z0 := box.Min.Z
	x1 := box.Max.X
	y1 := box.Max.Y
	z1 := box.Max.Z
	return NewLineMesh([]*Line{
		NewLineForPoints(Vector{x0, y0, z0}, Vector{x0, y0, z1}),
		NewLineForPoints(Vector{x0, y1, z0}, Vector{x0, y1, z1}),
		NewLineForPoints(Vector{x1, y0, z0}, Vector{x1, y0, z1}),
		NewLineForPoints(Vector{x1, y1, z0}, Vector{x1, y1, z1}),
		NewLineForPoints(Vector{x0, y0, z0}, Vector{x0, y1, z0}),
		NewLineForPoints(Vector{x0, y0, z1}, Vector{x0, y1, z1}),
		NewLineForPoints(Vector{x1, y0, z0}, Vector{x1, y1, z0}),
		NewLineForPoints(Vector{x1, y0, z1}, Vector{x1, y1, z1}),
		NewLineForPoints(Vector{x0, y0, z0}, Vector{x1, y0, z0}),
		NewLineForPoints(Vector{x0, y1, z0}, Vector{x1, y1, z0}),
		NewLineForPoints(Vector{x0, y0, z1}, Vector{x1, y0, z1}),
		NewLineForPoints(Vector{x0, y1, z1}, Vector{x1, y1, z1}),
	})
}

// NewSphere f
func NewSphere(detail int) *Mesh {
	var triangles []*Triangle
	ico := NewIcosahedron()
	for _, t := range ico.Triangles {
		v1 := t.V1.Position
		v2 := t.V2.Position
		v3 := t.V3.Position
		triangles = append(triangles, newSphereHelper(detail, v1, v2, v3)...)
	}
	return NewTriangleMesh(triangles)
}

func newSphereHelper(detail int, v1, v2, v3 Vector) []*Triangle {
	if detail == 0 {
		t := NewTriangleForPoints(v1, v2, v3)
		return []*Triangle{t}
	}
	var triangles []*Triangle
	v12 := v1.Add(v2).DivScalar(2).Normalize()
	v13 := v1.Add(v3).DivScalar(2).Normalize()
	v23 := v2.Add(v3).DivScalar(2).Normalize()
	triangles = append(triangles, newSphereHelper(detail-1, v1, v12, v13)...)
	triangles = append(triangles, newSphereHelper(detail-1, v2, v23, v12)...)
	triangles = append(triangles, newSphereHelper(detail-1, v3, v13, v23)...)
	triangles = append(triangles, newSphereHelper(detail-1, v12, v23, v13)...)
	return triangles
}

// NewCone f
func NewCone(step int, capped bool) *Mesh {
	var triangles []*Triangle
	for a0 := 0; a0 < 360; a0 += step {
		a1 := (a0 + step) % 360
		r0 := Radians(float64(a0))
		r1 := Radians(float64(a1))
		x0 := math.Cos(r0)
		y0 := math.Sin(r0)
		x1 := math.Cos(r1)
		y1 := math.Sin(r1)
		p00 := Vector{x0, y0, -0.5}
		p10 := Vector{x1, y1, -0.5}
		p1 := Vector{0, 0, 0.5}
		t1 := NewTriangleForPoints(p00, p10, p1)
		triangles = append(triangles, t1)
		if capped {
			p0 := Vector{0, 0, -0.5}
			t2 := NewTriangleForPoints(p0, p10, p00)
			triangles = append(triangles, t2)
		}
	}
	return NewTriangleMesh(triangles)
}

// NewIcosahedron f
func NewIcosahedron() *Mesh {
	const a = 0.8506507174597755
	const b = 0.5257312591858783
	vertices := []Vector{
		{-a, -b, 0},
		{-a, b, 0},
		{-b, 0, -a},
		{-b, 0, a},
		{0, -a, -b},
		{0, -a, b},
		{0, a, -b},
		{0, a, b},
		{b, 0, -a},
		{b, 0, a},
		{a, -b, 0},
		{a, b, 0},
	}
	indices := [][3]int{
		{0, 3, 1},
		{1, 3, 7},
		{2, 0, 1},
		{2, 1, 6},
		{4, 0, 2},
		{4, 5, 0},
		{5, 3, 0},
		{6, 1, 7},
		{6, 7, 11},
		{7, 3, 9},
		{8, 2, 6},
		{8, 4, 2},
		{8, 6, 11},
		{8, 10, 4},
		{8, 11, 10},
		{9, 3, 5},
		{10, 5, 4},
		{10, 9, 5},
		{11, 7, 9},
		{11, 9, 10},
	}
	triangles := make([]*Triangle, len(indices))
	for i, idx := range indices {
		p1 := vertices[idx[0]]
		p2 := vertices[idx[1]]
		p3 := vertices[idx[2]]
		triangles[i] = NewTriangleForPoints(p1, p2, p3)
	}
	return NewTriangleMesh(triangles)
}

// NewPlane creates a plane centered at the origin
func NewPlane(width, height float64) *Mesh {
	w := width / 2
	h := height / 2
	v := []Vector{
		{-w, 0, -h}, {w, 0, -h}, {w, 0, h}, {-w, 0, h},
	}
	return NewTriangleMesh([]*Triangle{
		NewTriangleForPoints(v[0], v[1], v[2]),
		NewTriangleForPoints(v[0], v[2], v[3]),
	})
}

// NewCylinder creates a cylinder with the specified parameters
func NewCylinder(radius, height float64, radialSegments, heightSegments int, openEnded bool) *Mesh {
	var triangles []*Triangle

	// Create vertices
	vertices := make([][]Vector, heightSegments+1)
	for y := 0; y <= heightSegments; y++ {
		vertices[y] = make([]Vector, radialSegments)
		v := float64(y)/float64(heightSegments)*height - height/2
		for x := 0; x < radialSegments; x++ {
			u := float64(x) / float64(radialSegments) * math.Pi * 2
			vertices[y][x] = Vector{math.Cos(u) * radius, v, math.Sin(u) * radius}
		}
	}

	// Create faces
	for y := 0; y < heightSegments; y++ {
		for x := 0; x < radialSegments; x++ {
			x1 := (x + 1) % radialSegments
			v1 := vertices[y][x]
			v2 := vertices[y+1][x]
			v3 := vertices[y][x1]
			v4 := vertices[y+1][x1]

			triangles = append(triangles, NewTriangleForPoints(v1, v2, v3))
			triangles = append(triangles, NewTriangleForPoints(v2, v4, v3))
		}
	}

	// Create top and bottom caps
	if !openEnded {
		topCenter := Vector{0, height / 2, 0}
		bottomCenter := Vector{0, -height / 2, 0}

		for x := 0; x < radialSegments; x++ {
			x1 := (x + 1) % radialSegments
			// Top cap
			triangles = append(triangles, NewTriangleForPoints(topCenter, vertices[heightSegments][x], vertices[heightSegments][x1]))
			// Bottom cap
			triangles = append(triangles, NewTriangleForPoints(bottomCenter, vertices[0][x1], vertices[0][x]))
		}
	}

	return NewTriangleMesh(triangles)
}

// NewTorus creates a torus with the specified parameters
func NewTorus(radius, tubeRadius float64, radialSegments, tubularSegments int) *Mesh {
	var triangles []*Triangle

	// Create vertices
	vertices := make([][]Vector, radialSegments)
	for i := 0; i < radialSegments; i++ {
		vertices[i] = make([]Vector, tubularSegments)
		u := float64(i) / float64(radialSegments) * math.Pi * 2
		for j := 0; j < tubularSegments; j++ {
			v := float64(j) / float64(tubularSegments) * math.Pi * 2
			vertex := Vector{
				(radius + tubeRadius*math.Cos(v)) * math.Cos(u),
				tubeRadius * math.Sin(v),
				(radius + tubeRadius*math.Cos(v)) * math.Sin(u),
			}
			vertices[i][j] = vertex
		}
	}

	// Create faces
	for i := 0; i < radialSegments; i++ {
		i1 := (i + 1) % radialSegments
		for j := 0; j < tubularSegments; j++ {
			j1 := (j + 1) % tubularSegments
			v1 := vertices[i][j]
			v2 := vertices[i][j1]
			v3 := vertices[i1][j]
			v4 := vertices[i1][j1]

			triangles = append(triangles, NewTriangleForPoints(v1, v2, v3))
			triangles = append(triangles, NewTriangleForPoints(v2, v4, v3))
		}
	}

	return NewTriangleMesh(triangles)
}

// NewCapsule creates a capsule (cylinder with hemispherical caps)
func NewCapsule(radius, height float64, radialSegments, heightSegments, capSegments int) *Mesh {
	var triangles []*Triangle

	// Create cylinder part
	cylinderHeight := height - 2*radius
	if cylinderHeight > 0 {
		cylinder := NewCylinder(radius, cylinderHeight, radialSegments, heightSegments, true)
		cylinder.Transform(Translate(Vector{0, 0, 0}))
		triangles = append(triangles, cylinder.Triangles...)
	}

	// Create top hemisphere
	topSphere := NewSphere(capSegments)
	topSphere.Transform(Scale(Vector{radius, radius, radius}).Translate(Vector{0, cylinderHeight / 2, 0}))
	triangles = append(triangles, topSphere.Triangles...)

	// Create bottom hemisphere
	bottomSphere := NewSphere(capSegments)
	bottomSphere.Transform(Scale(Vector{radius, radius, radius}).Translate(Vector{0, -cylinderHeight / 2, 0}))
	triangles = append(triangles, bottomSphere.Triangles...)

	return NewTriangleMesh(triangles)
}

// Subdivide subdivides a mesh using loop subdivision
func (m *Mesh) Subdivide() *Mesh {
	// This is a simplified subdivision implementation
	// A full implementation would be more complex

	// For now, we'll just add vertices at edge midpoints
	var newTriangles []*Triangle

	for _, t := range m.Triangles {
		// Calculate edge midpoints
		mid1 := InterpolateVertexes(t.V1, t.V2, t.V3, VectorW{0.5, 0.5, 0, 1})
		mid2 := InterpolateVertexes(t.V1, t.V2, t.V3, VectorW{0.5, 0, 0.5, 1})
		mid3 := InterpolateVertexes(t.V1, t.V2, t.V3, VectorW{0, 0.5, 0.5, 1})

		// Create four new triangles
		newTriangles = append(newTriangles, NewTriangle(t.V1, mid1, mid2))
		newTriangles = append(newTriangles, NewTriangle(t.V2, mid1, mid3))
		newTriangles = append(newTriangles, NewTriangle(t.V3, mid2, mid3))
		newTriangles = append(newTriangles, NewTriangle(mid1, mid2, mid3))
	}

	return NewTriangleMesh(newTriangles)
}

// Tessellate tessellates a mesh by splitting triangles
func (m *Mesh) Tessellate(maxEdgeLength float64) *Mesh {
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

	result := NewTriangleMesh(triangles)
	result.dirty()
	return result
}

// Smooth smooths a mesh by averaging vertex positions
func (m *Mesh) Smooth(iterations int) *Mesh {
	// Create a copy of the mesh
	result := &Mesh{
		Triangles: make([]*Triangle, len(m.Triangles)),
		Lines:     make([]*Line, len(m.Lines)),
	}

	// Copy triangles
	for i, t := range m.Triangles {
		result.Triangles[i] = &Triangle{
			V1: Vertex{
				Position: t.V1.Position,
				Normal:   t.V1.Normal,
				Texture:  t.V1.Texture,
				Color:    t.V1.Color,
				Output:   t.V1.Output,
			},
			V2: Vertex{
				Position: t.V2.Position,
				Normal:   t.V2.Normal,
				Texture:  t.V2.Texture,
				Color:    t.V2.Color,
				Output:   t.V2.Output,
			},
			V3: Vertex{
				Position: t.V3.Position,
				Normal:   t.V3.Normal,
				Texture:  t.V3.Texture,
				Color:    t.V3.Color,
				Output:   t.V3.Output,
			},
		}
	}

	// Copy lines
	for i, l := range m.Lines {
		result.Lines[i] = &Line{
			V1: Vertex{
				Position: l.V1.Position,
				Normal:   l.V1.Normal,
				Texture:  l.V1.Texture,
				Color:    l.V1.Color,
				Output:   l.V1.Output,
			},
			V2: Vertex{
				Position: l.V2.Position,
				Normal:   l.V2.Normal,
				Texture:  l.V2.Texture,
				Color:    l.V2.Color,
				Output:   l.V2.Output,
			},
		}
	}

	// Perform smoothing iterations
	for iter := 0; iter < iterations; iter++ {
		// Build vertex adjacency map
		vertexMap := make(map[Vector][]*Vertex)

		// Collect all vertices and their positions
		for _, t := range result.Triangles {
			vertices := []*Vertex{&t.V1, &t.V2, &t.V3}
			for _, v := range vertices {
				pos := v.Position
				vertexMap[pos] = append(vertexMap[pos], v)
			}
		}

		// Average positions for shared vertices
		for _, vertices := range vertexMap {
			if len(vertices) <= 1 {
				continue
			}

			// Calculate average position
			var avg Vector
			for _, v := range vertices {
				avg = avg.Add(v.Position)
			}
			avg = avg.DivScalar(float64(len(vertices)))

			// Update all shared vertices
			for _, v := range vertices {
				v.Position = avg
			}
		}
	}

	// Recalculate normals
	for _, t := range result.Triangles {
		// Calculate face normal
		normal := t.Normal()

		// Update vertex normals
		t.V1.Normal = normal
		t.V2.Normal = normal
		t.V3.Normal = normal
	}

	result.dirty()
	return result
}
