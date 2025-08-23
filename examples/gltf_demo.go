package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/swordkee/fauxgl-gltf"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run gltf_demo.go <path_to_gltf_file>")
		fmt.Println("Example: go run gltf_demo.go models/duck.gltf")
		return
	}

	gltfPath := os.Args[1]

	// Create rendering context
	width, height := 1920, 1080
	context := fauxgl.NewContext(width, height)
	context.ClearColor = fauxgl.Color{0.1, 0.1, 0.15, 1.0}

	fmt.Printf("Loading GLTF scene from: %s\n", gltfPath)

	// Load GLTF scene
	scene, err := fauxgl.LoadGLTFScene(gltfPath)
	if err != nil {
		fmt.Printf("Error loading GLTF scene: %v\n", err)
		return
	}

	// Print scene information
	printSceneInfo(scene)

	// Set up camera if none exists
	if scene.ActiveCamera == nil {
		bounds := scene.GetBounds()
		center := bounds.Center()
		size := bounds.Size()
		distance := size.MaxComponent() * 2.5

		camera := fauxgl.NewPerspectiveCamera(
			"default_camera",
			center.Add(fauxgl.Vector{distance, distance * 0.5, distance}),
			center,
			fauxgl.Vector{0, 1, 0},
			fauxgl.Radians(45),
			float64(width)/float64(height),
			0.1, distance*10,
		)

		scene.AddCamera(camera)
		fmt.Printf("Created default camera at distance %.2f from center %v\n", distance, center)
	}

	// Add default lights if none exist
	if len(scene.Lights) == 0 {
		// Key light (main directional light)
		keyLight := fauxgl.Light{
			Type:      fauxgl.DirectionalLight,
			Direction: fauxgl.Vector{-0.5, -1, -0.5}.Normalize(),
			Color:     fauxgl.Color{1.0, 0.95, 0.8, 1}, // Warm sunlight
			Intensity: 3.0,
		}
		scene.AddLight(keyLight)

		// Fill light (softer from opposite side)
		fillLight := fauxgl.Light{
			Type:      fauxgl.DirectionalLight,
			Direction: fauxgl.Vector{0.5, 0.2, 0.8}.Normalize(),
			Color:     fauxgl.Color{0.6, 0.7, 1.0, 1}, // Cool sky light
			Intensity: 1.0,
		}
		scene.AddLight(fillLight)

		// Rim light (back lighting)
		rimLight := fauxgl.Light{
			Type:      fauxgl.DirectionalLight,
			Direction: fauxgl.Vector{0, 0.5, -1}.Normalize(),
			Color:     fauxgl.Color{1.0, 0.9, 0.7, 1}, // Warm rim
			Intensity: 2.0,
		}
		scene.AddLight(rimLight)

		fmt.Println("Added default three-point lighting setup")
	}

	// Create scene renderer
	renderer := fauxgl.NewSceneRenderer(context)

	// Animation setup
	var animationPlayer *fauxgl.AnimationPlayer
	if len(scene.Animations) > 0 {
		animationPlayer = fauxgl.NewAnimationPlayer()
		for name, animation := range scene.Animations {
			animationPlayer.AddAnimation(name, animation)
			fmt.Printf("Added animation: %s (duration: %.2fs)\n", name, animation.Duration)
		}

		// Play the first animation
		for name := range scene.Animations {
			animationPlayer.Play(name)
			fmt.Printf("Playing animation: %s\n", name)
			break
		}
	}

	// Render multiple frames with camera animation
	frames := 120 // 4 seconds at 30 FPS
	startTime := time.Now()

	fmt.Printf("Rendering %d frames...\n", frames)

	for frame := 0; frame < frames; frame++ {
		currentTime := time.Since(startTime).Seconds()
		deltaTime := 1.0 / 30.0

		// Update animations
		if animationPlayer != nil {
			animationPlayer.Update(deltaTime)
		}

		// Animate camera around the scene
		if scene.ActiveCamera != nil {
			animateCamera(scene.ActiveCamera, scene.GetBounds(), currentTime)
		}

		// Clear buffers
		context.ClearColorBuffer()
		context.ClearDepthBuffer()

		// Render scene
		renderer.RenderScene(scene)

		// Save specific frames
		if frame == 0 || frame == frames/4 || frame == frames/2 || frame == frames-1 {
			filename := fmt.Sprintf("gltf_demo_frame_%03d.png", frame)
			err := fauxgl.SavePNG(filename, context.Image())
			if err != nil {
				fmt.Printf("Error saving frame %d: %v\n", frame, err)
			} else {
				fmt.Printf("Saved frame %d: %s\n", frame, filename)
			}
		}

		// Progress indicator
		if frame%30 == 0 {
			progress := float64(frame) / float64(frames) * 100
			fmt.Printf("Progress: %.1f%%\n", progress)
		}
	}

	fmt.Printf("GLTF demo completed! Rendered %d frames in %.2f seconds\n",
		frames, time.Since(startTime).Seconds())
}

// printSceneInfo prints information about the loaded scene
func printSceneInfo(scene *fauxgl.Scene) {
	fmt.Printf("=== Scene Information ===\n")
	fmt.Printf("Scene Name: %s\n", scene.Name)
	fmt.Printf("Cameras: %d\n", len(scene.Cameras))
	fmt.Printf("Lights: %d\n", len(scene.Lights))
	fmt.Printf("Materials: %d\n", len(scene.Materials))
	fmt.Printf("Textures: %d\n", len(scene.Textures))
	fmt.Printf("Meshes: %d\n", len(scene.Meshes))
	fmt.Printf("Animations: %d\n", len(scene.Animations))

	// Count renderable nodes
	renderables := scene.RootNode.GetRenderableNodes()
	fmt.Printf("Renderable nodes: %d\n", len(renderables))

	// Scene bounds
	bounds := scene.GetBounds()
	fmt.Printf("Scene bounds: min=%v, max=%v\n", bounds.Min, bounds.Max)
	fmt.Printf("Scene size: %v\n", bounds.Size())
	fmt.Printf("Scene center: %v\n", bounds.Center())

	// List materials
	if len(scene.Materials) > 0 {
		fmt.Printf("Materials:\n")
		for name, material := range scene.Materials {
			fmt.Printf("  - %s: metallic=%.2f, roughness=%.2f, baseColor=%v\n",
				name, material.MetallicFactor, material.RoughnessFactor, material.BaseColorFactor)
		}
	}

	fmt.Printf("========================\n\n")
}

// animateCamera animates the camera around the scene
func animateCamera(camera *fauxgl.Camera, bounds fauxgl.Box, time float64) {
	center := bounds.Center()
	size := bounds.Size()

	// Calculate orbit parameters
	radius := size.MaxComponent() * 2.0
	height := center.Y + size.Y*0.3

	// Animate camera position (orbit around Y axis)
	angle := time * 0.3 // Slow rotation
	x := center.X + math.Cos(angle)*radius
	z := center.Z + math.Sin(angle)*radius
	y := height + math.Sin(time*0.5)*size.Y*0.2 // Slight vertical movement

	camera.Position = fauxgl.Vector{x, y, z}
	camera.Target = center
	camera.Up = fauxgl.Vector{0, 1, 0}
}
