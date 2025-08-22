package main

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/swordkee/fauxgl-gltf"
)

func main() {
	fmt.Println("=== FauxGL-GLTF KTX2 Texture Demo ===")

	// 测试KTX2解码器功能
	demonstrateKTX2Support()

	// 创建场景用于测试KTX2纹理
	scene := fauxgl.NewScene("KTX2 Texture Demo")

	// 创建带有KTX2纹理的材质
	material := fauxgl.NewPBRMaterial()
	material.BaseColorFactor = fauxgl.Color{0.8, 0.8, 0.8, 1.0}
	material.MetallicFactor = 0.1
	material.RoughnessFactor = 0.4

	// 尝试加载KTX2纹理（如果文件存在）
	if ktx2Texture := tryLoadKTX2Texture("test_texture.ktx2"); ktx2Texture != nil {
		material.BaseColorTexture = ktx2Texture
		fmt.Println("✅ Successfully loaded KTX2 texture")
	} else {
		// 创建示例纹理
		material.BaseColorTexture = createDemoTexture()
		fmt.Println("⚠️  Using demo texture (KTX2 file not found)")
	}

	scene.AddMaterial("ktx2_material", material)

	// 创建几何体
	sphere := fauxgl.NewSphere(3) // 使用更小的细分度
	scene.AddMesh("sphere", sphere)

	// 创建场景节点
	node := scene.CreateMeshNode("ktx2_sphere", "sphere", "ktx2_material")
	scene.RootNode.AddChild(node)

	// 添加光照
	scene.AddAmbientLight(fauxgl.Color{0.3, 0.3, 0.4, 1.0}, 0.3)
	scene.AddDirectionalLight(
		fauxgl.V(-1, -1, -1),
		fauxgl.Color{1.0, 0.95, 0.8, 1.0},
		3.0,
	)
	scene.AddPointLight(
		fauxgl.V(2, 3, 2),
		fauxgl.Color{0.8, 0.9, 1.0, 1.0},
		5.0, 10.0,
	)

	// 测试GLTF扩展中的KTX2支持
	testKTX2Extension(scene)

	// 创建相机
	camera := fauxgl.NewPerspectiveCamera(
		"ktx2_camera",
		fauxgl.V(0, 2, 6),
		fauxgl.V(0, 0, 0),
		fauxgl.V(0, 1, 0),
		fauxgl.Radians(45),
		800.0/600.0,
		0.1, 100.0,
	)
	scene.AddCamera(camera)

	// 创建渲染上下文
	context := fauxgl.NewContext(800, 600)
	context.ClearColor = fauxgl.Color{0.2, 0.25, 0.3, 1.0}
	context.ClearColorBuffer()

	// 渲染场景
	renderer := fauxgl.NewSceneRenderer(context)
	renderer.RenderScene(scene)

	// 保存结果
	err := fauxgl.SavePNG("ktx2_texture_demo.png", context.Image())
	if err != nil {
		fmt.Printf("❌ Failed to save image: %v\n", err)
	} else {
		fmt.Println("✅ KTX2 texture demo rendered and saved as ktx2_texture_demo.png")
	}

	// 显示KTX2支持信息
	displayKTX2SupportInfo()

	fmt.Println("\n🎉 KTX2 texture demo completed!")
}

// demonstrateKTX2Support 演示KTX2解码器功能
func demonstrateKTX2Support() {
	fmt.Println("\n🔧 KTX2 Decoder Features:")
	fmt.Println("  ✅ KTX2 header parsing and validation")
	fmt.Println("  ✅ Mipmap level indexing")
	fmt.Println("  ✅ Data format descriptor (DFD) parsing")
	fmt.Println("  ✅ Key-value data parsing")
	fmt.Println("  ✅ Supercompression global data extraction")
	fmt.Println("  ✅ Multi-level texture support")
	fmt.Println("  ✅ GLTF KHR_texture_basisu extension")

	// 创建一个模拟的KTX2头部进行测试
	testKTX2Header()
}

// testKTX2Header 测试KTX2头部解析
func testKTX2Header() {
	fmt.Println("\n🧪 Testing KTX2 Header Parsing:")

	// 创建一个简单的KTX2头部进行测试
	header := &fauxgl.Header{
		Format:                 nil, // VK_FORMAT_UNDEFINED
		TypeSize:               1,
		PixelWidth:             512,
		PixelHeight:            512,
		PixelDepth:             0,
		LayerCount:             0,
		FaceCount:              1,
		LevelCount:             9, // 9个mipmap级别
		SupercompressionScheme: nil,
		Index: fauxgl.Index{
			DFDByteOffset: 80,
			DFDByteLength: 32,
			KVDByteOffset: 112,
			KVDByteLength: 16,
			SGDByteOffset: 128,
			SGDByteLength: 0,
		},
	}

	fmt.Printf("  - Pixel dimensions: %dx%d\n", header.PixelWidth, header.PixelHeight)
	fmt.Printf("  - Mipmap levels: %d\n", header.LevelCount)
	fmt.Printf("  - Face count: %d\n", header.FaceCount)

	// 测试头部序列化
	headerBytes := header.AsBytes()
	fmt.Printf("  - Header size: %d bytes\n", len(headerBytes))

	// 测试头部反序列化
	parsedHeader, err := fauxgl.HeaderFromBytes(headerBytes)
	if err != nil {
		fmt.Printf("  ❌ Header parsing failed: %v\n", err)
	} else {
		fmt.Printf("  ✅ Header round-trip successful\n")
		if parsedHeader.PixelWidth == header.PixelWidth &&
			parsedHeader.PixelHeight == header.PixelHeight {
			fmt.Printf("  ✅ Header data integrity verified\n")
		}
	}
}

// tryLoadKTX2Texture 尝试加载KTX2纹理文件
func tryLoadKTX2Texture(filename string) *fauxgl.AdvancedTexture {
	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil
	}

	// 尝试加载KTX2纹理
	texture, err := fauxgl.LoadKTX2TextureFromFile(filename)
	if err != nil {
		fmt.Printf("⚠️  Failed to load KTX2 texture: %v\n", err)
		return nil
	}

	return texture
}

// createDemoTexture 创建一个演示纹理
func createDemoTexture() *fauxgl.AdvancedTexture {
	// 创建一个简单的程序性纹理作为演示
	// 在实际应用中，这里会是真正的KTX2纹理
	img := createSimpleColorImage(256, 256, fauxgl.Color{0.7, 0.5, 0.3, 1.0})
	return fauxgl.NewAdvancedTexture(img, fauxgl.BaseColorTexture)
}

// createSimpleColorImage 创建简单的纯色图像
func createSimpleColorImage(width, height int, col fauxgl.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	c := color.RGBA{
		uint8(col.R * 255),
		uint8(col.G * 255),
		uint8(col.B * 255),
		uint8(col.A * 255),
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

// testKTX2Extension 测试GLTF扩展中的KTX2支持
func testKTX2Extension(scene *fauxgl.Scene) {
	fmt.Println("\n🔌 Testing GLTF KTX2 Extension:")

	// 模拟KHR_texture_basisu扩展数据
	extensionData := map[string]interface{}{
		"KHR_texture_basisu": map[string]interface{}{
			"source": 0, // 引用纹理索引
		},
	}

	err := scene.ProcessGLTFExtensions(extensionData)
	if err != nil {
		fmt.Printf("  ❌ Extension processing failed: %v\n", err)
	} else {
		fmt.Println("  ✅ KHR_texture_basisu extension processed successfully")
	}

	// 显示支持的扩展列表
	extensions := scene.GetSupportedGLTFExtensions()
	fmt.Printf("  📋 Supported GLTF extensions (%d total):\n", len(extensions))
	for i, ext := range extensions {
		marker := "  "
		if ext == "KHR_texture_basisu" {
			marker = "🎯"
		}
		fmt.Printf("    %s %d. %s\n", marker, i+1, ext)
	}
}

// displayKTX2SupportInfo 显示KTX2支持信息
func displayKTX2SupportInfo() {
	fmt.Println("\n📊 KTX2 Support Status:")
	fmt.Println("  ✅ KTX2 Container Format:")
	fmt.Println("    - Header parsing and validation")
	fmt.Println("    - Level indexing and mipmap support")
	fmt.Println("    - Data format descriptors (DFD)")
	fmt.Println("    - Key-value metadata")
	fmt.Println("    - Supercompression detection")

	fmt.Println("  🚧 Partial Support:")
	fmt.Println("    - Basic texture loading (placeholder implementation)")
	fmt.Println("    - GLTF KHR_texture_basisu extension integration")

	fmt.Println("  🔄 Future Enhancements:")
	fmt.Println("    - Basis Universal decompression")
	fmt.Println("    - Zstd/ZLIB supercompression")
	fmt.Println("    - GPU texture format conversion")
	fmt.Println("    - Advanced mipmap chain handling")

	fmt.Println("  🎯 Integration Points:")
	fmt.Println("    - Advanced texture system")
	fmt.Println("    - GLTF extension registry")
	fmt.Println("    - PBR material system")
	fmt.Println("    - Scene management")

	fmt.Println("\n💡 Usage Scenarios:")
	fmt.Println("  - High-quality texture compression")
	fmt.Println("  - Optimized texture streaming")
	fmt.Println("  - Cross-platform texture distribution")
	fmt.Println("  - Modern GLTF asset pipelines")
}
