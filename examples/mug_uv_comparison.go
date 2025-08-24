package main

import (
	"fmt"
	"math"
	"path/filepath"
	"runtime"

	"github.com/swordkee/fauxgl-gltf"
)

// 高质量渲染参数
const (
	scale  = 4    // 4x 超采样抗锯齿
	width  = 2000 // 输出宽度
	height = 2000 // 输出高度
	fovy   = 30   // 垂直视野角度
	near   = 1    // 近裁剪面
	far    = 20   // 远裁剪面
)

var (
	// 相机参数
	eye    = fauxgl.V(2.5, 4, 4.0)
	center = fauxgl.V(0, 1.14, 0.4)
	up     = fauxgl.V(0, 1, 0)
)

func main() {
	fmt.Println("=== FauxGL-GLTF UV映射渲染对比 ===")
	fmt.Println("对比原始渲染与修复版渲染")
	fmt.Println("")

	// 设置并行处理
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("使用 %d 个CPU核心进行并行渲染\n", runtime.NumCPU())

	// 获取当前工作目录
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	gltfPath := filepath.Join(dir, "./gltf/mug.gltf")

	fmt.Printf("GLTF文件路径: %s\n", gltfPath)

	// 加载GLTF场景
	scene, err := fauxgl.LoadGLTFScene(gltfPath)
	if err != nil {
		panic(err)
	}

	fmt.Printf("场景加载成功: %d材质, %d网格, %d纹理\n",
		len(scene.Materials), len(scene.Meshes), len(scene.Textures))

	// 1. 原始渲染（可能导致变形）
	fmt.Println("\n=== 1. 原始渲染（可能导致变形） ===")
	originalScene := cloneScene(scene)
	preprocessMeshesOriginal(originalScene)
	renderScene(originalScene, "mug_original.png", "原始渲染")

	// 2. 修复版渲染（避免变形）
	fmt.Println("\n=== 2. 修复版渲染（避免变形） ===")
	fixedScene := cloneScene(scene)
	preprocessMeshesFixed(fixedScene)
	setupUVMapping(fixedScene)
	renderScene(fixedScene, "mug_fixed_comparison.png", "修复版渲染")

	fmt.Println("\n=== 渲染对比完成 ===")
	fmt.Println("输出文件:")
	fmt.Println("  - mug_original.png: 原始渲染（可能导致变形）")
	fmt.Println("  - mug_fixed_comparison.png: 修复版渲染（避免变形）")
}

// cloneScene 克隆场景
func cloneScene(original *fauxgl.Scene) *fauxgl.Scene {
	// 创建新场景
	newScene := fauxgl.NewScene(original.Name)

	// 复制基本属性
	newScene.Cameras = make([]*fauxgl.Camera, len(original.Cameras))
	copy(newScene.Cameras, original.Cameras)

	newScene.Lights = make([]fauxgl.Light, len(original.Lights))
	copy(newScene.Lights, original.Lights)

	// 复制材质
	newScene.Materials = make(map[string]*fauxgl.PBRMaterial)
	for k, v := range original.Materials {
		newScene.Materials[k] = v
	}

	// 复制纹理
	newScene.Textures = make(map[string]*fauxgl.AdvancedTexture)
	for k, v := range original.Textures {
		newScene.Textures[k] = v
	}

	// 复制网格
	newScene.Meshes = make(map[string]*fauxgl.Mesh)
	for k, v := range original.Meshes {
		newScene.Meshes[k] = v
	}

	// 复制根节点
	newScene.RootNode = original.RootNode

	return newScene
}

// preprocessMeshesOriginal 原始网格预处理（可能导致变形）
func preprocessMeshesOriginal(scene *fauxgl.Scene) {
	fmt.Println("  原始网格预处理（可能导致变形）")
	for name, mesh := range scene.Meshes {
		fmt.Printf("    处理网格: %s\n", name)

		// 应用可能导致变形的预处理
		mesh.BiUnitCube() // 这可能导致模型变形
		mesh.SmoothNormalsThreshold(fauxgl.Radians(30))
	}
}

// preprocessMeshesFixed 修复版网格预处理（避免变形）
func preprocessMeshesFixed(scene *fauxgl.Scene) {
	fmt.Println("  修复版网格预处理（避免变形）")
	for name, mesh := range scene.Meshes {
		fmt.Printf("    处理网格: %s\n", name)

		// 只进行法线处理，避免几何变换导致的变形
		mesh.SmoothNormalsThreshold(fauxgl.Radians(30))
	}
}

// setupUVMapping 设置UV映射
func setupUVMapping(scene *fauxgl.Scene) {
	fmt.Println("  设置UV映射")

	// 为场景中的所有纹理设置UV映射
	for name, texture := range scene.Textures {
		fmt.Printf("    为纹理 %s 设置UV映射\n", name)

		// 创建UV修改器
		modifier := fauxgl.NewUVModifier()

		// 全局变换：保持原始UV坐标
		globalTransform := fauxgl.NewUVTransform()
		globalTransform.ScaleU = 1.0
		globalTransform.ScaleV = 1.0
		globalTransform.OffsetU = 0.0
		globalTransform.OffsetV = 0.0
		modifier.SetGlobalTransform(globalTransform)

		// 部分贴图 - 区域性映射
		fmt.Println("      设置部分区域UV映射")

		// 前面板标志区域
		frontLogoMapping := &fauxgl.UVMapping{
			Name:    "front_logo",
			Enabled: true,
			Region: fauxgl.UVRegion{
				MinU: 0.25, MaxU: 0.75,
				MinV: 0.35, MaxV: 0.65,
				MaskType: fauxgl.UVMaskRectangle,
			},
			Transform: &fauxgl.UVTransform{
				ScaleU: 1.0, ScaleV: 1.0,
				OffsetU: 0.0, OffsetV: 0.0,
				PivotU: 0.5, PivotV: 0.5,
			},
			BlendMode: fauxgl.UVBlendReplace,
			Priority:  3,
		}
		modifier.AddMapping(frontLogoMapping)

		// 上部装饰带
		upperBandMapping := &fauxgl.UVMapping{
			Name:    "upper_band",
			Enabled: true,
			Region: fauxgl.UVRegion{
				MinU: 0.1, MaxU: 0.9,
				MinV: 0.75, MaxV: 0.85,
				MaskType: fauxgl.UVMaskRectangle,
			},
			Transform: &fauxgl.UVTransform{
				ScaleU: 2.0, ScaleV: 0.5,
				OffsetU: -0.5, OffsetV: 0.6,
			},
			BlendMode: fauxgl.UVBlendReplace,
			Priority:  2,
		}
		modifier.AddMapping(upperBandMapping)

		// 应用UV修改器到纹理
		texture.UVModifier = modifier
		fmt.Printf("      ✓ UV映射设置完成\n")
	}
}

// renderScene 渲染场景
func renderScene(scene *fauxgl.Scene, filename, description string) {
	fmt.Printf("  渲染场景: %s -> %s\n", description, filename)

	// 创建渲染上下文
	context := fauxgl.NewContext(width*scale, height*scale)
	context.ClearColor = fauxgl.Color{0.1, 0.12, 0.15, 1.0} // 深色背景
	context.ClearColorBuffer()
	context.ClearDepthBuffer()

	// 设置相机矩阵
	aspect := float64(width) / float64(height)
	matrix := fauxgl.LookAt(eye, center, up).Perspective(fovy, aspect, near, far)

	// 设置多光源系统
	lightSystem := setupLightingSystem()

	// 获取可渲染节点
	renderableNodes := scene.RootNode.GetRenderableNodes()
	fmt.Printf("    渲染 %d 个节点\n", len(renderableNodes))

	// 渲染每个节点
	for i, node := range renderableNodes {
		if node.Mesh == nil || node.Material == nil {
			continue
		}

		// 查找材质名称
		materialName := "unknown"
		for name, mat := range scene.Materials {
			if mat == node.Material {
				materialName = name
				break
			}
		}

		fmt.Printf("      渲染节点 %d: %s (材质: %s)\n", i+1, node.Name, materialName)

		// 创建着色器
		shader := fauxgl.NewPhongShader(matrix, lightSystem.MainLight, eye)

		// 根据材质类型设置参数
		switch materialName {
		case "material_0": // 杯底纯色材质
			shader.DiffuseColor = fauxgl.Color{0.8, 0.75, 0.7, 1.0}
			shader.SpecularColor = fauxgl.Color{0.3, 0.3, 0.3, 1.0}
			shader.SpecularPower = 16
		case "material_1": // 主体纹理材质
			shader.DiffuseColor = fauxgl.Color{1.0, 1.0, 1.0, 1.0}
			shader.SpecularColor = fauxgl.Color{0.8, 0.8, 0.8, 1.0}
			shader.SpecularPower = 64
			// 应用纹理
			if node.Material.BaseColorTexture != nil {
				shader.Texture = node.Material.BaseColorTexture
			}
		case "material_2": // 绿色装饰带
			shader.DiffuseColor = fauxgl.Color{0.3, 0.7, 0.4, 1.0}
			shader.SpecularColor = fauxgl.Color{0.2, 0.4, 0.2, 1.0}
			shader.SpecularPower = 32
		case "material_3": // 蓝色装饰带
			shader.DiffuseColor = fauxgl.Color{0.3, 0.5, 0.8, 1.0}
			shader.SpecularColor = fauxgl.Color{0.2, 0.3, 0.5, 1.0}
			shader.SpecularPower = 32
		case "material_4": // 黄色杯口
			shader.DiffuseColor = fauxgl.Color{0.9, 0.8, 0.2, 1.0}
			shader.SpecularColor = fauxgl.Color{0.6, 0.5, 0.1, 1.0}
			shader.SpecularPower = 48
		default:
			shader.DiffuseColor = fauxgl.Color{0.8, 0.8, 0.8, 1.0}
			shader.SpecularColor = fauxgl.Color{0.4, 0.4, 0.4, 1.0}
			shader.SpecularPower = 32
		}

		// 应用材质颜色
		baseColor := fauxgl.Color{
			node.Material.BaseColorFactor.R,
			node.Material.BaseColorFactor.G,
			node.Material.BaseColorFactor.B,
			node.Material.BaseColorFactor.A,
		}
		shader.ObjectColor = baseColor

		// 混合环境光
		enhancedColor := fauxgl.Color{
			baseColor.R + lightSystem.AmbientColor.R*lightSystem.AmbientStrength,
			baseColor.G + lightSystem.AmbientColor.G*lightSystem.AmbientStrength,
			baseColor.B + lightSystem.AmbientColor.B*lightSystem.AmbientStrength,
			baseColor.A,
		}
		// 确保颜色值在合理范围内
		enhancedColor.R = math.Min(enhancedColor.R, 1.0)
		enhancedColor.G = math.Min(enhancedColor.G, 1.0)
		enhancedColor.B = math.Min(enhancedColor.B, 1.0)
		shader.ObjectColor = enhancedColor

		// 渲染
		context.Shader = shader
		context.DrawMesh(node.Mesh)
	}

	// 保存结果
	err := fauxgl.SavePNG(filename, context.Image())
	if err != nil {
		fmt.Printf("      保存失败: %v\n", err)
	} else {
		fmt.Printf("      ✓ 渲染结果已保存: %s\n", filename)
	}
}

// LightingSystem 光照系统
type LightingSystem struct {
	MainLight       fauxgl.Vector // 主光源
	FillLight       fauxgl.Vector // 补光
	RimLight        fauxgl.Vector // 边缘光
	AmbientColor    fauxgl.Color  // 环境光颜色
	AmbientStrength float64       // 环境光强度
}

// setupLightingSystem 设置光照系统
func setupLightingSystem() *LightingSystem {
	lightSystem := &LightingSystem{
		// 主光源：从右上方照射
		MainLight: fauxgl.V(-0.4, -0.6, -0.8).Normalize(),
		// 补光：从左侧补光
		FillLight: fauxgl.V(0.7, -0.2, -0.3).Normalize(),
		// 边缘光：从背后照射
		RimLight: fauxgl.V(0.2, 0.3, 0.9).Normalize(),
		// 环境光
		AmbientColor:    fauxgl.Color{0.4, 0.45, 0.5, 1.0},
		AmbientStrength: 0.3,
	}
	return lightSystem
}
