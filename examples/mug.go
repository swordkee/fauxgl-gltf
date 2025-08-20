package main

import . "github.com/swordkee/fauxgl"

const (
	width  = 2000
	height = 1000
	fovy   = 20
	near   = 1
	far    = 50
)

var (
	eye    = V(-1, -2, 2)
	center = V(-0, 0, 0)
	up     = V(0, 0, 1)
)

func main() {
	mesh, err := LoadGLTF("examples/mug.gltf")
	if err != nil {
		panic(err)
	}
	mesh.BiUnitCube()
	mesh.SmoothNormalsThreshold(Radians(30))

	context := NewContext(width, height)
	context.ClearColor = White
	context.ClearColorBuffer()

	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	light := V(0, 0, 1).Normalize()
	color := Color{0.5, 1, 0.65, 1}

	shader := NewPhongShader(matrix, light, eye)
	shader.ObjectColor = color
	context.Shader = shader
	context.DrawMesh(mesh)

	SavePNG("out.png", context.Image())
}
