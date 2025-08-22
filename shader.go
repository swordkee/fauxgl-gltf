package fauxgl

import (
	"math"
)

type Shader interface {
	Vertex(Vertex) Vertex
	Fragment(Vertex) Color
}

// SolidColorShader renders with a single, solid color.
type SolidColorShader struct {
	Matrix Matrix
	Color  Color
}

func NewSolidColorShader(matrix Matrix, color Color) *SolidColorShader {
	return &SolidColorShader{matrix, color}
}

func (shader *SolidColorShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

func (shader *SolidColorShader) Fragment(v Vertex) Color {
	return shader.Color
}

// TextureShader renders with a texture and no lighting.
type TextureShader struct {
	Matrix  Matrix
	Texture Texture
}

func NewTextureShader(matrix Matrix, texture Texture) *TextureShader {
	return &TextureShader{matrix, texture}
}

func (shader *TextureShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

func (shader *TextureShader) Fragment(v Vertex) Color {
	return shader.Texture.BilinearSample(v.Texture.X, v.Texture.Y)
}

// PhongShader implements Phong shading with an optional texture.
type PhongShader struct {
	Matrix         Matrix
	LightDirection Vector
	CameraPosition Vector
	ObjectColor    Color
	AmbientColor   Color
	DiffuseColor   Color
	SpecularColor  Color
	Texture        Texture
	SpecularPower  float64
}

// NewPhongShader f
func NewPhongShader(matrix Matrix, lightDirection, cameraPosition Vector) *PhongShader {
	ambient := Color{0.2, 0.2, 0.2, 1}
	diffuse := Color{0.8, 0.8, 0.8, 1}
	specular := Color{1, 1, 1, 1}
	return &PhongShader{
		matrix, lightDirection, cameraPosition,
		Discard, ambient, diffuse, specular, nil, 32}
}

// Vertex f
func (shader *PhongShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

func (shader *PhongShader) Fragment(v Vertex) Color {
	light := shader.AmbientColor
	color := v.Color
	if shader.ObjectColor != Discard {
		color = shader.ObjectColor
	}
	if shader.Texture != nil {
		color = shader.Texture.BilinearSample(v.Texture.X, v.Texture.Y)
	}
	diffuse := math.Max(v.Normal.Dot(shader.LightDirection), 0)
	light = light.Add(shader.DiffuseColor.MulScalar(diffuse))
	if diffuse > 0 && shader.SpecularPower > 0 {
		camera := shader.CameraPosition.Sub(v.Position).Normalize()
		reflected := shader.LightDirection.Negate().Reflect(v.Normal)
		specular := math.Max(camera.Dot(reflected), 0)
		if specular > 0 {
			specular = math.Pow(specular, shader.SpecularPower)
			light = light.Add(shader.SpecularColor.MulScalar(specular))
		}
	}
	return color.Mul(light).Min(White).Alpha(color.A)
}

// PBRShader implements physically-based rendering
type PBRShader struct {
	Matrix         Matrix
	Material       *PBRMaterial
	Lights         []Light
	AmbientColor   Color
	CameraPosition Vector
	pbrLighting    *PBRLighting
}

// NewPBRShader creates a new PBR shader
func NewPBRShader(matrix Matrix, material *PBRMaterial, lights []Light, cameraPos Vector) *PBRShader {
	return &PBRShader{
		Matrix:         matrix,
		Material:       material,
		Lights:         lights,
		AmbientColor:   Color{0.1, 0.1, 0.1, 1.0},
		CameraPosition: cameraPos,
		pbrLighting:    &PBRLighting{},
	}
}

// Vertex processes a vertex through the PBR shader pipeline
func (shader *PBRShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

// Fragment performs PBR shading calculations
func (shader *PBRShader) Fragment(v Vertex) Color {
	if shader.Material == nil {
		return Color{1, 0, 1, 1} // Magenta for missing material
	}

	// Sample material properties at current texture coordinates
	sampledMaterial := shader.Material.Sample(v.Texture.X, v.Texture.Y)

	// Transform normal from tangent space to world space
	// For simplicity, we'll use the vertex normal directly
	// In a full implementation, you'd calculate TBN matrix
	worldNormal := v.Normal.Normalize()

	// Calculate view direction
	viewDir := shader.CameraPosition.Sub(v.Position).Normalize()

	// Perform PBR lighting calculation
	finalColor := shader.pbrLighting.CalculatePBR(
		sampledMaterial,
		v.Position,
		worldNormal,
		viewDir,
		shader.Lights,
		shader.AmbientColor,
	)

	// Handle alpha mode
	switch shader.Material.AlphaMode {
	case AlphaMask:
		if finalColor.A < shader.Material.AlphaCutoff {
			return Discard // Discard fragment
		}
		finalColor.A = 1.0
	case AlphaBlend:
		// Keep original alpha
	default: // AlphaOpaque
		finalColor.A = 1.0
	}

	return finalColor
}

// MetallicRoughnessShader is a specialized PBR shader for metallic-roughness workflow
type MetallicRoughnessShader struct {
	*PBRShader
	BaseColorTexture         *AdvancedTexture
	MetallicRoughnessTexture *AdvancedTexture
	NormalTexture            *AdvancedTexture
	OcclusionTexture         *AdvancedTexture
	EmissiveTexture          *AdvancedTexture
}

// NewMetallicRoughnessShader creates a metallic-roughness PBR shader
func NewMetallicRoughnessShader(matrix Matrix, lights []Light, cameraPos Vector) *MetallicRoughnessShader {
	pbrShader := &PBRShader{
		Matrix:         matrix,
		Material:       NewPBRMaterial(),
		Lights:         lights,
		AmbientColor:   Color{0.1, 0.1, 0.1, 1.0},
		CameraPosition: cameraPos,
		pbrLighting:    &PBRLighting{},
	}

	return &MetallicRoughnessShader{
		PBRShader: pbrShader,
	}
}

// Fragment performs metallic-roughness PBR shading
func (shader *MetallicRoughnessShader) Fragment(v Vertex) Color {
	u, v_coord := v.Texture.X, v.Texture.Y

	// Sample base color
	baseColor := shader.Material.BaseColorFactor
	if shader.BaseColorTexture != nil {
		baseColor = baseColor.Mul(shader.BaseColorTexture.Sample(u, v_coord))
	}

	// Sample metallic and roughness
	metallic := shader.Material.MetallicFactor
	roughness := shader.Material.RoughnessFactor
	if shader.MetallicRoughnessTexture != nil {
		mr := shader.MetallicRoughnessTexture.Sample(u, v_coord)
		metallic *= mr.B  // Blue channel
		roughness *= mr.G // Green channel
	}

	// Sample normal
	normal := v.Normal.Normalize()
	if shader.NormalTexture != nil {
		tangentNormal := shader.NormalTexture.SampleNormal(u, v_coord)
		// For simplicity, just use the normal directly
		// In practice, you'd transform from tangent to world space
		normal = tangentNormal
	}

	// Sample occlusion
	occlusion := 1.0
	if shader.OcclusionTexture != nil {
		occlusionColor := shader.OcclusionTexture.Sample(u, v_coord)
		occlusion = occlusionColor.R
	}

	// Sample emissive
	emissive := shader.Material.EmissiveFactor
	if shader.EmissiveTexture != nil {
		emissive = emissive.Mul(shader.EmissiveTexture.Sample(u, v_coord))
	}

	// Create sampled material
	sampledMaterial := &SampledMaterial{
		BaseColor: baseColor,
		Metallic:  metallic,
		Roughness: roughness,
		Normal:    normal,
		Occlusion: occlusion,
		Emissive:  emissive,
	}

	// Calculate view direction
	viewDir := shader.CameraPosition.Sub(v.Position).Normalize()

	// Perform PBR lighting calculation
	finalColor := shader.pbrLighting.CalculatePBR(
		sampledMaterial,
		v.Position,
		normal,
		viewDir,
		shader.Lights,
		shader.AmbientColor,
	)

	return finalColor
}

// EnvironmentShader renders environment mapping and reflections
type EnvironmentShader struct {
	Matrix         Matrix
	CubeMap        *CubeMapTexture
	CameraPosition Vector
	Reflectance    float64
}

// NewEnvironmentShader creates a new environment shader
func NewEnvironmentShader(matrix Matrix, cubeMap *CubeMapTexture, cameraPos Vector) *EnvironmentShader {
	return &EnvironmentShader{
		Matrix:         matrix,
		CubeMap:        cubeMap,
		CameraPosition: cameraPos,
		Reflectance:    1.0,
	}
}

// Vertex processes vertices for environment mapping
func (shader *EnvironmentShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

// Fragment performs environment mapping
func (shader *EnvironmentShader) Fragment(v Vertex) Color {
	if shader.CubeMap == nil {
		return Color{0, 0, 0, 1}
	}

	// Calculate reflection direction
	viewDir := shader.CameraPosition.Sub(v.Position).Normalize()
	reflectionDir := viewDir.Reflect(v.Normal)

	// Sample cube map
	reflectionColor := shader.CubeMap.SampleCubeMap(reflectionDir)

	return reflectionColor.MulScalar(shader.Reflectance)
}
