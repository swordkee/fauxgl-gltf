package main

import (
	"fmt"
	"math"

	"github.com/swordkee/fauxgl-gltf"
)

const (
	width  = 1024
	height = 768
	scale  = 1
)

func main() {
	// 创建场景
	scene := fauxgl.NewScene("Geometry and Camera Demo")

	// 创建各种几何体
	cube := fauxgl.NewCube()
	sphere := fauxgl.NewSphere(4)
	cone := fauxgl.NewCone(10, true)
	cylinder := createCylinder(16, true)
	torus := createTorus(20, 12, 1.0, 0.3)

	// 创建材质
	materials := createMaterials()

	// 添加网格和材质到场景
	scene.Meshes["cube"] = cube
	scene.Meshes["sphere"] = sphere
	scene.Meshes["cone"] = cone
	scene.Meshes["cylinder"] = cylinder
	scene.Meshes["torus"] = torus

	scene.Materials["red_material"] = materials[0]
	scene.Materials["green_material"] = materials[1]
	scene.Materials["blue_material"] = materials[2]
	scene.Materials["yellow_material"] = materials[3]
	scene.Materials["purple_material"] = materials[4]

	// 创建场景节点并排列几何体
	nodes := createSceneNodes(cube, sphere, cone, cylinder, torus, materials)
	for _, node := range nodes {
		scene.RootNode.AddChild(node)
	}

	// 创建多个光源
	lights := createLights()
	scene.Lights = lights

	// 创建多个相机
	cameras := createCameras()
	for _, camera := range cameras {
		scene.AddCamera(camera)
	}
	scene.SetActiveCamera("perspective_camera")

	// 创建渲染上下文
	context := fauxgl.NewContext(width*scale, height*scale)
	context.ClearColor = fauxgl.Color{0.1, 0.1, 0.2, 1} // 深蓝色背景

	// 渲染不同视角的场景
	fmt.Println("渲染不同视角的场景...")
	renderFromDifferentCameras(context, scene)

	fmt.Println("所有渲染完成！")
}

// createCylinder 创建圆柱体
func createCylinder(segments int, capped bool) *fauxgl.Mesh {
	var triangles []*fauxgl.Triangle

	// 创建顶面和底面顶点
	topCenter := fauxgl.Vector{0, 0.5, 0}
	bottomCenter := fauxgl.Vector{0, -0.5, 0}

	topVertices := make([]fauxgl.Vector, segments)
	bottomVertices := make([]fauxgl.Vector, segments)

	for i := 0; i < segments; i++ {
		angle := float64(i) / float64(segments) * 2 * math.Pi
		x := math.Cos(angle)
		z := math.Sin(angle)
		topVertices[i] = fauxgl.Vector{x, 0.5, z}
		bottomVertices[i] = fauxgl.Vector{x, -0.5, z}
	}

	// 创建侧面
	for i := 0; i < segments; i++ {
		next := (i + 1) % segments

		// 侧面三角形
		t1 := fauxgl.NewTriangleForPoints(topVertices[i], bottomVertices[i], topVertices[next])
		t2 := fauxgl.NewTriangleForPoints(topVertices[next], bottomVertices[i], bottomVertices[next])

		triangles = append(triangles, t1, t2)
	}

	// 创建顶面和底面
	if capped {
		for i := 0; i < segments-2; i++ {
			// 顶面三角形
			t1 := fauxgl.NewTriangleForPoints(topCenter, topVertices[i], topVertices[i+1])
			// 底面三角形
			t2 := fauxgl.NewTriangleForPoints(bottomCenter, bottomVertices[i+1], bottomVertices[i])

			triangles = append(triangles, t1, t2)
		}
		// 最后一个面
		t1 := fauxgl.NewTriangleForPoints(topCenter, topVertices[segments-1], topVertices[0])
		t2 := fauxgl.NewTriangleForPoints(bottomCenter, bottomVertices[0], bottomVertices[segments-1])
		triangles = append(triangles, t1, t2)
	}

	return fauxgl.NewTriangleMesh(triangles)
}

// createTorus 创建圆环体
func createTorus(majorSegments, minorSegments int, majorRadius, minorRadius float64) *fauxgl.Mesh {
	var triangles []*fauxgl.Triangle

	for i := 0; i < majorSegments; i++ {
		for j := 0; j < minorSegments; j++ {
			// 计算四个顶点
			nextI := (i + 1) % majorSegments
			nextJ := (j + 1) % minorSegments

			v0 := torusVertex(i, j, majorSegments, minorSegments, majorRadius, minorRadius)
			v1 := torusVertex(nextI, j, majorSegments, minorSegments, majorRadius, minorRadius)
			v2 := torusVertex(i, nextJ, majorSegments, minorSegments, majorRadius, minorRadius)
			v3 := torusVertex(nextI, nextJ, majorSegments, minorSegments, majorRadius, minorRadius)

			// 创建两个三角形
			t1 := fauxgl.NewTriangleForPoints(v0, v1, v2)
			t2 := fauxgl.NewTriangleForPoints(v1, v3, v2)

			triangles = append(triangles, t1, t2)
		}
	}

	return fauxgl.NewTriangleMesh(triangles)
}

// torusVertex 计算圆环体顶点
func torusVertex(i, j, majorSegments, minorSegments int, majorRadius, minorRadius float64) fauxgl.Vector {
	majorAngle := float64(i) / float64(majorSegments) * 2 * math.Pi
	minorAngle := float64(j) / float64(minorSegments) * 2 * math.Pi

	centerX := math.Cos(majorAngle) * majorRadius
	centerZ := math.Sin(majorAngle) * majorRadius

	x := centerX + math.Cos(majorAngle)*math.Cos(minorAngle)*minorRadius
	y := math.Sin(minorAngle) * minorRadius
	z := centerZ + math.Sin(majorAngle)*math.Cos(minorAngle)*minorRadius

	return fauxgl.Vector{x, y, z}
}

// createMaterials 创建材质
func createMaterials() []*fauxgl.PBRMaterial {
	materials := make([]*fauxgl.PBRMaterial, 5)

	// 红色材质
	materials[0] = fauxgl.NewPBRMaterial()
	materials[0].BaseColorFactor = fauxgl.Color{1, 0, 0, 1}
	materials[0].MetallicFactor = 0.2
	materials[0].RoughnessFactor = 0.3

	// 绿色材质
	materials[1] = fauxgl.NewPBRMaterial()
	materials[1].BaseColorFactor = fauxgl.Color{0, 1, 0, 1}
	materials[1].MetallicFactor = 0.3
	materials[1].RoughnessFactor = 0.4

	// 蓝色材质
	materials[2] = fauxgl.NewPBRMaterial()
	materials[2].BaseColorFactor = fauxgl.Color{0, 0, 1, 1}
	materials[2].MetallicFactor = 0.4
	materials[2].RoughnessFactor = 0.5

	// 黄色材质
	materials[3] = fauxgl.NewPBRMaterial()
	materials[3].BaseColorFactor = fauxgl.Color{1, 1, 0, 1}
	materials[3].MetallicFactor = 0.1
	materials[3].RoughnessFactor = 0.2

	// 紫色材质
	materials[4] = fauxgl.NewPBRMaterial()
	materials[4].BaseColorFactor = fauxgl.Color{1, 0, 1, 1}
	materials[4].MetallicFactor = 0.5
	materials[4].RoughnessFactor = 0.6

	return materials
}

// createSceneNodes 创建场景节点并排列几何体
func createSceneNodes(cube, sphere, cone, cylinder, torus *fauxgl.Mesh, materials []*fauxgl.PBRMaterial) []*fauxgl.SceneNode {
	nodes := make([]*fauxgl.SceneNode, 0)

	// 排列几何体在一个网格中
	positions := []fauxgl.Vector{
		{-2, 0, -2}, // 立方体
		{2, 0, -2},  // 球体
		{-2, 0, 2},  // 圆锥体
		{2, 0, 2},   // 圆柱体
		{0, 0, 0},   // 圆环体
	}

	meshes := []*fauxgl.Mesh{cube, sphere, cone, cylinder, torus}
	names := []string{"cube", "sphere", "cone", "cylinder", "torus"}

	for i := 0; i < len(positions) && i < len(meshes) && i < len(materials); i++ {
		node := fauxgl.NewSceneNode(names[i])
		node.Mesh = meshes[i]
		node.Material = materials[i]
		node.SetTransform(fauxgl.Translate(positions[i]))
		nodes = append(nodes, node)
	}

	return nodes
}

// createLights 创建光源
func createLights() []fauxgl.Light {
	lights := make([]fauxgl.Light, 3)

	// 主光源
	lights[0] = fauxgl.Light{
		Type:      fauxgl.DirectionalLight,
		Direction: fauxgl.Vector{-1, -1, -1}.Normalize(),
		Color:     fauxgl.Color{1, 1, 1, 1},
		Intensity: 1.0,
	}

	// 补光
	lights[1] = fauxgl.Light{
		Type:      fauxgl.DirectionalLight,
		Direction: fauxgl.Vector{1, -0.5, -0.5}.Normalize(),
		Color:     fauxgl.Color{0.7, 0.7, 0.8, 1},
		Intensity: 0.5,
	}

	// 环境光
	lights[2] = fauxgl.Light{
		Type:      fauxgl.AmbientLight,
		Color:     fauxgl.Color{0.3, 0.3, 0.4, 1},
		Intensity: 0.3,
	}

	return lights
}

// createCameras 创建多个相机
func createCameras() []*fauxgl.Camera {
	cameras := make([]*fauxgl.Camera, 0)

	// 透视相机
	perspectiveCamera := fauxgl.NewPerspectiveCamera(
		"perspective_camera",
		fauxgl.Vector{0, 2, 5}, // 相机位置
		fauxgl.Vector{0, 0, 0}, // 目标点
		fauxgl.Vector{0, 1, 0}, // 上方向
		math.Pi/4,              // 45度视野
		float64(1024)/float64(768),
		0.1, 100.0,
	)
	cameras = append(cameras, perspectiveCamera)

	// 正交相机
	orthographicCamera := fauxgl.NewOrthographicCamera(
		"orthographic_camera",
		fauxgl.Vector{0, 2, 5}, // 相机位置
		fauxgl.Vector{0, 0, 0}, // 目标点
		fauxgl.Vector{0, 1, 0}, // 上方向
		5.0,                    // 正交大小
		float64(1024)/float64(768),
		0.1, 100.0,
	)
	cameras = append(cameras, orthographicCamera)

	// 侧面相机
	sideCamera := fauxgl.NewPerspectiveCamera(
		"side_camera",
		fauxgl.Vector{5, 0, 0}, // 相机位置
		fauxgl.Vector{0, 0, 0}, // 目标点
		fauxgl.Vector{0, 1, 0}, // 上方向
		math.Pi/4,              // 45度视野
		float64(1024)/float64(768),
		0.1, 100.0,
	)
	cameras = append(cameras, sideCamera)

	return cameras
}

// renderFromDifferentCameras 从不同相机视角渲染场景
func renderFromDifferentCameras(context *fauxgl.Context, scene *fauxgl.Scene) {
	cameraNames := []string{"perspective_camera", "orthographic_camera", "side_camera"}

	for _, cameraName := range cameraNames {
		// 设置活动相机
		scene.SetActiveCamera(cameraName)

		// 清除缓冲区
		context.ClearColorBuffer()
		context.ClearDepthBuffer()

		// 创建场景渲染器
		renderer := fauxgl.NewSceneRenderer(context)
		renderer.RenderScene(scene)

		// 保存图像
		filename := fmt.Sprintf("geometry_demo_%s.png", cameraName)
		fauxgl.SavePNG(filename, context.Image())
		fmt.Printf("保存图像到 %s\n", filename)
	}
}
