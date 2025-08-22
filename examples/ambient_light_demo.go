package main

import (
	"fmt"

	"github.com/swordkee/fauxgl-gltf"
)

func main() {
	fmt.Println("Testing AmbientLight implementation...")

	// 创建场景
	scene := fauxgl.NewScene("AmbientLight Demo Scene")

	// 创建基本PBR材质
	material := fauxgl.NewPBRMaterial()
	material.BaseColorFactor = fauxgl.Color{0.8, 0.3, 0.3, 1.0} // 红色
	material.MetallicFactor = 0.1
	material.RoughnessFactor = 0.6
	scene.AddMaterial("demo_material", material)

	// 创建一个简单的球体网格
	sphere := fauxgl.NewSphere(1.0)
	scene.AddMesh("sphere", sphere)

	// 创建场景节点
	node := scene.CreateMeshNode("sphere_node", "sphere", "demo_material")
	scene.RootNode.AddChild(node)

	// 添加环境光
	scene.AddAmbientLight(fauxgl.Color{0.2, 0.3, 0.4, 1.0}, 0.8)
	fmt.Println("Added ambient light with blue tint and intensity 0.8")

	// 添加方向光作为对比
	scene.AddDirectionalLight(
		fauxgl.V(-1, -1, -1),
		fauxgl.Color{1.0, 0.9, 0.8, 1.0},
		2.0,
	)
	fmt.Println("Added directional light for comparison")

	// 创建相机
	camera := fauxgl.NewPerspectiveCamera(
		"main_camera",
		fauxgl.V(0, 0, 4),  // 位置
		fauxgl.V(0, 0, 0),  // 目标
		fauxgl.V(0, 1, 0),  // 上方向
		fauxgl.Radians(45), // 视野角度
		1.0,                // 宽高比
		0.1, 10.0,          // 近远平面
	)
	scene.AddCamera(camera)

	// 创建渲染上下文
	context := fauxgl.NewContext(800, 600)
	context.ClearColor = fauxgl.Color{0.1, 0.1, 0.15, 1.0}
	context.ClearColorBuffer()

	// 获取环境光
	ambientLights := scene.GetLightsByType(fauxgl.AmbientLight)
	fmt.Printf("Found %d ambient lights in scene\n", len(ambientLights))

	// 检查所有光源
	fmt.Printf("Total lights in scene: %d\n", len(scene.Lights))
	for i, light := range scene.Lights {
		switch light.Type {
		case fauxgl.AmbientLight:
			fmt.Printf("  Light %d: Ambient - Color: %.2f,%.2f,%.2f Intensity: %.2f\n",
				i, light.Color.R, light.Color.G, light.Color.B, light.Intensity)
		case fauxgl.DirectionalLight:
			fmt.Printf("  Light %d: Directional - Direction: %.2f,%.2f,%.2f Intensity: %.2f\n",
				i, light.Direction.X, light.Direction.Y, light.Direction.Z, light.Intensity)
		case fauxgl.PointLight:
			fmt.Printf("  Light %d: Point - Position: %.2f,%.2f,%.2f Intensity: %.2f\n",
				i, light.Position.X, light.Position.Y, light.Position.Z, light.Intensity)
		case fauxgl.SpotLight:
			fmt.Printf("  Light %d: Spot - Position: %.2f,%.2f,%.2f Intensity: %.2f\n",
				i, light.Position.X, light.Position.Y, light.Position.Z, light.Intensity)
		}
	}

	// 渲染场景
	renderer := fauxgl.NewSceneRenderer(context)
	renderer.RenderScene(scene)

	// 保存结果
	err := fauxgl.SavePNG("ambient_light_demo.png", context.Image())
	if err != nil {
		fmt.Printf("Failed to save image: %v\n", err)
	} else {
		fmt.Println("Rendered scene with ambient light saved as ambient_light_demo.png")
	}

	fmt.Println("AmbientLight demo completed successfully!")
}
