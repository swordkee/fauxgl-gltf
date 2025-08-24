package main

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/swordkee/fauxgl-gltf"
)

const (
	// 使用texture.png作为主体区域纹理
	MAIN_TEXTURE_FILE = "texture.png"

	// 渲染参数
	scale  = 4    // 4x 超采样抗锯齿
	width  = 2000 // 高分辨率渲染
	height = 2000 // 高分辨率渲染
	fovy   = 30   // 垂直视野角度
	near   = 1    // 近裁剪面
	far    = 20   // 远裁剪面
)

var (
	// 相机参数
	eye    = fauxgl.V(2.5, 1.5, 2.5) // 相机位置
	center = fauxgl.V(0, 0.6, 0)     // 对准杯子中心
	up     = fauxgl.V(0, 1, 0)       // 上方向向量
)

func main() {
	fmt.Println("=== 简化版GLTF渲染 ===")
	fmt.Printf("📝 主体区域纹理文件: %s\n", MAIN_TEXTURE_FILE)
	fmt.Println("")

	// 设置并行处理
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("使用 %d 个CPU核心进行并行渲染\n", runtime.NumCPU())

	// 获取当前工作目录
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	gltfPath := filepath.Join(dir, "./gltf/mug.gltf")
	texturePath := filepath.Join(dir, MAIN_TEXTURE_FILE)

	fmt.Printf("GLTF文件路径: %s\n", gltfPath)
	fmt.Printf("纹理文件路径: %s\n", texturePath)

	// 使用GLTF场景加载器
	scene, err := fauxgl.LoadGLTFScene(gltfPath)
	if err != nil {
		panic(err)
	}

	// 加载并应用纹理
	applyTextures(scene, texturePath)

	fmt.Printf("场景加载成功:\n")
	fmt.Printf("  材质数量: %d\n", len(scene.Materials))
	fmt.Printf("  纹理数量: %d\n", len(scene.Textures))

	// 设置光源系统
	lightSystem := setupLightingSystem()

	// 执行渲染
	renderScene(scene, lightSystem, "mug_simple.png")

	fmt.Println("\n=== 渲染完成 ===")
	fmt.Println("渲染结果已保存为 mug_simple.png")
	printMaterialInfo()
}

// applyTextures 应用纹理到场景
func applyTextures(scene *fauxgl.Scene, texturePath string) {
	fmt.Println("\n=== 应用纹理 ===")

	// 加载主体区域纹理
	mainTexture, err := fauxgl.LoadAdvancedTexture(texturePath, fauxgl.BaseColorTexture)
	if err != nil {
		fmt.Printf(" ✗ 无法加载纹理 %s: %v\n", texturePath, err)
		return
	}
	fmt.Printf("✓ 纹理加载成功 (%dx%d)\n", mainTexture.Width, mainTexture.Height)

	// 为每个材质配置适当的纹理和颜色
	greenColor := fauxgl.Color{0.58, 0.78, 0.03, 1.0} // #94C808 绿色

	// 设置材质
	for name, material := range scene.Materials {
		if name == "material_1" {
			// 主体区域应用纹理
			material.BaseColorTexture = mainTexture
			material.BaseColorFactor = fauxgl.Color{1.0, 1.0, 1.0, 1.0} // 白色基础色
			fmt.Printf("✓ 已为主体区域(%s)设置纹理\n", name)
		} else {
			// 其他区域设置为绿色
			material.BaseColorTexture = nil // 移除纹理
			material.BaseColorFactor = greenColor
			fmt.Printf("✓ 已为区域(%s)设置绿色#94C808\n", name)
		}
	}

	// 添加纹理到场景
	scene.Textures["main_texture"] = mainTexture
	fmt.Println("✓ 纹理应用完成!")
}

// setupLightingSystem 设置光源系统
func setupLightingSystem() *fauxgl.Vector {
	fmt.Println("\n=== 设置光源系统 ===")

	// 使用简单的主光源
	light := fauxgl.V(-0.5, -0.6, -0.6).Normalize()
	fmt.Printf("主光源方向: %v\n", light)

	return &light
}

// renderScene 渲染场景
func renderScene(scene *fauxgl.Scene, light *fauxgl.Vector, filename string) {
	fmt.Println("\n=== 渲染场景 ===")

	// 创建渲染上下文
	context := fauxgl.NewContext(width*scale, height*scale)
	context.ClearColor = fauxgl.Color{0.05, 0.05, 0.05, 1.0} // 深色背景
	context.ClearColorBuffer()
	context.ClearDepthBuffer()

	// 设置相机矩阵
	aspect := float64(width) / float64(height)
	matrix := fauxgl.LookAt(eye, center, up).Perspective(fovy, aspect, near, far)

	// 获取所有可渲染节点
	renderableNodes := scene.RootNode.GetRenderableNodes()
	fmt.Printf("开始渲染，共 %d 个节点...\n", len(renderableNodes))

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

		fmt.Printf("渲染节点 %d: %s (材质: %s)\n", i+1, node.Name, materialName)

		// 创建着色器
		shader := fauxgl.NewPhongShader(matrix, *light, eye)

		// 根据材质类型调整参数
		if materialName == "material_1" {
			// 主体区域使用纹理
			shader.DiffuseColor = fauxgl.Color{1.0, 1.0, 1.0, 1.0}
			shader.SpecularColor = fauxgl.Color{0.8, 0.8, 0.8, 1.0}
			shader.SpecularPower = 64
			if node.Material.BaseColorTexture != nil {
				shader.Texture = node.Material.BaseColorTexture
			}
		} else {
			// 其他区域使用绿色
			shader.DiffuseColor = fauxgl.Color{0.58, 0.78, 0.03, 1.0} // #94C808
			shader.SpecularColor = fauxgl.Color{0.6, 0.7, 0.3, 1.0}
			shader.SpecularPower = 48
		}

		// 设置材质颜色
		baseColor := fauxgl.Color{
			node.Material.BaseColorFactor.R,
			node.Material.BaseColorFactor.G,
			node.Material.BaseColorFactor.B,
			node.Material.BaseColorFactor.A,
		}
		shader.ObjectColor = baseColor

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

	fmt.Printf("\n渲染完成，共处理 %d 个节点\n", len(renderableNodes))
}

// printMaterialInfo 打印材质信息
func printMaterialInfo() {
	fmt.Println("\n=== 材质信息 ===")
	fmt.Println("material_0: 绿色材质（杯底区域）- #94C808")
	fmt.Println("material_1: 纹理材质（主体区域，使用texture.png）")
	fmt.Println("material_2: 绿色材质（装饰带）- #94C808")
	fmt.Println("material_3: 绿色材质（装饰带）- #94C808")
	fmt.Println("material_4: 绿色材质（杯口区域）- #94C808")
	fmt.Println("\n✨ 最终效果：")
	fmt.Println("  🎨 主体区域：texture.png贴图")
	fmt.Println("  🟩 其他区域：#94C808绿色")
	fmt.Println("  🛠️ 保持原生GLTF效果")
}
