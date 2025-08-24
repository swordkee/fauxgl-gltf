package main

import (
	"fmt"
	"image"
	"math"

	"github.com/swordkee/fauxgl-gltf"
)

const (
	width  = 800
	height = 600
	scale  = 1
)

func main() {
	// 创建场景
	scene := fauxgl.NewScene("Shadow and Postprocessing Demo")

	// 创建几何体
	cube := fauxgl.NewCube()
	sphere := fauxgl.NewSphere(3)
	plane := fauxgl.NewCube()

	// 调整平面大小作为地面
	plane.Transform(fauxgl.Scale(fauxgl.Vector{10, 0.1, 10}).Translate(fauxgl.Vector{0, -2, 0}))

	// 创建材质
	cubeMaterial := fauxgl.NewPBRMaterial()
	cubeMaterial.BaseColorFactor = fauxgl.Color{1, 0, 0, 1} // 红色立方体

	sphereMaterial := fauxgl.NewPBRMaterial()
	sphereMaterial.BaseColorFactor = fauxgl.Color{0, 1, 0, 1} // 绿色球体

	planeMaterial := fauxgl.NewPBRMaterial()
	planeMaterial.BaseColorFactor = fauxgl.Color{0.5, 0.5, 0.5, 1} // 灰色地面

	// 添加网格和材质到场景
	scene.Meshes["cube"] = cube
	scene.Meshes["sphere"] = sphere
	scene.Meshes["plane"] = plane

	scene.Materials["cube_material"] = cubeMaterial
	scene.Materials["sphere_material"] = sphereMaterial
	scene.Materials["plane_material"] = planeMaterial

	// 创建场景节点
	cubeNode := fauxgl.NewSceneNode("cube")
	cubeNode.Mesh = cube
	cubeNode.Material = cubeMaterial
	cubeNode.SetTransform(fauxgl.Translate(fauxgl.Vector{-2, 0, 0}))

	sphereNode := fauxgl.NewSceneNode("sphere")
	sphereNode.Mesh = sphere
	sphereNode.Material = sphereMaterial
	sphereNode.SetTransform(fauxgl.Translate(fauxgl.Vector{2, 0, 0}))

	planeNode := fauxgl.NewSceneNode("plane")
	planeNode.Mesh = plane
	planeNode.Material = planeMaterial

	// 添加节点到场景
	scene.RootNode.AddChild(cubeNode)
	scene.RootNode.AddChild(sphereNode)
	scene.RootNode.AddChild(planeNode)

	// 创建光源
	light := fauxgl.Light{
		Type:      fauxgl.DirectionalLight,
		Direction: fauxgl.Vector{-1, -1, -1}.Normalize(),
		Color:     fauxgl.Color{1, 1, 1, 1},
		Intensity: 1.0,
	}
	scene.Lights = append(scene.Lights, light)

	// 创建相机
	camera := fauxgl.NewPerspectiveCamera(
		"main_camera",
		fauxgl.Vector{0, 2, 5}, // 相机位置
		fauxgl.Vector{0, 0, 0}, // 目标点
		fauxgl.Vector{0, 1, 0}, // 上方向
		math.Pi/4,              // 45度视野
		float64(width)/float64(height),
		0.1, 100.0,
	)
	scene.AddCamera(camera)
	scene.SetActiveCamera("main_camera")

	// 创建渲染上下文
	context := fauxgl.NewContext(width*scale, height*scale)
	context.ClearColor = fauxgl.Color{0.1, 0.1, 0.1, 1} // 深灰色背景

	// 渲染场景（无阴影）
	fmt.Println("渲染无阴影场景...")
	renderScene(context, scene, "demo_no_shadows.png")

	// 渲染带阴影的场景
	fmt.Println("渲染带阴影场景...")
	renderSceneWithShadows(context, scene, light, "demo_with_shadows.png")

	// 渲染带后期处理的场景
	fmt.Println("渲染带后期处理场景...")
	renderSceneWithPostProcessing(context, scene, "demo_postprocessing.png")

	fmt.Println("所有渲染完成！")
}

// renderScene 渲染基本场景
func renderScene(context *fauxgl.Context, scene *fauxgl.Scene, filename string) {
	// 清除缓冲区
	context.ClearColorBuffer()
	context.ClearDepthBuffer()

	// 创建场景渲染器
	renderer := fauxgl.NewSceneRenderer(context)
	renderer.RenderScene(scene)

	// 保存图像
	fauxgl.SavePNG(filename, context.Image())
	fmt.Printf("保存图像到 %s\n", filename)
}

// renderSceneWithShadows 渲染带阴影的场景
func renderSceneWithShadows(context *fauxgl.Context, scene *fauxgl.Scene, light fauxgl.Light, filename string) {
	// 清除缓冲区
	context.ClearColorBuffer()
	context.ClearDepthBuffer()

	// 创建阴影渲染器
	shadowRenderer := fauxgl.NewShadowRenderer(context, 1024, light)
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
		shadowShader := fauxgl.NewShadowReceiverShader(
			finalMatrix,
			lightMatrix,
			light.Direction,
			scene.ActiveCamera.Position,
			shadowMap,
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

// renderSceneWithPostProcessing 渲染带后期处理的场景
func renderSceneWithPostProcessing(context *fauxgl.Context, scene *fauxgl.Scene, filename string) {
	// 清除缓冲区
	context.ClearColorBuffer()
	context.ClearDepthBuffer()

	// 创建场景渲染器
	renderer := fauxgl.NewSceneRenderer(context)
	renderer.RenderScene(scene)

	// 创建后期处理管线
	pipeline := fauxgl.NewPostProcessingPipeline()

	// 添加模糊效果
	blurEffect := fauxgl.NewBlurEffect(2)
	pipeline.AddEffect(blurEffect)

	// 添加辉光效果
	bloomEffect := fauxgl.NewBloomEffect(0.7, 3, 0.5)
	pipeline.AddEffect(bloomEffect)

	// 添加色调映射
	toneMapEffect := fauxgl.NewToneMappingEffect(1.0, 2.2)
	pipeline.AddEffect(toneMapEffect)

	// 添加FXAA抗锯齿
	fxaaEffect := fauxgl.NewFXAAEffect()
	pipeline.AddEffect(fxaaEffect)

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
