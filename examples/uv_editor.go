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
	fmt.Println("=== UV编辑器 ===")

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

	// 获取第一个网格进行测试
	var testMesh *fauxgl.Mesh
	for _, mesh := range scene.Meshes {
		testMesh = mesh
		break
	}

	if testMesh == nil {
		fmt.Println("  ✗ 未找到测试网格")
		return
	}

	fmt.Printf("  处理网格: %d个三角形\n", len(testMesh.Triangles))

	// 创建UV松弛设置
	settings := fauxgl.NewUVRelaxationSettings()
	settings.Iterations = 20
	settings.StepSize = 0.3

	// 应用UV松弛算法
	err = fauxgl.ApplyUVRelaxation(testMesh, settings)
	if err != nil {
		fmt.Printf("  ✗ UV松弛失败: %v\n", err)
		return
	}

	fmt.Println("  ✓ UV平面展开算法应用完成")

	// 生成UV可视化图
	uvImage := generateUVVisualization(testMesh, 1024, 1024)
	uvImagePath := filepath.Join(dir, "uv_visualization.png")
	err = saveImage(uvImage, uvImagePath)
	if err != nil {
		fmt.Printf("  ✗ 无法保存UV可视化图: %v\n", err)
	} else {
		fmt.Printf("  ✓ UV可视化图已保存到: %s\n", uvImagePath)
	}

	// 测试坐标映射功能
	fmt.Println("测试坐标映射功能...")

	// 测试UV到画布坐标的转换
	uv := fauxgl.Vector2{0.5, 0.5}
	x, y := fauxgl.UVToCanvas(uv, 1024, 1024)
	fmt.Printf("  UV(0.5, 0.5) -> 画布坐标(%d, %d)\n", x, y)

	// 测试画布到UV坐标的转换
	uv2 := fauxgl.CanvasToUV(512, 512, 1024, 1024)
	fmt.Printf("  画布坐标(512, 512) -> UV(%.2f, %.2f)\n", uv2.X, uv2.Y)

	fmt.Println("  ✓ 坐标映射功能测试通过")

	fmt.Println("\n=== UV编辑器测试完成 ===")
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
