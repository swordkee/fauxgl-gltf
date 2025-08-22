package fauxgl

import (
	"math"
)

// PBRWorkflow represents the type of PBR workflow
type PBRWorkflow int

const (
	// MetallicRoughness is the metallic-roughness workflow
	MetallicRoughness PBRWorkflow = iota
	// SpecularGlossiness is the specular-glossiness workflow (legacy)
	SpecularGlossiness
)

// PBRMaterial represents a physically-based rendering material
type PBRMaterial struct {
	// Base color and alpha
	BaseColorFactor  Color
	BaseColorTexture Texture

	// Metallic-Roughness workflow
	MetallicFactor           float64
	RoughnessFactor          float64
	MetallicRoughnessTexture Texture

	// Specular-Glossiness workflow (legacy support)
	DiffuseFactor             Color
	SpecularFactor            Color
	GlossinessFactor          float64
	DiffuseTexture            Texture
	SpecularGlossinessTexture Texture

	// Normal mapping
	NormalTexture Texture
	NormalScale   float64

	// Occlusion mapping
	OcclusionTexture  Texture
	OcclusionStrength float64

	// Emissive mapping
	EmissiveFactor  Color
	EmissiveTexture Texture

	// Extended material properties (GLTF Extensions)
	// KHR_materials_emissive_strength
	EmissiveStrength float64

	// KHR_materials_ior
	IOR float64 // Index of Refraction

	// KHR_materials_specular
	SpecularColorFactor  Color
	SpecularColorTexture Texture
	SpecularTexture      Texture

	// KHR_materials_transmission
	TransmissionFactor  float64
	TransmissionTexture Texture

	// KHR_materials_volume
	ThicknessFactor     float64
	ThicknessTexture    Texture
	AttenuationDistance float64
	AttenuationColor    Color

	// KHR_materials_anisotropy
	AnisotropyStrength float64
	AnisotropyRotation float64
	AnisotropyTexture  Texture

	// KHR_materials_sheen
	SheenColorFactor      Color
	SheenRoughnessFactor  float64
	SheenColorTexture     Texture
	SheenRoughnessTexture Texture

	// KHR_materials_iridescence
	IridescenceFactor           float64
	IridescenceIor              float64
	IridescenceThicknessMinimum float64
	IridescenceThicknessMaximum float64
	IridescenceTexture          Texture
	IridescenceThicknessTexture Texture

	// KHR_materials_dispersion
	DispersionFactor float64

	// KHR_materials_clearcoat
	ClearcoatFactor           float64
	ClearcoatRoughnessFactor  float64
	ClearcoatTexture          Texture
	ClearcoatRoughnessTexture Texture
	ClearcoatNormalTexture    Texture

	// Additional properties
	AlphaCutoff float64
	AlphaMode   AlphaMode
	DoubleSided bool
	Workflow    PBRWorkflow
}

// AlphaMode represents how alpha blending should be handled
type AlphaMode int

const (
	// AlphaOpaque - fully opaque
	AlphaOpaque AlphaMode = iota
	// AlphaMask - alpha testing
	AlphaMask
	// AlphaBlend - alpha blending
	AlphaBlend
)

// NewPBRMaterial creates a new PBR material with default values
func NewPBRMaterial() *PBRMaterial {
	return &PBRMaterial{
		BaseColorFactor:   Color{1, 1, 1, 1},
		MetallicFactor:    1.0,
		RoughnessFactor:   1.0,
		NormalScale:       1.0,
		OcclusionStrength: 1.0,
		EmissiveFactor:    Color{0, 0, 0, 1},

		// Extended properties defaults
		EmissiveStrength:    1.0,               // KHR_materials_emissive_strength
		IOR:                 1.5,               // KHR_materials_ior (typical for glass/plastic)
		SpecularColorFactor: Color{1, 1, 1, 1}, // KHR_materials_specular
		TransmissionFactor:  0.0,               // KHR_materials_transmission (opaque by default)
		ThicknessFactor:     0.0,               // KHR_materials_volume
		AttenuationDistance: math.Inf(1),       // Infinite attenuation distance
		AttenuationColor:    Color{1, 1, 1, 1}, // White attenuation

		// Anisotropy defaults
		AnisotropyStrength: 0.0, // No anisotropy by default
		AnisotropyRotation: 0.0, // No rotation

		// Sheen defaults
		SheenColorFactor:     Color{0, 0, 0, 1}, // No sheen by default
		SheenRoughnessFactor: 0.0,

		// Iridescence defaults
		IridescenceFactor:           0.0,   // No iridescence by default
		IridescenceIor:              1.3,   // Typical iridescence IOR
		IridescenceThicknessMinimum: 100.0, // nm
		IridescenceThicknessMaximum: 400.0, // nm

		// Dispersion defaults
		DispersionFactor: 0.0, // No dispersion by default

		// Clearcoat defaults
		ClearcoatFactor:          0.0, // No clearcoat by default
		ClearcoatRoughnessFactor: 0.0,

		AlphaCutoff: 0.5,
		AlphaMode:   AlphaOpaque,
		DoubleSided: false,
		Workflow:    MetallicRoughness,
	}
}

// Sample samples the material at given texture coordinates
func (m *PBRMaterial) Sample(u, v float64) *SampledMaterial {
	result := &SampledMaterial{}

	// Sample base color
	result.BaseColor = m.BaseColorFactor
	if m.BaseColorTexture != nil {
		textureColor := m.BaseColorTexture.BilinearSample(u, v)
		result.BaseColor = result.BaseColor.Mul(textureColor)
	}

	// Sample metallic and roughness
	result.Metallic = m.MetallicFactor
	result.Roughness = m.RoughnessFactor
	if m.MetallicRoughnessTexture != nil {
		mr := m.MetallicRoughnessTexture.BilinearSample(u, v)
		result.Metallic *= mr.B  // Blue channel for metallic
		result.Roughness *= mr.G // Green channel for roughness
	}

	// Sample normal
	result.Normal = Vector{0, 0, 1} // Default normal in tangent space
	if m.NormalTexture != nil {
		normalColor := m.NormalTexture.BilinearSample(u, v)
		// Convert from [0,1] to [-1,1] range
		result.Normal = Vector{
			(normalColor.R*2.0 - 1.0) * m.NormalScale,
			(normalColor.G*2.0 - 1.0) * m.NormalScale,
			normalColor.B*2.0 - 1.0,
		}.Normalize()
	}

	// Sample occlusion
	result.Occlusion = 1.0
	if m.OcclusionTexture != nil {
		occlusionColor := m.OcclusionTexture.BilinearSample(u, v)
		result.Occlusion = 1.0 - (1.0-occlusionColor.R)*m.OcclusionStrength
	}

	// Sample emissive
	result.Emissive = m.EmissiveFactor
	if m.EmissiveTexture != nil {
		emissiveColor := m.EmissiveTexture.BilinearSample(u, v)
		result.Emissive = result.Emissive.Mul(emissiveColor)
	}

	// Sample extended properties
	result.EmissiveStrength = m.EmissiveStrength
	result.IOR = m.IOR

	// Sample specular color (KHR_materials_specular)
	result.SpecularColor = m.SpecularColorFactor
	if m.SpecularColorTexture != nil {
		specularColor := m.SpecularColorTexture.BilinearSample(u, v)
		result.SpecularColor = result.SpecularColor.Mul(specularColor)
	}

	// Sample transmission (KHR_materials_transmission)
	result.Transmission = m.TransmissionFactor
	if m.TransmissionTexture != nil {
		transmissionColor := m.TransmissionTexture.BilinearSample(u, v)
		result.Transmission *= transmissionColor.R // Red channel for transmission
	}

	// Sample thickness (KHR_materials_volume)
	result.Thickness = m.ThicknessFactor
	if m.ThicknessTexture != nil {
		thicknessColor := m.ThicknessTexture.BilinearSample(u, v)
		result.Thickness *= thicknessColor.G // Green channel for thickness
	}
	result.AttenuationColor = m.AttenuationColor
	result.AttenuationDistance = m.AttenuationDistance

	// Sample anisotropy (KHR_materials_anisotropy)
	result.AnisotropyStrength = m.AnisotropyStrength
	result.AnisotropyRotation = m.AnisotropyRotation
	if m.AnisotropyTexture != nil {
		anisotropyColor := m.AnisotropyTexture.BilinearSample(u, v)
		result.AnisotropyStrength *= anisotropyColor.R
		result.AnisotropyRotation += (anisotropyColor.G*2.0 - 1.0) * math.Pi
	}

	// Sample sheen (KHR_materials_sheen)
	result.SheenColor = m.SheenColorFactor
	if m.SheenColorTexture != nil {
		sheenColor := m.SheenColorTexture.BilinearSample(u, v)
		result.SheenColor = result.SheenColor.Mul(sheenColor)
	}
	result.SheenRoughness = m.SheenRoughnessFactor
	if m.SheenRoughnessTexture != nil {
		sheenRoughnessColor := m.SheenRoughnessTexture.BilinearSample(u, v)
		result.SheenRoughness *= sheenRoughnessColor.A
	}

	// Sample iridescence (KHR_materials_iridescence)
	result.Iridescence = m.IridescenceFactor
	if m.IridescenceTexture != nil {
		iridescenceColor := m.IridescenceTexture.BilinearSample(u, v)
		result.Iridescence *= iridescenceColor.R
	}
	result.IridescenceIor = m.IridescenceIor

	// Calculate iridescence thickness
	thicknessRange := m.IridescenceThicknessMaximum - m.IridescenceThicknessMinimum
	result.IridescenceThickness = m.IridescenceThicknessMinimum
	if m.IridescenceThicknessTexture != nil {
		thicknessColor := m.IridescenceThicknessTexture.BilinearSample(u, v)
		result.IridescenceThickness += thicknessColor.G * thicknessRange
	} else {
		result.IridescenceThickness += thicknessRange * 0.5 // Use middle value
	}

	// Sample dispersion (KHR_materials_dispersion)
	result.Dispersion = m.DispersionFactor

	// Sample clearcoat (KHR_materials_clearcoat)
	result.Clearcoat = m.ClearcoatFactor
	if m.ClearcoatTexture != nil {
		clearcoatColor := m.ClearcoatTexture.BilinearSample(u, v)
		result.Clearcoat *= clearcoatColor.R
	}
	result.ClearcoatRoughness = m.ClearcoatRoughnessFactor
	if m.ClearcoatRoughnessTexture != nil {
		clearcoatRoughnessColor := m.ClearcoatRoughnessTexture.BilinearSample(u, v)
		result.ClearcoatRoughness *= clearcoatRoughnessColor.G
	}

	// Sample clearcoat normal
	result.ClearcoatNormal = Vector{0, 0, 1} // Default normal
	if m.ClearcoatNormalTexture != nil {
		clearcoatNormalColor := m.ClearcoatNormalTexture.BilinearSample(u, v)
		result.ClearcoatNormal = Vector{
			clearcoatNormalColor.R*2.0 - 1.0,
			clearcoatNormalColor.G*2.0 - 1.0,
			clearcoatNormalColor.B*2.0 - 1.0,
		}.Normalize()
	}

	return result
}

// SampledMaterial represents material properties sampled at a specific point
type SampledMaterial struct {
	BaseColor Color
	Metallic  float64
	Roughness float64
	Normal    Vector
	Occlusion float64
	Emissive  Color

	// Extended properties
	EmissiveStrength    float64
	IOR                 float64
	SpecularColor       Color
	Transmission        float64
	Thickness           float64
	AttenuationColor    Color
	AttenuationDistance float64

	// Advanced material properties
	AnisotropyStrength   float64
	AnisotropyRotation   float64
	SheenColor           Color
	SheenRoughness       float64
	Iridescence          float64
	IridescenceIor       float64
	IridescenceThickness float64
	Dispersion           float64
	Clearcoat            float64
	ClearcoatRoughness   float64
	ClearcoatNormal      Vector
}

// Light represents a light source
type Light struct {
	Type      LightType
	Position  Vector
	Direction Vector
	Color     Color
	Intensity float64
	Range     float64
	InnerCone float64 // For spot lights
	OuterCone float64 // For spot lights
}

// LightType represents the type of light
type LightType int

const (
	// DirectionalLight - parallel rays (sun)
	DirectionalLight LightType = iota
	// PointLight - omnidirectional point light
	PointLight
	// SpotLight - cone-shaped light
	SpotLight
	// AmbientLight - uniform ambient lighting
	AmbientLight
)

// PBRLighting contains PBR lighting calculation functions
type PBRLighting struct{}

// CalculatePBR performs PBR lighting calculation
func (pbrL *PBRLighting) CalculatePBR(
	material *SampledMaterial,
	worldPos Vector,
	worldNormal Vector,
	viewDir Vector,
	lights []Light,
	ambientColor Color,
) Color {
	// Convert roughness to alpha for calculations
	alpha := material.Roughness * material.Roughness

	// Calculate F0 (base reflectance)
	f0 := Vector{0.04, 0.04, 0.04} // Non-metallic base reflectance
	if material.Metallic > 0 {
		// Metallic materials use base color as F0
		metallic := Vector{material.BaseColor.R, material.BaseColor.G, material.BaseColor.B}
		f0 = f0.Lerp(metallic, material.Metallic)
	}

	// Check if we have ambient lights in the light array
	hasAmbientLights := false
	for _, light := range lights {
		if light.Type == AmbientLight {
			hasAmbientLights = true
			break
		}
	}

	// Initialize final color with emissive
	finalColor := material.Emissive

	// Add legacy ambient color only if no AmbientLight sources are present
	if !hasAmbientLights && (ambientColor.R > 0 || ambientColor.G > 0 || ambientColor.B > 0) {
		ambientContrib := material.BaseColor.Mul(ambientColor).MulScalar(material.Occlusion)
		finalColor = finalColor.Add(ambientContrib)
	}

	// Process each light
	for _, light := range lights {
		lightContrib := pbrL.calculateLightContribution(
			material, worldPos, worldNormal, viewDir, light, f0, alpha)
		finalColor = finalColor.Add(lightContrib)
	}

	return finalColor
}

// calculateLightContribution calculates the contribution of a single light
func (pbrL *PBRLighting) calculateLightContribution(
	material *SampledMaterial,
	worldPos Vector,
	normal Vector,
	viewDir Vector,
	light Light,
	f0 Vector,
	alpha float64,
) Color {
	var lightDir Vector
	var lightColor Color
	var attenuation float64 = 1.0

	switch light.Type {
	case DirectionalLight:
		lightDir = light.Direction.Negate().Normalize()
		lightColor = light.Color.MulScalar(light.Intensity)

	case PointLight:
		lightVec := light.Position.Sub(worldPos)
		distance := lightVec.Length()
		lightDir = lightVec.Normalize()

		// Distance attenuation
		if light.Range > 0 {
			attenuation = math.Max(0, 1.0-(distance/light.Range))
			attenuation = attenuation * attenuation
		}
		lightColor = light.Color.MulScalar(light.Intensity * attenuation)

	case SpotLight:
		lightVec := light.Position.Sub(worldPos)
		distance := lightVec.Length()
		lightDir = lightVec.Normalize()

		// Distance attenuation
		if light.Range > 0 {
			attenuation = math.Max(0, 1.0-(distance/light.Range))
			attenuation = attenuation * attenuation
		}

		// Spot cone attenuation
		spotEffect := lightDir.Dot(light.Direction.Negate())
		innerCos := math.Cos(light.InnerCone)
		outerCos := math.Cos(light.OuterCone)

		if spotEffect < outerCos {
			attenuation = 0
		} else if spotEffect > innerCos {
			attenuation *= 1.0
		} else {
			attenuation *= (spotEffect - outerCos) / (innerCos - outerCos)
		}

		lightColor = light.Color.MulScalar(light.Intensity * attenuation)

	case AmbientLight:
		// Ambient light provides uniform illumination to all surfaces
		// It contributes directly to the base color without BRDF calculations
		ambientContrib := material.BaseColor.Mul(light.Color).MulScalar(light.Intensity * material.Occlusion)
		return Color{ambientContrib.R, ambientContrib.G, ambientContrib.B, 0}
	}

	// Calculate lighting terms
	NdotL := math.Max(0, normal.Dot(lightDir))
	if NdotL <= 0 {
		return Color{0, 0, 0, 0}
	}

	halfVector := lightDir.Add(viewDir).Normalize()
	NdotV := math.Max(0, normal.Dot(viewDir))
	NdotH := math.Max(0, normal.Dot(halfVector))
	VdotH := math.Max(0, viewDir.Dot(halfVector))

	// BRDF calculations
	D := pbrL.distributionGGX(NdotH, alpha)
	G := pbrL.geometrySmith(NdotV, NdotL, alpha)
	F := pbrL.fresnelSchlick(VdotH, f0)

	// Cook-Torrance BRDF
	numerator := D * G
	denominator := 4.0*NdotV*NdotL + 0.001 // Add small value to prevent division by zero
	specular := numerator / denominator

	// Calculate kS and kD for energy conservation
	kS := Vector{F.X, F.Y, F.Z}
	kD := Vector{1.0, 1.0, 1.0}.Sub(kS)
	kD = kD.MulScalar(1.0 - material.Metallic) // Metallic materials have no diffuse

	// Combine diffuse and specular
	diffuse := Vector{
		material.BaseColor.R / math.Pi,
		material.BaseColor.G / math.Pi,
		material.BaseColor.B / math.Pi,
	}

	brdf := kD.Mul(diffuse).Add(Vector{specular * F.X, specular * F.Y, specular * F.Z})

	// Final color contribution
	radiance := Vector{lightColor.R, lightColor.G, lightColor.B}
	contribution := brdf.Mul(radiance).MulScalar(NdotL)

	return Color{contribution.X, contribution.Y, contribution.Z, 0}
}

// distributionGGX calculates the normal distribution function using GGX/Trowbridge-Reitz
func (pbrL *PBRLighting) distributionGGX(NdotH, alpha float64) float64 {
	a2 := alpha * alpha
	NdotH2 := NdotH * NdotH

	num := a2
	denom := NdotH2*(a2-1.0) + 1.0
	denom = math.Pi * denom * denom

	return num / denom
}

// geometrySchlickGGX calculates geometry occlusion for a single direction
func (pbrL *PBRLighting) geometrySchlickGGX(NdotV, alpha float64) float64 {
	r := alpha + 1.0
	k := (r * r) / 8.0

	num := NdotV
	denom := NdotV*(1.0-k) + k

	return num / denom
}

// geometrySmith calculates the geometry function using Smith's method
func (pbrL *PBRLighting) geometrySmith(NdotV, NdotL, alpha float64) float64 {
	ggx2 := pbrL.geometrySchlickGGX(NdotV, alpha)
	ggx1 := pbrL.geometrySchlickGGX(NdotL, alpha)

	return ggx1 * ggx2
}

// fresnelSchlick calculates the Fresnel reflection using Schlick's approximation
func (pbrL *PBRLighting) fresnelSchlick(cosTheta float64, F0 Vector) Vector {
	f := math.Pow(1.0-cosTheta, 5.0)
	one := Vector{1.0, 1.0, 1.0}
	return F0.Add(one.Sub(F0).MulScalar(f))
}
