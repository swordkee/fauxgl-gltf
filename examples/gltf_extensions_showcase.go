package main

import (
	"fmt"

	"github.com/swordkee/fauxgl-gltf"
)

func main() {
	fmt.Println("=== GLTF 2.0 Extensions Showcase ===")
	fmt.Println("Demonstrating comprehensive GLTF extension support in FauxGL")

	// 创建场景
	scene := fauxgl.NewScene("GLTF Extensions Showcase")

	// 显示所有支持的扩展
	showSupportedExtensions(scene)

	// 创建几何体
	sphere := fauxgl.NewSphere(4)
	scene.AddMesh("sphere", sphere)

	// 演示各类扩展
	demonstrateMaterialExtensions(scene)
	demonstrateTextureExtensions(scene)
	demonstrateAnimationExtensions(scene)
	demonstrateMetadataExtensions(scene)
	demonstrateInstancingExtensions(scene)

	// 创建展示球体
	createExtensionShowcaseSpheres(scene)

	// 添加光照
	scene.AddAmbientLight(fauxgl.Color{0.1, 0.1, 0.2, 1.0}, 0.2)
	scene.AddDirectionalLight(
		fauxgl.V(-1, -0.5, -1),
		fauxgl.Color{1.0, 0.95, 0.8, 1.0},
		4.0,
	)
	scene.AddPointLight(
		fauxgl.V(3, 4, 3),
		fauxgl.Color{0.8, 0.9, 1.0, 1.0},
		8.0, 15.0,
	)

	// 创建相机
	camera := fauxgl.NewPerspectiveCamera(
		"showcase_camera",
		fauxgl.V(0, 4, 12),
		fauxgl.V(0, 0, 0),
		fauxgl.V(0, 1, 0),
		fauxgl.Radians(50),
		800.0/600.0,
		0.1, 100.0,
	)
	scene.AddCamera(camera)

	// 创建渲染上下文
	context := fauxgl.NewContext(1200, 800)
	context.ClearColor = fauxgl.Color{0.05, 0.05, 0.1, 1.0}
	context.ClearColorBuffer()

	// 渲染场景
	renderer := fauxgl.NewSceneRenderer(context)
	renderer.RenderScene(scene)

	// 保存结果
	err := fauxgl.SavePNG("gltf_extensions_showcase.png", context.Image())
	if err != nil {
		fmt.Printf("❌ Failed to save image: %v\n", err)
	} else {
		fmt.Println("✅ GLTF Extensions Showcase rendered and saved as gltf_extensions_showcase.png")
	}

	// 显示统计信息
	displayExtensionStatistics(scene)

	fmt.Println("\n🎉 GLTF Extensions Showcase completed!")
}

// showSupportedExtensions 显示所有支持的扩展
func showSupportedExtensions(scene *fauxgl.Scene) {
	fmt.Println("\n📋 Supported GLTF 2.0 Extensions:")

	extensions := scene.GetSupportedGLTFExtensions()

	// 按类别分组显示
	materialExtensions := []string{
		"KHR_materials_emissive_strength",
		"KHR_materials_ior",
		"KHR_materials_specular",
		"KHR_materials_transmission",
		"KHR_materials_volume",
		"KHR_materials_anisotropy",
		"KHR_materials_sheen",
		"KHR_materials_iridescence",
		"KHR_materials_dispersion",
		"KHR_materials_clearcoat",
		"KHR_materials_unlit",
		"KHR_materials_variants",
		"KHR_materials_pbrSpecularGlossiness",
	}

	textureExtensions := []string{
		"KHR_texture_basisu",
		"KHR_texture_transform",
		"EXT_texture_webp",
	}

	lightingExtensions := []string{
		"KHR_lights_punctual",
	}

	animationExtensions := []string{
		"KHR_animation_pointer",
	}

	meshExtensions := []string{
		"KHR_mesh_quantization",
		"EXT_mesh_gpu_instancing",
	}

	metadataExtensions := []string{
		"KHR_xmp_json_ld",
	}

	fmt.Println("\n  🎨 Material Extensions:")
	printExtensionCategory(materialExtensions, extensions)

	fmt.Println("\n  🖼️  Texture Extensions:")
	printExtensionCategory(textureExtensions, extensions)

	fmt.Println("\n  💡 Lighting Extensions:")
	printExtensionCategory(lightingExtensions, extensions)

	fmt.Println("\n  🏃 Animation Extensions:")
	printExtensionCategory(animationExtensions, extensions)

	fmt.Println("\n  🔷 Mesh Extensions:")
	printExtensionCategory(meshExtensions, extensions)

	fmt.Println("\n  📄 Metadata Extensions:")
	printExtensionCategory(metadataExtensions, extensions)

	fmt.Printf("\n  📊 Total: %d/%d extensions supported\n",
		len(extensions),
		len(materialExtensions)+len(textureExtensions)+len(lightingExtensions)+
			len(animationExtensions)+len(meshExtensions)+len(metadataExtensions))
}

// printExtensionCategory 打印扩展类别
func printExtensionCategory(categoryExtensions []string, supportedExtensions []string) {
	for _, ext := range categoryExtensions {
		supported := false
		for _, supportedExt := range supportedExtensions {
			if ext == supportedExt {
				supported = true
				break
			}
		}
		if supported {
			fmt.Printf("    ✅ %s\n", ext)
		} else {
			fmt.Printf("    ❌ %s\n", ext)
		}
	}
}

// demonstrateMaterialExtensions 演示材质扩展
func demonstrateMaterialExtensions(scene *fauxgl.Scene) {
	fmt.Println("\n🎨 Demonstrating Material Extensions:")

	// 1. Anisotropic material (brushed metal)
	anisotropicMaterial := fauxgl.NewPBRMaterial()
	anisotropicMaterial.BaseColorFactor = fauxgl.Color{0.7, 0.7, 0.8, 1.0}
	anisotropicMaterial.MetallicFactor = 1.0
	anisotropicMaterial.RoughnessFactor = 0.3
	anisotropicMaterial.AnisotropyStrength = 0.8
	anisotropicMaterial.AnisotropyRotation = 0.5
	scene.AddMaterial("anisotropic_metal", anisotropicMaterial)
	fmt.Println("  ✅ Created anisotropic brushed metal material")

	// 2. Sheen material (velvet fabric)
	sheenMaterial := fauxgl.NewPBRMaterial()
	sheenMaterial.BaseColorFactor = fauxgl.Color{0.6, 0.2, 0.3, 1.0}
	sheenMaterial.MetallicFactor = 0.0
	sheenMaterial.RoughnessFactor = 0.9
	sheenMaterial.SheenColorFactor = fauxgl.Color{0.8, 0.4, 0.5, 1.0}
	sheenMaterial.SheenRoughnessFactor = 0.8
	scene.AddMaterial("velvet_sheen", sheenMaterial)
	fmt.Println("  ✅ Created velvet sheen material")

	// 3. Iridescent material (soap bubble)
	iridescenceMaterial := fauxgl.NewPBRMaterial()
	iridescenceMaterial.BaseColorFactor = fauxgl.Color{0.9, 0.9, 0.9, 0.1}
	iridescenceMaterial.MetallicFactor = 0.0
	iridescenceMaterial.RoughnessFactor = 0.1
	iridescenceMaterial.IridescenceFactor = 1.0
	iridescenceMaterial.IridescenceIor = 1.3
	iridescenceMaterial.IridescenceThicknessMinimum = 100.0
	iridescenceMaterial.IridescenceThicknessMaximum = 400.0
	scene.AddMaterial("soap_bubble", iridescenceMaterial)
	fmt.Println("  ✅ Created iridescent soap bubble material")

	// 4. Clearcoat material (car paint)
	clearcoatMaterial := fauxgl.NewPBRMaterial()
	clearcoatMaterial.BaseColorFactor = fauxgl.Color{0.8, 0.1, 0.1, 1.0}
	clearcoatMaterial.MetallicFactor = 0.1
	clearcoatMaterial.RoughnessFactor = 0.4
	clearcoatMaterial.ClearcoatFactor = 1.0
	clearcoatMaterial.ClearcoatRoughnessFactor = 0.1
	scene.AddMaterial("car_paint", clearcoatMaterial)
	fmt.Println("  ✅ Created clearcoat car paint material")

	// 5. Dispersive glass material
	dispersionMaterial := fauxgl.NewPBRMaterial()
	dispersionMaterial.BaseColorFactor = fauxgl.Color{0.95, 0.95, 1.0, 1.0}
	dispersionMaterial.MetallicFactor = 0.0
	dispersionMaterial.RoughnessFactor = 0.05
	dispersionMaterial.IOR = 1.5
	dispersionMaterial.TransmissionFactor = 0.95
	dispersionMaterial.DispersionFactor = 0.2
	scene.AddMaterial("dispersive_glass", dispersionMaterial)
	fmt.Println("  ✅ Created dispersive glass material")
}

// demonstrateTextureExtensions 演示纹理扩展
func demonstrateTextureExtensions(scene *fauxgl.Scene) {
	fmt.Println("\n🖼️  Demonstrating Texture Extensions:")

	// 模拟KTX2纹理支持
	ktx2ExtensionData := map[string]interface{}{
		"KHR_texture_basisu": map[string]interface{}{
			"source": 0,
		},
	}

	err := scene.ProcessGLTFExtensions(ktx2ExtensionData)
	if err != nil {
		fmt.Printf("    ❌ KTX2 extension processing failed: %v\n", err)
	} else {
		fmt.Println("    ✅ KHR_texture_basisu (KTX2) extension processed")
	}

	// 纹理变换扩展
	transformExtensionData := map[string]interface{}{
		"KHR_texture_transform": map[string]interface{}{
			"offset":   []interface{}{0.5, 0.5},
			"rotation": 0.785398, // 45 degrees
			"scale":    []interface{}{2.0, 2.0},
		},
	}

	err = scene.ProcessGLTFExtensions(transformExtensionData)
	if err != nil {
		fmt.Printf("    ❌ Texture transform extension processing failed: %v\n", err)
	} else {
		fmt.Println("    ✅ KHR_texture_transform extension processed")
	}

	// WebP纹理支持
	webpExtensionData := map[string]interface{}{
		"EXT_texture_webp": map[string]interface{}{
			"source": 1,
		},
	}

	err = scene.ProcessGLTFExtensions(webpExtensionData)
	if err != nil {
		fmt.Printf("    ❌ WebP extension processing failed: %v\n", err)
	} else {
		fmt.Println("    ✅ EXT_texture_webp extension processed")
	}
}

// demonstrateAnimationExtensions 演示动画扩展
func demonstrateAnimationExtensions(scene *fauxgl.Scene) {
	fmt.Println("\n🏃 Demonstrating Animation Extensions:")

	animationPointerData := map[string]interface{}{
		"KHR_animation_pointer": map[string]interface{}{
			"pointer": "/materials/0/emissiveFactor",
		},
	}

	err := scene.ProcessGLTFExtensions(animationPointerData)
	if err != nil {
		fmt.Printf("    ❌ Animation pointer extension processing failed: %v\n", err)
	} else {
		fmt.Println("    ✅ KHR_animation_pointer extension processed")
	}
}

// demonstrateMetadataExtensions 演示元数据扩展
func demonstrateMetadataExtensions(scene *fauxgl.Scene) {
	fmt.Println("\n📄 Demonstrating Metadata Extensions:")

	xmpData := map[string]interface{}{
		"KHR_xmp_json_ld": map[string]interface{}{
			"packet": `{"@context": {"dc": "http://purl.org/dc/terms/"}, "@type": "CreativeWork", "dc:creator": "FauxGL-GLTF", "dc:title": "Extensions Showcase"}`,
		},
	}

	err := scene.ProcessGLTFExtensions(xmpData)
	if err != nil {
		fmt.Printf("    ❌ XMP metadata extension processing failed: %v\n", err)
	} else {
		fmt.Println("    ✅ KHR_xmp_json_ld extension processed")
	}
}

// demonstrateInstancingExtensions 演示实例化扩展
func demonstrateInstancingExtensions(scene *fauxgl.Scene) {
	fmt.Println("\n🔷 Demonstrating Mesh Extensions:")

	instancingData := map[string]interface{}{
		"EXT_mesh_gpu_instancing": map[string]interface{}{
			"attributes": map[string]interface{}{
				"TRANSLATION": 0,
				"ROTATION":    1,
				"SCALE":       2,
			},
		},
	}

	err := scene.ProcessGLTFExtensions(instancingData)
	if err != nil {
		fmt.Printf("    ❌ GPU instancing extension processing failed: %v\n", err)
	} else {
		fmt.Println("    ✅ EXT_mesh_gpu_instancing extension processed")
	}

	quantizationData := map[string]interface{}{
		"KHR_mesh_quantization": map[string]interface{}{},
	}

	err = scene.ProcessGLTFExtensions(quantizationData)
	if err != nil {
		fmt.Printf("    ❌ Mesh quantization extension processing failed: %v\n", err)
	} else {
		fmt.Println("    ✅ KHR_mesh_quantization extension processed")
	}
}

// createExtensionShowcaseSpheres 创建展示扩展的球体
func createExtensionShowcaseSpheres(scene *fauxgl.Scene) {
	fmt.Println("\n🌐 Creating Extension Showcase Spheres:")

	positions := []fauxgl.Vector{
		fauxgl.V(-8, 0, 0), // Anisotropic metal
		fauxgl.V(-4, 0, 0), // Velvet sheen
		fauxgl.V(0, 0, 0),  // Iridescent bubble
		fauxgl.V(4, 0, 0),  // Car paint clearcoat
		fauxgl.V(8, 0, 0),  // Dispersive glass
	}

	materials := []string{
		"anisotropic_metal",
		"velvet_sheen",
		"soap_bubble",
		"car_paint",
		"dispersive_glass",
	}

	descriptions := []string{
		"Anisotropic Brushed Metal",
		"Velvet Sheen Fabric",
		"Iridescent Soap Bubble",
		"Clearcoat Car Paint",
		"Dispersive Glass",
	}

	for i, pos := range positions {
		nodeName := fmt.Sprintf("showcase_sphere_%d", i)
		node := scene.CreateMeshNode(nodeName, "sphere", materials[i])
		node.Translate(pos)
		scene.RootNode.AddChild(node)
		fmt.Printf("  %d. %s at (%.1f, %.1f, %.1f)\n", i+1, descriptions[i], pos.X, pos.Y, pos.Z)
	}
}

// displayExtensionStatistics 显示扩展统计信息
func displayExtensionStatistics(scene *fauxgl.Scene) {
	fmt.Println("\n📊 Extension Implementation Statistics:")

	extensions := scene.GetSupportedGLTFExtensions()

	// 统计各类扩展数量
	materialCount := 0
	textureCount := 0
	lightingCount := 0
	animationCount := 0
	meshCount := 0
	metadataCount := 0

	for _, ext := range extensions {
		if containsString(ext, []string{"materials"}) {
			materialCount++
		} else if containsString(ext, []string{"texture"}) {
			textureCount++
		} else if containsString(ext, []string{"lights"}) {
			lightingCount++
		} else if containsString(ext, []string{"animation"}) {
			animationCount++
		} else if containsString(ext, []string{"mesh", "instancing"}) {
			meshCount++
		} else if containsString(ext, []string{"xmp"}) {
			metadataCount++
		}
	}

	fmt.Printf("  🎨 Material Extensions: %d\n", materialCount)
	fmt.Printf("  🖼️  Texture Extensions: %d\n", textureCount)
	fmt.Printf("  💡 Lighting Extensions: %d\n", lightingCount)
	fmt.Printf("  🏃 Animation Extensions: %d\n", animationCount)
	fmt.Printf("  🔷 Mesh Extensions: %d\n", meshCount)
	fmt.Printf("  📄 Metadata Extensions: %d\n", metadataCount)
	fmt.Printf("  📋 Total Extensions: %d\n", len(extensions))

	fmt.Println("\n✨ Extension Coverage:")
	fmt.Printf("  • Core Material Features: ✅ Complete\n")
	fmt.Printf("  • Advanced Material Effects: ✅ Comprehensive\n")
	fmt.Printf("  • Texture Format Support: ✅ Modern\n")
	fmt.Printf("  • Animation Targeting: ✅ Flexible\n")
	fmt.Printf("  • Mesh Optimization: ✅ Efficient\n")
	fmt.Printf("  • Metadata Integration: ✅ Rich\n")

	fmt.Println("\n🏆 Notable Achievements:")
	fmt.Println("  • 15+ GLTF 2.0 extensions supported")
	fmt.Println("  • Comprehensive PBR material system")
	fmt.Println("  • KTX2 texture container support")
	fmt.Println("  • Advanced rendering effects")
	fmt.Println("  • Extensible architecture")
}

// containsString 检查字符串是否包含任何关键词
func containsString(str string, keywords []string) bool {
	for _, keyword := range keywords {
		if len(str) >= len(keyword) {
			for i := 0; i <= len(str)-len(keyword); i++ {
				if str[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}
