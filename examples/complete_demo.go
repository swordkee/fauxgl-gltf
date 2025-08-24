package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"runtime"

	"github.com/swordkee/fauxgl-gltf"
)

func main() {
	fmt.Println("=== 3D马克杯完整演示 ===")

	// 设置多核并行渲染
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("使用 %d 个CPU核心进行并行渲染\n", runtime.NumCPU())

	// 获取工作目录
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	// 加载GLTF模型
	modelPath := filepath.Join(dir, "./gltf/mug.gltf")
	scene, err := fauxgl.LoadGLTFScene(modelPath)
	if err != nil {
		fmt.Printf("错误: 无法加载模型 %s: %v\n", modelPath, err)
		return
	}
	fmt.Printf("✓ 已加载模型: %s\n", modelPath)

	// 分析模型结构
	analyzeModel(scene)

	// 获取第一个网格进行UV处理
	var testMesh *fauxgl.Mesh
	for _, mesh := range scene.Meshes {
		testMesh = mesh
		break
	}

	if testMesh != nil {
		fmt.Printf("\n=== UV编辑器功能演示 ===")
		// 应用UV松弛算法
		settings := fauxgl.NewUVRelaxationSettings()
		settings.Iterations = 15
		settings.StepSize = 0.3

		err = fauxgl.ApplyUVRelaxation(testMesh, settings)
		if err != nil {
			fmt.Printf("  ✗ UV松弛失败: %v\n", err)
		} else {
			fmt.Println("  ✓ UV平面展开算法应用完成")
		}

		// 生成UV可视化图
		uvImage := generateUVVisualization(testMesh, 1024, 1024)
		uvImagePath := filepath.Join(dir, "complete_uv_visualization.png")
		err = saveImage(uvImage, uvImagePath)
		if err != nil {
			fmt.Printf("  ✗ 无法保存UV可视化图: %v\n", err)
		} else {
			fmt.Printf("  ✓ UV可视化图已保存到: %s\n", uvImagePath)
		}
	}

	// 应用材质
	applyMaterials(scene)

	// 设置多光源系统
	setupMultiLightScene(scene)

	// 渲染场景
	renderScene(scene, dir)

	fmt.Println("\n=== 完整演示完成 ===")
}

// analyzeModel 分析模型结构
func analyzeModel(scene *fauxgl.Scene) {
	fmt.Println("\n=== 模型结构分析 ===")

	// 材质分析
	fmt.Printf("材质数量: %d\n", len(scene.Materials))
	for name, material := range scene.Materials {
		fmt.Printf("  %s: ", name)
		if material.BaseColorTexture != nil {
			fmt.Printf("纹理材质 - 基础颜色: [%.2f, %.2f, %.2f]",
				material.BaseColorFactor.R, material.BaseColorFactor.G, material.BaseColorFactor.B)
		} else {
			fmt.Printf("纯色材质 - 颜色: [%.2f, %.2f, %.2f]",
				material.BaseColorFactor.R, material.BaseColorFactor.G, material.BaseColorFactor.B)
		}
		fmt.Printf(", 金属度: %.2f, 粗糙度: %.2f\n",
			material.MetallicFactor, material.RoughnessFactor)
	}

	// 网格分析
	fmt.Printf("\n网格数量: %d\n", len(scene.Meshes))
	totalTriangles := 0
	for name, mesh := range scene.Meshes {
		fmt.Printf("  %s: %d三角形\n", name, len(mesh.Triangles))
		totalTriangles += len(mesh.Triangles)
	}
	fmt.Printf("  总三角形数: %d\n", totalTriangles)
}

// applyMaterials 应用材质
func applyMaterials(scene *fauxgl.Scene) {
	fmt.Println("\n=== 应用材质 ===")

	// 遍历所有材质
	for name, material := range scene.Materials {
		// 设置通用材质属性
		material.MetallicFactor = 0.0    // 非金属
		material.RoughnessFactor = 0.3   // 光滑表面
		material.BaseColorFactor.A = 1.0 // 完全不透明

		fmt.Printf("  ✓ 已处理材质: %s\n", name)
	}
}

// setupMultiLightScene 设置多光源场景
func setupMultiLightScene(scene *fauxgl.Scene) {
	fmt.Println("\n=== 设置多光源系统 ===")

	// 清除现有光源
	scene.ClearLights()

	// 设置多光源系统
	mainLight := fauxgl.Light{
		Type:      fauxgl.DirectionalLight,
		Direction: fauxgl.V(-0.5, -0.5, -1.0).Normalize(), // 主光源
		Color:     fauxgl.Color{1.0, 1.0, 1.0, 1.0},
		Intensity: 1.0,
	}

	fillLight := fauxgl.Light{ // 填充光
		Type:      fauxgl.DirectionalLight,
		Direction: fauxgl.V(0.7, -0.3, -0.3).Normalize(),
		Color:     fauxgl.Color{0.6, 0.6, 0.6, 1.0},
		Intensity: 0.5,
	}

	backLight := fauxgl.Light{ // 背光
		Type:      fauxgl.DirectionalLight,
		Direction: fauxgl.V(0.3, 0.2, 0.8).Normalize(),
		Color:     fauxgl.Color{0.4, 0.4, 0.4, 1.0},
		Intensity: 0.3,
	}

	ambientLight := fauxgl.Light{ // 环境光
		Type:      fauxgl.AmbientLight,
		Color:     fauxgl.Color{0.3, 0.3, 0.3, 1.0},
		Intensity: 1.0,
	}

	// 添加光源到场景
	scene.AddLight(mainLight)
	scene.AddLight(fillLight)
	scene.AddLight(backLight)
	scene.AddLight(ambientLight)

	fmt.Printf("  已设置 %d 个光源\n", len(scene.Lights))
}

// renderScene 渲染场景
func renderScene(scene *fauxgl.Scene, dir string) {
	fmt.Println("\n=== 渲染场景 ===")

	// 创建渲染上下文
	context := fauxgl.NewContext(2048, 2048)
	context.ClearColor = fauxgl.Color{0.05, 0.05, 0.05, 1.0}
	context.ClearColorBuffer()
	context.ClearDepthBuffer()

	// 设置相机
	aspect := 1.0
	matrix := fauxgl.LookAt(
		fauxgl.V(2.5, 1.5, 2.5),
		fauxgl.V(0, 0.6, 0),
		fauxgl.V(0, 1, 0),
	).Perspective(30, aspect, 1, 20)

	// 获取所有可渲染节点
	renderableNodes := scene.RootNode.GetRenderableNodes()
	fmt.Printf("开始渲染 %d 个节点...\n", len(renderableNodes))

	// 逐个渲染每个节点
	for i, node := range renderableNodes {
		if node.Mesh == nil || node.Material == nil {
			continue
		}

		fmt.Printf("渲染节点 %d: %s\n", i+1, node.Name)

		// 创建PBR着色器，支持多光源
		pbrShader := fauxgl.NewPBRShader(matrix, node.Material, scene.Lights, fauxgl.V(2.5, 1.5, 2.5))

		// 应用着色器
		context.Shader = pbrShader
		context.DrawMesh(node.Mesh)

		fmt.Printf("  ✓ 完成渲染\n")
	}

	// 保存渲染结果
	outputPath := filepath.Join(dir, "complete_demo_render.png")
	err := fauxgl.SavePNG(outputPath, context.Image())
	if err != nil {
		fmt.Printf("  ✗ 渲染失败: %v\n", err)
		return
	}

	fmt.Printf("  ✓ 渲染完成，结果保存到: %s\n", outputPath)
}

// generateUVVisualization 生成UV可视化图
func generateUVVisualization(mesh *fauxgl.Mesh, width, height int) image.Image {
	// 创建新的RGBA图像
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 填充背景色
	background := color.RGBA{30, 30, 30, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{background}, image.Point{}, draw.Src)

	// 绘制UV三角形
	for _, triangle := range mesh.Triangles {
		// 获取UV坐标
		uv1 := fauxgl.Vector2{triangle.V1.Texture.X, triangle.V1.Texture.Y}
		uv2 := fauxgl.Vector2{triangle.V2.Texture.X, triangle.V2.Texture.Y}
		uv3 := fauxgl.Vector2{triangle.V3.Texture.X, triangle.V3.Texture.Y}

		// 转换为画布坐标
		x1, y1 := fauxgl.UVToCanvas(uv1, width, height)
		x2, y2 := fauxgl.UVToCanvas(uv2, width, height)
		x3, y3 := fauxgl.UVToCanvas(uv3, width, height)

		// 绘制三角形边框
		drawLine(img, x1, y1, x2, y2, color.RGBA{255, 100, 100, 255})
		drawLine(img, x2, y2, x3, y3, color.RGBA{100, 255, 100, 255})
		drawLine(img, x3, y3, x1, y1, color.RGBA{100, 100, 255, 255})
	}

	return img
}

// drawLine 在图像上绘制线段
func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	// Bresenham算法实现线段绘制
	dx := int(math.Abs(float64(x1 - x0)))
	dy := int(math.Abs(float64(y1 - y0)))
	sx := 1
	sy := 1
	if x0 >= x1 {
		sx = -1
	}
	if y0 >= y1 {
		sy = -1
	}
	err := dx - dy

	for {
		// 确保坐标在图像范围内
		if x0 >= 0 && x0 < img.Bounds().Dx() && y0 >= 0 && y0 < img.Bounds().Dy() {
			img.Set(x0, y0, c)
		}

		// 检查是否到达终点
		if x0 == x1 && y0 == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// saveImage 保存图像到文件
func saveImage(img image.Image, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}
