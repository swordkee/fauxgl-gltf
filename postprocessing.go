package fauxgl

import (
	"image"
	"image/color"
	"math"
)

// PostProcessingEffect represents a post-processing effect
type PostProcessingEffect interface {
	Apply(input *image.NRGBA) *image.NRGBA
}

// PostProcessingPipeline represents a chain of post-processing effects
type PostProcessingPipeline struct {
	Effects []PostProcessingEffect
}

// NewPostProcessingPipeline creates a new post-processing pipeline
func NewPostProcessingPipeline() *PostProcessingPipeline {
	return &PostProcessingPipeline{
		Effects: make([]PostProcessingEffect, 0),
	}
}

// AddEffect adds an effect to the pipeline
func (pp *PostProcessingPipeline) AddEffect(effect PostProcessingEffect) {
	pp.Effects = append(pp.Effects, effect)
}

// Process applies all effects in the pipeline
func (pp *PostProcessingPipeline) Process(input *image.NRGBA) *image.NRGBA {
	result := input
	for _, effect := range pp.Effects {
		result = effect.Apply(result)
	}
	return result
}

// BlurEffect implements a simple blur effect
type BlurEffect struct {
	Radius int
}

// NewBlurEffect creates a new blur effect
func NewBlurEffect(radius int) *BlurEffect {
	return &BlurEffect{Radius: radius}
}

// Apply applies the blur effect to the input image
func (be *BlurEffect) Apply(input *image.NRGBA) *image.NRGBA {
	bounds := input.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	output := image.NewNRGBA(bounds)

	// Horizontal blur pass
	temp := image.NewNRGBA(bounds)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var r, g, b, a float64
			var count float64

			for dx := -be.Radius; dx <= be.Radius; dx++ {
				nx := x + dx
				if nx >= 0 && nx < width {
					c := input.NRGBAAt(nx+bounds.Min.X, y+bounds.Min.Y)
					weight := gaussian(float64(dx), float64(be.Radius)/2.0)
					r += float64(c.R) * weight
					g += float64(c.G) * weight
					b += float64(c.B) * weight
					a += float64(c.A) * weight
					count += weight
				}
			}

			if count > 0 {
				r /= count
				g /= count
				b /= count
				a /= count
			}

			temp.SetNRGBA(x+bounds.Min.X, y+bounds.Min.Y, color.NRGBA{
				R: uint8(math.Min(255, math.Max(0, r))),
				G: uint8(math.Min(255, math.Max(0, g))),
				B: uint8(math.Min(255, math.Max(0, b))),
				A: uint8(math.Min(255, math.Max(0, a))),
			})
		}
	}

	// Vertical blur pass
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var r, g, b, a float64
			var count float64

			for dy := -be.Radius; dy <= be.Radius; dy++ {
				ny := y + dy
				if ny >= 0 && ny < height {
					c := temp.NRGBAAt(x+bounds.Min.X, ny+bounds.Min.Y)
					weight := gaussian(float64(dy), float64(be.Radius)/2.0)
					r += float64(c.R) * weight
					g += float64(c.G) * weight
					b += float64(c.B) * weight
					a += float64(c.A) * weight
					count += weight
				}
			}

			if count > 0 {
				r /= count
				g /= count
				b /= count
				a /= count
			}

			output.SetNRGBA(x+bounds.Min.X, y+bounds.Min.Y, color.NRGBA{
				R: uint8(math.Min(255, math.Max(0, r))),
				G: uint8(math.Min(255, math.Max(0, g))),
				B: uint8(math.Min(255, math.Max(0, b))),
				A: uint8(math.Min(255, math.Max(0, a))),
			})
		}
	}

	return output
}

// gaussian calculates the Gaussian weight for a given distance and sigma
func gaussian(x, sigma float64) float64 {
	return math.Exp(-(x * x) / (2 * sigma * sigma))
}

// BloomEffect implements a bloom effect
type BloomEffect struct {
	Threshold  float64
	BlurRadius int
	Intensity  float64
}

// NewBloomEffect creates a new bloom effect
func NewBloomEffect(threshold float64, blurRadius int, intensity float64) *BloomEffect {
	return &BloomEffect{
		Threshold:  threshold,
		BlurRadius: blurRadius,
		Intensity:  intensity,
	}
}

// Apply applies the bloom effect to the input image
func (be *BloomEffect) Apply(input *image.NRGBA) *image.NRGBA {
	bounds := input.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Extract bright parts
	bright := image.NewNRGBA(bounds)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := input.NRGBAAt(x+bounds.Min.X, y+bounds.Min.Y)
			brightness := (float64(c.R) + float64(c.G) + float64(c.B)) / (3.0 * 255.0)
			if brightness > be.Threshold {
				bright.SetNRGBA(x+bounds.Min.X, y+bounds.Min.Y, c)
			} else {
				bright.SetNRGBA(x+bounds.Min.X, y+bounds.Min.Y, color.NRGBA{0, 0, 0, c.A})
			}
		}
	}

	// Blur the bright parts
	blur := NewBlurEffect(be.BlurRadius)
	blurred := blur.Apply(bright)

	// Combine original with blurred bright parts
	output := image.NewNRGBA(bounds)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			original := input.NRGBAAt(x+bounds.Min.X, y+bounds.Min.Y)
			bloom := blurred.NRGBAAt(x+bounds.Min.X, y+bounds.Min.Y)

			r := float64(original.R) + float64(bloom.R)*be.Intensity
			g := float64(original.G) + float64(bloom.G)*be.Intensity
			b := float64(original.B) + float64(bloom.B)*be.Intensity
			a := float64(original.A)

			output.SetNRGBA(x+bounds.Min.X, y+bounds.Min.Y, color.NRGBA{
				R: uint8(math.Min(255, math.Max(0, r))),
				G: uint8(math.Min(255, math.Max(0, g))),
				B: uint8(math.Min(255, math.Max(0, b))),
				A: uint8(math.Min(255, math.Max(0, a))),
			})
		}
	}

	return output
}

// ToneMappingEffect implements tone mapping
type ToneMappingEffect struct {
	Exposure float64
	Gamma    float64
}

// NewToneMappingEffect creates a new tone mapping effect
func NewToneMappingEffect(exposure, gamma float64) *ToneMappingEffect {
	return &ToneMappingEffect{
		Exposure: exposure,
		Gamma:    gamma,
	}
}

// Apply applies tone mapping to the input image
func (tme *ToneMappingEffect) Apply(input *image.NRGBA) *image.NRGBA {
	bounds := input.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	output := image.NewNRGBA(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := input.NRGBAAt(x+bounds.Min.X, y+bounds.Min.Y)

			// Convert to linear space
			r := float64(c.R) / 255.0
			g := float64(c.G) / 255.0
			b := float64(c.B) / 255.0

			// Apply exposure
			r *= math.Pow(2.0, tme.Exposure)
			g *= math.Pow(2.0, tme.Exposure)
			b *= math.Pow(2.0, tme.Exposure)

			// Reinhard tone mapping
			r = r / (r + 1.0)
			g = g / (g + 1.0)
			b = b / (b + 1.0)

			// Apply gamma correction
			r = math.Pow(r, 1.0/tme.Gamma)
			g = math.Pow(g, 1.0/tme.Gamma)
			b = math.Pow(b, 1.0/tme.Gamma)

			output.SetNRGBA(x+bounds.Min.X, y+bounds.Min.Y, color.NRGBA{
				R: uint8(r * 255),
				G: uint8(g * 255),
				B: uint8(b * 255),
				A: c.A,
			})
		}
	}

	return output
}

// FXAAEffect implements Fast Approximate Anti-Aliasing
type FXAAEffect struct {
	SpanMax   float64
	ReduceMul float64
}

// NewFXAAEffect creates a new FXAA effect
func NewFXAAEffect() *FXAAEffect {
	return &FXAAEffect{
		SpanMax:   8.0,
		ReduceMul: 1.0 / 8.0,
	}
}

// Apply applies FXAA to the input image
func (fxaa *FXAAEffect) Apply(input *image.NRGBA) *image.NRGBA {
	bounds := input.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	output := image.NewNRGBA(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Sample the current pixel and its neighbors
			rgbM := getColor(input, x, y, bounds)
			rgbN := getColor(input, x, y-1, bounds)
			rgbS := getColor(input, x, y+1, bounds)
			rgbE := getColor(input, x+1, y, bounds)
			rgbW := getColor(input, x-1, y, bounds)

			// Compute local contrast
			rgbL := minColor(rgbN, rgbW)
			rgbU := maxColor(rgbN, rgbW)
			rgbL = minColor(rgbL, rgbM)
			rgbU = maxColor(rgbU, rgbM)

			rgbL = minColor(rgbL, rgbS)
			rgbU = maxColor(rgbU, rgbS)
			rgbL = minColor(rgbL, rgbE)
			rgbU = maxColor(rgbU, rgbE)

			// Compute edge direction
			dir := subColor(rgbS, rgbN)
			dir = addColor(dir, subColor(rgbE, rgbW))

			// Compute gradient
			dirAbs := absColor(dir)
			dirAbs = addColor(dirAbs, dirAbs)

			// Reduce gradient
			temp := addScalarColor(dirAbs, fxaa.ReduceMul)
			dirAbs = maxColor(dirAbs, temp)

			// Compute edge lerp
			dirAbs = invColor(dirAbs)
			dirAbs = mulScalarColor(dirAbs, 1.0/16.0)
			dir = mulColor(dir, dirAbs)

			// Clamp direction
			dir = clampColor(dir, -fxaa.SpanMax, fxaa.SpanMax)

			// Sample along the gradient
			rgbA := getColorBilinear(input, float64(x)+dir.X, float64(y)+dir.Y, bounds)
			rgbB := getColorBilinear(input, float64(x)-dir.X, float64(y)-dir.Y, bounds)

			// Compute final color
			rgbF := addColor(rgbA, rgbB)
			rgbF = mulScalarColor(rgbF, 0.5)

			// Blend with original
			blend := dotColor(absColor(subColor(rgbF, rgbM)), 1.0)
			blend = math.Min(1.0, blend*4.0)

			rgbR := lerpColor(rgbM, rgbF, blend)

			output.SetNRGBA(x+bounds.Min.X, y+bounds.Min.Y, color.NRGBA{
				R: uint8(math.Min(255, math.Max(0, rgbR.X*255))),
				G: uint8(math.Min(255, math.Max(0, rgbR.Y*255))),
				B: uint8(math.Min(255, math.Max(0, rgbR.Z*255))),
				A: 255,
			})
		}
	}

	return output
}

// Helper functions for FXAA
func getColor(img *image.NRGBA, x, y int, bounds image.Rectangle) Vector {
	if x < 0 || x >= bounds.Dx() || y < 0 || y >= bounds.Dy() {
		return Vector{0, 0, 0}
	}
	c := img.NRGBAAt(x+bounds.Min.X, y+bounds.Min.Y)
	return Vector{float64(c.R) / 255.0, float64(c.G) / 255.0, float64(c.B) / 255.0}
}

func getColorBilinear(img *image.NRGBA, x, y float64, bounds image.Rectangle) Vector {
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1 := x0 + 1
	y1 := y0 + 1

	c00 := getColor(img, x0, y0, bounds)
	c01 := getColor(img, x0, y1, bounds)
	c10 := getColor(img, x1, y0, bounds)
	c11 := getColor(img, x1, y1, bounds)

	dx := x - float64(x0)
	dy := y - float64(y0)

	c0 := lerpColor(c00, c01, dy)
	c1 := lerpColor(c10, c11, dy)

	return lerpColor(c0, c1, dx)
}

func minColor(a, b Vector) Vector {
	return Vector{math.Min(a.X, b.X), math.Min(a.Y, b.Y), math.Min(a.Z, b.Z)}
}

func maxColor(a, b Vector) Vector {
	return Vector{math.Max(a.X, b.X), math.Max(a.Y, b.Y), math.Max(a.Z, b.Z)}
}

func subColor(a, b Vector) Vector {
	return Vector{a.X - b.X, a.Y - b.Y, a.Z - b.Z}
}

func addColor(a, b Vector) Vector {
	return Vector{a.X + b.X, a.Y + b.Y, a.Z + b.Z}
}

func mulColor(a, b Vector) Vector {
	return Vector{a.X * b.X, a.Y * b.Y, a.Z * b.Z}
}

func mulScalarColor(a Vector, s float64) Vector {
	return Vector{a.X * s, a.Y * s, a.Z * s}
}

func addScalarColor(a Vector, s float64) Vector {
	return Vector{a.X + s, a.Y + s, a.Z + s}
}

func absColor(a Vector) Vector {
	return Vector{math.Abs(a.X), math.Abs(a.Y), math.Abs(a.Z)}
}

func invColor(a Vector) Vector {
	return Vector{1.0 / (a.X + 0.0001), 1.0 / (a.Y + 0.0001), 1.0 / (a.Z + 0.0001)}
}

func clampColor(a Vector, min, max float64) Vector {
	return Vector{
		math.Min(math.Max(a.X, min), max),
		math.Min(math.Max(a.Y, min), max),
		math.Min(math.Max(a.Z, min), max),
	}
}

func dotColor(a Vector, b float64) float64 {
	return a.X*b + a.Y*b + a.Z*b
}

func lerpColor(a, b Vector, t float64) Vector {
	return Vector{
		a.X + t*(b.X-a.X),
		a.Y + t*(b.Y-a.Y),
		a.Z + t*(b.Z-a.Z),
	}
}

// ChromaticAberrationEffect implements chromatic aberration
type ChromaticAberrationEffect struct {
	RedOffset   Vector
	GreenOffset Vector
	BlueOffset  Vector
}

// NewChromaticAberrationEffect creates a new chromatic aberration effect
func NewChromaticAberrationEffect(redOffset, greenOffset, blueOffset Vector) *ChromaticAberrationEffect {
	return &ChromaticAberrationEffect{
		RedOffset:   redOffset,
		GreenOffset: greenOffset,
		BlueOffset:  blueOffset,
	}
}

// Apply applies chromatic aberration to the input image
func (cae *ChromaticAberrationEffect) Apply(input *image.NRGBA) *image.NRGBA {
	bounds := input.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	output := image.NewNRGBA(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Sample red channel with offset
			redX := x + int(cae.RedOffset.X)
			redY := y + int(cae.RedOffset.Y)
			var r uint8
			if redX >= 0 && redX < width && redY >= 0 && redY < height {
				r = input.NRGBAAt(redX+bounds.Min.X, redY+bounds.Min.Y).R
			}

			// Sample green channel with offset
			greenX := x + int(cae.GreenOffset.X)
			greenY := y + int(cae.GreenOffset.Y)
			var g uint8
			if greenX >= 0 && greenX < width && greenY >= 0 && greenY < height {
				g = input.NRGBAAt(greenX+bounds.Min.X, greenY+bounds.Min.Y).G
			}

			// Sample blue channel with offset
			blueX := x + int(cae.BlueOffset.X)
			blueY := y + int(cae.BlueOffset.Y)
			var b uint8
			if blueX >= 0 && blueX < width && blueY >= 0 && blueY < height {
				b = input.NRGBAAt(blueX+bounds.Min.X, blueY+bounds.Min.Y).B
			}

			// Get original alpha
			a := input.NRGBAAt(x+bounds.Min.X, y+bounds.Min.Y).A

			output.SetNRGBA(x+bounds.Min.X, y+bounds.Min.Y, color.NRGBA{r, g, b, a})
		}
	}

	return output
}

// VignetteEffect implements a vignette effect
type VignetteEffect struct {
	Strength float64
}

// NewVignetteEffect creates a new vignette effect
func NewVignetteEffect(strength float64) *VignetteEffect {
	return &VignetteEffect{Strength: strength}
}

// Apply applies vignette to the input image
func (ve *VignetteEffect) Apply(input *image.NRGBA) *image.NRGBA {
	bounds := input.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	centerX := float64(width) / 2.0
	centerY := float64(height) / 2.0
	maxDist := math.Sqrt(centerX*centerX + centerY*centerY)

	output := image.NewNRGBA(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := input.NRGBAAt(x+bounds.Min.X, y+bounds.Min.Y)

			// Calculate distance from center
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx + dy*dy)

			// Calculate vignette factor
			factor := 1.0 - (dist/maxDist)*ve.Strength

			// Apply vignette
			r := float64(c.R) * factor
			g := float64(c.G) * factor
			b := float64(c.B) * factor

			output.SetNRGBA(x+bounds.Min.X, y+bounds.Min.Y, color.NRGBA{
				R: uint8(math.Min(255, math.Max(0, r))),
				G: uint8(math.Min(255, math.Max(0, g))),
				B: uint8(math.Min(255, math.Max(0, b))),
				A: c.A,
			})
		}
	}

	return output
}

// ColorGradingEffect implements color grading
type ColorGradingEffect struct {
	Brightness float64
	Contrast   float64
	Saturation float64
	HueShift   float64
}

// NewColorGradingEffect creates a new color grading effect
func NewColorGradingEffect(brightness, contrast, saturation, hueShift float64) *ColorGradingEffect {
	return &ColorGradingEffect{
		Brightness: brightness,
		Contrast:   contrast,
		Saturation: saturation,
		HueShift:   hueShift,
	}
}

// Apply applies color grading to the input image
func (cge *ColorGradingEffect) Apply(input *image.NRGBA) *image.NRGBA {
	bounds := input.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	output := image.NewNRGBA(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := input.NRGBAAt(x+bounds.Min.X, y+bounds.Min.Y)

			// Convert to [0,1] range
			r := float64(c.R) / 255.0
			g := float64(c.G) / 255.0
			b := float64(c.B) / 255.0

			// Apply brightness
			r += cge.Brightness
			g += cge.Brightness
			b += cge.Brightness

			// Apply contrast
			r = (r-0.5)*cge.Contrast + 0.5
			g = (g-0.5)*cge.Contrast + 0.5
			b = (b-0.5)*cge.Contrast + 0.5

			// Apply saturation
			lum := 0.299*r + 0.587*g + 0.114*b
			r = lum + (r-lum)*cge.Saturation
			g = lum + (g-lum)*cge.Saturation
			b = lum + (b-lum)*cge.Saturation

			// Apply hue shift (simplified)
			if cge.HueShift != 0 {
				// Simple hue rotation approximation
				r, g, b = rotateHue(r, g, b, cge.HueShift)
			}

			// Clamp and convert back to [0,255] range
			r = math.Max(0, math.Min(1, r))
			g = math.Max(0, math.Min(1, g))
			b = math.Max(0, math.Min(1, b))

			output.SetNRGBA(x+bounds.Min.X, y+bounds.Min.Y, color.NRGBA{
				R: uint8(r * 255),
				G: uint8(g * 255),
				B: uint8(b * 255),
				A: c.A,
			})
		}
	}

	return output
}

// rotateHue performs a simple hue rotation
func rotateHue(r, g, b, hueShift float64) (float64, float64, float64) {
	// Simplified hue rotation
	// In a full implementation, you would convert to HSV, rotate hue, and convert back
	cos := math.Cos(hueShift)
	sin := math.Sin(hueShift)

	// Simple RGB rotation matrix approximation
	newR := r*cos + g*(-sin) + b*0
	newG := r*sin + g*cos + b*0
	newB := r*0 + g*0 + b*1

	return newR, newG, newB
}

// MotionBlurEffect implements motion blur
type MotionBlurEffect struct {
	Angle   float64
	Length  float64
	Samples int
}

// NewMotionBlurEffect creates a new motion blur effect
func NewMotionBlurEffect(angle, length float64, samples int) *MotionBlurEffect {
	return &MotionBlurEffect{
		Angle:   angle,
		Length:  length,
		Samples: samples,
	}
}

// Apply applies motion blur to the input image
func (mbe *MotionBlurEffect) Apply(input *image.NRGBA) *image.NRGBA {
	bounds := input.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	output := image.NewNRGBA(bounds)

	// Calculate blur direction vector
	dx := math.Cos(mbe.Angle) * mbe.Length
	dy := math.Sin(mbe.Angle) * mbe.Length

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var r, g, b, a float64
			var count float64

			// Sample along the blur direction
			for i := 0; i < mbe.Samples; i++ {
				t := float64(i) / float64(mbe.Samples-1)
				sampleX := float64(x) + dx*t
				sampleY := float64(y) + dy*t

				// Bilinear sampling
				sampleColor := getColorBilinear(input, sampleX, sampleY, bounds)
				r += sampleColor.X
				g += sampleColor.Y
				b += sampleColor.Z
				a += 1.0
				count += 1.0
			}

			if count > 0 {
				r /= count
				g /= count
				b /= count
				a /= count
			}

			output.SetNRGBA(x+bounds.Min.X, y+bounds.Min.Y, color.NRGBA{
				R: uint8(math.Min(255, math.Max(0, r*255))),
				G: uint8(math.Min(255, math.Max(0, g*255))),
				B: uint8(math.Min(255, math.Max(0, b*255))),
				A: uint8(math.Min(255, math.Max(0, a*255))),
			})
		}
	}

	return output
}

// DepthOfFieldEffect implements depth of field
type DepthOfFieldEffect struct {
	FocusDepth float64
	Aperture   float64
	Samples    int
}

// NewDepthOfFieldEffect creates a new depth of field effect
func NewDepthOfFieldEffect(focusDepth, aperture float64, samples int) *DepthOfFieldEffect {
	return &DepthOfFieldEffect{
		FocusDepth: focusDepth,
		Aperture:   aperture,
		Samples:    samples,
	}
}

// Apply applies depth of field to the input image
// Note: This is a simplified implementation that simulates DoF based on pixel position
func (dof *DepthOfFieldEffect) Apply(input *image.NRGBA) *image.NRGBA {
	bounds := input.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	output := image.NewNRGBA(bounds)

	// Create a blurred version of the input for bokeh effect
	blurEffect := NewBlurEffect(int(dof.Aperture * 10))
	blurred := blurEffect.Apply(input)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Simulate depth based on Y position (closer to center = in focus)
			centerY := float64(height) / 2.0
			distanceFromCenter := math.Abs(float64(y)-centerY) / centerY
			depth := distanceFromCenter // Simple depth simulation

			// Calculate blur amount based on distance from focus depth
			blurAmount := math.Abs(depth-dof.FocusDepth) * dof.Aperture

			// Mix original and blurred based on blur amount
			original := input.NRGBAAt(x+bounds.Min.X, y+bounds.Min.Y)
			blurredPixel := blurred.NRGBAAt(x+bounds.Min.X, y+bounds.Min.Y)

			mix := math.Min(1.0, blurAmount)
			r := float64(original.R)*(1-mix) + float64(blurredPixel.R)*mix
			g := float64(original.G)*(1-mix) + float64(blurredPixel.G)*mix
			b := float64(original.B)*(1-mix) + float64(blurredPixel.B)*mix
			a := float64(original.A)*(1-mix) + float64(blurredPixel.A)*mix

			output.SetNRGBA(x+bounds.Min.X, y+bounds.Min.Y, color.NRGBA{
				R: uint8(math.Min(255, math.Max(0, r))),
				G: uint8(math.Min(255, math.Max(0, g))),
				B: uint8(math.Min(255, math.Max(0, b))),
				A: uint8(math.Min(255, math.Max(0, a))),
			})
		}
	}

	return output
}

// CompositeEffect combines multiple effects
type CompositeEffect struct {
	Effects []PostProcessingEffect
}

// NewCompositeEffect creates a new composite effect
func NewCompositeEffect(effects ...PostProcessingEffect) *CompositeEffect {
	return &CompositeEffect{
		Effects: effects,
	}
}

// Apply applies all effects in the composite
func (ce *CompositeEffect) Apply(input *image.NRGBA) *image.NRGBA {
	result := input
	for _, effect := range ce.Effects {
		result = effect.Apply(result)
	}
	return result
}
