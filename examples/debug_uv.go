package main

import (
	"fmt"

	"github.com/swordkee/fauxgl-gltf"
)

func main() {
	fmt.Println("=== UV修改器调试测试 ===")

	// 测试UV修改器本身的功能
	testUVModifierLogic()

	// 测试纹理采样中的UV修改器
	testTextureUVModifier()

	fmt.Println("调试测试完成！")
}

func testUVModifierLogic() {
	fmt.Println("\n--- 测试UV修改器逻辑 ---")

	modifier := fauxgl.NewUVModifier()

	// 设置全局变换
	globalTransform := fauxgl.NewUVTransform()
	globalTransform.ScaleU = 2.0
	globalTransform.ScaleV = 0.5
	globalTransform.OffsetU = 0.1
	globalTransform.OffsetV = 0.2
	modifier.SetGlobalTransform(globalTransform)

	// 测试一些UV坐标的变换
	testCases := [][2]float64{
		{0.0, 0.0},   // 左下角
		{0.5, 0.5},   // 中心
		{1.0, 1.0},   // 右上角
		{0.25, 0.75}, // 随机点
	}

	fmt.Println("UV变换测试:")
	for i, uv := range testCases {
		origU, origV := uv[0], uv[1]
		newU, newV := modifier.TransformUV(origU, origV)
		fmt.Printf("  测试 %d: (%.3f, %.3f) -> (%.3f, %.3f)\n",
			i+1, origU, origV, newU, newV)
	}
}

func testTextureUVModifier() {
	fmt.Println("\n--- 测试纹理UV修改器 ---")

	// 加载场景
	scene, err := fauxgl.LoadGLTFScene("mug.gltf")
	if err != nil {
		panic(err)
	}

	// 获取第一个纹理
	var firstTexture *fauxgl.AdvancedTexture
	var textureName string
	for name, texture := range scene.Textures {
		firstTexture = texture
		textureName = name
		break
	}

	if firstTexture == nil {
		fmt.Println("未找到纹理")
		return
	}

	fmt.Printf("测试纹理: %s (尺寸: %dx%d)\n",
		textureName, firstTexture.Width, firstTexture.Height)

	// 测试原始采样
	fmt.Println("\n原始采样测试:")
	testUVs := [][2]float64{
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
	}

	for _, uv := range testUVs {
		color := firstTexture.SampleWithFilter(uv[0], uv[1], fauxgl.FilterLinear)
		fmt.Printf("  UV(%.1f, %.1f) -> RGBA(%.3f, %.3f, %.3f, %.3f)\n",
			uv[0], uv[1], color.R, color.G, color.B, color.A)
	}

	// 创建UV修改器
	modifier := fauxgl.NewUVModifier()
	globalTransform := fauxgl.NewUVTransform()
	globalTransform.ScaleU = 0.5 // 缩小采样范围
	globalTransform.ScaleV = 0.5
	modifier.SetGlobalTransform(globalTransform)

	// 应用UV修改器
	firstTexture.UVModifier = modifier
	fmt.Printf("\n应用UV修改器 (ScaleU=%.1f, ScaleV=%.1f)\n",
		globalTransform.ScaleU, globalTransform.ScaleV)

	// 测试修改后的采样
	fmt.Println("修改后采样测试:")
	for _, uv := range testUVs {
		color := firstTexture.SampleWithFilter(uv[0], uv[1], fauxgl.FilterLinear)
		fmt.Printf("  UV(%.1f, %.1f) -> RGBA(%.3f, %.3f, %.3f, %.3f)\n",
			uv[0], uv[1], color.R, color.G, color.B, color.A)
	}

	// 创建更极端的修改器
	modifier2 := fauxgl.NewUVModifier()
	globalTransform2 := fauxgl.NewUVTransform()
	globalTransform2.OffsetU = 0.5 // 偏移50%
	globalTransform2.OffsetV = 0.5
	modifier2.SetGlobalTransform(globalTransform2)

	firstTexture.UVModifier = modifier2
	fmt.Printf("\n应用偏移修改器 (OffsetU=%.1f, OffsetV=%.1f)\n",
		globalTransform2.OffsetU, globalTransform2.OffsetV)

	// 测试偏移后的采样
	fmt.Println("偏移后采样测试:")
	for _, uv := range testUVs {
		color := firstTexture.SampleWithFilter(uv[0], uv[1], fauxgl.FilterLinear)
		fmt.Printf("  UV(%.1f, %.1f) -> RGBA(%.3f, %.3f, %.3f, %.3f)\n",
			uv[0], uv[1], color.R, color.G, color.B, color.A)
	}

	// 验证UV修改器是否被正确调用
	fmt.Println("\n验证UV修改器调用:")
	fmt.Printf("  修改器是否为nil: %t\n", firstTexture.UVModifier == nil)
	if firstTexture.UVModifier != nil {
		fauxgl.PrintUVModifierInfo(firstTexture.UVModifier)
	}
}
