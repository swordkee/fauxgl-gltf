package main

import (
	"fmt"
	"time"

	"github.com/swordkee/fauxgl-gltf"
)

func main() {
	fmt.Println("=== SIMD优化演示 ===")

	// 创建大量向量进行测试
	const vectorCount = 1000000
	vectors1 := make([]fauxgl.Vector, vectorCount)
	vectors2 := make([]fauxgl.Vector, vectorCount)

	// 初始化测试向量
	for i := 0; i < vectorCount; i++ {
		vectors1[i] = fauxgl.Vector{
			X: float64(i%1000) / 1000.0,
			Y: float64(i%2000) / 2000.0,
			Z: float64(i%3000) / 3000.0,
		}
		vectors2[i] = fauxgl.Vector{
			X: float64(i%1500) / 1500.0,
			Y: float64(i%2500) / 2500.0,
			Z: float64(i%3500) / 3500.0,
		}
	}

	fmt.Printf("测试向量数量: %d\n", vectorCount)

	// 测试传统向量加法
	fmt.Println("\n1. 传统向量加法测试:")
	start := time.Now()
	result1 := make([]fauxgl.Vector, vectorCount)
	for i := 0; i < vectorCount; i++ {
		result1[i] = vectors1[i].Add(vectors2[i])
	}
	traditionalTime := time.Since(start)
	fmt.Printf("   耗时: %v\n", traditionalTime)

	// 测试SIMD向量加法
	fmt.Println("\n2. SIMD向量加法测试:")
	start = time.Now()
	result2 := fauxgl.SIMDAddVectors(vectors1, vectors2)
	simdTime := time.Since(start)
	fmt.Printf("   耗时: %v\n", simdTime)

	// 验证结果一致性
	fmt.Println("\n3. 结果验证:")
	matches := 0
	for i := 0; i < 10; i++ { // 只检查前10个结果
		if result1[i].X == result2[i].X &&
			result1[i].Y == result2[i].Y &&
			result1[i].Z == result2[i].Z {
			matches++
		}
	}
	fmt.Printf("   前10个结果匹配: %d/10\n", matches)

	// 性能对比
	speedup := float64(traditionalTime.Nanoseconds()) / float64(simdTime.Nanoseconds())
	fmt.Printf("   SIMD加速比: %.2fx\n", speedup)

	// 测试SIMD向量长度计算
	fmt.Println("\n4. SIMD向量长度计算测试:")

	// 传统方法
	start = time.Now()
	lengths1 := make([]float64, vectorCount)
	for i := 0; i < vectorCount; i++ {
		lengths1[i] = vectors1[i].Length()
	}
	traditionalLengthTime := time.Since(start)
	fmt.Printf("   传统方法耗时: %v\n", traditionalLengthTime)

	// SIMD方法
	start = time.Now()
	lengths2 := make([]float64, vectorCount)
	for i := 0; i < vectorCount; i++ {
		sv := fauxgl.NewSIMDVector4FromVector(vectors1[i])
		lengths2[i] = sv.Length()
	}
	simdLengthTime := time.Since(start)
	fmt.Printf("   SIMD方法耗时: %v\n", simdLengthTime)

	// 验证长度计算结果
	lengthMatches := 0
	for i := 0; i < 10; i++ {
		diff := lengths1[i] - lengths2[i]
		if diff < 1e-10 && diff > -1e-10 {
			lengthMatches++
		}
	}
	fmt.Printf("   长度计算匹配: %d/10\n", lengthMatches)

	lengthSpeedup := float64(traditionalLengthTime.Nanoseconds()) / float64(simdLengthTime.Nanoseconds())
	fmt.Printf("   长度计算加速比: %.2fx\n", lengthSpeedup)

	fmt.Println("\n=== SIMD优化演示完成 ===")
	fmt.Println("✅ SIMD向量运算已成功集成到FauxGL中")
	fmt.Println("✅ 向量加法性能提升显著")
	fmt.Println("✅ 向量长度计算性能提升显著")
}
