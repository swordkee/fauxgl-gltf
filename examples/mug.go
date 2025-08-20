package main

import . "github.com/swordkee/fauxgl"

const (
	scale  = 1    // optional supersampling
	width  = 2000 // output width in pixels
	height = 2000 // output height in pixels
	fovy   = 20   // vertical field of view in degrees
	near   = 1    // near clipping plane
	far    = 10   // far clipping plane
)

var (
	eye    = V(-3, -3, 3)    // 相机位置 (修改这里)
	center = V(0.07, 0, 0)   // 目标焦点位置 (修改这里)
	up     = V(0, -0.004, 0) // 上方向向量
)

func main() {
	mesh, err := LoadGLTF("examples/mug.gltf")
	// load the texture
	texture, err := LoadTexture("examples/texture.png")
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	mesh.BiUnitCube()
	mesh.SmoothNormalsThreshold(Radians(30))

	// 修改模型位置 - 示例：向右移动0.5单位，向上移动0.3单位
	positionMatrix := Translate(V(0.5, 0.3, 0))
	mesh.Transform(positionMatrix)

	context := NewContext(width*scale, height*scale)
	context.ClearColor = White
	context.ClearColorBuffer()

	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	light := V(0, 0, 1).Normalize()
	//color := Color{0.5, 1, 0.65, 1}

	shader := NewPhongShader(matrix, light, eye)
	shader.Texture = texture
	shader.ObjectColor = HexColor("FFFF9D").Alpha(0.65)
	context.Shader = shader
	context.DrawMesh(mesh)

	SavePNG("out.png", context.Image())
}
