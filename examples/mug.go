package main

import (
	"fmt"

	"github.com/swordkee/fauxgl-gltf"
)

const (
	scale  = 1    // optional supersampling
	width  = 2000 // output width in pixels
	height = 2000 // output height in pixels
	fovy   = 20   // vertical field of view in degrees
	near   = 1    // near clipping plane
	far    = 10   // far clipping plane
)

var (
	// 原始参数，确保与mug.go完全一致
	eye    = fauxgl.V(2.8, 2.8, 4) // 拉远相机距离，显示完整杯子
	center = fauxgl.V(0, 0.6, 0)   // 略微向上调整焦点
	up     = fauxgl.V(0, 1, 0)     // 标准上方向向量
)

func main() {
	fmt.Println("Loading mug.gltf using traditional method...")

	// 使用GLTF场景加载器（更新的方法）
	scene, err := fauxgl.LoadGLTFScene("mug.gltf")
	if err != nil {
		panic(err)
	}

	// 从场景中合并所有网格为单个网格（保持向后兼容）
	var allTriangles []*fauxgl.Triangle
	scene.RootNode.VisitNodes(func(node *fauxgl.SceneNode) {
		if node.Mesh != nil {
			allTriangles = append(allTriangles, node.Mesh.Triangles...)
		}
	})
	mesh := fauxgl.NewTriangleMesh(allTriangles)

	// 加载纹理
	texture, err := fauxgl.LoadTexture("texture.jpg")
	if err != nil {
		fmt.Printf("Warning: Could not load texture: %v\n", err)
	}

	fmt.Printf("Mesh loaded: %d triangles\n", len(mesh.Triangles))

	// 获取原始边界
	bounds := mesh.BoundingBox()
	fmt.Printf("Original bounds: min=%v, max=%v\n", bounds.Min, bounds.Max)
	fmt.Printf("Original center: %v\n", bounds.Center())

	// 应用预处理（与原始mug.go相同）
	mesh.BiUnitCube()
	mesh.SmoothNormalsThreshold(fauxgl.Radians(30))

	// 检查BiUnitCube后的边界
	bounds = mesh.BoundingBox()
	fmt.Printf("After BiUnitCube: min=%v, max=%v\n", bounds.Min, bounds.Max)
	fmt.Printf("After BiUnitCube center: %v\n", bounds.Center())

	// 应用变换（与原始mug.go相同）
	//positionMatrix := Translate(V(0.5, 0.3, 0))
	//mesh.Transform(positionMatrix)

	// 检查变换后的边界
	bounds = mesh.BoundingBox()
	fmt.Printf("After transform: min=%v, max=%v\n", bounds.Min, bounds.Max)
	fmt.Printf("After transform center: %v\n", bounds.Center())

	// 创建渲染上下文
	context := fauxgl.NewContext(width*scale, height*scale)
	context.ClearColor = fauxgl.White
	context.ClearColorBuffer()

	// 设置相机矩阵（与原始mug.go相同）
	aspect := float64(width) / float64(height)
	matrix := fauxgl.LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	light := fauxgl.V(0, 0, 1).Normalize()

	// 创建着色器（与原始mug.go相同）
	shader := fauxgl.NewPhongShader(matrix, light, eye)
	if texture != nil {
		shader.Texture = texture
	}
	shader.ObjectColor = fauxgl.HexColor("FFFF9D").Alpha(0.65)
	context.Shader = shader

	fmt.Println("Rendering mesh...")
	context.DrawMesh(mesh)

	// 保存图像
	err = fauxgl.SavePNG("mug_fixed.png", context.Image())
	if err != nil {
		fmt.Printf("Failed to save image: %v\n", err)
	} else {
		fmt.Println("Fixed mug rendering saved as mug_fixed.png")
	}
}
