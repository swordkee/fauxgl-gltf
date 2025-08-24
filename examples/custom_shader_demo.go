package main

import (
	"fmt"
	"math"

	"github.com/swordkee/fauxgl-gltf"
)

const (
	width  = 800
	height = 600
	scale  = 1
)

// CustomShader 自定义着色器示例
type CustomShader struct {
	Matrix         fauxgl.Matrix
	LightDirection fauxgl.Vector
	CameraPosition fauxgl.Vector
	Time           float64
}

// NewCustomShader 创建新的自定义着色器
func NewCustomShader(matrix fauxgl.Matrix, lightDirection, cameraPosition fauxgl.Vector, time float64) *CustomShader {
	return &CustomShader{
		Matrix:         matrix,
		LightDirection: lightDirection,
		CameraPosition: cameraPosition,
		Time:           time,
	}
}

// Vertex 顶点着色器
func (shader *CustomShader) Vertex(v fauxgl.Vertex) fauxgl.Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

// Fragment 片段着色器
func (shader *CustomShader) Fragment(v fauxgl.Vertex) fauxgl.Color {
	// 基于时间的颜色变化
	red := 0.5 + 0.5*math.Sin(shader.Time+v.Position.X)
	green := 0.5 + 0.5*math.Sin(shader.Time*1.2+v.Position.Y)
	blue := 0.5 + 0.5*math.Sin(shader.Time*0.8+v.Position.Z)

	baseColor := fauxgl.Color{red, green, blue, 1.0}

	// 计算光照
	diffuse := math.Max(v.Normal.Dot(shader.LightDirection), 0)
	lightIntensity := 0.3 + 0.7*diffuse

	// 计算边缘光
	viewDir := shader.CameraPosition.Sub(v.Position).Normalize()
	rim := math.Pow(1.0-math.Max(v.Normal.Dot(viewDir), 0), 2.0)
	rimLight := 0.3 * rim

	// 组合最终颜色
	finalColor := baseColor.MulScalar(lightIntensity)
	finalColor = finalColor.Add(fauxgl.Color{1, 1, 1, 1}.MulScalar(rimLight))

	return finalColor.Alpha(1.0)
}

// WaveShader 波浪着色器示例
type WaveShader struct {
	Matrix         fauxgl.Matrix
	LightDirection fauxgl.Vector
	CameraPosition fauxgl.Vector
	Time           float64
	Frequency      float64
	Amplitude      float64
}

// NewWaveShader 创建新的波浪着色器
func NewWaveShader(matrix fauxgl.Matrix, lightDirection, cameraPosition fauxgl.Vector, time, frequency, amplitude float64) *WaveShader {
	return &WaveShader{
		Matrix:         matrix,
		LightDirection: lightDirection,
		CameraPosition: cameraPosition,
		Time:           time,
		Frequency:      frequency,
		Amplitude:      amplitude,
	}
}

// Vertex 顶点着色器（添加波浪变形）
func (shader *WaveShader) Vertex(v fauxgl.Vertex) fauxgl.Vertex {
	// 应用波浪变形
	waveOffset := shader.Amplitude * math.Sin(v.Position.X*shader.Frequency+shader.Time)
	v.Position.Y += waveOffset

	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

// Fragment 片段着色器
func (shader *WaveShader) Fragment(v fauxgl.Vertex) fauxgl.Color {
	// 基于高度的颜色变化
	heightFactor := (v.Position.Y + shader.Amplitude) / (2 * shader.Amplitude)
	blue := 0.5 + 0.5*heightFactor
	green := 0.3 + 0.7*(1.0-heightFactor)

	baseColor := fauxgl.Color{0, green, blue, 1.0}

	// 计算光照
	diffuse := math.Max(v.Normal.Dot(shader.LightDirection), 0)
	lightIntensity := 0.4 + 0.6*diffuse

	// 计算边缘光
	viewDir := shader.CameraPosition.Sub(v.Position).Normalize()
	rim := math.Pow(1.0-math.Max(v.Normal.Dot(viewDir), 0), 3.0)
	rimLight := 0.4 * rim

	// 组合最终颜色
	finalColor := baseColor.MulScalar(lightIntensity)
	finalColor = finalColor.Add(fauxgl.Color{1, 1, 1, 1}.MulScalar(rimLight))

	return finalColor.Alpha(1.0)
}

func main() {
	// 创建场景
	scene := fauxgl.NewScene("Custom Shader Demo")

	// 创建几何体
	sphere := fauxgl.NewSphere(4)
	plane := fauxgl.NewCube()

	// 调整平面大小作为地面
	plane.Transform(fauxgl.Scale(fauxgl.Vector{5, 0.1, 5}).Translate(fauxgl.Vector{0, -2, 0}))

	// 创建材质
	sphereMaterial := fauxgl.NewPBRMaterial()
	sphereMaterial.BaseColorFactor = fauxgl.Color{1, 1, 1, 1}

	planeMaterial := fauxgl.NewPBRMaterial()
	planeMaterial.BaseColorFactor = fauxgl.Color{0.5, 0.5, 0.5, 1}

	// 添加网格和材质到场景
	scene.Meshes["sphere"] = sphere
	scene.Meshes["plane"] = plane

	scene.Materials["sphere_material"] = sphereMaterial
	scene.Materials["plane_material"] = planeMaterial

	// 创建场景节点
	sphereNode := fauxgl.NewSceneNode("sphere")
	sphereNode.Mesh = sphere
	sphereNode.Material = sphereMaterial
	sphereNode.SetTransform(fauxgl.Translate(fauxgl.Vector{0, 0, 0}))

	planeNode := fauxgl.NewSceneNode("plane")
	planeNode.Mesh = plane
	planeNode.Material = planeMaterial

	// 添加节点到场景
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
	context.ClearColor = fauxgl.Color{0.05, 0.05, 0.1, 1} // 深蓝色背景

	// 渲染不同时间点的动画帧
	fmt.Println("渲染自定义着色器动画...")
	renderCustomShaderAnimation(context, scene, camera, light)

	fmt.Println("自定义着色器演示完成！")
}

// renderCustomShaderAnimation 渲染自定义着色器动画
func renderCustomShaderAnimation(context *fauxgl.Context, scene *fauxgl.Scene, camera *fauxgl.Camera, light fauxgl.Light) {
	// 获取相机矩阵
	cameraMatrix := camera.GetCameraMatrix()

	// 获取所有可渲染节点
	renderables := scene.RootNode.GetRenderableNodes()

	// 渲染多个动画帧
	for frame := 0; frame < 30; frame++ {
		time := float64(frame) * 0.2

		// 清除缓冲区
		context.ClearColorBuffer()
		context.ClearDepthBuffer()

		// 渲染每个节点
		for _, node := range renderables {
			if node.Mesh == nil || node.Material == nil {
				continue
			}

			// 计算最终变换矩阵
			modelMatrix := node.WorldTransform
			finalMatrix := cameraMatrix.Mul(modelMatrix)

			// 根据节点名称选择不同的着色器
			if node.Name == "sphere" {
				// 使用波浪着色器
				waveShader := NewWaveShader(
					finalMatrix,
					light.Direction,
					camera.Position,
					time,
					3.0, // 频率
					0.3, // 振幅
				)
				context.Shader = waveShader
			} else {
				// 使用自定义着色器
				customShader := NewCustomShader(
					finalMatrix,
					light.Direction,
					camera.Position,
					time,
				)
				context.Shader = customShader
			}

			// 渲染网格
			context.DrawMesh(node.Mesh)
		}

		// 保存图像
		filename := fmt.Sprintf("custom_shader_frame_%03d.png", frame)
		fauxgl.SavePNG(filename, context.Image())
		fmt.Printf("保存动画帧 %s\n", filename)
	}
}
