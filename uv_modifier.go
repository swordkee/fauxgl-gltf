package fauxgl

import (
	"fmt"
	"math"
)

// UVTransform represents a UV coordinate transformation
type UVTransform struct {
	// Basic transformations
	OffsetU, OffsetV float64 // UV偏移
	ScaleU, ScaleV   float64 // UV缩放
	Rotation         float64 // 旋转角度(弧度)

	// Advanced transformations
	SkewU, SkewV   float64 // UV剪切
	PivotU, PivotV float64 // 旋转中心点

	// Animation parameters
	AnimationTime float64 // 动画时间
	RotationSpeed float64 // 旋转速度
	ScrollSpeedU  float64 // U方向滚动速度
	ScrollSpeedV  float64 // V方向滚动速度
}

// NewUVTransform creates a new UV transform with default values
func NewUVTransform() *UVTransform {
	return &UVTransform{
		ScaleU: 1.0,
		ScaleV: 1.0,
		PivotU: 0.5,
		PivotV: 0.5,
	}
}

// UVMapping defines a UV mapping configuration
type UVMapping struct {
	Name      string       // 映射名称
	Enabled   bool         // 是否启用
	Region    UVRegion     // UV区域定义
	Transform *UVTransform // UV变换
	BlendMode UVBlendMode  // 混合模式
	Priority  int          // 优先级(数值越高优先级越高)
}

// UVRegion defines a UV coordinate region
type UVRegion struct {
	MinU, MaxU float64    // U坐标范围 [0.0, 1.0]
	MinV, MaxV float64    // V坐标范围 [0.0, 1.0]
	MaskType   UVMaskType // 区域遮罩类型
}

// UVBlendMode defines how UV transformations are blended
type UVBlendMode int

const (
	UVBlendReplace  UVBlendMode = iota // 替换模式
	UVBlendAdd                         // 加法混合
	UVBlendMultiply                    // 乘法混合
	UVBlendOverlay                     // 叠加模式
)

// UVMaskType defines the type of UV region mask
type UVMaskType int

const (
	UVMaskRectangle UVMaskType = iota // 矩形遮罩
	UVMaskCircle                      // 圆形遮罩
	UVMaskGradient                    // 渐变遮罩
)

// UVModifier provides dynamic UV coordinate modification capabilities
type UVModifier struct {
	mappings         []*UVMapping
	globalTransform  *UVTransform
	animationEnabled bool
}

// NewUVModifier creates a new UV modifier
func NewUVModifier() *UVModifier {
	return &UVModifier{
		mappings:         make([]*UVMapping, 0),
		globalTransform:  NewUVTransform(),
		animationEnabled: false,
	}
}

// AddMapping adds a UV mapping configuration
func (modifier *UVModifier) AddMapping(mapping *UVMapping) {
	modifier.mappings = append(modifier.mappings, mapping)
	// 按优先级排序(高优先级在前)
	for i := len(modifier.mappings) - 1; i > 0; i-- {
		if modifier.mappings[i].Priority > modifier.mappings[i-1].Priority {
			modifier.mappings[i], modifier.mappings[i-1] = modifier.mappings[i-1], modifier.mappings[i]
		}
	}
}

// RemoveMapping removes a mapping by name
func (modifier *UVModifier) RemoveMapping(name string) bool {
	for i, mapping := range modifier.mappings {
		if mapping.Name == name {
			modifier.mappings = append(modifier.mappings[:i], modifier.mappings[i+1:]...)
			return true
		}
	}
	return false
}

// GetMapping gets a mapping by name
func (modifier *UVModifier) GetMapping(name string) *UVMapping {
	for _, mapping := range modifier.mappings {
		if mapping.Name == name {
			return mapping
		}
	}
	return nil
}

// SetGlobalTransform sets the global UV transformation
func (modifier *UVModifier) SetGlobalTransform(transform *UVTransform) {
	modifier.globalTransform = transform
}

// EnableAnimation enables/disables UV animation
func (modifier *UVModifier) EnableAnimation(enabled bool) {
	modifier.animationEnabled = enabled
}

// UpdateAnimation updates animation time for all transforms
func (modifier *UVModifier) UpdateAnimation(deltaTime float64) {
	if !modifier.animationEnabled {
		return
	}

	// 更新全局变换动画
	modifier.updateTransformAnimation(modifier.globalTransform, deltaTime)

	// 更新所有映射的动画
	for _, mapping := range modifier.mappings {
		if mapping.Enabled && mapping.Transform != nil {
			modifier.updateTransformAnimation(mapping.Transform, deltaTime)
		}
	}
}

// updateTransformAnimation updates a single transform's animation
func (modifier *UVModifier) updateTransformAnimation(transform *UVTransform, deltaTime float64) {
	transform.AnimationTime += deltaTime

	// 更新旋转
	if transform.RotationSpeed != 0 {
		transform.Rotation += transform.RotationSpeed * deltaTime
		// 保持在[0, 2π]范围内
		transform.Rotation = math.Mod(transform.Rotation, 2*math.Pi)
	}

	// 更新滚动
	if transform.ScrollSpeedU != 0 {
		transform.OffsetU += transform.ScrollSpeedU * deltaTime
	}
	if transform.ScrollSpeedV != 0 {
		transform.OffsetV += transform.ScrollSpeedV * deltaTime
	}
}

// TransformUV applies all UV transformations to input coordinates
func (modifier *UVModifier) TransformUV(u, v float64) (float64, float64) {
	// 首先应用全局变换
	transformedU, transformedV := modifier.applyTransform(u, v, modifier.globalTransform)

	// 然后按优先级应用所有启用的映射变换
	for _, mapping := range modifier.mappings {
		if !mapping.Enabled || mapping.Transform == nil {
			continue
		}

		// 检查是否在映射区域内
		if !modifier.isInRegion(transformedU, transformedV, &mapping.Region) {
			continue
		}

		// 计算区域权重
		weight := modifier.calculateRegionWeight(transformedU, transformedV, &mapping.Region)

		// 应用变换
		newU, newV := modifier.applyTransform(transformedU, transformedV, mapping.Transform)

		// 根据混合模式合并结果
		transformedU, transformedV = modifier.blendUV(
			transformedU, transformedV,
			newU, newV,
			mapping.BlendMode,
			weight,
		)
	}

	return transformedU, transformedV
}

// applyTransform applies a single UV transform
func (modifier *UVModifier) applyTransform(u, v float64, transform *UVTransform) (float64, float64) {
	if transform == nil {
		return u, v
	}

	// 1. 移动到旋转中心
	u -= transform.PivotU
	v -= transform.PivotV

	// 2. 应用缩放
	u *= transform.ScaleU
	v *= transform.ScaleV

	// 3. 应用剪切
	u += v * transform.SkewU
	v += u * transform.SkewV

	// 4. 应用旋转
	if transform.Rotation != 0 {
		cos := math.Cos(transform.Rotation)
		sin := math.Sin(transform.Rotation)
		newU := u*cos - v*sin
		newV := u*sin + v*cos
		u, v = newU, newV
	}

	// 5. 移回原位置
	u += transform.PivotU
	v += transform.PivotV

	// 6. 应用偏移
	u += transform.OffsetU
	v += transform.OffsetV

	return u, v
}

// isInRegion checks if UV coordinates are within a region
func (modifier *UVModifier) isInRegion(u, v float64, region *UVRegion) bool {
	switch region.MaskType {
	case UVMaskRectangle:
		return u >= region.MinU && u <= region.MaxU &&
			v >= region.MinV && v <= region.MaxV

	case UVMaskCircle:
		centerU := (region.MinU + region.MaxU) * 0.5
		centerV := (region.MinV + region.MaxV) * 0.5
		radiusU := (region.MaxU - region.MinU) * 0.5
		radiusV := (region.MaxV - region.MinV) * 0.5

		// 椭圆内部检测
		du := (u - centerU) / radiusU
		dv := (v - centerV) / radiusV
		return du*du+dv*dv <= 1.0

	case UVMaskGradient:
		// 渐变遮罩：始终返回true，权重在calculateRegionWeight中计算
		return true

	default:
		return true
	}
}

// calculateRegionWeight calculates the influence weight for a region
func (modifier *UVModifier) calculateRegionWeight(u, v float64, region *UVRegion) float64 {
	switch region.MaskType {
	case UVMaskRectangle:
		return 1.0 // 矩形区域内权重为1

	case UVMaskCircle:
		centerU := (region.MinU + region.MaxU) * 0.5
		centerV := (region.MinV + region.MaxV) * 0.5
		radiusU := (region.MaxU - region.MinU) * 0.5
		radiusV := (region.MaxV - region.MinV) * 0.5

		du := (u - centerU) / radiusU
		dv := (v - centerV) / radiusV
		distance := math.Sqrt(du*du + dv*dv)

		// 边缘软化
		if distance <= 1.0 {
			return 1.0 - distance*distance // 二次衰减
		}
		return 0.0

	case UVMaskGradient:
		// 线性渐变：从MinU到MaxU
		if u <= region.MinU {
			return 0.0
		} else if u >= region.MaxU {
			return 1.0
		} else {
			return (u - region.MinU) / (region.MaxU - region.MinU)
		}

	default:
		return 1.0
	}
}

// blendUV blends two UV coordinate sets using the specified blend mode
func (modifier *UVModifier) blendUV(u1, v1, u2, v2 float64, mode UVBlendMode, weight float64) (float64, float64) {
	switch mode {
	case UVBlendReplace:
		return u1*(1-weight) + u2*weight, v1*(1-weight) + v2*weight

	case UVBlendAdd:
		return u1 + u2*weight, v1 + v2*weight

	case UVBlendMultiply:
		return u1*(1-weight) + u1*u2*weight, v1*(1-weight) + v1*v2*weight

	case UVBlendOverlay:
		// 覆盖模式：基于原始值决定是否使用multiply或screen
		overlayU := u1
		overlayV := v1
		if u1 < 0.5 {
			overlayU = 2 * u1 * u2
		} else {
			overlayU = 1 - 2*(1-u1)*(1-u2)
		}
		if v1 < 0.5 {
			overlayV = 2 * v1 * v2
		} else {
			overlayV = 1 - 2*(1-v1)*(1-v2)
		}
		return u1*(1-weight) + overlayU*weight, v1*(1-weight) + overlayV*weight

	default:
		return u1, v1
	}
}

// CreateScrollingUVMapping creates a scrolling UV animation mapping
func CreateScrollingUVMapping(name string, speedU, speedV float64, region UVRegion) *UVMapping {
	transform := NewUVTransform()
	transform.ScrollSpeedU = speedU
	transform.ScrollSpeedV = speedV

	return &UVMapping{
		Name:      name,
		Enabled:   true,
		Region:    region,
		Transform: transform,
		BlendMode: UVBlendAdd,
		Priority:  1,
	}
}

// CreateRotatingUVMapping creates a rotating UV animation mapping
func CreateRotatingUVMapping(name string, speed float64, pivotU, pivotV float64, region UVRegion) *UVMapping {
	transform := NewUVTransform()
	transform.RotationSpeed = speed
	transform.PivotU = pivotU
	transform.PivotV = pivotV

	return &UVMapping{
		Name:      name,
		Enabled:   true,
		Region:    region,
		Transform: transform,
		BlendMode: UVBlendReplace,
		Priority:  2,
	}
}

// CreateScalingUVMapping creates a scaling UV mapping
func CreateScalingUVMapping(name string, scaleU, scaleV float64, region UVRegion) *UVMapping {
	transform := NewUVTransform()
	transform.ScaleU = scaleU
	transform.ScaleV = scaleV

	return &UVMapping{
		Name:      name,
		Enabled:   true,
		Region:    region,
		Transform: transform,
		BlendMode: UVBlendReplace,
		Priority:  0,
	}
}

// ApplyUVModifierToMesh applies UV modifications to all triangles in a mesh
func ApplyUVModifierToMesh(mesh *Mesh, modifier *UVModifier) {
	if modifier == nil {
		return
	}

	for _, triangle := range mesh.Triangles {
		// 修改第一个顶点的UV坐标
		triangle.V1.Texture.X, triangle.V1.Texture.Y = modifier.TransformUV(
			triangle.V1.Texture.X, triangle.V1.Texture.Y,
		)

		// 修改第二个顶点的UV坐标
		triangle.V2.Texture.X, triangle.V2.Texture.Y = modifier.TransformUV(
			triangle.V2.Texture.X, triangle.V2.Texture.Y,
		)

		// 修改第三个顶点的UV坐标
		triangle.V3.Texture.X, triangle.V3.Texture.Y = modifier.TransformUV(
			triangle.V3.Texture.X, triangle.V3.Texture.Y,
		)
	}
}

// PrintUVModifierInfo prints detailed information about a UV modifier
func PrintUVModifierInfo(modifier *UVModifier) {
	fmt.Println("=== UV Modifier Information ===")
	fmt.Printf("Animation Enabled: %t\n", modifier.animationEnabled)
	fmt.Printf("Global Transform: Offset(%.3f, %.3f) Scale(%.3f, %.3f) Rotation(%.3f)\n",
		modifier.globalTransform.OffsetU, modifier.globalTransform.OffsetV,
		modifier.globalTransform.ScaleU, modifier.globalTransform.ScaleV,
		modifier.globalTransform.Rotation)

	fmt.Printf("Mappings Count: %d\n", len(modifier.mappings))
	for i, mapping := range modifier.mappings {
		fmt.Printf("  %d. %s (Priority: %d, Enabled: %t)\n",
			i+1, mapping.Name, mapping.Priority, mapping.Enabled)
		if mapping.Transform != nil {
			fmt.Printf("     Transform: Offset(%.3f, %.3f) Scale(%.3f, %.3f)\n",
				mapping.Transform.OffsetU, mapping.Transform.OffsetV,
				mapping.Transform.ScaleU, mapping.Transform.ScaleV)
		}
		fmt.Printf("     Region: U(%.3f-%.3f) V(%.3f-%.3f)\n",
			mapping.Region.MinU, mapping.Region.MaxU,
			mapping.Region.MinV, mapping.Region.MaxV)
	}
}
