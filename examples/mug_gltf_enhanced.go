package main

import (
	"fmt"
	"github.com/swordkee/fauxgl-gltf"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
)

const (
	scale  = 1    // optional supersampling
	width  = 2000 // output width in pixels
	height = 2000 // output height in pixels
	fovy   = 20   // vertical field of view in degrees
	near   = 1    // near clipping plane
	far    = 10   // far clipping plane
)

var (
	eye    = fauxgl.V(-3, -3, 3)    // 相机位置
	center = fauxgl.V(0.07, 0, 0)   // 目标焦点位置
	up     = fauxgl.V(0, -0.004, 0) // 上方向向量
)

// UV区域定义
type UVRegion struct {
	Name        string
	MinU, MaxU  float64
	MinV, MaxV  float64
	Color       fauxgl.Color
	TexturePath string
	Enabled     bool
}

// 杯子UV区域配置
var mugUVRegions = []UVRegion{
	{
		Name: "02 - Default",
		MinU: 0.2, MaxU: 0.8,
		MinV: 0.3, MaxV: 0.7,
		Color:       fauxgl.Color{0.9, 0.9, 0.9, 1.0}, // 白色主体
		TexturePath: "examples/texture.png",
		Enabled:     true,
	},
	{
		Name: "手柄区域",
		MinU: 0.0, MaxU: 0.2,
		MinV: 0.4, MaxV: 0.6,
		Color:       fauxgl.Color{0.8, 0.6, 0.4, 1.0}, // 棕色手柄
		TexturePath: "",
		Enabled:     true,
	},
	{
		Name: "底部区域",
		MinU: 0.3, MaxU: 0.7,
		MinV: 0.0, MaxV: 0.3,
		Color:       fauxgl.Color{0.7, 0.7, 0.8, 1.0}, // 淡蓝色底部
		TexturePath: "",
		Enabled:     true,
	},
	{
		Name: "顶部边缘",
		MinU: 0.2, MaxU: 0.8,
		MinV: 0.7, MaxV: 1.0,
		Color:       fauxgl.Color{0.9, 0.8, 0.6, 1.0}, // 金色边缘
		TexturePath: "",
		Enabled:     true,
	},
}

// createCustomTexture 创建自定义纹理，支持UV区域映射
func createCustomTexture(regions []UVRegion, textureSize int) *fauxgl.AdvancedTexture {
	// 创建空白的RGBA图像
	img := image.NewRGBA(image.Rect(0, 0, textureSize, textureSize))

	// 填充背景色为白色
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{255, 255, 255, 255}}, image.Point{}, draw.Src)

	// 遍历每个UV区域
	for _, region := range regions {
		if !region.Enabled {
			continue
		}

		// 计算纹理中的像素位置
		startX := int(region.MinU * float64(textureSize))
		endX := int(region.MaxU * float64(textureSize))
		startY := int(region.MinV * float64(textureSize))
		endY := int(region.MaxV * float64(textureSize))

		// 如果有纹理路径，尝试加载纹理
		if region.TexturePath != "" {
			if customImg, err := fauxgl.LoadImage(region.TexturePath); err == nil {
				// 缩放并粘贴自定义纹理
				scaledImg := resizeImage(customImg, endX-startX, endY-startY)
				pasteImage(img, scaledImg, startX, startY)
				continue
			}
		}

		// 使用纯色填充区域
		c := color.RGBA{
			uint8(region.Color.R * 255),
			uint8(region.Color.G * 255),
			uint8(region.Color.B * 255),
			uint8(region.Color.A * 255),
		}

		for y := startY; y < endY; y++ {
			for x := startX; x < endX; x++ {
				if x >= 0 && x < textureSize && y >= 0 && y < textureSize {
					img.Set(x, y, c)
				}
			}
		}
	}

	// 创建高级纹理
	return fauxgl.NewAdvancedTexture(img, fauxgl.BaseColorTexture)
}

// resizeImage 简单的图像缩放函数
func resizeImage(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	srcBounds := src.Bounds()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// 简单的最近邻采样
			srcX := int(float64(x) / float64(width) * float64(srcBounds.Dx()))
			srcY := int(float64(y) / float64(height) * float64(srcBounds.Dy()))
			dst.Set(x, y, src.At(srcX+srcBounds.Min.X, srcY+srcBounds.Min.Y))
		}
	}

	return dst
}

// pasteImage 将一个图像粘贴到另一个图像上
func pasteImage(dst *image.RGBA, src image.Image, startX, startY int) {
	srcBounds := src.Bounds()
	for y := 0; y < srcBounds.Dy(); y++ {
		for x := 0; x < srcBounds.Dx(); x++ {
			dstX := startX + x
			dstY := startY + y
			if dstX >= 0 && dstX < dst.Bounds().Dx() && dstY >= 0 && dstY < dst.Bounds().Dy() {
				dst.Set(dstX, dstY, src.At(x+srcBounds.Min.X, y+srcBounds.Min.Y))
			}
		}
	}
}

// createCustomLogo 创建一个简单的logo纹理作为示例
func createCustomLogo(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// 创建一个渐变圆形logo
	centerX := float64(size) / 2
	centerY := float64(size) / 2
	radius := float64(size) / 3

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist <= radius {
				// 在圆形内，创建渐变效果
				alpha := 1.0 - (dist / radius)
				intensity := uint8(alpha * 255)

				// 创建蓝色到绿色的渐变
				r := uint8((1.0 - alpha) * 100)
				g := uint8(alpha*200 + 55)
				b := uint8(alpha*150 + 105)

				img.Set(x, y, color.RGBA{r, g, b, intensity})
			} else {
				// 在圆形外，使用透明背景
				img.Set(x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	return img
}

// saveCustomLogo 保存自定义logo作为示例
func saveCustomLogo() {
	logo := createCustomLogo(256)

	file, err := os.Create("examples/custom_logo.png")
	if err != nil {
		fmt.Printf("Warning: Could not create custom logo: %v\n", err)
		return
	}
	defer file.Close()

	err = png.Encode(file, logo)
	if err != nil {
		fmt.Printf("Warning: Could not save custom logo: %v\n", err)
	} else {
		fmt.Println("Custom logo saved as examples/custom_logo.png")
	}
}

func main() {
	fmt.Println("Loading mug.gltf with enhanced UV mapping and texture support...")

	// 创建自定义logo作为示例
	// saveCustomLogo()

	// 使用增强的GLTF场景加载器
	scene, err := fauxgl.LoadGLTFScene("examples/mug.gltf")
	if err != nil {
		fmt.Printf("Failed to load GLTF scene: %v\n", err)
		// 回退到传统的网格加载方式
		loadTraditionalMesh()
		return
	}

	// 打印场景信息
	printSceneInfo(scene)

	// 应用与原始mug.go相同的变换到所有网格
	applyMugTransforms(scene)

	// 创建和应用自定义UV贴图
	applyCustomUVMapping(scene)

	// 渲染多个版本：原始、自定义UV、不同配色方案
	renderMultipleVersions(scene)
}

func printSceneInfo(scene *fauxgl.Scene) {
	fmt.Printf("=== GLTF Scene Analysis ===\n")
	fmt.Printf("Scene Name: %s\n", scene.Name)
	fmt.Printf("Materials: %d\n", len(scene.Materials))
	fmt.Printf("Textures: %d\n", len(scene.Textures))
	fmt.Printf("Meshes: %d\n", len(scene.Meshes))
	fmt.Printf("Cameras: %d\n", len(scene.Cameras))
	fmt.Printf("Lights: %d\n", len(scene.Lights))

	// 分析材质
	fmt.Printf("\nMaterials Analysis:\n")
	for name, material := range scene.Materials {
		fmt.Printf("  - %s:\n", name)
		fmt.Printf("    Base Color: %v\n", material.BaseColorFactor)
		fmt.Printf("    Metallic: %.2f\n", material.MetallicFactor)
		fmt.Printf("    Roughness: %.2f\n", material.RoughnessFactor)
		fmt.Printf("    Double Sided: %t\n", material.DoubleSided)

		if material.BaseColorTexture != nil {
			fmt.Printf("    Has Base Color Texture: Yes\n")
		}
		if material.EmissiveFactor.R > 0 || material.EmissiveFactor.G > 0 || material.EmissiveFactor.B > 0 {
			fmt.Printf("    Emissive: %v\n", material.EmissiveFactor)
		}
	}

	// 分析场景节点
	renderables := scene.RootNode.GetRenderableNodes()
	fmt.Printf("\nRenderable Nodes: %d\n", len(renderables))

	bounds := scene.GetBounds()
	fmt.Printf("Scene Bounds: min=%v, max=%v\n", bounds.Min, bounds.Max)
	fmt.Printf("Scene Center: %v\n", bounds.Center())
	fmt.Printf("=========================\n\n")
}

// applyMugTransforms applies the same transforms as the original mug.go
func applyMugTransforms(scene *fauxgl.Scene) {
	fmt.Println("Applying original mug.go transforms...")

	// 遍历所有网格并应用与原始mug.go相同的变换
	for _, mesh := range scene.Meshes {
		// 获取原始边界
		bounds := mesh.BoundingBox()
		fmt.Printf("Original mesh bounds: min=%v, max=%v\n", bounds.Min, bounds.Max)

		// 应用BiUnitCube变换（与原始mug.go相同）
		mesh.BiUnitCube()
		mesh.SmoothNormalsThreshold(fauxgl.Radians(30))

		// 检查BiUnitCube后的边界
		bounds = mesh.BoundingBox()
		fmt.Printf("After BiUnitCube: min=%v, max=%v\n", bounds.Min, bounds.Max)

		// 应用变换（与原始mug.go相同）
		positionMatrix := fauxgl.Translate(fauxgl.V(0.5, 0.3, 0))
		mesh.Transform(positionMatrix)

		// 检查变换后的边界
		bounds = mesh.BoundingBox()
		fmt.Printf("After transform: min=%v, max=%v\n", bounds.Min, bounds.Max)
		fmt.Printf("Final center: %v\n", bounds.Center())
	}
}

// setupOriginalCamera uses the same camera setup as original mug.go
func setupOriginalCamera(scene *fauxgl.Scene, width, height int) {
	// 使用与原始mug.go完全相同的相机参数
	camera := fauxgl.NewPerspectiveCamera(
		"original_camera",
		eye,    // V(-3, -3, 3)
		center, // V(0.07, 0, 0)
		up,     // V(0, -0.004, 0)
		fauxgl.Radians(float64(fovy)),
		float64(width)/float64(height),
		near, far,
	)

	scene.AddCamera(camera)
	fmt.Printf("Set original camera: eye=%v, center=%v, up=%v\n", eye, center, up)
}

func setupCamera(scene *fauxgl.Scene, width, height int) {
	if scene.ActiveCamera == nil {
		bounds := scene.GetBounds()
		center := bounds.Center()
		size := bounds.Size()

		// 根据场景大小调整相机位置
		distance := size.MaxComponent() * 2.5

		camera := fauxgl.NewPerspectiveCamera(
			"mug_camera",
			center.Add(fauxgl.V(distance*0.8, distance*0.6, distance*0.8)),
			center,
			fauxgl.V(0, 1, 0),
			fauxgl.Radians(float64(fovy)),
			float64(width)/float64(height),
			near, far,
		)

		scene.AddCamera(camera)
		fmt.Printf("Created camera at distance %.2f from scene center %v\n", distance, center)
	}
}

func setupLights(scene *fauxgl.Scene) {
	if len(scene.Lights) == 0 {
		// 主光源 - 暖色调
		keyLight := fauxgl.Light{
			Type:      fauxgl.DirectionalLight,
			Direction: fauxgl.V(-0.5, -1, -0.5).Normalize(),
			Color:     fauxgl.Color{1.0, 0.95, 0.8, 1}, // 温暖的阳光
			Intensity: 2.5,
		}
		scene.AddLight(keyLight)

		// 填充光 - 冷色调
		fillLight := fauxgl.Light{
			Type:      fauxgl.DirectionalLight,
			Direction: fauxgl.V(0.5, 0.2, 0.8).Normalize(),
			Color:     fauxgl.Color{0.6, 0.7, 1.0, 1}, // 冷色天空光
			Intensity: 0.8,
		}
		scene.AddLight(fillLight)

		// 边缘光
		rimLight := fauxgl.Light{
			Type:      fauxgl.DirectionalLight,
			Direction: fauxgl.V(0, 0.3, -1).Normalize(),
			Color:     fauxgl.Color{1.0, 0.9, 0.7, 1},
			Intensity: 1.5,
		}
		scene.AddLight(rimLight)

		fmt.Println("Added three-point lighting setup")
	}
}

func enhanceMaterials(scene *fauxgl.Scene) {
	fmt.Println("Enhancing materials based on GLTF specifications...")

	for name, material := range scene.Materials {
		// 根据原始mug.go的逻辑调整材质
		switch name {
		case "material_0": // "01 - Default" - 底部材质
			material.BaseColorFactor = fauxgl.Color{0.9, 0.9, 0.95, 1.0}
			material.MetallicFactor = 0.1
			material.RoughnessFactor = 0.7

		case "material_1": // "02 - Default" - 带纹理的主体
			// 保持原有纹理，调整金属度和粗糙度
			material.MetallicFactor = 0.0
			material.RoughnessFactor = 0.6

		case "material_2": // "03 - Default" - 绿色部分
			material.BaseColorFactor = fauxgl.Color{0.0, 0.8, 0.2, 1.0}
			material.MetallicFactor = 0.0
			material.RoughnessFactor = 0.8

		case "material_3": // "04 - Default" - 蓝紫色部分
			material.BaseColorFactor = fauxgl.Color{0.3, 0.0, 0.9, 1.0}
			material.MetallicFactor = 0.0
			material.RoughnessFactor = 0.7

		case "material_4": // "05 - Default" - 黄色部分
			material.BaseColorFactor = fauxgl.Color{1.0, 0.8, 0.0, 1.0}
			material.MetallicFactor = 0.0
			material.RoughnessFactor = 0.6
		}

		// 添加一些发光效果到特定材质
		if name == "material_2" || name == "material_3" || name == "material_4" {
			material.EmissiveFactor = fauxgl.Color{
				material.BaseColorFactor.R * 0.1,
				material.BaseColorFactor.G * 0.1,
				material.BaseColorFactor.B * 0.1,
				1.0,
			}
		}

		fmt.Printf("Enhanced material %s\n", name)
	}
}

func generateAnimationSequence(scene *fauxgl.Scene) {
	fmt.Println("Generating animation sequence...")

	// 创建360度旋转动画
	frames := 36 // 36帧，每帧10度

	for frame := 0; frame < frames; frame++ {
		angle := float64(frame) * 10.0 * math.Pi / 180.0 // 转换为弧度

		// 更新相机位置（围绕Y轴旋转）
		if scene.ActiveCamera != nil {
			bounds := scene.GetBounds()
			center := bounds.Center()
			distance := bounds.Size().MaxComponent() * 2.5

			x := center.X + math.Cos(angle)*distance*0.8
			z := center.Z + math.Sin(angle)*distance*0.8
			y := center.Y + distance*0.6

			scene.ActiveCamera.Position = fauxgl.V(x, y, z)
			scene.ActiveCamera.Target = center
		}

		// 渲染当前帧
		context := fauxgl.NewContext(width, height)
		context.ClearColor = fauxgl.Color{0.95, 0.95, 0.98, 1.0} // 淡蓝色背景
		context.ClearColorBuffer()
		context.ClearDepthBuffer()

		renderer := fauxgl.NewSceneRenderer(context)
		renderer.RenderScene(scene)

		// 保存帧
		filename := fmt.Sprintf("mug_animation_frame_%02d.png", frame)
		err := fauxgl.SavePNG(filename, context.Image())
		if err != nil {
			fmt.Printf("Failed to save frame %d: %v\n", frame, err)
		}

		if frame%6 == 0 {
			fmt.Printf("Generated frame %d/%d\n", frame+1, frames)
		}
	}

	fmt.Printf("Animation sequence completed! Generated %d frames.\n", frames)
	fmt.Println("You can use tools like ffmpeg to create a video from these frames:")
	fmt.Println("ffmpeg -r 6 -i mug_animation_frame_%02d.png -vcodec libx264 -crf 25 -pix_fmt yuv420p mug_rotation.mp4")
}

func loadTraditionalMesh() {
	fmt.Println("Using traditional mesh loading as fallback...")

	// 传统的网格加载方式（原始mug.go的逻辑）
	mesh, err := fauxgl.LoadGLTF("examples/mug.gltf")
	if err != nil {
		panic(err)
	}

	// 加载纹理
	texture, err := fauxgl.LoadTexture("examples/texture.png")
	if err != nil {
		fmt.Printf("Warning: Could not load texture: %v\n", err)
	}

	mesh.BiUnitCube()
	mesh.SmoothNormalsThreshold(fauxgl.Radians(30))

	// 应用变换
	positionMatrix := fauxgl.Translate(fauxgl.V(0.5, 0.3, 0))
	mesh.Transform(positionMatrix)

	context := fauxgl.NewContext(width*scale, height*scale)
	context.ClearColor = fauxgl.White
	context.ClearColorBuffer()

	aspect := float64(width) / float64(height)
	matrix := fauxgl.LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	light := fauxgl.V(0, 0, 1).Normalize()

	shader := fauxgl.NewPhongShader(matrix, light, eye)
	if texture != nil {
		shader.Texture = texture
	}
	shader.ObjectColor = fauxgl.HexColor("FFFF9D").Alpha(0.65)
	context.Shader = shader
	context.DrawMesh(mesh)

	fauxgl.SavePNG("mug_traditional.png", context.Image())
	fmt.Println("Traditional mug rendering saved as mug_traditional.png")
}

// applyCustomUVMapping 应用自定义UV贴图到场景
func applyCustomUVMapping(scene *fauxgl.Scene) {
	fmt.Println("Applying custom UV mapping...")

	// 创建自定义纹理
	customTexture := createCustomTexture(mugUVRegions, 1024)

	// 保存自定义纹理作为参考
	file, err := os.Create("mug_custom_texture.png")
	if err == nil {
		png.Encode(file, customTexture.Image)
		file.Close()
		fmt.Println("Custom texture saved as mug_custom_texture.png")
	}

	// 将自定义纹理添加到场景
	scene.AddTexture("custom_uv_texture", customTexture)

	// 创建使用自定义纹理的材质
	customMaterial := fauxgl.NewPBRMaterial()
	customMaterial.BaseColorTexture = customTexture
	customMaterial.MetallicFactor = 0.1
	customMaterial.RoughnessFactor = 0.6
	customMaterial.DoubleSided = true

	scene.AddMaterial("custom_uv_material", customMaterial)

	fmt.Printf("Applied custom UV mapping with %d regions\n", len(mugUVRegions))
}

// renderMultipleVersions 渲染多个版本的杯子
func renderMultipleVersions(scene *fauxgl.Scene) {
	fmt.Println("Rendering multiple mug versions...")

	// 版本1：原始GLTF材质
	renderVersion(scene, "original", "mug_original.png", func(s *fauxgl.Scene) {
		setupOriginalCamera(s, width, height)
		setupLights(s)
		enhanceMaterials(s)
	})

	// 版本2：自定义UV贴图
	renderVersion(scene, "custom_uv", "mug_custom_uv.png", func(s *fauxgl.Scene) {
		setupOriginalCamera(s, width, height)
		setupLights(s)
		applyCustomMaterials(s)
	})

	// 版本3：蓝色主题
	renderVersion(scene, "blue_theme", "mug_blue_theme.png", func(s *fauxgl.Scene) {
		setupOriginalCamera(s, width, height)
		setupLights(s)
		applyBlueTheme(s)
	})

	// 版本4：金属主题
	renderVersion(scene, "metal_theme", "mug_metal_theme.png", func(s *fauxgl.Scene) {
		setupOriginalCamera(s, width, height)
		setupLights(s)
		applyMetalTheme(s)
	})

	// 版本5：彩虹主题
	renderVersion(scene, "rainbow_theme", "mug_rainbow_theme.png", func(s *fauxgl.Scene) {
		setupOriginalCamera(s, width, height)
		setupLights(s)
		applyRainbowTheme(s)
	})

	fmt.Println("All mug versions rendered successfully!")
}

// renderVersion 渲染单个版本
func renderVersion(scene *fauxgl.Scene, name, filename string, setupFunc func(*fauxgl.Scene)) {
	fmt.Printf("Rendering %s version...\n", name)

	// 设置渲染上下文
	context := fauxgl.NewContext(width*scale, height*scale)
	context.ClearColor = fauxgl.Color{0.95, 0.95, 0.98, 1.0} // 淡蓝灰色背景
	context.ClearColorBuffer()
	context.ClearDepthBuffer()

	// 应用特定的设置
	setupFunc(scene)

	// 渲染场景
	renderer := fauxgl.NewSceneRenderer(context)
	renderer.RenderScene(scene)

	// 保存图像
	err := fauxgl.SavePNG(filename, context.Image())
	if err != nil {
		fmt.Printf("Failed to save %s: %v\n", filename, err)
	} else {
		fmt.Printf("%s saved as %s\n", name, filename)
	}
}

// applyCustomMaterials 应用自定义材质
func applyCustomMaterials(scene *fauxgl.Scene) {
	// 为主要材质应用自定义纹理
	if customMaterial := scene.GetMaterial("custom_uv_material"); customMaterial != nil {
		for name := range scene.Materials {
			if name == "material_1" { // 主体材质
				scene.Materials[name] = customMaterial
				break
			}
		}
	}
}

// applyBlueTheme 应用蓝色主题
func applyBlueTheme(scene *fauxgl.Scene) {
	for name, material := range scene.Materials {
		switch name {
		case "material_0":
			material.BaseColorFactor = fauxgl.Color{0.2, 0.4, 0.8, 1.0} // 深蓝色
		case "material_1":
			material.BaseColorFactor = fauxgl.Color{0.4, 0.6, 0.9, 1.0} // 亮蓝色
		case "material_2":
			material.BaseColorFactor = fauxgl.Color{0.1, 0.3, 0.7, 1.0} // 深海蓝
		case "material_3":
			material.BaseColorFactor = fauxgl.Color{0.6, 0.8, 1.0, 1.0} // 天空蓝
		case "material_4":
			material.BaseColorFactor = fauxgl.Color{0.0, 0.2, 0.6, 1.0} // 午夜蓝
		}
		material.MetallicFactor = 0.2
		material.RoughnessFactor = 0.4
	}
}

// applyMetalTheme 应用金属主题
func applyMetalTheme(scene *fauxgl.Scene) {
	for name, material := range scene.Materials {
		switch name {
		case "material_0":
			material.BaseColorFactor = fauxgl.Color{0.8, 0.8, 0.9, 1.0} // 银色
			material.MetallicFactor = 0.9
		case "material_1":
			material.BaseColorFactor = fauxgl.Color{0.9, 0.8, 0.6, 1.0} // 金色
			material.MetallicFactor = 0.8
		case "material_2":
			material.BaseColorFactor = fauxgl.Color{0.7, 0.5, 0.3, 1.0} // 铜色
			material.MetallicFactor = 0.7
		case "material_3":
			material.BaseColorFactor = fauxgl.Color{0.6, 0.6, 0.7, 1.0} // 钢色
			material.MetallicFactor = 0.9
		case "material_4":
			material.BaseColorFactor = fauxgl.Color{0.9, 0.9, 0.8, 1.0} // 铂金色
			material.MetallicFactor = 0.95
		}
		material.RoughnessFactor = 0.2 // 金属表面较光滑
	}
}

// applyRainbowTheme 应用彩虹主题
func applyRainbowTheme(scene *fauxgl.Scene) {
	colors := []fauxgl.Color{
		{1.0, 0.0, 0.0, 1.0}, // 红色
		{1.0, 0.5, 0.0, 1.0}, // 橙色
		{1.0, 1.0, 0.0, 1.0}, // 黄色
		{0.0, 1.0, 0.0, 1.0}, // 绿色
		{0.0, 0.0, 1.0, 1.0}, // 蓝色
	}

	i := 0
	for _, material := range scene.Materials {
		material.BaseColorFactor = colors[i%len(colors)]
		material.MetallicFactor = 0.1
		material.RoughnessFactor = 0.7
		// 添加轻微发光效果
		material.EmissiveFactor = fauxgl.Color{
			colors[i%len(colors)].R * 0.1,
			colors[i%len(colors)].G * 0.1,
			colors[i%len(colors)].B * 0.1,
			1.0,
		}
		i++
	}
}
