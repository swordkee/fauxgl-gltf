package main

import (
	"fmt"
	"math"
	"time"

	"github.com/swordkee/fauxgl-gltf"
)

// 高质量渲染参数
const (
	scale  = 1
	width  = 1600
	height = 1600
	fovy   = 30
	near   = 1
	far    = 20
)

var (
	eye    = fauxgl.V(2.5, 4, 4.0)
	center = fauxgl.V(0, 1.14, 0.4)
	up     = fauxgl.V(0, 1, 0)
)

func main() {
	fmt.Println("=== FauxGL-GLTF 动态UV修改器 - 完整演示 ===")

	// 加载GLTF场景
	scene, err := fauxgl.LoadGLTFScene("mug.gltf")
	if err != nil {
		panic(err)
	}

	fmt.Printf("场景加载成功: %d材质, %d网格, %d纹理\n",
		len(scene.Materials), len(scene.Meshes), len(scene.Textures))

	// 预处理网格
	preprocessMeshes(scene)

	// 演示1: 基础UV变换效果
	demonstrateBasicUVTransforms(scene)

	// 演示2: 区域性UV修改
	demonstrateRegionalUVModification(scene)

	// 演示3: 动画UV效果
	demonstrateAnimatedUVEffects(scene)

	// 演示4: 复合UV变换
	demonstrateCompositeUVTransforms(scene)

	// 演示5: 实时UV修改（模拟）
	demonstrateRealtimeUVModification(scene)

	fmt.Println("\n=== FauxGL-GLTF 动态UV修改器演示完成 ===")
	fmt.Println("✅ UV修改器已成功整合到FauxGL-GLTF引擎中")
	fmt.Println("✅ 支持实时动态UV坐标修改")
	fmt.Println("✅ 支持多种UV变换：缩放、旋转、偏移、剪切")
	fmt.Println("✅ 支持区域遮罩和混合模式")
	fmt.Println("✅ 支持动画和时间线控制")
}

// preprocessMeshes 预处理网格
func preprocessMeshes(scene *fauxgl.Scene) {
	fmt.Println("\n=== 网格预处理 ===")
	for name, mesh := range scene.Meshes {
		fmt.Printf("处理网格: %s (%d三角形)\n", name, len(mesh.Triangles))
		mesh.SmoothNormalsThreshold(fauxgl.Radians(30))
	}
}

// demonstrateBasicUVTransforms 演示基础UV变换
func demonstrateBasicUVTransforms(scene *fauxgl.Scene) {
	fmt.Println("\n=== 演示1: 基础UV变换效果 ===")

	// UV缩放效果
	fmt.Println("1.1 UV缩放效果")
	modifier1 := createScaleUVModifier(2.0, 1.5)
	renderWithUVModifier(scene, modifier1, "demo1_uv_scale.png", "UV缩放(2x, 1.5x)")

	// UV旋转效果
	fmt.Println("1.2 UV旋转效果")
	modifier2 := createRotationUVModifier(math.Pi / 4) // 45度
	renderWithUVModifier(scene, modifier2, "demo1_uv_rotation.png", "UV旋转(45度)")

	// UV偏移效果
	fmt.Println("1.3 UV偏移效果")
	modifier3 := createOffsetUVModifier(0.3, 0.2)
	renderWithUVModifier(scene, modifier3, "demo1_uv_offset.png", "UV偏移(0.3, 0.2)")
}

// demonstrateRegionalUVModification 演示区域性UV修改
func demonstrateRegionalUVModification(scene *fauxgl.Scene) {
	fmt.Println("\n=== 演示2: 区域性UV修改 ===")

	modifier := fauxgl.NewUVModifier()

	// 上半部分：缩放效果
	upperMapping := &fauxgl.UVMapping{
		Name:    "upper_scale",
		Enabled: true,
		Region: fauxgl.UVRegion{
			MinU: 0.0, MaxU: 1.0,
			MinV: 0.5, MaxV: 1.0,
			MaskType: fauxgl.UVMaskRectangle,
		},
		Transform: &fauxgl.UVTransform{
			ScaleU: 1.5, ScaleV: 1.5,
			PivotU: 0.5, PivotV: 0.75,
		},
		BlendMode: fauxgl.UVBlendReplace,
		Priority:  1,
	}
	modifier.AddMapping(upperMapping)

	// 下半部分：旋转效果
	lowerMapping := &fauxgl.UVMapping{
		Name:    "lower_rotation",
		Enabled: true,
		Region: fauxgl.UVRegion{
			MinU: 0.0, MaxU: 1.0,
			MinV: 0.0, MaxV: 0.5,
			MaskType: fauxgl.UVMaskRectangle,
		},
		Transform: &fauxgl.UVTransform{
			ScaleU: 1.0, ScaleV: 1.0,
			PivotU: 0.5, PivotV: 0.25,
			Rotation: math.Pi / 6, // 30度
		},
		BlendMode: fauxgl.UVBlendReplace,
		Priority:  1,
	}
	modifier.AddMapping(lowerMapping)

	// 中心圆形区域：特殊效果
	centerMapping := &fauxgl.UVMapping{
		Name:    "center_effect",
		Enabled: true,
		Region: fauxgl.UVRegion{
			MinU: 0.25, MaxU: 0.75,
			MinV: 0.25, MaxV: 0.75,
			MaskType: fauxgl.UVMaskCircle,
		},
		Transform: &fauxgl.UVTransform{
			ScaleU: 0.8, ScaleV: 0.8,
			PivotU: 0.5, PivotV: 0.5,
			Rotation: -math.Pi / 4, // -45度
		},
		BlendMode: fauxgl.UVBlendOverlay,
		Priority:  2,
	}
	modifier.AddMapping(centerMapping)

	renderWithUVModifier(scene, modifier, "demo2_regional_uv.png", "区域性UV修改")
}

// demonstrateAnimatedUVEffects 演示动画UV效果
func demonstrateAnimatedUVEffects(scene *fauxgl.Scene) {
	fmt.Println("\n=== 演示3: 动画UV效果 ===")

	// 生成动画序列
	for frame := 0; frame < 8; frame++ {
		modifier := fauxgl.NewUVModifier()
		modifier.EnableAnimation(true)

		// 时间参数
		t := float64(frame) / 7.0 // 0到1

		// 旋转动画
		rotationMapping := &fauxgl.UVMapping{
			Name:    "rotation_anim",
			Enabled: true,
			Region: fauxgl.UVRegion{
				MinU: 0.0, MaxU: 1.0,
				MinV: 0.0, MaxV: 1.0,
				MaskType: fauxgl.UVMaskCircle,
			},
			Transform: &fauxgl.UVTransform{
				ScaleU: 1.0, ScaleV: 1.0,
				PivotU: 0.5, PivotV: 0.5,
				Rotation: t * 2 * math.Pi, // 完整旋转
			},
			BlendMode: fauxgl.UVBlendReplace,
			Priority:  1,
		}
		modifier.AddMapping(rotationMapping)

		// 滚动动画
		scrollMapping := &fauxgl.UVMapping{
			Name:    "scroll_anim",
			Enabled: true,
			Region: fauxgl.UVRegion{
				MinU: 0.0, MaxU: 1.0,
				MinV: 0.6, MaxV: 1.0,
				MaskType: fauxgl.UVMaskRectangle,
			},
			Transform: &fauxgl.UVTransform{
				OffsetU: t * 0.5, // 滚动偏移
				OffsetV: t * 0.3,
			},
			BlendMode: fauxgl.UVBlendAdd,
			Priority:  0,
		}
		modifier.AddMapping(scrollMapping)

		filename := fmt.Sprintf("demo3_anim_frame_%02d.png", frame)
		renderWithUVModifier(scene, modifier, filename, fmt.Sprintf("动画帧%d", frame))
	}
}

// demonstrateCompositeUVTransforms 演示复合UV变换
func demonstrateCompositeUVTransforms(scene *fauxgl.Scene) {
	fmt.Println("\n=== 演示4: 复合UV变换 ===")

	modifier := fauxgl.NewUVModifier()

	// 全局变换：缩放和旋转
	globalTransform := &fauxgl.UVTransform{
		ScaleU: 1.2, ScaleV: 1.2,
		PivotU: 0.5, PivotV: 0.5,
		Rotation: math.Pi / 8, // 22.5度
	}
	modifier.SetGlobalTransform(globalTransform)

	// 渐变效果
	gradientMapping := &fauxgl.UVMapping{
		Name:    "gradient_transform",
		Enabled: true,
		Region: fauxgl.UVRegion{
			MinU: 0.0, MaxU: 1.0,
			MinV: 0.0, MaxV: 1.0,
			MaskType: fauxgl.UVMaskGradient,
		},
		Transform: &fauxgl.UVTransform{
			SkewU: 0.2, // 剪切变换
			SkewV: 0.1,
		},
		BlendMode: fauxgl.UVBlendAdd,
		Priority:  1,
	}
	modifier.AddMapping(gradientMapping)

	// 局部扭曲效果
	distortionMapping := &fauxgl.UVMapping{
		Name:    "distortion_effect",
		Enabled: true,
		Region: fauxgl.UVRegion{
			MinU: 0.3, MaxU: 0.7,
			MinV: 0.3, MaxV: 0.7,
			MaskType: fauxgl.UVMaskCircle,
		},
		Transform: &fauxgl.UVTransform{
			ScaleU: 0.6, ScaleV: 0.6,
			PivotU: 0.5, PivotV: 0.5,
			Rotation: math.Pi / 3, // 60度
			OffsetU:  0.1, OffsetV: -0.1,
		},
		BlendMode: fauxgl.UVBlendMultiply,
		Priority:  2,
	}
	modifier.AddMapping(distortionMapping)

	renderWithUVModifier(scene, modifier, "demo4_composite_uv.png", "复合UV变换")
}

// demonstrateRealtimeUVModification 演示实时UV修改
func demonstrateRealtimeUVModification(scene *fauxgl.Scene) {
	fmt.Println("\n=== 演示5: 实时UV修改（模拟） ===")

	// 模拟实时修改：每秒更新一次UV参数
	for step := 0; step < 5; step++ {
		fmt.Printf("实时步骤 %d: ", step+1)

		modifier := fauxgl.NewUVModifier()

		// 根据时间动态调整参数
		currentTime := float64(step)

		// 动态全局变换
		globalTransform := &fauxgl.UVTransform{
			ScaleU:   1.0 + 0.2*math.Sin(currentTime),
			ScaleV:   1.0 + 0.2*math.Cos(currentTime),
			Rotation: currentTime * 0.3,
			PivotU:   0.5, PivotV: 0.5,
		}
		modifier.SetGlobalTransform(globalTransform)

		// 动态波形效果
		waveMapping := &fauxgl.UVMapping{
			Name:    "realtime_wave",
			Enabled: true,
			Region: fauxgl.UVRegion{
				MinU: 0.0, MaxU: 1.0,
				MinV: 0.0, MaxV: 1.0,
				MaskType: fauxgl.UVMaskRectangle,
			},
			Transform: &fauxgl.UVTransform{
				OffsetU: 0.1 * math.Sin(currentTime*2),
				OffsetV: 0.1 * math.Cos(currentTime*2),
			},
			BlendMode: fauxgl.UVBlendAdd,
			Priority:  1,
		}
		modifier.AddMapping(waveMapping)

		filename := fmt.Sprintf("demo5_realtime_step_%d.png", step+1)
		description := fmt.Sprintf("实时步骤%d(t=%.1f)", step+1, currentTime)
		renderWithUVModifier(scene, modifier, filename, description)

		// 模拟时间延迟
		time.Sleep(100 * time.Millisecond)
	}
}

// 辅助函数

// createScaleUVModifier 创建缩放UV修改器
func createScaleUVModifier(scaleU, scaleV float64) *fauxgl.UVModifier {
	modifier := fauxgl.NewUVModifier()
	transform := fauxgl.NewUVTransform()
	transform.ScaleU = scaleU
	transform.ScaleV = scaleV
	modifier.SetGlobalTransform(transform)
	return modifier
}

// createRotationUVModifier 创建旋转UV修改器
func createRotationUVModifier(rotation float64) *fauxgl.UVModifier {
	modifier := fauxgl.NewUVModifier()
	transform := fauxgl.NewUVTransform()
	transform.Rotation = rotation
	transform.PivotU = 0.5
	transform.PivotV = 0.5
	modifier.SetGlobalTransform(transform)
	return modifier
}

// createOffsetUVModifier 创建偏移UV修改器
func createOffsetUVModifier(offsetU, offsetV float64) *fauxgl.UVModifier {
	modifier := fauxgl.NewUVModifier()
	transform := fauxgl.NewUVTransform()
	transform.OffsetU = offsetU
	transform.OffsetV = offsetV
	modifier.SetGlobalTransform(transform)
	return modifier
}

// renderWithUVModifier 使用UV修改器渲染场景
func renderWithUVModifier(scene *fauxgl.Scene, modifier *fauxgl.UVModifier, filename, description string) {
	fmt.Printf("渲染 %s -> %s\n", description, filename)

	// 应用UV修改器到所有纹理
	for _, texture := range scene.Textures {
		texture.UVModifier = modifier
	}

	// 执行高质量渲染
	context := fauxgl.NewContext(width*scale, height*scale)
	context.ClearColor = fauxgl.Color{0.95, 0.95, 0.95, 1.0}
	context.ClearColorBuffer()
	context.ClearDepthBuffer()

	// 设置相机和光照
	aspect := float64(width) / float64(height)
	matrix := fauxgl.LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	mainLight := fauxgl.V(-0.6, -0.8, -0.4).Normalize()

	// 渲染所有节点
	renderableNodes := scene.RootNode.GetRenderableNodes()
	for _, node := range renderableNodes {
		if node.Mesh == nil || node.Material == nil {
			continue
		}

		// 创建增强着色器
		shader := fauxgl.NewPhongShader(matrix, mainLight, eye)
		shader.DiffuseColor = fauxgl.Color{1.0, 0.95, 0.8, 1.0}
		shader.SpecularColor = fauxgl.Color{1.0, 1.0, 1.0, 1.0}
		shader.SpecularPower = 64

		// 应用材质
		if node.Material.BaseColorTexture != nil {
			shader.Texture = node.Material.BaseColorTexture
		}

		shader.ObjectColor = fauxgl.Color{
			node.Material.BaseColorFactor.R,
			node.Material.BaseColorFactor.G,
			node.Material.BaseColorFactor.B,
			node.Material.BaseColorFactor.A,
		}

		// 渲染
		context.Shader = shader
		context.DrawMesh(node.Mesh)
	}

	// 保存结果
	err := fauxgl.SavePNG(filename, context.Image())
	if err != nil {
		fmt.Printf("保存失败: %v\n", err)
	} else {
		fmt.Printf("已保存: %s\n", filename)
	}

	// 清理UV修改器
	for _, texture := range scene.Textures {
		texture.UVModifier = nil
	}
}
