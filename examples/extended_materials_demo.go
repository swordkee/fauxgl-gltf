package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/swordkee/fauxgl-gltf"
)

func main() {
	fmt.Println("=== FauxGL-GLTF Extended Materials Demo ===")

	// 创建场景
	scene := fauxgl.NewScene("Extended Materials Demo")

	// 演示新增的材质扩展支持
	demonstrateExtendedMaterials(scene)

	// 测试GLTF扩展
	testExtendedGLTFExtensions(scene)

	// 创建几何体
	sphere := fauxgl.NewSphere(3)
	scene.AddMesh("sphere", sphere)

	// 创建多个材质球体来展示不同的材质扩展
	createMaterialSpheres(scene)

	// 添加光照
	scene.AddAmbientLight(fauxgl.Color{0.2, 0.2, 0.3, 1.0}, 0.3)
	scene.AddDirectionalLight(
		fauxgl.V(-1, -1, -1),
		fauxgl.Color{1.0, 0.95, 0.8, 1.0},
		3.0,
	)

	// 创建相机
	camera := fauxgl.NewPerspectiveCamera(
		"materials_camera",
		fauxgl.V(0, 3, 10),
		fauxgl.V(0, 0, 0),
		fauxgl.V(0, 1, 0),
		fauxgl.Radians(45),
		800.0/600.0,
		0.1, 100.0,
	)
	scene.AddCamera(camera)

	// 创建渲染上下文
	context := fauxgl.NewContext(800, 600)
	context.ClearColor = fauxgl.Color{0.1, 0.1, 0.15, 1.0}
	context.ClearColorBuffer()

	// 渲染场景
	renderer := fauxgl.NewSceneRenderer(context)
	renderer.RenderScene(scene)

	// 保存结果
	err := fauxgl.SavePNG("extended_materials_demo.png", context.Image())
	if err != nil {
		fmt.Printf("❌ Failed to save image: %v\n", err)
	} else {
		fmt.Println("✅ Extended materials demo rendered and saved as extended_materials_demo.png")
	}

	fmt.Println("\n🎉 Extended materials demo completed!")
}

// demonstrateExtendedMaterials 演示扩展的材质属性
func demonstrateExtendedMaterials(scene *fauxgl.Scene) {
	fmt.Println("\n🎨 Extended Material Properties:")

	// 创建增强自发光材质
	emissiveMaterial := fauxgl.NewPBRMaterial()
	emissiveMaterial.BaseColorFactor = fauxgl.Color{0.2, 0.2, 0.2, 1.0}
	emissiveMaterial.EmissiveFactor = fauxgl.Color{1.0, 0.5, 0.0, 1.0}
	emissiveMaterial.EmissiveStrength = 3.0 // KHR_materials_emissive_strength
	fmt.Println("  ✅ Enhanced Emissive Strength: 3.0x intensity")

	// 创建高折射率材质
	glassMaterial := fauxgl.NewPBRMaterial()
	glassMaterial.BaseColorFactor = fauxgl.Color{0.9, 0.95, 1.0, 1.0}
	glassMaterial.MetallicFactor = 0.0
	glassMaterial.RoughnessFactor = 0.1
	glassMaterial.IOR = 1.52               // KHR_materials_ior (glass)
	glassMaterial.TransmissionFactor = 0.9 // KHR_materials_transmission
	fmt.Println("  ✅ Glass Material: IOR=1.52, Transmission=0.9")

	// 创建体积材质
	volumeMaterial := fauxgl.NewPBRMaterial()
	volumeMaterial.BaseColorFactor = fauxgl.Color{0.8, 0.2, 0.2, 1.0}
	volumeMaterial.ThicknessFactor = 2.0 // KHR_materials_volume
	volumeMaterial.AttenuationColor = fauxgl.Color{0.9, 0.1, 0.1, 1.0}
	volumeMaterial.AttenuationDistance = 1.0
	fmt.Println("  ✅ Volume Material: Thickness=2.0, Attenuation")

	// 创建镜面材质
	specularMaterial := fauxgl.NewPBRMaterial()
	specularMaterial.BaseColorFactor = fauxgl.Color{0.1, 0.1, 0.1, 1.0}
	specularMaterial.MetallicFactor = 0.0
	specularMaterial.RoughnessFactor = 0.2
	specularMaterial.SpecularColorFactor = fauxgl.Color{1.0, 0.8, 0.6, 1.0} // KHR_materials_specular
	fmt.Println("  ✅ Enhanced Specular: Colored specular reflection")

	// 添加材质到场景
	scene.AddMaterial("emissive_enhanced", emissiveMaterial)
	scene.AddMaterial("glass_material", glassMaterial)
	scene.AddMaterial("volume_material", volumeMaterial)
	scene.AddMaterial("specular_material", specularMaterial)
}

// testExtendedGLTFExtensions 测试扩展的GLTF扩展
func testExtendedGLTFExtensions(scene *fauxgl.Scene) {
	fmt.Println("\n🔌 Testing Extended GLTF Extensions:")

	// 测试新的材质扩展
	extensionData := map[string]interface{}{
		"KHR_materials_emissive_strength": map[string]interface{}{
			"emissiveStrength": 2.5,
		},
		"KHR_materials_ior": map[string]interface{}{
			"ior": 1.33, // Water
		},
		"KHR_materials_specular": map[string]interface{}{
			"specularFactor":      1.0,
			"specularColorFactor": []interface{}{1.0, 0.9, 0.8},
		},
		"KHR_materials_transmission": map[string]interface{}{
			"transmissionFactor": 0.8,
		},
		"KHR_materials_volume": map[string]interface{}{
			"thicknessFactor":     1.5,
			"attenuationDistance": 2.0,
			"attenuationColor":    []interface{}{0.8, 0.2, 0.2},
		},
	}

	err := scene.ProcessGLTFExtensions(extensionData)
	if err != nil {
		fmt.Printf("  ❌ Extension processing failed: %v\n", err)
	} else {
		fmt.Println("  ✅ All material extensions processed successfully")
	}

	// 显示支持的扩展列表
	extensions := scene.GetSupportedGLTFExtensions()
	fmt.Printf("  📋 Total supported extensions: %d\n", len(extensions))

	fmt.Println("  🆕 New material extensions:")
	newExtensions := []string{
		"KHR_materials_emissive_strength",
		"KHR_materials_ior",
		"KHR_materials_specular",
		"KHR_materials_transmission",
		"KHR_materials_volume",
	}

	for _, newExt := range newExtensions {
		found := false
		for _, ext := range extensions {
			if ext == newExt {
				found = true
				break
			}
		}
		if found {
			fmt.Printf("    ✅ %s\n", newExt)
		} else {
			fmt.Printf("    ❌ %s (not found)\n", newExt)
		}
	}
}

// createMaterialSpheres 创建展示不同材质的球体
func createMaterialSpheres(scene *fauxgl.Scene) {
	fmt.Println("\n🌐 Creating Material Demonstration Spheres:")

	positions := []fauxgl.Vector{
		fauxgl.V(-6, 0, 0), // 增强自发光
		fauxgl.V(-2, 0, 0), // 玻璃材质
		fauxgl.V(2, 0, 0),  // 体积材质
		fauxgl.V(6, 0, 0),  // 镜面材质
	}

	materials := []string{
		"emissive_enhanced",
		"glass_material",
		"volume_material",
		"specular_material",
	}

	descriptions := []string{
		"Enhanced Emissive (3.0x strength)",
		"Glass (IOR=1.52, Transmission=0.9)",
		"Volume (Thickness=2.0, Attenuation)",
		"Enhanced Specular (Colored reflection)",
	}

	for i, pos := range positions {
		nodeName := fmt.Sprintf("material_sphere_%d", i)
		node := scene.CreateMeshNode(nodeName, "sphere", materials[i])
		node.Translate(pos)
		scene.RootNode.AddChild(node)
		fmt.Printf("  %d. %s at (%.1f, %.1f, %.1f)\n", i+1, descriptions[i], pos.X, pos.Y, pos.Z)
	}
}

// createSimpleTexture 创建简单的程序纹理
func createSimpleTexture(width, height int, col fauxgl.Color) *fauxgl.AdvancedTexture {
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
	return fauxgl.NewAdvancedTexture(img, fauxgl.BaseColorTexture)
}
