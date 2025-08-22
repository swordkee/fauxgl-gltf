package fauxgl

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
)

// Texture interface defines the basic texture functionality
// This interface provides backward compatibility with the old texture system
type Texture interface {
	// BilinearSample samples the texture using bilinear interpolation
	BilinearSample(u, v float64) Color
}

// TextureType represents different types of textures
type TextureType int

const (
	// BaseColorTexture - diffuse/albedo texture
	BaseColorTexture TextureType = iota
	// NormalTexture - normal map for surface detail
	NormalTexture
	// MetallicTexture - metallic values
	MetallicTexture
	// RoughnessTexture - surface roughness
	RoughnessTexture
	// OcclusionTexture - ambient occlusion
	OcclusionTexture
	// EmissiveTexture - emissive/glow texture
	EmissiveTexture
	// HeightTexture - height/displacement map
	HeightTexture
	// SpecularTexture - specular values (legacy)
	SpecularTexture
	// GlossinessTexture - glossiness values (legacy)
	GlossinessTexture
	// KTX2Texture - KTX2 container format texture
	KTX2Texture
)

// TextureWrap defines texture wrapping behavior
type TextureWrap int

const (
	// WrapRepeat - repeat the texture
	WrapRepeat TextureWrap = iota
	// WrapClamp - clamp to edge
	WrapClamp
	// WrapMirror - mirror repeat
	WrapMirror
)

// TextureFilter defines texture filtering method
type TextureFilter int

const (
	// FilterNearest - nearest neighbor filtering
	FilterNearest TextureFilter = iota
	// FilterLinear - bilinear filtering
	FilterLinear
	// FilterMipmap - mipmap filtering
	FilterMipmap
)

// AdvancedTexture extends the basic texture interface with advanced features
type AdvancedTexture struct {
	Image     image.Image
	Width     int
	Height    int
	Type      TextureType
	WrapS     TextureWrap
	WrapT     TextureWrap
	MinFilter TextureFilter
	MagFilter TextureFilter
	MipLevels []image.Image // For mipmap support
	Transform Matrix        // Texture coordinate transformation
}

// NewAdvancedTexture creates a new advanced texture from an image
func NewAdvancedTexture(img image.Image, textureType TextureType) *AdvancedTexture {
	bounds := img.Bounds()
	texture := &AdvancedTexture{
		Image:     img,
		Width:     bounds.Dx(),
		Height:    bounds.Dy(),
		Type:      textureType,
		WrapS:     WrapRepeat,
		WrapT:     WrapRepeat,
		MinFilter: FilterLinear,
		MagFilter: FilterLinear,
		Transform: Identity(),
	}

	// Generate mipmaps for better quality
	texture.GenerateMipmaps()

	return texture
}

// LoadAdvancedTexture loads an advanced texture from a file path
func LoadAdvancedTexture(path string, textureType TextureType) (*AdvancedTexture, error) {
	img, err := LoadImage(path)
	if err != nil {
		return nil, err
	}
	return NewAdvancedTexture(img, textureType), nil
}

// LoadTexture loads a texture from a file path (legacy compatibility)
func LoadTexture(path string) (Texture, error) {
	img, err := LoadImage(path)
	if err != nil {
		return nil, err
	}
	return NewAdvancedTexture(img, BaseColorTexture), nil
}

// Sample samples the texture with basic UV coordinates
func (t *AdvancedTexture) Sample(u, v float64) Color {
	return t.SampleWithFilter(u, v, t.MagFilter)
}

// BilinearSample performs bilinear sampling
func (t *AdvancedTexture) BilinearSample(u, v float64) Color {
	return t.SampleWithFilter(u, v, FilterLinear)
}

// SampleWithFilter samples the texture with specified filtering
func (t *AdvancedTexture) SampleWithFilter(u, v float64, filter TextureFilter) Color {
	// Apply texture coordinate transformation
	uv := Vector{u, v, 0}
	transformedUV := t.Transform.MulPosition(uv)
	u, v = transformedUV.X, transformedUV.Y

	// Handle texture wrapping
	u = t.wrapCoordinate(u, t.WrapS)
	v = t.wrapCoordinate(v, t.WrapT)

	// Flip V coordinate (OpenGL convention)
	v = 1.0 - v

	switch filter {
	case FilterNearest:
		return t.sampleNearest(u, v)
	case FilterLinear:
		return t.sampleBilinear(u, v)
	case FilterMipmap:
		// For now, fall back to bilinear
		// TODO: Implement proper mipmap sampling with derivatives
		return t.sampleBilinear(u, v)
	default:
		return t.sampleBilinear(u, v)
	}
}

// wrapCoordinate applies texture wrapping to a coordinate
func (t *AdvancedTexture) wrapCoordinate(coord float64, wrap TextureWrap) float64 {
	switch wrap {
	case WrapRepeat:
		return coord - math.Floor(coord)
	case WrapClamp:
		return math.Max(0, math.Min(1, coord))
	case WrapMirror:
		coord = coord - math.Floor(coord)
		if int(math.Floor(coord*2))%2 == 1 {
			coord = 1.0 - coord
		}
		return coord
	default:
		return coord - math.Floor(coord)
	}
}

// sampleNearest performs nearest neighbor sampling
func (t *AdvancedTexture) sampleNearest(u, v float64) Color {
	x := int(u*float64(t.Width-1) + 0.5)
	y := int(v*float64(t.Height-1) + 0.5)
	x = ClampInt(x, 0, t.Width-1)
	y = ClampInt(y, 0, t.Height-1)
	return MakeColor(t.Image.At(x, y))
}

// sampleBilinear performs bilinear sampling
func (t *AdvancedTexture) sampleBilinear(u, v float64) Color {
	x := u * float64(t.Width-1)
	y := v * float64(t.Height-1)

	x0 := int(x)
	y0 := int(y)
	x1 := x0 + 1
	y1 := y0 + 1

	// Clamp coordinates
	x0 = ClampInt(x0, 0, t.Width-1)
	y0 = ClampInt(y0, 0, t.Height-1)
	x1 = ClampInt(x1, 0, t.Width-1)
	y1 = ClampInt(y1, 0, t.Height-1)

	// Fractional parts
	fx := x - float64(int(x))
	fy := y - float64(int(y))

	// Sample four corners
	c00 := MakeColor(t.Image.At(x0, y0))
	c01 := MakeColor(t.Image.At(x0, y1))
	c10 := MakeColor(t.Image.At(x1, y0))
	c11 := MakeColor(t.Image.At(x1, y1))

	// Bilinear interpolation
	top := c00.Lerp(c10, fx)
	bottom := c01.Lerp(c11, fx)
	return top.Lerp(bottom, fy)
}

// GenerateMipmaps generates mipmap levels for the texture
func (t *AdvancedTexture) GenerateMipmaps() {
	// Clear existing mipmaps
	t.MipLevels = nil
	t.MipLevels = append(t.MipLevels, t.Image)

	currentImg := t.Image
	currentWidth := t.Width
	currentHeight := t.Height

	// Generate mipmaps until 1x1
	for currentWidth > 1 || currentHeight > 1 {
		newWidth := int(math.Max(1, float64(currentWidth)/2))
		newHeight := int(math.Max(1, float64(currentHeight)/2))

		// For simplicity, we'll skip actual mipmap generation here
		// In a real implementation, you'd want to use proper downsampling
		// For now, just store the original image at each level
		t.MipLevels = append(t.MipLevels, currentImg)

		currentWidth = newWidth
		currentHeight = newHeight

		if newWidth == 1 && newHeight == 1 {
			break
		}
	}
}

// SampleNormal samples a normal map and returns the normal in tangent space
func (t *AdvancedTexture) SampleNormal(u, v float64) Vector {
	if t.Type != NormalTexture {
		return Vector{0, 0, 1} // Default normal
	}

	color := t.SampleWithFilter(u, v, t.MagFilter)

	// Convert from [0,1] to [-1,1] range
	normal := Vector{
		color.R*2.0 - 1.0,
		color.G*2.0 - 1.0,
		color.B*2.0 - 1.0,
	}

	return normal.Normalize()
}

// SampleHeight samples a height map and returns the height value
func (t *AdvancedTexture) SampleHeight(u, v float64) float64 {
	if t.Type != HeightTexture {
		return 0.0
	}

	color := t.SampleWithFilter(u, v, t.MagFilter)

	// Use grayscale value as height
	return color.R*0.299 + color.G*0.587 + color.B*0.114
}

// SampleSingleChannel samples a single channel texture (metallic, roughness, etc.)
func (t *AdvancedTexture) SampleSingleChannel(u, v float64, channel int) float64 {
	color := t.SampleWithFilter(u, v, t.MagFilter)

	switch channel {
	case 0: // Red
		return color.R
	case 1: // Green
		return color.G
	case 2: // Blue
		return color.B
	case 3: // Alpha
		return color.A
	default:
		return color.R
	}
}

// TextureAtlas represents a texture atlas for batching multiple textures
type TextureAtlas struct {
	Image   image.Image
	Width   int
	Height  int
	Regions map[string]TextureRegion
}

// TextureRegion represents a region within a texture atlas
type TextureRegion struct {
	Name   string
	X, Y   int
	Width  int
	Height int
	U0, V0 float64 // Top-left UV
	U1, V1 float64 // Bottom-right UV
}

// NewTextureAtlas creates a new texture atlas
func NewTextureAtlas(img image.Image) *TextureAtlas {
	bounds := img.Bounds()
	return &TextureAtlas{
		Image:   img,
		Width:   bounds.Dx(),
		Height:  bounds.Dy(),
		Regions: make(map[string]TextureRegion),
	}
}

// AddRegion adds a texture region to the atlas
func (atlas *TextureAtlas) AddRegion(name string, x, y, width, height int) {
	u0 := float64(x) / float64(atlas.Width)
	v0 := float64(y) / float64(atlas.Height)
	u1 := float64(x+width) / float64(atlas.Width)
	v1 := float64(y+height) / float64(atlas.Height)

	atlas.Regions[name] = TextureRegion{
		Name:   name,
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
		U0:     u0,
		V0:     v0,
		U1:     u1,
		V1:     v1,
	}
}

// SampleRegion samples a specific region in the atlas
func (atlas *TextureAtlas) SampleRegion(regionName string, u, v float64) Color {
	region, exists := atlas.Regions[regionName]
	if !exists {
		return Color{1, 0, 1, 1} // Magenta for missing texture
	}

	// Map UV to region coordinates
	regionU := region.U0 + u*(region.U1-region.U0)
	regionV := region.V0 + v*(region.V1-region.V0)

	// Sample the atlas
	x := int(regionU * float64(atlas.Width))
	y := int(regionV * float64(atlas.Height))
	x = ClampInt(x, 0, atlas.Width-1)
	y = ClampInt(y, 0, atlas.Height-1)

	return MakeColor(atlas.Image.At(x, y))
}

// CubeMapTexture represents a cube map texture for environment mapping
type CubeMapTexture struct {
	Faces [6]*AdvancedTexture // +X, -X, +Y, -Y, +Z, -Z
}

// NewCubeMapTexture creates a new cube map texture
func NewCubeMapTexture(faces [6]*AdvancedTexture) *CubeMapTexture {
	return &CubeMapTexture{Faces: faces}
}

// SampleCubeMap samples the cube map using a direction vector
func (cubemap *CubeMapTexture) SampleCubeMap(direction Vector) Color {
	direction = direction.Normalize()

	// Determine which face to use
	absX := math.Abs(direction.X)
	absY := math.Abs(direction.Y)
	absZ := math.Abs(direction.Z)

	var faceIndex int
	var u, v float64

	if absX >= absY && absX >= absZ {
		// X face
		if direction.X > 0 {
			faceIndex = 0 // +X
			u = (-direction.Z/absX + 1.0) * 0.5
			v = (-direction.Y/absX + 1.0) * 0.5
		} else {
			faceIndex = 1 // -X
			u = (direction.Z/absX + 1.0) * 0.5
			v = (-direction.Y/absX + 1.0) * 0.5
		}
	} else if absY >= absZ {
		// Y face
		if direction.Y > 0 {
			faceIndex = 2 // +Y
			u = (direction.X/absY + 1.0) * 0.5
			v = (direction.Z/absY + 1.0) * 0.5
		} else {
			faceIndex = 3 // -Y
			u = (direction.X/absY + 1.0) * 0.5
			v = (-direction.Z/absY + 1.0) * 0.5
		}
	} else {
		// Z face
		if direction.Z > 0 {
			faceIndex = 4 // +Z
			u = (direction.X/absZ + 1.0) * 0.5
			v = (-direction.Y/absZ + 1.0) * 0.5
		} else {
			faceIndex = 5 // -Z
			u = (-direction.X/absZ + 1.0) * 0.5
			v = (-direction.Y/absZ + 1.0) * 0.5
		}
	}

	if cubemap.Faces[faceIndex] != nil {
		return cubemap.Faces[faceIndex].Sample(u, v)
	}

	return Color{0, 0, 0, 1} // Black for missing face
}

// KTX2TextureLoader handles loading of KTX2 format textures
type KTX2TextureLoader struct {
	reader *Reader
}

// LoadKTX2Texture loads a KTX2 texture from file data
func LoadKTX2Texture(data []byte) (*AdvancedTexture, error) {
	reader, err := NewKTX2Reader(data)
	if err != nil {
		return nil, fmt.Errorf("failed to create KTX2 reader: %w", err)
	}

	header := reader.Header()
	levels, err := reader.Levels()
	if err != nil {
		return nil, fmt.Errorf("failed to get KTX2 levels: %w", err)
	}

	if len(levels) == 0 {
		return nil, fmt.Errorf("no texture levels found in KTX2 file")
	}

	// For now, we'll create a basic texture from the first level
	// TODO: Implement proper KTX2 decoding with supercompression support
	// firstLevel := levels[0] // 暂时不使用第一级数据

	// Create a placeholder implementation
	// In a real implementation, you'd need to:
	// 1. Check the format and supercompression scheme
	// 2. Decompress the data if needed (Basis Universal, Zstd, etc.)
	// 3. Convert to a standard image format

	// For demonstration, create a simple colored texture
	img := createPlaceholderKTX2Image(int(header.PixelWidth), int(header.PixelHeight))

	texture := &AdvancedTexture{
		Image:     img,
		Width:     int(header.PixelWidth),
		Height:    int(header.PixelHeight),
		Type:      KTX2Texture,
		WrapS:     WrapRepeat,
		WrapT:     WrapRepeat,
		MinFilter: FilterLinear,
		MagFilter: FilterLinear,
		Transform: Identity(),
	}

	// Store KTX2 specific data
	texture.MipLevels = make([]image.Image, len(levels))
	for i, level := range levels {
		// In a real implementation, decode each level
		texture.MipLevels[i] = createPlaceholderKTX2Image(
			int(header.PixelWidth)>>i,
			int(header.PixelHeight)>>i,
		)
		_ = level // 避免未使用变量警告
	}

	return texture, nil
}

// LoadKTX2TextureFromFile loads a KTX2 texture from a file path
func LoadKTX2TextureFromFile(path string) (*AdvancedTexture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read KTX2 file: %w", err)
	}

	return LoadKTX2Texture(data)
}

// createPlaceholderKTX2Image creates a placeholder image for KTX2 textures
// TODO: Replace with proper KTX2 decoding
func createPlaceholderKTX2Image(width, height int) image.Image {
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}

	// Create a simple gradient image as placeholder
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8(float64(x) / float64(width) * 255)
			g := uint8(float64(y) / float64(height) * 255)
			b := uint8(128) // 固定蓝色分量
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	return img
}

// IsKTX2File checks if the given data represents a KTX2 file
func IsKTX2File(data []byte) bool {
	if len(data) < len(KTX2_MAGIC) {
		return false
	}

	for i, b := range KTX2_MAGIC {
		if data[i] != b {
			return false
		}
	}

	return true
}

// GetKTX2Info extracts basic information from KTX2 file without full parsing
func GetKTX2Info(data []byte) (*Header, error) {
	if !IsKTX2File(data) {
		return nil, fmt.Errorf("not a valid KTX2 file")
	}

	if len(data) < HeaderLength {
		return nil, UnexpectedEnd
	}

	return HeaderFromBytes(data[0:HeaderLength])
}
