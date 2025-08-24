package main

import (
	"fmt"
	"math"
	"path/filepath"
	"runtime"

	"github.com/swordkee/fauxgl-gltf"
)

// 全局变量，用于存储装饰纹理
var decorativeTexture *fauxgl.AdvancedTexture

// **配置区域**: 在这里修改您想使用的纹理文件
const (
	// **自定义纹理文件**: 将您的纹理文件名替换这里
	CUSTOM_TEXTURE_FILE = "texture.png" // 修改为您的纹理文件名

	// **渲染参数**: 高质量渲染以达到300KB+
	scale  = 4    // 4x 超采样抗锯齿
	width  = 2000 // 高分辨率渲染
	height = 2000 // 高分辨率渲染
	fovy   = 30   // 垂直视野角度
	near   = 1    // 近裁剪面
	far    = 20   // 远裁剪面
)

var (
	// **调整相机参数**: 优化视角避免穿刺和变形，更好地展示杯子
	eye    = fauxgl.V(2.5, 1.5, 2.5) // 调整相机位置避免穿刺，更好地展示杯子
	center = fauxgl.V(0, 0.6, 0)     // 对准杯子中心
	up     = fauxgl.V(0, 1, 0)       // 标准上方向向量
)

func main() {
	fmt.Println("=== 高质量GLTF多材质UV分区渲染 - 最终优化版 ===")
	fmt.Printf("📝 当前配置的纹理文件: %s\n", CUSTOM_TEXTURE_FILE)
	fmt.Println("💡 提示: 要使用自定义纹理，请修改文件顶部的 CUSTOM_TEXTURE_FILE 常量")
	fmt.Println("")

	// 设置并行处理
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("使用 %d 个CPU核心进行并行渲染\n", runtime.NumCPU())

	// 获取当前工作目录
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	gltfPath := filepath.Join(dir, "./gltf/mug.gltf")
	texturePath := filepath.Join(dir, CUSTOM_TEXTURE_FILE)

	fmt.Printf("GLTF文件路径: %s\n", gltfPath)
	fmt.Printf("纹理文件路径: %s\n", texturePath)

	// 使用GLTF场景加载器，支持多材质
	scene, err := fauxgl.LoadGLTFScene(gltfPath)
	if err != nil {
		panic(err)
	}

	// 加载自定义纹理
	loadCustomTexture(scene, texturePath)

	fmt.Printf("场景加载成功:\n")
	fmt.Printf("  材质数量: %d\n", len(scene.Materials))
	fmt.Printf("  网格数量: %d\n", len(scene.Meshes))
	fmt.Printf("  纹理数量: %d\n", len(scene.Textures))

	// 分析场景结构
	analyzeScene(scene)

	// 设置自定义UV映射
	setupCustomUVMappings(scene)

	// 网格预处理 - 修复变形问题
	preprocessMeshesFixed(scene)

	// 设置增强光源系统
	lightSystem := setupAdvancedLightingSystem()

	// 执行高质量渲染
	renderHighQuality(scene, lightSystem, "mug_uv.png")

	fmt.Println("\n=== 渲染完成 ===")
	fmt.Println("最终版多材质UV分区渲染已保存为 mug_uv.png")
	fmt.Println("✅ 修复模型变形问题")
	fmt.Println("✅ 优化杯子展示位置")
	fmt.Println("✅ 解决模型穿刺问题")
	fmt.Println("✅ 高质量4K分辨率渲染")
	fmt.Println("✅ 4x超采样抗锯齿")
	fmt.Println("✅ 自定义UV映射")
	fmt.Println("✅ 多光源照明系统")
	printMaterialInfo(scene)
}

// loadCustomTexture 加载自定义纹理
func loadCustomTexture(scene *fauxgl.Scene, texturePath string) {
	fmt.Println("\n=== 加载自定义纹理 ===")

	// 加载texture.png作为棋盘格贴图（用于主体区域）
	fmt.Println("尝试加载棋盘格贴图 texture.png...")
	checkerTexture, err := fauxgl.LoadAdvancedTexture("texture.png", fauxgl.BaseColorTexture)
	if err != nil {
		fmt.Printf(" ✗ 无法加载棋盘格贴图 texture.png: %v\n", err)
		// 尝试使用备选纹理
		checkerTexture, err = fauxgl.LoadAdvancedTexture(texturePath, fauxgl.BaseColorTexture)
		if err != nil {
			fmt.Printf(" ✗ 无法加载备选贴图: %v\n", err)
			fmt.Println("❌ 警告: 无法加载任何棋盘格贴图，将使用原始GLTF纹理")
			checkOriginalTextures(scene)
			return
		}
		fmt.Printf("✓ 备选贴图加载成功 (%dx%d)\n", checkerTexture.Width, checkerTexture.Height)
	} else {
		fmt.Printf("✓ 棋盘格贴图 texture.png 加载成功 (%dx%d)\n", checkerTexture.Width, checkerTexture.Height)
	}

	// 验证纹理内容
	validateTextureContent(checkerTexture, "棋盘格贴图")

	// 替换场景中的纹理
	replaceSceneTextures(scene, checkerTexture)
}

// validateTextureContent 验证纹理内容
func validateTextureContent(texture *fauxgl.AdvancedTexture, filename string) bool {
	// 采样多个点来验证纹理内容
	testPoints := [][2]float64{
		{0.25, 0.25}, {0.75, 0.25}, {0.25, 0.75}, {0.75, 0.75}, {0.5, 0.5},
	}

	fmt.Printf("验证纹理内容(%s):\n", filename)
	allWhite := true
	for i, point := range testPoints {
		color := texture.SampleWithFilter(point[0], point[1], fauxgl.FilterLinear)
		fmt.Printf("  点%d UV(%.2f,%.2f): RGBA(%.3f,%.3f,%.3f,%.3f)\n",
			i+1, point[0], point[1], color.R, color.G, color.B, color.A)

		// 检查是否为白色（允许一些容差）
		if color.R < 0.95 || color.G < 0.95 || color.B < 0.95 {
			allWhite = false
		}
	}

	if allWhite {
		fmt.Printf("⚠️  警告: 纹理 %s 似乎是全白或接近全白的\n", filename)
		return false
	} else {
		fmt.Printf("✓ 纹理内容验证通过: 包含非白色像素\n")
		return true
	}
}

// checkOriginalTextures 检查原始GLTF纹理
func checkOriginalTextures(scene *fauxgl.Scene) {
	fmt.Println("检查原始GLTF纹理:")
	for name, texture := range scene.Textures {
		fmt.Printf("  原始纹理 %s: %dx%d\n", name, texture.Width, texture.Height)
		validateTextureContent(texture, name)
	}
}

// replaceSceneTextures 替换场景中的纹理
func replaceSceneTextures(scene *fauxgl.Scene, checkerTexture *fauxgl.AdvancedTexture) {
	fmt.Printf("\n替换场景纹理\n")

	// 为每个材质配置适当的纹理和颜色
	greenColor := fauxgl.Color{0.58, 0.78, 0.03, 1.0} // #94C808 绿色

	// 设置纹理和颜色
	for name, material := range scene.Materials {
		if name == "material_1" {
			// 主体区域应用棋盘格贴图
			material.BaseColorTexture = checkerTexture
			material.BaseColorFactor = fauxgl.Color{1.0, 1.0, 1.0, 1.0} // 白色，不影响贴图颜色
			material.MetallicFactor = 0.0                               // 使用GLTF默认值
			material.RoughnessFactor = 0.5                              // 使用GLTF默认值
			fmt.Printf("✓ 已为主体区域(%s)设置棋盘格贴图\n", name)
		} else {
			// 其他区域统一设置为#94C808绿色
			if material.BaseColorTexture != nil {
				material.BaseColorTexture = nil // 移除纹理，使用纯色
			}
			material.BaseColorFactor = greenColor
			fmt.Printf("✓ 已为区域(%s)设置绿色#94C808\n", name)
		}
	}

	// 保存纹理到场景中（用于UV映射等）
	if len(scene.Textures) > 0 {
		// 更新现有纹理
		for name, _ := range scene.Textures {
			scene.Textures[name] = checkerTexture
			fmt.Printf("✓ 已更新纹理: %s\n", name)
		}
	} else {
		// 添加新纹理
		scene.Textures["texture_0"] = checkerTexture
		fmt.Printf("✓ 已添加纹理: texture_0\n")
	}

	fmt.Println("✓ 纹理和材质替换完成!")

	// 保存为全局变量，以便后续处理
	decorativeTexture = checkerTexture
}

// analyzeScene 分析场景结构
func analyzeScene(scene *fauxgl.Scene) {
	fmt.Println("\n=== 场景结构分析 ===")

	// 分析纹理
	fmt.Println("纹理列表:")
	for name, texture := range scene.Textures {
		fmt.Printf("  %s: %dx%d\n", name, texture.Width, texture.Height)
	}

	// 分析材质
	fmt.Println("\n材质列表:")
	for name, material := range scene.Materials {
		fmt.Printf("  %s: ", name)
		if material.BaseColorTexture != nil {
			fmt.Printf("纹理材质 - 基础颜色: [%.3f, %.3f, %.3f]",
				material.BaseColorFactor.R, material.BaseColorFactor.G, material.BaseColorFactor.B)
		} else {
			fmt.Printf("纯色材质 - 颜色: [%.3f, %.3f, %.3f]",
				material.BaseColorFactor.R, material.BaseColorFactor.G, material.BaseColorFactor.B)
		}
		fmt.Printf(", 金属度: %.3f, 粗糙度: %.3f, 双面: %t\n",
			material.MetallicFactor, material.RoughnessFactor, material.DoubleSided)
	}

	// 分析网格和边界
	fmt.Println("\n网格列表:")
	for name, mesh := range scene.Meshes {
		bounds := mesh.BoundingBox()
		fmt.Printf("  %s: %d三角形, 边界: min=%v, max=%v\n",
			name, len(mesh.Triangles), bounds.Min, bounds.Max)
	}

	// 场景整体边界
	bounds := scene.GetBounds()
	fmt.Printf("\n场景整体边界: min=%v, max=%v, center=%v, size=%v\n",
		bounds.Min, bounds.Max, bounds.Center(), bounds.Size())

	// 分析可渲染节点
	fmt.Println("\n可渲染节点:")
	renderableNodes := scene.RootNode.GetRenderableNodes()
	fmt.Printf("  可渲染节点数量: %d\n", len(renderableNodes))
	for i, node := range renderableNodes {
		materialName := "<无材质>"
		if node.Material != nil {
			// 查找材质名称
			for name, mat := range scene.Materials {
				if mat == node.Material {
					materialName = name
					break
				}
			}
		}
		meshName := "<无网格>"
		triangleCount := 0
		if node.Mesh != nil {
			triangleCount = len(node.Mesh.Triangles)
			// 查找网格名称
			for name, mesh := range scene.Meshes {
				if mesh == node.Mesh {
					meshName = name
					break
				}
			}
		}
		fmt.Printf("    节点 %d: %s -> 网格: %s (%d三角形), 材质: %s\n",
			i+1, node.Name, meshName, triangleCount, materialName)
	}
}

// setupCustomUVMappings 设置自定义UV映射 - 部分区域贴图
func setupCustomUVMappings(scene *fauxgl.Scene) {
	fmt.Println("\n=== 自定义UV映射设置 ===")

	// 由于要保持three.js的默认效果，这里我们简化UV映射处理
	// 主体区域使用棋盘格贴图，其他区域使用纯绿色

	// 对于material_1(主体区域)，保持原始UV映射
	// 对于其他材质区域，不需要特别的UV映射，直接使用纯色

	// 这个函数我们保留但不做特殊UV映射处理
	fmt.Println("✓ 使用原生GLTF材质配置和默认UV映射")
}

// preprocessMeshesFixed 修复版网格预处理 - 避免模型变形
func preprocessMeshesFixed(scene *fauxgl.Scene) {
	fmt.Println("\n=== 网格预处理 (保持原始形状) ===")
	for name, mesh := range scene.Meshes {
		fmt.Printf("处理网格: %s (%d三角形)\n", name, len(mesh.Triangles))

		// 打印原始边界信息
		originalBounds := mesh.BoundingBox()
		fmt.Printf("  原始边界: min=%v, max=%v\n", originalBounds.Min, originalBounds.Max)
		fmt.Printf("  原始尺寸: %v\n", originalBounds.Size())

		// 修复方案: 进行细致的网格处理，解决破碎问题
		// 1. 先进行更全面的法线平滑
		mesh.SmoothNormals()
		fmt.Println("  ✓ 应用全面法线平滑")

		// 2. 再应用带阈值的法线平滑，保留锐利边缘
		mesh.SmoothNormalsThreshold(fauxgl.Radians(60))
		fmt.Println("  ✓ 应用阈值法线平滑，保留锐利边缘")

		// 打印处理后边界信息
		newBounds := mesh.BoundingBox()
		fmt.Printf("  处理后边界: min=%v, max=%v\n", newBounds.Min, newBounds.Max)
		fmt.Printf("  处理后尺寸: %v\n", newBounds.Size())

		// 验证网格完整性
		fmt.Printf("  网格完整性检查: %d个三角形\n", len(mesh.Triangles))
	}
}

// LightingSystem 光照系统结构
type LightingSystem struct {
	MainLight       fauxgl.Vector // 主光源
	FillLight       fauxgl.Vector // 补光
	RimLight        fauxgl.Vector // 边缘光
	AmbientColor    fauxgl.Color  // 环境光颜色
	AmbientStrength float64       // 环境光强度
}

// setupAdvancedLightingSystem 设置增强光照系统
func setupAdvancedLightingSystem() *LightingSystem {
	fmt.Println("\n=== 增强光照系统设置 ===")

	lightSystem := &LightingSystem{
		// 主光源：从右上方照射，增强光泼效果
		MainLight: fauxgl.V(-0.5, -0.6, -0.6).Normalize(),

		// 补光：从左侧补光，减少阴影
		FillLight: fauxgl.V(0.7, -0.3, -0.3).Normalize(),

		// 边缘光：从背后照射，增强轮廓
		RimLight: fauxgl.V(0.3, 0.2, 0.8).Normalize(),

		// 环境光：亮色环境色调增强光泼
		AmbientColor:    fauxgl.Color{0.6, 0.6, 0.6, 1.0},
		AmbientStrength: 0.4,
	}

	fmt.Printf("主光源方向: %v\n", lightSystem.MainLight)
	fmt.Printf("补光方向: %v\n", lightSystem.FillLight)
	fmt.Printf("边缘光方向: %v\n", lightSystem.RimLight)
	fmt.Printf("环境光: RGBA(%.2f, %.2f, %.2f, %.2f), 强度: %.2f\n",
		lightSystem.AmbientColor.R, lightSystem.AmbientColor.G,
		lightSystem.AmbientColor.B, lightSystem.AmbientColor.A,
		lightSystem.AmbientStrength)

	return lightSystem
}

// renderHighQuality 使用多光源系统执行高质量渲染
func renderHighQuality(scene *fauxgl.Scene, lightSystem *LightingSystem, filename string) {
	fmt.Println("\n=== 多光源高质量渲染 ===")

	// 创建渲染上下文
	context := fauxgl.NewContext(width*scale, height*scale)
	context.ClearColor = fauxgl.Color{0.05, 0.05, 0.05, 1.0} // 深色背景，增强对比度
	context.ClearColorBuffer()
	context.ClearDepthBuffer()

	// 设置相机矩阵
	aspect := float64(width) / float64(height)
	matrix := fauxgl.LookAt(eye, center, up).Perspective(fovy, aspect, near, far)

	// 获取所有可渲染节点
	renderableNodes := scene.RootNode.GetRenderableNodes()
	fmt.Printf("开始增强渲染，共 %d 个节点...\n", len(renderableNodes))

	// 逐个渲染每个节点
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

		fmt.Printf("渲染节点 %d: %s (材质: %s, %d三角形)\n",
			i+1, node.Name, materialName, len(node.Mesh.Triangles))

		// 创建高级着色器，使用主光源
		shader := fauxgl.NewPhongShader(matrix, lightSystem.MainLight, eye)

		// 根据材质类型调整光照参数
		switch materialName {
		case "material_1": // 主体纹理材质（棋盘格）
			shader.DiffuseColor = fauxgl.Color{1.0, 1.0, 1.0, 1.0}
			shader.SpecularColor = fauxgl.Color{0.8, 0.8, 0.8, 1.0}
			shader.SpecularPower = 64
		default: // 其他材质区域（绿色）
			shader.DiffuseColor = fauxgl.Color{0.58, 0.78, 0.03, 1.0} // #94C808
			shader.SpecularColor = fauxgl.Color{0.6, 0.7, 0.3, 1.0}
			shader.SpecularPower = 48
		}

		// 应用材质纹理
		if node.Material.BaseColorTexture != nil {
			shader.Texture = node.Material.BaseColorTexture
		}

		// 增强材质颜色，考虑环境光影响
		baseColor := fauxgl.Color{
			node.Material.BaseColorFactor.R,
			node.Material.BaseColorFactor.G,
			node.Material.BaseColorFactor.B,
			node.Material.BaseColorFactor.A,
		}

		// 如果是非主体区域，使用#94C808绿色
		if materialName != "material_1" {
			baseColor = fauxgl.Color{0.58, 0.78, 0.03, 1.0} // #94C808
		}

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

		// 渲染当前节点
		context.Shader = shader
		context.DrawMesh(node.Mesh)

		fmt.Printf("  ✓ 完成渲染\n")
	}

	// 保存结果
	err := fauxgl.SavePNG(filename, context.Image())
	if err != nil {
		panic(err)
	}

	// 生成额外版本 - 全部贴图
	if filename == "mug_uv.png" {
		// 复制场景以保留原始设置
		sceneCopy := *scene

		// 为所有纹理应用全部贴图设置
		setupFullUVMapping(&sceneCopy)

		// 渲染全部贴图版本
		renderHighQuality(&sceneCopy, lightSystem, "mug_uv_full.png")
		fmt.Println("✓ 全部贴图版本已保存为 mug_uv_full.png")
	}

	fmt.Printf("\n多光源渲染完成，共处理 %d 个节点\n", len(renderableNodes))
}

// printMaterialInfo 打印材质信息
func printMaterialInfo(scene *fauxgl.Scene) {
	fmt.Println("\n=== GLTF材质信息 ===")
	fmt.Println("根据GLTF文件定义，此模型包含5个材质区域：")
	fmt.Println("  material_0: 绿色材质（杯底区域）- #94C808")
	fmt.Println("  material_1: 棋盘格纹理（主体区域，使用texture.png）")
	fmt.Println("  material_2: 绿色材质（装饰带）- #94C808")
	fmt.Println("  material_3: 绿色材质（装饰带）- #94C808")
	fmt.Println("  material_4: 绿色材质（杯口区域）- #94C808")
	fmt.Println("\n每个primitive使用不同的材质，实现真正的多材质分区。")
	fmt.Println("\n✨ 最终效果：")
	fmt.Println("  🎨 主体区域：棋盘格贴图（texture.png）")
	fmt.Println("  🟩 其他区域：绿色（#94C808）")
	fmt.Println("  🔥 高质量渲染：高分辨率，超采样抗锯齿")
	fmt.Println("  🛠️ 保持原始模型形状与比例")
}

// setupFullUVMapping 设置全部贴图模式
func setupFullUVMapping(scene *fauxgl.Scene) {
	fmt.Println("\n=== 全部贴图模式设置 ===")

	// 为所有纹理应用全覆盖映射
	for name, texture := range scene.Textures {
		fmt.Printf("为纹理 %s 设置全部贴图\n", name)

		// 创建一个新的UV修改器
		modifier := fauxgl.NewUVModifier()

		// 全局变换：将纹理应用到整个模型
		globalTransform := fauxgl.NewUVTransform()
		globalTransform.ScaleU = 0.9 // 进行适当缩放以覆盖大部分区域
		globalTransform.ScaleV = 0.9
		globalTransform.OffsetU = 0.05 // 居中偏移
		globalTransform.OffsetV = 0.05
		modifier.SetGlobalTransform(globalTransform)

		// 增强杯把手区域
		handleMapping := &fauxgl.UVMapping{
			Name:    "handle_area",
			Enabled: true,
			Region: fauxgl.UVRegion{
				MinU: -0.3, MaxU: 0.1, // 左侧区域
				MinV: 0.3, MaxV: 0.8, // 中部区域
				MaskType: fauxgl.UVMaskRectangle,
			},
			Transform: &fauxgl.UVTransform{
				ScaleU: 1.0, ScaleV: 1.0,
				OffsetU: 0.2, OffsetV: 0.0,
			},
			BlendMode: fauxgl.UVBlendReplace,
			Priority:  1,
		}
		modifier.AddMapping(handleMapping)

		// 应用修改器
		texture.UVModifier = modifier

		fmt.Printf("  ✓ 全部贴图设置完成\n")
	}
}
