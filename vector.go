package fauxgl

import (
	"math"
)

type Vector struct {
	X, Y, Z float64
}

func V(x, y, z float64) Vector {
	return Vector{x, y, z}
}

func (a Vector) VectorW() VectorW {
	return VectorW{a.X, a.Y, a.Z, 1}
}

func (a Vector) IsDegenerate() bool {
	nan := math.IsNaN(a.X) || math.IsNaN(a.Y) || math.IsNaN(a.Z)
	inf := math.IsInf(a.X, 0) || math.IsInf(a.Y, 0) || math.IsInf(a.Z, 0)
	return nan || inf
}

func (a Vector) Length() float64 {
	// 使用SIMD优化的长度计算
	sv := NewSIMDVector4FromVector(a)
	return sv.Length()
}

func (a Vector) Less(b Vector) bool {
	if a.X != b.X {
		return a.X < b.X
	}
	if a.Y != b.Y {
		return a.Y < b.Y
	}
	return a.Z < b.Z
}

func (a Vector) Distance(b Vector) float64 {
	// 使用SIMD优化的距离计算
	return SIMDVectorDistance(a, b)
}

func (a Vector) LengthSquared() float64 {
	return a.X*a.X + a.Y*a.Y + a.Z*a.Z
}

func (a Vector) DistanceSquared(b Vector) float64 {
	return a.Sub(b).LengthSquared()
}

func (a Vector) Lerp(b Vector, t float64) Vector {
	// 使用SIMD优化的线性插值
	sv1 := NewSIMDVector4FromVector(a)
	sv2 := NewSIMDVector4FromVector(b)
	diff := sv2.Sub(sv1).MulScalar(t)
	return sv1.Add(diff).ToVector()
}

func (a Vector) LerpDistance(b Vector, d float64) Vector {
	return a.Add(b.Sub(a).Normalize().MulScalar(d))
}

func (a Vector) Dot(b Vector) float64 {
	// 使用SIMD优化的点积计算
	return SIMDVectorDot(a, b)
}

func (a Vector) Cross(b Vector) Vector {
	// 使用SIMD优化的叉积计算
	return SIMDVectorCross(a, b)
}

func (a Vector) Normalize() Vector {
	// 使用SIMD优化的归一化
	sv := NewSIMDVector4FromVector(a)
	return sv.Normalize().ToVector()
}

func (a Vector) Negate() Vector {
	return Vector{-a.X, -a.Y, -a.Z}
}

func (a Vector) Abs() Vector {
	return Vector{math.Abs(a.X), math.Abs(a.Y), math.Abs(a.Z)}
}

func (a Vector) Add(b Vector) Vector {
	// 使用SIMD优化的向量加法
	return SIMDAddVectors([]Vector{a}, []Vector{b})[0]
}

func (a Vector) Sub(b Vector) Vector {
	// 使用SIMD优化的向量减法
	sv1 := NewSIMDVector4FromVector(a)
	sv2 := NewSIMDVector4FromVector(b)
	return sv1.Sub(sv2).ToVector()
}

func (a Vector) Mul(b Vector) Vector {
	return Vector{a.X * b.X, a.Y * b.Y, a.Z * b.Z}
}

func (a Vector) Div(b Vector) Vector {
	return Vector{a.X / b.X, a.Y / b.Y, a.Z / b.Z}
}

func (a Vector) AddScalar(b float64) Vector {
	return Vector{a.X + b, a.Y + b, a.Z + b}
}

func (a Vector) SubScalar(b float64) Vector {
	return Vector{a.X - b, a.Y - b, a.Z - b}
}

func (a Vector) MulScalar(b float64) Vector {
	// 使用SIMD优化的标量乘法
	return SIMDMulScalarVectors([]Vector{a}, b)[0]
}

func (a Vector) DivScalar(b float64) Vector {
	return Vector{a.X / b, a.Y / b, a.Z / b}
}

func (a Vector) PowScalar(b float64) Vector {
	return Vector{math.Pow(a.X, b), math.Pow(a.Y, b), math.Pow(a.Z, b)}
}

func (a Vector) Min(b Vector) Vector {
	// 使用SIMD优化的分量最小值
	sv1 := NewSIMDVector4FromVector(a)
	sv2 := NewSIMDVector4FromVector(b)
	return sv1.Min(sv2).ToVector()
}

func (a Vector) Max(b Vector) Vector {
	// 使用SIMD优化的分量最大值
	sv1 := NewSIMDVector4FromVector(a)
	sv2 := NewSIMDVector4FromVector(b)
	return sv1.Max(sv2).ToVector()
}

func (a Vector) Floor() Vector {
	return Vector{math.Floor(a.X), math.Floor(a.Y), math.Floor(a.Z)}
}

func (a Vector) Ceil() Vector {
	return Vector{math.Ceil(a.X), math.Ceil(a.Y), math.Ceil(a.Z)}
}

func (a Vector) Round() Vector {
	return a.RoundPlaces(0)
}

func (a Vector) RoundPlaces(n int) Vector {
	x := RoundPlaces(a.X, n)
	y := RoundPlaces(a.Y, n)
	z := RoundPlaces(a.Z, n)
	return Vector{x, y, z}
}

func (a Vector) MinComponent() float64 {
	return math.Min(math.Min(a.X, a.Y), a.Z)
}

func (a Vector) MaxComponent() float64 {
	return math.Max(math.Max(a.X, a.Y), a.Z)
}

func (a Vector) Reflect(n Vector) Vector {
	return a.Sub(n.MulScalar(2 * n.Dot(a)))
}

func (a Vector) Perpendicular() Vector {
	if a.X == 0 && a.Y == 0 {
		if a.Z == 0 {
			return Vector{}
		}
		return Vector{0, 1, 0}
	}
	return Vector{-a.Y, a.X, 0}.Normalize()
}

func (a Vector) SegmentDistance(v Vector, w Vector) float64 {
	l2 := v.DistanceSquared(w)
	if l2 == 0 {
		return a.Distance(v)
	}
	t := a.Sub(v).Dot(w.Sub(v)) / l2
	if t < 0 {
		return a.Distance(v)
	}
	if t > 1 {
		return a.Distance(w)
	}
	return v.Add(w.Sub(v).MulScalar(t)).Distance(a)
}
