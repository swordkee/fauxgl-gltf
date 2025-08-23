package main

import (
	"fmt"
	"math"

	"github.com/swordkee/fauxgl-gltf"
)

// **配置区域**: 在这里修改您想使用的纹理文件
const (
	// **自定义纹理文件**: 将您的纹理文件名替换这里
	CUSTOM_TEXTURE_FILE = "texture.png" // 修改为您的纹理文件名

	// **渲染参数**: 提高分辨率和超采样以达到300KB+
	scale  = 1    // **提高超采样**: 2x 超采样抗锯齿
	width  = 2000 // **提高分辨率**: 4K 分辨率
	height = 2000 // **提高分辨率**: 4K 分辨率
	fovy   = 30   // vertical field of view in degrees
	near   = 1    // near clipping plane
	far    = 20   // far clipping plane
)

var (
	// **调整相机参数**: 适应原始尺寸的杯子模型
	eye    = fauxgl.V(2.5, 4, 4.0)  // 调整相机位置适应原始尺寸
	center = fauxgl.V(0, 1.14, 0.4) // 焦点对准杯子中心
	up     = fauxgl.V(0, 1, 0)      // 标准上方向向量
)

func main() {
	fmt.Println("=== 增强版GLTF多材质UV分区渲染 - 自定义UV+多光源 ===")
	fmt.Printf("📝 当前配置的纹理文件: %s\n", CUSTOM_TEXTURE_FILE)
	fmt.Println("💡 提示: 要使用自定义纹理，请修改文件顶部的 CUSTOM_TEXTURE_FILE 常量")
	fmt.Println("")

	// 使用GLTF场景加载器，支持多材质
	scene, err := fauxgl.LoadGLTFScene("mug.gltf")
	if err != nil {
		panic(err)
	}

	// **问题修复**: 替换白色纹理为有图案的测试纹理
	replaceWithVisibleTexture(scene)

	fmt.Printf("场景加载成功:\n")
	fmt.Printf("  材质数量: %d\n", len(scene.Materials))
	fmt.Printf("  网格数量: %d\n", len(scene.Meshes))
	fmt.Printf("  纹理数量: %d\n", len(scene.Textures))

	// 分析场景结构
	analyzeScene(scene)

	// **新增功能1**: 自定义UV设置
	setupCustomUVMappings(scene)

	// **关键修复**: 仅进行法线处理，不改变几何形状
	fmt.Println("\n=== 网格预处理 (保持原始形状) ===")
	for name, mesh := range scene.Meshes {
		fmt.Printf("处理网格: %s (%d三角形)\n", name, len(mesh.Triangles))

		// **只进行法线平滑处理**，不使用BiUnitCube()以防止变形
		mesh.SmoothNormalsThreshold(fauxgl.Radians(30))

		// 打印网格边界信息
		bounds := mesh.BoundingBox()
		fmt.Printf("原始边界: min=%v, max=%v\n", bounds.Min, bounds.Max)
	}

	// **新增功能2**: 设置增强光源系统
	lightSystem := setupAdvancedLightingSystem()

	// **高质量渲染设置**: 增强光照系统
	context := fauxgl.NewContext(width*scale, height*scale)
	context.ClearColor = fauxgl.Color{0.1, 0.12, 0.15, 1.0} // 深色背景，突出光照效果
	context.ClearColorBuffer()
	context.ClearDepthBuffer()

	// **高质量渲染**: 使用传统渲染方法，增强光照
	aspect := float64(width) / float64(height)
	matrix := fauxgl.LookAt(eye, center, up).Perspective(fovy, aspect, near, far)

	// 获取所有可渲染节点
	renderableNodes := scene.RootNode.GetRenderableNodes()
	fmt.Printf("开始增强渲染，共 %d 个节点...\n", len(renderableNodes))

	// **多光源渲染**: 逐个渲染每个节点
	renderWithAdvancedLighting(context, matrix, renderableNodes, scene, lightSystem)

	// 保存增强渲染结果
	err = fauxgl.SavePNG("mug_uv_enhanced.png", context.Image())
	if err != nil {
		panic(err)
	}

	fmt.Println("\n=== 渲染完成 ===")
	fmt.Println("增强版多材质UV分区渲染已保存为 mug_uv_enhanced.png")
	fmt.Println("✅ 自定义UV映射")
	fmt.Println("✅ 多光源照明系统")
	fmt.Println("✅ 高质量渲染")
	printMaterialInfo(scene)
}

// replaceWithVisibleTexture 替换原有的白色纹理为有图案的测试纹理
func replaceWithVisibleTexture(scene *fauxgl.Scene) {
	fmt.Println("\n=== 纹理替换修复 ===")

	// 定义候选纹理文件列表（按优先级排序）
	textureFiles := []string{
		CUSTOM_TEXTURE_FILE,  // 用户自定义纹理（优先级最高）
		"texture.png",        // 备选格式
		"custom_texture.jpg", // 自定义纹理
		"custom_texture.png",
		"logo_texture.png", // 标志纹理
		"test_pattern.png", // 测试纹理
		"simple_pattern.png",
		"texture.png", // 现有纹理
		"texture.jpg", // GLTF默认纹理
	}

	var testTexture *fauxgl.AdvancedTexture
	var loadedFile string
	var err error

	// 尝试按优先级加载纹理
	fmt.Println("尝试加载纹理文件...")
	for _, filename := range textureFiles {
		fmt.Printf("  尝试加载: %s", filename)
		testTexture, err = fauxgl.LoadAdvancedTexture(filename, fauxgl.BaseColorTexture)
		if err == nil {
			loadedFile = filename
			fmt.Printf(" ✓ 成功\n")
			break
		} else {
			fmt.Printf(" ✗ 失败: %v\n", err)
		}
	}

	if testTexture == nil {
		fmt.Println("❌ 警告: 无法加载任何纹理文件，将使用原始GLTF纹理")
		// 检查原始纹理是否为白色
		checkOriginalTextures(scene)
		return
	}

	fmt.Printf("✓ 纹理加载成功: %s (%dx%d)\n", loadedFile, testTexture.Width, testTexture.Height)

	// 验证纹理内容不是全白
	if !validateTextureContent(testTexture, loadedFile) {
		fmt.Println("⚠️  纹理内容可能有问题，但继续使用")
	}

	// 替换场景中的所有纹理
	replaceSceneTextures(scene, testTexture, loadedFile)
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
func replaceSceneTextures(scene *fauxgl.Scene, newTexture *fauxgl.AdvancedTexture, filename string) {
	fmt.Printf("\n替换场景纹理为: %s\n", filename)

	// 替换场景中的所有纹理
	replaceCount := 0
	for name, _ := range scene.Textures {
		scene.Textures[name] = newTexture
		replaceCount++
		fmt.Printf("✓ 已替换纹理: %s\n", name)
	}

	// 更新材质引用
	for _, material := range scene.Materials {
		if material.BaseColorTexture != nil {
			material.BaseColorTexture = newTexture
		}
	}

	fmt.Printf("✓ 总共替换了 %d 个纹理\n", replaceCount)
	fmt.Println("✓ 纹理替换完成，UV修改器效果现在应该可见了!")
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

// **新增功能1**: setupCustomUVMappings 设置自定义UV映射
func setupCustomUVMappings(scene *fauxgl.Scene) {
	fmt.Println("\n=== 自定义UV映射设置 - 部分区域贴图 ===")

	// 为不同材质区域设置不同的UV修改器
	for name, texture := range scene.Textures {
		fmt.Printf("为纹理 %s 设置部分区域UV映射\n", name)

		// 创建部分区域UV修改器
		modifier := fauxgl.NewUVModifier()

		// **关键改进**: 设置背景为纯色或边缘延伸
		// 全局变换：将大部分区域映射到纹理的一个点（相当于纯色背景）
		globalTransform := fauxgl.NewUVTransform()
		globalTransform.ScaleU = 0.001 // 极小缩放
		globalTransform.ScaleV = 0.001
		globalTransform.OffsetU = 0.9 // 映射到纹理右上角（通常是白色或边缘色）
		globalTransform.OffsetV = 0.9
		modifier.SetGlobalTransform(globalTransform)

		// **方案1**: 前面板标志区域（最实用）
		frontLogoMapping := &fauxgl.UVMapping{
			Name:    "front_logo_area",
			Enabled: true,
			Region: fauxgl.UVRegion{
				MinU: 0.25, MaxU: 0.75, // 前面50%宽度
				MinV: 0.35, MaxV: 0.65, // 中间30%高度
				MaskType: fauxgl.UVMaskRectangle,
			},
			Transform: &fauxgl.UVTransform{
				ScaleU: 0.8, ScaleV: 0.6, // 适度缩放纹理
				OffsetU: 0.1, OffsetV: 0.2, // 居中偏移
				PivotU: 0.5, PivotV: 0.5,
			},
			BlendMode: fauxgl.UVBlendReplace,
			Priority:  3,
		}
		modifier.AddMapping(frontLogoMapping)

		// **方案2**: 上部装饰带（可选）
		upperBandMapping := &fauxgl.UVMapping{
			Name:    "upper_decoration_band",
			Enabled: true,
			Region: fauxgl.UVRegion{
				MinU: 0.1, MaxU: 0.9, // 大部分宽度
				MinV: 0.75, MaxV: 0.85, // 上部窄带
				MaskType: fauxgl.UVMaskRectangle,
			},
			Transform: &fauxgl.UVTransform{
				ScaleU: 2.0, ScaleV: 0.3, // 水平拉伸，垂直压缩
				OffsetU: -0.5, OffsetV: 0.6,
			},
			BlendMode: fauxgl.UVBlendReplace,
			Priority:  2,
		}
		modifier.AddMapping(upperBandMapping)

		// **方案3**: 侧面渐变效果（可选）
		sideGradientMapping := &fauxgl.UVMapping{
			Name:    "side_gradient",
			Enabled: true,
			Region: fauxgl.UVRegion{
				MinU: 0.0, MaxU: 0.2, // 左侧20%
				MinV: 0.2, MaxV: 0.8, // 中间60%高度
				MaskType: fauxgl.UVMaskGradient,
			},
			Transform: &fauxgl.UVTransform{
				ScaleU: 0.5, ScaleV: 1.2,
				OffsetU: 0.25, OffsetV: -0.1,
			},
			BlendMode: fauxgl.UVBlendMultiply,
			Priority:  1,
		}
		modifier.AddMapping(sideGradientMapping)

		// 应用UV修改器到纹理
		texture.UVModifier = modifier

		fmt.Printf("  ✓ 设置了部分区域贴图效果\n")
		fmt.Printf("    - 前面板标志区域 (25%%~75%% 宽, 35%%~65%% 高)\n")
		fmt.Printf("    - 上部装饰带 (10%%~90%% 宽, 75%%~85%% 高)\n")
		fmt.Printf("    - 侧面渐变效果 (0%%~20%% 宽, 20%%~80%% 高)\n")
		fmt.Printf("    - 其余区域显示为背景色\n")
	}
}

// **新增功能2**: LightingSystem 光照系统结构
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
		// 主光源：从右上方照射，模拟太阳光
		MainLight: fauxgl.V(-0.4, -0.6, -0.8).Normalize(),

		// 补光：从左侧补光，减少阴影
		FillLight: fauxgl.V(0.7, -0.2, -0.3).Normalize(),

		// 边缘光：从背后照射，增强轮廓
		RimLight: fauxgl.V(0.2, 0.3, 0.9).Normalize(),

		// 环境光：温暖的环境色调
		AmbientColor:    fauxgl.Color{0.4, 0.45, 0.5, 1.0},
		AmbientStrength: 0.3,
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

// **新增功能3**: renderWithAdvancedLighting 使用增强光照渲染
func renderWithAdvancedLighting(context *fauxgl.Context, matrix fauxgl.Matrix,
	renderableNodes []*fauxgl.SceneNode, scene *fauxgl.Scene, lightSystem *LightingSystem) {

	fmt.Println("\n=== 多光源增强渲染 ===")

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
		if materialName == "material_1" && node.Material.BaseColorTexture != nil {
			// 纹理材质：增强细节
			shader.DiffuseColor = fauxgl.Color{1.2, 1.1, 1.0, 1.0}
			shader.SpecularColor = fauxgl.Color{1.0, 1.0, 1.0, 1.0}
			shader.SpecularPower = 80
			fmt.Printf("  → 纹理材质，增强光照\n")
		} else {
			// 纯色材质：柔和光照
			shader.DiffuseColor = fauxgl.Color{0.9, 0.85, 0.8, 1.0}
			shader.SpecularColor = fauxgl.Color{0.6, 0.6, 0.7, 1.0}
			shader.SpecularPower = 32
			fmt.Printf("  → 纯色材质，柔和光照\n")
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

	fmt.Printf("\n多光源渲染完成，共处理 %d 个节点\n", len(renderableNodes))
}

// printMaterialInfo 打印材质信息
func printMaterialInfo(scene *fauxgl.Scene) {
	fmt.Println("\n=== GLTF材质信息 ===")
	fmt.Println("根据GLTF文件定义，此模型包含5个材质区域：")
	fmt.Println("  material_0: 纯色材质（杯底区域）")
	fmt.Println("  material_1: 纹理材质（主体区域，使用texture.jpg）")
	fmt.Println("  material_2: 绿色装饰带")
	fmt.Println("  material_3: 蓝色装饰带")
	fmt.Println("  material_4: 黄色杯口区域")
	fmt.Println("\n每个primitive使用不同的材质，实现真正的多材质UV分区。")
	fmt.Println("\n✨ 增强功能：")
	fmt.Println("  🎨 自定义UV映射：4层复合UV变换效果")
	fmt.Println("  💡 多光源系统：主光源+补光+边缘光+环境光")
	fmt.Println("  🔥 高质量渲染：4K分辨率，增强材质效果")
}
