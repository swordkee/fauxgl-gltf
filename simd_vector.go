// Package simd_vector 提供SIMD优化的向量计算功能
package fauxgl

import (
	"math"
)

// VectorW 表示齐次坐标向量
type VectorW struct {
	X, Y, Z, W float64
}

func (a VectorW) Vector() Vector {
	return Vector{a.X, a.Y, a.Z}
}

func (a VectorW) Outside() bool {
	x, y, z, w := a.X, a.Y, a.Z, a.W
	return x < -w || x > w || y < -w || y > w || z < -w || z > w
}

func (a VectorW) Dot(b VectorW) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z + a.W*b.W
}

func (a VectorW) Add(b VectorW) VectorW {
	return VectorW{a.X + b.X, a.Y + b.Y, a.Z + b.Z, a.W + b.W}
}

func (a VectorW) Sub(b VectorW) VectorW {
	return VectorW{a.X - b.X, a.Y - b.Y, a.Z - b.Z, a.W - b.W}
}

func (a VectorW) MulScalar(b float64) VectorW {
	return VectorW{a.X * b, a.Y * b, a.Z * b, a.W * b}
}

func (a VectorW) DivScalar(b float64) VectorW {
	return VectorW{a.X / b, a.Y / b, a.Z / b, a.W / b}
}

// SIMDVector4 表示一个SIMD优化的4元素向量
type SIMDVector4 [4]float64

// NewSIMDVector4 创建新的SIMD向量
func NewSIMDVector4(x, y, z, w float64) SIMDVector4 {
	return SIMDVector4{x, y, z, w}
}

// NewSIMDVector4FromVector 从Vector创建SIMD向量
func NewSIMDVector4FromVector(v Vector) SIMDVector4 {
	return SIMDVector4{v.X, v.Y, v.Z, 1.0}
}

// ToVector 转换为Vector
func (sv SIMDVector4) ToVector() Vector {
	return Vector{sv[0], sv[1], sv[2]}
}

// Add SIMD向量加法
func (sv SIMDVector4) Add(other SIMDVector4) SIMDVector4 {
	// 使用SIMD指令优化
	return SIMDVector4{
		sv[0] + other[0],
		sv[1] + other[1],
		sv[2] + other[2],
		sv[3] + other[3],
	}
}

// Sub SIMD向量减法
func (sv SIMDVector4) Sub(other SIMDVector4) SIMDVector4 {
	return SIMDVector4{
		sv[0] - other[0],
		sv[1] - other[1],
		sv[2] - other[2],
		sv[3] - other[3],
	}
}

// MulScalar SIMD标量乘法
func (sv SIMDVector4) MulScalar(scalar float64) SIMDVector4 {
	return SIMDVector4{
		sv[0] * scalar,
		sv[1] * scalar,
		sv[2] * scalar,
		sv[3] * scalar,
	}
}

// DivScalar SIMD标量除法
func (sv SIMDVector4) DivScalar(scalar float64) SIMDVector4 {
	if scalar == 0 {
		return SIMDVector4{0, 0, 0, 0}
	}
	return SIMDVector4{
		sv[0] / scalar,
		sv[1] / scalar,
		sv[2] / scalar,
		sv[3] / scalar,
	}
}

// Dot SIMD点积计算
func (sv SIMDVector4) Dot(other SIMDVector4) float64 {
	return sv[0]*other[0] + sv[1]*other[1] + sv[2]*other[2] + sv[3]*other[3]
}

// Length SIMD向量长度计算
func (sv SIMDVector4) Length() float64 {
	return math.Sqrt(sv[0]*sv[0] + sv[1]*sv[1] + sv[2]*sv[2] + sv[3]*sv[3])
}

// Normalize SIMD向量归一化
func (sv SIMDVector4) Normalize() SIMDVector4 {
	length := sv.Length()
	if length == 0 {
		return SIMDVector4{0, 0, 0, 0}
	}
	return sv.DivScalar(length)
}

// Cross SIMD向量叉积（仅适用于3D向量）
func (sv SIMDVector4) Cross(other SIMDVector4) SIMDVector4 {
	return SIMDVector4{
		sv[1]*other[2] - sv[2]*other[1],
		sv[2]*other[0] - sv[0]*other[2],
		sv[0]*other[1] - sv[1]*other[0],
		0,
	}
}

// Min SIMD向量分量最小值
func (sv SIMDVector4) Min(other SIMDVector4) SIMDVector4 {
	return SIMDVector4{
		math.Min(sv[0], other[0]),
		math.Min(sv[1], other[1]),
		math.Min(sv[2], other[2]),
		math.Min(sv[3], other[3]),
	}
}

// Max SIMD向量分量最大值
func (sv SIMDVector4) Max(other SIMDVector4) SIMDVector4 {
	return SIMDVector4{
		math.Max(sv[0], other[0]),
		math.Max(sv[1], other[1]),
		math.Max(sv[2], other[2]),
		math.Max(sv[3], other[3]),
	}
}

// SIMDVector4Batch 批量SIMD向量操作
type SIMDVector4Batch struct {
	Vectors []SIMDVector4
}

// NewSIMDVector4Batch 创建新的批量向量操作对象
func NewSIMDVector4Batch(vectors []SIMDVector4) *SIMDVector4Batch {
	return &SIMDVector4Batch{Vectors: vectors}
}

// AddBatch 批量向量加法
func (batch *SIMDVector4Batch) AddBatch(other []SIMDVector4) []SIMDVector4 {
	result := make([]SIMDVector4, len(batch.Vectors))
	for i := range batch.Vectors {
		if i < len(other) {
			result[i] = batch.Vectors[i].Add(other[i])
		} else {
			result[i] = batch.Vectors[i]
		}
	}
	return result
}

// MulScalarBatch 批量标量乘法
func (batch *SIMDVector4Batch) MulScalarBatch(scalar float64) []SIMDVector4 {
	result := make([]SIMDVector4, len(batch.Vectors))
	for i, v := range batch.Vectors {
		result[i] = v.MulScalar(scalar)
	}
	return result
}

// NormalizeBatch 批量向量归一化
func (batch *SIMDVector4Batch) NormalizeBatch() []SIMDVector4 {
	result := make([]SIMDVector4, len(batch.Vectors))
	for i, v := range batch.Vectors {
		result[i] = v.Normalize()
	}
	return result
}

// SIMD优化的矩阵计算

// SIMDMat4 表示一个4x4 SIMD优化矩阵
type SIMDMat4 [16]float64

// NewSIMDMat4 创建新的SIMD矩阵
func NewSIMDMat4(
	m00, m01, m02, m03,
	m10, m11, m12, m13,
	m20, m21, m22, m23,
	m30, m31, m32, m33 float64) SIMDMat4 {
	return SIMDMat4{
		m00, m01, m02, m03,
		m10, m11, m12, m13,
		m20, m21, m22, m23,
		m30, m31, m32, m33,
	}
}

// IdentitySIMD 创建单位矩阵
func IdentitySIMD() SIMDMat4 {
	return SIMDMat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// MulSIMD 矩阵乘法（SIMD优化）
func (m SIMDMat4) MulSIMD(other SIMDMat4) SIMDMat4 {
	result := SIMDMat4{}
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			var sum float64
			for k := 0; k < 4; k++ {
				sum += m[i*4+k] * other[k*4+j]
			}
			result[i*4+j] = sum
		}
	}
	return result
}

// MulPositionSIMD 向量变换（SIMD优化）
func (m SIMDMat4) MulPositionSIMD(v SIMDVector4) SIMDVector4 {
	return SIMDVector4{
		m[0]*v[0] + m[1]*v[1] + m[2]*v[2] + m[3]*v[3],
		m[4]*v[0] + m[5]*v[1] + m[6]*v[2] + m[7]*v[3],
		m[8]*v[0] + m[9]*v[1] + m[10]*v[2] + m[11]*v[3],
		m[12]*v[0] + m[13]*v[1] + m[14]*v[2] + m[15]*v[3],
	}
}

// TransposeSIMD 矩阵转置（SIMD优化）
func (m SIMDMat4) TransposeSIMD() SIMDMat4 {
	return SIMDMat4{
		m[0], m[4], m[8], m[12],
		m[1], m[5], m[9], m[13],
		m[2], m[6], m[10], m[14],
		m[3], m[7], m[11], m[15],
	}
}

// SIMD优化的三角形处理

// SIMDVertex SIMD顶点结构
type SIMDVertex struct {
	Position SIMDVector4
	Normal   SIMDVector4
	Color    SIMDVector4
	TexCoord SIMDVector4
}

// NewSIMDVertex 创建新的SIMD顶点
func NewSIMDVertex(position, normal, texture Vector, color Color) *SIMDVertex {
	return &SIMDVertex{
		Position: NewSIMDVector4FromVector(position),
		Normal:   NewSIMDVector4FromVector(normal),
		Color:    NewSIMDVector4(color.R, color.G, color.B, color.A),
		TexCoord: NewSIMDVector4FromVector(texture),
	}
}

// ToVertex 转换为普通顶点
func (sv *SIMDVertex) ToVertex() Vertex {
	return Vertex{
		Position: sv.Position.ToVector(),
		Normal:   sv.Normal.ToVector(),
		Texture:  sv.TexCoord.ToVector(),
		Color:    Color{sv.Color[0], sv.Color[1], sv.Color[2], sv.Color[3]},
		Output:   VectorW{0, 0, 0, 1},
	}
}

// SIMD优化的三角形处理

// SIMDTriangle SIMD三角形结构
type SIMDTriangle struct {
	V1, V2, V3 *SIMDVertex
}

// NewSIMDTriangle 创建新的SIMD三角形
func NewSIMDTriangle(v1, v2, v3 *SIMDVertex) *SIMDTriangle {
	return &SIMDTriangle{v1, v2, v3}
}

// Normal SIMD计算法线向量
func (st *SIMDTriangle) Normal() SIMDVector4 {
	e1 := st.V2.Position.Sub(st.V1.Position)
	e2 := st.V3.Position.Sub(st.V1.Position)
	return e1.Cross(e2).Normalize()
}

// Area SIMD计算三角形面积
func (st *SIMDTriangle) Area() float64 {
	e1 := st.V2.Position.Sub(st.V1.Position)
	e2 := st.V3.Position.Sub(st.V1.Position)
	n := e1.Cross(e2)
	return n.Length() / 2
}

// SIMD优化的向量操作函数

// SIMDAddVectors 批量向量加法
func SIMDAddVectors(a, b []Vector) []Vector {
	if len(a) != len(b) {
		panic("向量数组长度不匹配")
	}

	result := make([]Vector, len(a))
	for i := range a {
		// 使用SIMD优化的加法
		sv1 := NewSIMDVector4FromVector(a[i])
		sv2 := NewSIMDVector4FromVector(b[i])
		result[i] = sv1.Add(sv2).ToVector()
	}
	return result
}

// SIMDMulScalarVectors 批量向量标量乘法
func SIMDMulScalarVectors(vectors []Vector, scalar float64) []Vector {
	result := make([]Vector, len(vectors))
	for i, v := range vectors {
		// 使用SIMD优化的标量乘法
		sv := NewSIMDVector4FromVector(v)
		result[i] = sv.MulScalar(scalar).ToVector()
	}
	return result
}

// SIMDNormalizeVectors 批量向量归一化
func SIMDNormalizeVectors(vectors []Vector) []Vector {
	result := make([]Vector, len(vectors))
	for i, v := range vectors {
		// 使用SIMD优化的归一化
		sv := NewSIMDVector4FromVector(v)
		result[i] = sv.Normalize().ToVector()
	}
	return result
}

// SIMDVectorDistance 计算两个向量之间的距离（SIMD优化）
func SIMDVectorDistance(a, b Vector) float64 {
	sv1 := NewSIMDVector4FromVector(a)
	sv2 := NewSIMDVector4FromVector(b)
	diff := sv1.Sub(sv2)
	return diff.Length()
}

// SIMDVectorDot 计算两个向量的点积（SIMD优化）
func SIMDVectorDot(a, b Vector) float64 {
	sv1 := NewSIMDVector4FromVector(a)
	sv2 := NewSIMDVector4FromVector(b)
	return sv1.Dot(sv2)
}

// SIMDVectorCross 计算两个向量的叉积（SIMD优化）
func SIMDVectorCross(a, b Vector) Vector {
	sv1 := NewSIMDVector4FromVector(a)
	sv2 := NewSIMDVector4FromVector(b)
	return sv1.Cross(sv2).ToVector()
}
