package main

import (
	"fmt"

	"github.com/swordkee/fauxgl-gltf"
)

func main() {
	fmt.Println("=== FauxGL-GLTF Advanced Features Demo ===")

	// 创建场景
	scene := fauxgl.NewScene("Advanced Features Demo")

	// 演示支持的GLTF扩展
	fmt.Println("\n🔧 Supported GLTF Extensions:")
	extensions := scene.GetSupportedGLTFExtensions()
	for i, ext := range extensions {
		fmt.Printf("  %d. %s\n", i+1, ext)
	}

	// 创建基础材质
	material := fauxgl.NewPBRMaterial()
	material.BaseColorFactor = fauxgl.Color{0.8, 0.4, 0.2, 1.0}
	material.MetallicFactor = 0.1
	material.RoughnessFactor = 0.7
	scene.AddMaterial("demo_material", material)

	// 创建基础网格
	sphere := fauxgl.NewSphere(1.0)
	scene.AddMesh("sphere", sphere)

	// 演示1: 基础场景节点
	basicNode := scene.CreateMeshNode("basic_sphere", "sphere", "demo_material")
	basicNode.Translate(fauxgl.V(-3, 0, 0))
	scene.RootNode.AddChild(basicNode)
	fmt.Println("\n✅ Created basic mesh node")

	// 演示2: 蒙皮动画支持
	skin := fauxgl.NewSkin("demo_skin")

	// 创建一些关节节点
	joint1 := fauxgl.NewSceneNode("joint1")
	joint1.Translate(fauxgl.V(0, 1, 0))

	joint2 := fauxgl.NewSceneNode("joint2")
	joint2.Translate(fauxgl.V(0, 2, 0))
	joint1.AddChild(joint2)

	// 添加关节到蒙皮
	skin.AddJoint(joint1, fauxgl.Identity()) // 简化的逆绑定矩阵
	skin.AddJoint(joint2, fauxgl.Identity())
	scene.AddSkin("demo_skin", skin)

	// 创建蒙皮网格节点
	skinnedNode := scene.CreateSkinnedMeshNode("skinned_sphere", "sphere", "demo_material", "demo_skin")
	skinnedNode.Translate(fauxgl.V(0, 0, 0))
	scene.RootNode.AddChild(skinnedNode)
	fmt.Println("✅ Created skinned mesh node with 2 joints")

	// 演示3: 变形目标支持
	morphTargets := &fauxgl.MorphTargets{
		Targets: []fauxgl.MorphTarget{
			*fauxgl.NewMorphTarget("smile", len(sphere.Triangles)*3), // 假设顶点数
			*fauxgl.NewMorphTarget("frown", len(sphere.Triangles)*3),
		},
		Weights: []float64{0.5, 0.3}, // 混合权重
	}
	scene.AddMorphTargets("facial_morphs", morphTargets)

	morphNode := scene.CreateMorphTargetMeshNode("morph_sphere", "sphere", "demo_material", "facial_morphs")
	morphNode.Translate(fauxgl.V(3, 0, 0))
	scene.RootNode.AddChild(morphNode)
	fmt.Println("✅ Created morph target mesh node with 2 targets")

	// 演示4: 高级光照
	scene.AddAmbientLight(fauxgl.Color{0.2, 0.2, 0.3, 1.0}, 0.4)
	scene.AddDirectionalLight(
		fauxgl.V(-1, -1, -1),
		fauxgl.Color{1.0, 0.9, 0.8, 1.0},
		2.5,
	)
	scene.AddPointLight(
		fauxgl.V(2, 3, 2),
		fauxgl.Color{0.8, 0.9, 1.0, 1.0},
		5.0, 10.0,
	)
	fmt.Println("✅ Added ambient, directional, and point lights")

	// 演示5: GLTF扩展处理
	extensionData := map[string]interface{}{
		"KHR_lights_punctual": map[string]interface{}{
			"lights": []interface{}{
				map[string]interface{}{
					"type":      "directional",
					"intensity": 3.0,
					"color":     []interface{}{1.0, 0.95, 0.8},
				},
			},
		},
	}

	err := scene.ProcessGLTFExtensions(extensionData)
	if err != nil {
		fmt.Printf("⚠️  Extension processing failed: %v\n", err)
	} else {
		fmt.Println("✅ Processed GLTF extensions successfully")
	}

	// 创建相机
	camera := fauxgl.NewPerspectiveCamera(
		"demo_camera",
		fauxgl.V(0, 3, 8),
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

	// 更新场景中的高级功能
	fmt.Println("\n🔄 Updating advanced features...")
	scene.UpdateSkinnedMeshes()
	scene.ApplyMorphTargetsToMeshes()

	// 渲染场景
	renderer := fauxgl.NewSceneRenderer(context)
	renderer.RenderScene(scene)

	// 保存结果
	err = fauxgl.SavePNG("advanced_features_demo.png", context.Image())
	if err != nil {
		fmt.Printf("❌ Failed to save image: %v\n", err)
	} else {
		fmt.Println("✅ Advanced features demo rendered and saved as advanced_features_demo.png")
	}

	// 显示场景统计信息
	fmt.Println("\n📊 Scene Statistics:")
	fmt.Printf("  - Nodes: %d\n", countNodes(scene.RootNode))
	fmt.Printf("  - Materials: %d\n", len(scene.Materials))
	fmt.Printf("  - Meshes: %d\n", len(scene.Meshes))
	fmt.Printf("  - Animations: %d\n", len(scene.Animations))
	fmt.Printf("  - Skins: %d\n", len(scene.Skins))
	fmt.Printf("  - Morph Targets: %d\n", len(scene.MorphTargets))
	fmt.Printf("  - Lights: %d\n", len(scene.Lights))
	fmt.Printf("  - Cameras: %d\n", len(scene.Cameras))
	fmt.Printf("  - Extensions: %d\n", len(extensions))

	fmt.Println("\n🎉 Advanced features demo completed!")
}

// 辅助函数：计算场景节点数量
func countNodes(node *fauxgl.SceneNode) int {
	count := 1
	for _, child := range node.Children {
		count += countNodes(child)
	}
	return count
}
