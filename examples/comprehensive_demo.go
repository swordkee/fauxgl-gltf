package main

import (
	"fmt"
	"image"
	"math"

	"github.com/swordkee/fauxgl-gltf"
)

const (
	width  = 1024
	height = 768
	scale  = 1
)

func main() {
	fmt.Println("开始综合演示...")

	// 创建场景
	scene := fauxgl.NewScene("Comprehensive Demo")

	// 创建各种几何体
	geometries := createGeometries()
	materials := createMaterials()

	// 添加网格和材质到场景
	for i, mesh := range geometries {
		scene.Meshes[fmt.Sprintf("mesh_%d", i)] = mesh
		scene.Materials[fmt.Sprintf("material_%d", i)] = materials[i%len(materials)]
	}

	// 创建场景节点并排列几何体
	nodes := createSceneNodes(geometries, materials)
	for _, node := range nodes {
		scene.RootNode.AddChild(node)
	}

	// 创建光源
	lights := createLights()
	scene.Lights = lights

	// 创建相机
	camera := fauxgl.NewOrbitCamera(
		"orbit_camera",
		fauxgl.Vector{0, 0, 0}, // 目标点
		8.0,                    // 距离
		math.Pi/4,              // 45度视野
		float64(width)/float64(height),
		0.1, 100.0,
	)
	camera.Update()
	scene.AddCamera(camera.Camera)
	scene.SetActiveCamera("orbit_camera")

	// 创建渲染上下文
	context := fauxgl.NewContext(width*scale, height*scale)
	context.ClearColor = fauxgl.Color{0.1, 0.1, 0.2, 1} // 深蓝色背景

	// 演示不同功能
	demonstrateFeatures(context, scene, camera, lights)

	fmt.Println("综合演示完成！")
}

// createGeometries 创建各种几何体
func createGeometries() []*fauxgl.Mesh {
	geometries := make([]*fauxgl.Mesh, 0)

	// 立方体
	cube := fauxgl.NewCube()
	geometries = append(geometries, cube)

	// 球体
	sphere := fauxgl.NewSphere(3)
	geometries = append(geometries, sphere)

	// 圆锥体
	cone := fauxgl.NewCone(10, true)
	geometries = append(geometries, cone)

	// 圆柱体
	cylinder := fauxgl.NewCylinder(0.5, 1.0, 16, 1, false)
	geometries = append(geometries, cylinder)

	// 平面
	plane := fauxgl.NewPlane(2, 2)
	geometries = append(geometries, plane)

	// 圆环体
	torus := fauxgl.NewTorus(1.0, 0.3, 20, 12)
	geometries = append(geometries, torus)

	// 胶囊体
	capsule := fauxgl.NewCapsule(0.5, 1.5, 12, 1, 2)
	geometries = append(geometries, capsule)

	return geometries
}

// createMaterials 创建材质
func createMaterials() []*fauxgl.PBRMaterial {
	materials := make([]*fauxgl.PBRMaterial, 6)

	// 红色金属材质
	materials[0] = fauxgl.NewPBRMaterial()
	materials[0].BaseColorFactor = fauxgl.Color{1, 0.2, 0.2, 1}
	materials[0].MetallicFactor = 0.9
	materials[0].RoughnessFactor = 0.1

	// 绿色粗糙材质
	materials[1] = fauxgl.NewPBRMaterial()
	materials[1].BaseColorFactor = fauxgl.Color{0.2, 1, 0.2, 1}
	materials[1].MetallicFactor = 0.1
	materials[1].RoughnessFactor = 0.8

	// 蓝色半金属材质
	materials[2] = fauxgl.NewPBRMaterial()
	materials[2].BaseColorFactor = fauxgl.Color{0.2, 0.2, 1, 1}
	materials[2].MetallicFactor = 0.5
	materials[2].RoughnessFactor = 0.3

	// 黄色光滑材质
	materials[3] = fauxgl.NewPBRMaterial()
	materials[3].BaseColorFactor = fauxgl.Color{1, 1, 0.2, 1}
	materials[3].MetallicFactor = 0.8
	materials[3].RoughnessFactor = 0.2

	// 紫色材质
	materials[4] = fauxgl.NewPBRMaterial()
	materials[4].BaseColorFactor = fauxgl.Color{0.8, 0.2, 1, 1}
	materials[4].MetallicFactor = 0.3
	materials[4].RoughnessFactor = 0.5

	// 白色陶瓷材质
	materials[5] = fauxgl.NewPBRMaterial()
	materials[5].BaseColorFactor = fauxgl.Color{0.9, 0.9, 0.9, 1}
	materials[5].MetallicFactor = 0.0
	materials[5].RoughnessFactor = 0.9

	return materials
}

// createSceneNodes 创建场景节点并排列几何体
func createSceneNodes(geometries []*fauxgl.Mesh, materials []*fauxgl.PBRMaterial) []*fauxgl.SceneNode {
	nodes := make([]*fauxgl.SceneNode, 0)

	// 排列几何体在一个网格中
	positions := []fauxgl.Vector{
		{-3, 0, -3}, // 立方体
		{0, 0, -3},  // 球体
		{3, 0, -3},  // 圆锥体
		{-3, 0, 0},  // 圆柱体
		{0, -1, 0},  // 平面
		{3, 0, 0},   // 圆环体
		{0, 0, 3},   // 胶囊体
	}

	names := []string{"cube", "sphere", "cone", "cylinder", "plane", "torus", "capsule"}

	for i := 0; i < len(positions) && i < len(geometries) && i < len(materials); i++ {
		node := fauxgl.NewSceneNode(names[i])
		node.Mesh = geometries[i]
		node.Material = materials[i%len(materials)]

		// 应用位置变换
		node.SetTransform(fauxgl.Translate(positions[i]))

		// 对某些几何体应用额外的变换
		switch names[i] {
		case "sphere":
			// 缩放球体
			node.SetTransform(fauxgl.Scale(fauxgl.Vector{1.2, 1.2, 1.2}).Translate(positions[i]))
		case "cone":
			// 旋转圆锥体
			node.SetTransform(fauxgl.Rotate(fauxgl.Vector{1, 0, 0}, math.Pi/2).Translate(positions[i]))
		}

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

// demonstrateFeatures 演示不同功能
func demonstrateFeatures(context *fauxgl.Context, scene *fauxgl.Scene, camera *fauxgl.OrbitCamera, lights []fauxgl.Light) {
	// 1. 基本渲染
	fmt.Println("1. 基本渲染...")
	renderBasic(context, scene, "demo_basic.png")

	// 2. 阴影映射
	fmt.Println("2. 阴影映射...")
	renderWithShadows(context, scene, lights[0], "demo_shadows.png")

	// 3. 后期处理
	fmt.Println("3. 后期处理...")
	renderWithPostProcessing(context, scene, "demo_postprocessing.png")

	// 4. 相机控制动画
	fmt.Println("4. 相机控制动画...")
	renderCameraAnimation(context, scene, camera, "demo_camera_animation.gif")

	// 5. 几何体细分
	fmt.Println("5. 几何体细分...")
	demonstrateSubdivision(context, scene, "demo_subdivision.png")

	// 6. 视锥体剔除
	fmt.Println("6. 视锥体剔除...")
	demonstrateCulling(context, scene, "demo_culling.png")
}

// renderBasic 基本渲染
func renderBasic(context *fauxgl.Context, scene *fauxgl.Scene, filename string) {
	// 清除缓冲区
	context.ClearColorBuffer()
	context.ClearDepthBuffer()

	// 创建场景渲染器
	renderer := fauxgl.NewSceneRenderer(context)
	renderer.RenderScene(scene)

	// 保存图像
	fauxgl.SavePNG(filename, context.Image())
	fmt.Printf("保存基本渲染图像到 %s\n", filename)
}

// renderWithShadows 渲染带阴影的场景
func renderWithShadows(context *fauxgl.Context, scene *fauxgl.Scene, light fauxgl.Light, filename string) {
	// 清除缓冲区
	context.ClearColorBuffer()
	context.ClearDepthBuffer()

	// 创建阴影渲染器
	shadowRenderer := fauxgl.NewShadowMapRenderer(context, 1024, light, fauxgl.PCFShadow)
	shadowMap := shadowRenderer.GenerateShadowMap(scene)

	// 获取光照矩阵
	lightMatrix := shadowRenderer.GetLightMatrix()

	// 获取相机矩阵
	cameraMatrix := scene.ActiveCamera.GetCameraMatrix()

	// 获取所有可渲染节点
	renderables := scene.RootNode.GetRenderableNodes()

	// 渲染每个节点（带阴影）
	for _, node := range renderables {
		if node.Mesh == nil || node.Material == nil {
			continue
		}

		// 计算最终变换矩阵
		modelMatrix := node.WorldTransform
		finalMatrix := cameraMatrix.Mul(modelMatrix)

		// 创建阴影接收着色器
		shadowShader := fauxgl.NewSoftShadowReceiverShader(
			finalMatrix,
			lightMatrix,
			light.Direction,
			scene.ActiveCamera.Position,
			shadowMap,
			fauxgl.PCFShadow,
		)
		shadowShader.ObjectColor = node.Material.BaseColorFactor

		// 设置着色器并渲染
		context.Shader = shadowShader
		context.DrawMesh(node.Mesh)
	}

	// 保存图像
	fauxgl.SavePNG(filename, context.Image())
	fmt.Printf("保存带阴影图像到 %s\n", filename)
}

// renderWithPostProcessing 渲染带后期处理的场景
func renderWithPostProcessing(context *fauxgl.Context, scene *fauxgl.Scene, filename string) {
	// 清除缓冲区
	context.ClearColorBuffer()
	context.ClearDepthBuffer()

	// 创建场景渲染器
	renderer := fauxgl.NewSceneRenderer(context)
	renderer.RenderScene(scene)

	// 创建后期处理管线
	pipeline := fauxgl.NewPostProcessingPipeline()

	// 添加辉光效果
	bloomEffect := fauxgl.NewBloomEffect(0.7, 3, 0.5)
	pipeline.AddEffect(bloomEffect)

	// 添加色调映射
	toneMapEffect := fauxgl.NewToneMappingEffect(1.0, 2.2)
	pipeline.AddEffect(toneMapEffect)

	// 添加FXAA抗锯齿
	fxaaEffect := fauxgl.NewFXAAEffect()
	pipeline.AddEffect(fxaaEffect)

	// 添加色差效果
	chromaEffect := fauxgl.NewChromaticAberrationEffect(
		fauxgl.Vector{-2, 0, 0},
		fauxgl.Vector{0, 0, 0},
		fauxgl.Vector{2, 0, 0},
	)
	pipeline.AddEffect(chromaEffect)

	// 添加暗角效果
	vignetteEffect := fauxgl.NewVignetteEffect(0.3)
	pipeline.AddEffect(vignetteEffect)

	// 应用后期处理
	img := context.Image()
	nrgba, ok := img.(*image.NRGBA)
	if !ok {
		// 如果不是*NRGBA，需要转换
		bounds := img.Bounds()
		nrgba = image.NewNRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				nrgba.Set(x, y, img.At(x, y))
			}
		}
	}
	result := pipeline.Process(nrgba)

	// 保存图像
	fauxgl.SavePNG(filename, result)
	fmt.Printf("保存带后期处理图像到 %s\n", filename)
}

// renderCameraAnimation 渲染相机控制动画
func renderCameraAnimation(context *fauxgl.Context, scene *fauxgl.Scene, camera *fauxgl.OrbitCamera, filename string) {
	// 渲染多个动画帧
	for frame := 0; frame < 20; frame++ {
		// 旋转相机
		camera.Rotate(0.2, 0.05)

		// 清除缓冲区
		context.ClearColorBuffer()
		context.ClearDepthBuffer()

		// 创建场景渲染器
		renderer := fauxgl.NewSceneRenderer(context)
		renderer.RenderScene(scene)

		// 保存图像帧
		frameFilename := fmt.Sprintf("demo_camera_frame_%03d.png", frame)
		fauxgl.SavePNG(frameFilename, context.Image())
		fmt.Printf("保存相机动画帧 %s\n", frameFilename)
	}

	fmt.Printf("保存相机控制动画到 %s\n", filename)
}

// demonstrateSubdivision 演示几何体细分
func demonstrateSubdivision(context *fauxgl.Context, scene *fauxgl.Scene, filename string) {
	// 找到球体节点并进行细分
	sphereNode := scene.RootNode.FindChild("sphere")
	if sphereNode != nil && sphereNode.Mesh != nil {
		// 创建细分后的球体
		subdividedSphere := sphereNode.Mesh.Subdivide()

		// 替换原球体
		originalMesh := sphereNode.Mesh
		sphereNode.Mesh = subdividedSphere

		// 渲染
		context.ClearColorBuffer()
		context.ClearDepthBuffer()
		renderer := fauxgl.NewSceneRenderer(context)
		renderer.RenderScene(scene)
		fauxgl.SavePNG(filename, context.Image())
		fmt.Printf("保存细分演示图像到 %s\n", filename)

		// 恢复原球体
		sphereNode.Mesh = originalMesh
	} else {
		fmt.Println("未找到球体节点")
	}
}

// demonstrateCulling 演示视锥体剔除
func demonstrateCulling(context *fauxgl.Context, scene *fauxgl.Scene, filename string) {
	// 清除缓冲区
	context.ClearColorBuffer()
	context.ClearDepthBuffer()

	// 创建带剔除的场景渲染器
	cullingRenderer := fauxgl.NewCullingSceneRenderer(context)
	cullingRenderer.RenderScene(scene)

	// 保存图像
	fauxgl.SavePNG(filename, context.Image())
	fmt.Printf("保存视锥体剔除演示图像到 %s\n", filename)
}
