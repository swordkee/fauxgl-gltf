package main

import (
	"math"
	"time"

	"github.com/swordkee/fauxgl-gltf"
)

func main() {
	// Create rendering context
	width, height := 1024, 768
	context := fauxgl.NewContext(width, height)
	context.ClearColor = fauxgl.HexColor("1a1a2e")

	// Create scene
	scene := fauxgl.NewScene("PBR Demo Scene")

	// Create camera
	camera := fauxgl.NewPerspectiveCamera(
		"main_camera",
		fauxgl.Vector{0, 0, 5},         // position
		fauxgl.Vector{0, 0, 0},         // target
		fauxgl.Vector{0, 1, 0},         // up
		fauxgl.Radians(60),             // fov
		float64(width)/float64(height), // aspect ratio
		0.1, 100.0,                     // near, far
	)
	scene.AddCamera(camera)

	// Create lights
	directionalLight := fauxgl.Light{
		Type:      fauxgl.DirectionalLight,
		Direction: fauxgl.Vector{-1, -1, -1}.Normalize(),
		Color:     fauxgl.Color{1, 1, 0.9, 1}, // Warm white
		Intensity: 3.0,
	}
	scene.AddLight(directionalLight)

	pointLight := fauxgl.Light{
		Type:      fauxgl.PointLight,
		Position:  fauxgl.Vector{2, 2, 2},
		Color:     fauxgl.Color{0.8, 0.8, 1, 1}, // Cool blue
		Intensity: 2.0,
		Range:     10.0,
	}
	scene.AddLight(pointLight)

	// Create PBR materials with different properties
	materials := createPBRMaterials()
	for name, material := range materials {
		scene.AddMaterial(name, material)
	}

	// Create test meshes
	sphere := fauxgl.NewSphere(2)
	cube := fauxgl.NewCube()

	scene.AddMesh("sphere", sphere)
	scene.AddMesh("cube", cube)

	// Create scene nodes with different materials
	positions := []fauxgl.Vector{
		{-3, 0, 0}, {-1, 0, 0}, {1, 0, 0}, {3, 0, 0},
		{-3, 2, 0}, {-1, 2, 0}, {1, 2, 0}, {3, 2, 0},
	}

	materialNames := []string{
		"metal_rough", "metal_smooth", "plastic_rough", "plastic_smooth",
		"gold", "copper", "dielectric", "emissive",
	}

	for i, pos := range positions {
		nodeName := "sphere_" + materialNames[i]
		node := scene.CreateMeshNode(nodeName, "sphere", materialNames[i])
		node.Translate(pos)
		scene.RootNode.AddChild(node)
	}

	// Create animated cube
	animatedCube := scene.CreateMeshNode("animated_cube", "cube", "metal_rough")
	animatedCube.Translate(fauxgl.Vector{0, -2, 0})
	scene.RootNode.AddChild(animatedCube)

	// Create animation
	animation := createRotationAnimation(animatedCube, 5.0)
	scene.Animations["cube_rotation"] = animation

	// Animation player
	player := fauxgl.NewAnimationPlayer()
	player.AddAnimation("cube_rotation", animation)
	player.Play("cube_rotation")

	// Create scene renderer
	renderer := fauxgl.NewSceneRenderer(context)

	// Animation loop
	startTime := time.Now()
	frames := 60

	for frame := 0; frame < frames; frame++ {
		// Calculate time
		currentTime := time.Since(startTime).Seconds()
		deltaTime := 1.0 / 30.0 // 30 FPS

		// Update animation
		player.Update(deltaTime)

		// Orbit camera around scene
		angle := currentTime * 0.5
		distance := 8.0
		camera.Position = fauxgl.Vector{
			math.Cos(angle) * distance,
			2.0 + math.Sin(currentTime)*0.5,
			math.Sin(angle) * distance,
		}
		camera.LookAt(camera.Position, fauxgl.Vector{0, 0, 0}, fauxgl.Vector{0, 1, 0})

		// Clear buffers
		context.ClearColorBuffer()
		context.ClearDepthBuffer()

		// Render scene
		renderer.RenderScene(scene)

		// Save frame
		if frame == frames/2 { // Save middle frame
			filename := "pbr_demo.png"
			fauxgl.SavePNG(filename, context.Image())
			println("Saved:", filename)
		}
	}

	println("PBR Demo completed successfully!")
}

// createPBRMaterials creates various PBR materials for testing
func createPBRMaterials() map[string]*fauxgl.PBRMaterial {
	materials := make(map[string]*fauxgl.PBRMaterial)

	// Metallic rough
	metalRough := fauxgl.NewPBRMaterial()
	metalRough.BaseColorFactor = fauxgl.Color{0.7, 0.7, 0.7, 1}
	metalRough.MetallicFactor = 1.0
	metalRough.RoughnessFactor = 0.8
	materials["metal_rough"] = metalRough

	// Metallic smooth
	metalSmooth := fauxgl.NewPBRMaterial()
	metalSmooth.BaseColorFactor = fauxgl.Color{0.9, 0.9, 0.9, 1}
	metalSmooth.MetallicFactor = 1.0
	metalSmooth.RoughnessFactor = 0.1
	materials["metal_smooth"] = metalSmooth

	// Plastic rough
	plasticRough := fauxgl.NewPBRMaterial()
	plasticRough.BaseColorFactor = fauxgl.Color{0.8, 0.2, 0.2, 1}
	plasticRough.MetallicFactor = 0.0
	plasticRough.RoughnessFactor = 0.9
	materials["plastic_rough"] = plasticRough

	// Plastic smooth
	plasticSmooth := fauxgl.NewPBRMaterial()
	plasticSmooth.BaseColorFactor = fauxgl.Color{0.2, 0.8, 0.2, 1}
	plasticSmooth.MetallicFactor = 0.0
	plasticSmooth.RoughnessFactor = 0.1
	materials["plastic_smooth"] = plasticSmooth

	// Gold
	gold := fauxgl.NewPBRMaterial()
	gold.BaseColorFactor = fauxgl.Color{1.0, 0.86, 0.57, 1}
	gold.MetallicFactor = 1.0
	gold.RoughnessFactor = 0.2
	materials["gold"] = gold

	// Copper
	copper := fauxgl.NewPBRMaterial()
	copper.BaseColorFactor = fauxgl.Color{0.95, 0.64, 0.54, 1}
	copper.MetallicFactor = 1.0
	copper.RoughnessFactor = 0.25
	materials["copper"] = copper

	// Dielectric (glass-like)
	dielectric := fauxgl.NewPBRMaterial()
	dielectric.BaseColorFactor = fauxgl.Color{0.95, 0.95, 1.0, 0.8}
	dielectric.MetallicFactor = 0.0
	dielectric.RoughnessFactor = 0.05
	dielectric.AlphaMode = fauxgl.AlphaBlend
	materials["dielectric"] = dielectric

	// Emissive
	emissive := fauxgl.NewPBRMaterial()
	emissive.BaseColorFactor = fauxgl.Color{0.2, 0.2, 0.8, 1}
	emissive.MetallicFactor = 0.0
	emissive.RoughnessFactor = 0.5
	emissive.EmissiveFactor = fauxgl.Color{0.5, 0.8, 1.0, 1}
	materials["emissive"] = emissive

	return materials
}

// createRotationAnimation creates a simple rotation animation
func createRotationAnimation(target *fauxgl.SceneNode, duration float64) *fauxgl.Animation {
	animation := fauxgl.NewAnimation("rotation", duration)

	// Create rotation keyframes
	channel := fauxgl.AnimationChannel{
		Target:        target,
		Property:      fauxgl.Rotation,
		Interpolation: fauxgl.Linear,
	}

	// Add keyframes for full rotation
	steps := 8
	for i := 0; i <= steps; i++ {
		time := float64(i) / float64(steps) * duration
		angle := float64(i) / float64(steps) * 2 * math.Pi

		quat := fauxgl.QuaternionFromAxisAngle(fauxgl.Vector{0, 1, 0}, angle)

		keyframe := fauxgl.Keyframe{
			Time:  time,
			Value: quat,
		}

		channel.Keyframes = append(channel.Keyframes, keyframe)
	}

	animation.AddChannel(channel)
	return animation
}
