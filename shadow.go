package fauxgl

import (
	"image"
	"math"
)

// ShadowMap represents a shadow map for a light source
type ShadowMap struct {
	Width     int
	Height    int
	DepthMap  []float64
	LightView Matrix
}

// NewShadowMap creates a new shadow map with the specified dimensions
func NewShadowMap(width, height int) *ShadowMap {
	return &ShadowMap{
		Width:    width,
		Height:   height,
		DepthMap: make([]float64, width*height),
	}
}

// Clear clears the shadow map with the specified depth value
func (sm *ShadowMap) Clear(depth float64) {
	for i := range sm.DepthMap {
		sm.DepthMap[i] = depth
	}
}

// GetDepth retrieves the depth value at the specified coordinates
func (sm *ShadowMap) GetDepth(x, y int) float64 {
	if x < 0 || x >= sm.Width || y < 0 || y >= sm.Height {
		return math.MaxFloat64
	}
	return sm.DepthMap[y*sm.Width+x]
}

// SetDepth sets the depth value at the specified coordinates
func (sm *ShadowMap) SetDepth(x, y int, depth float64) {
	if x < 0 || x >= sm.Width || y < 0 || y >= sm.Height {
		return
	}
	sm.DepthMap[y*sm.Width+x] = depth
}

// ShadowMapShader is a shader that renders depth information for shadow mapping
type ShadowMapShader struct {
	Matrix Matrix
}

// NewShadowMapShader creates a new shadow map shader
func NewShadowMapShader(matrix Matrix) *ShadowMapShader {
	return &ShadowMapShader{Matrix: matrix}
}

// Vertex processes a vertex for shadow mapping
func (shader *ShadowMapShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

// Fragment returns the depth value for shadow mapping
func (shader *ShadowMapShader) Fragment(v Vertex) Color {
	// Return the depth value as a color
	depth := v.Output.Z / v.Output.W
	return Color{depth, depth, depth, 1}
}

// ShadowReceiverShader is a shader that receives shadows
type ShadowReceiverShader struct {
	Matrix         Matrix
	LightDirection Vector
	CameraPosition Vector
	ObjectColor    Color
	Texture        Texture
	ShadowMap      *ShadowMap
	LightMatrix    Matrix
	ShadowBias     float64
	ShadowStrength float64
	PCFSize        int // Percentage Closer Filtering size
}

// NewShadowReceiverShader creates a new shadow receiver shader
func NewShadowReceiverShader(matrix, lightMatrix Matrix, lightDirection, cameraPosition Vector, shadowMap *ShadowMap) *ShadowReceiverShader {
	return &ShadowReceiverShader{
		Matrix:         matrix,
		LightDirection: lightDirection,
		CameraPosition: cameraPosition,
		ObjectColor:    Color{0.8, 0.8, 0.8, 1},
		ShadowMap:      shadowMap,
		LightMatrix:    lightMatrix,
		ShadowBias:     0.005,
		ShadowStrength: 0.7,
		PCFSize:        2, // 2x2 PCF
	}
}

// Vertex processes a vertex for shadow receiving
func (shader *ShadowReceiverShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

// Fragment performs shading with shadow calculation
func (shader *ShadowReceiverShader) Fragment(v Vertex) Color {
	// Sample base color
	color := shader.ObjectColor
	if shader.Texture != nil {
		color = shader.Texture.BilinearSample(v.Texture.X, v.Texture.Y)
	}

	// Calculate lighting without shadows first
	light := Color{0.2, 0.2, 0.2, 1} // Ambient
	diffuse := math.Max(v.Normal.Dot(shader.LightDirection), 0)
	light = light.Add(Color{0.8, 0.8, 0.8, 1}.MulScalar(diffuse))

	// Calculate shadow
	shadowFactor := shader.calculateShadow(v)

	// Apply shadow to lighting
	shadowedLight := light.MulScalar(1.0 - shadowFactor*shader.ShadowStrength)

	return color.Mul(shadowedLight).Min(White).Alpha(color.A)
}

// calculateShadow computes the shadow factor for a fragment
func (shader *ShadowReceiverShader) calculateShadow(v Vertex) float64 {
	if shader.ShadowMap == nil {
		return 0.0
	}

	// Transform fragment position to light space
	lightSpacePos := shader.LightMatrix.MulPositionW(v.Position)

	// Perspective divide
	lightSpacePos = lightSpacePos.DivScalar(lightSpacePos.W)

	// Transform to shadow map coordinates
	x := int((lightSpacePos.X*0.5 + 0.5) * float64(shader.ShadowMap.Width))
	y := int((lightSpacePos.Y*0.5 + 0.5) * float64(shader.ShadowMap.Height))

	// Get depth from shadow map
	shadowDepth := shader.ShadowMap.GetDepth(x, y)

	// Current fragment depth
	currentDepth := lightSpacePos.Z

	// Add bias to reduce shadow acne
	currentDepth -= shader.ShadowBias

	// Percentage Closer Filtering (PCF)
	if shader.PCFSize > 0 {
		return shader.pcfFilter(x, y, currentDepth)
	}

	// Simple shadow test
	if currentDepth > shadowDepth {
		return 1.0
	}
	return 0.0
}

// pcfFilter applies Percentage Closer Filtering for softer shadows
func (shader *ShadowReceiverShader) pcfFilter(x, y int, currentDepth float64) float64 {
	var shadow float64
	var samples float64

	for dx := -shader.PCFSize; dx <= shader.PCFSize; dx++ {
		for dy := -shader.PCFSize; dy <= shader.PCFSize; dy++ {
			sampleDepth := shader.ShadowMap.GetDepth(x+dx, y+dy)
			if currentDepth > sampleDepth {
				shadow += 1.0
			}
			samples += 1.0
		}
	}

	return shadow / samples
}

// PCSShadowReceiverShader implements Percentage Closer Soft Shadows
type PCSShadowReceiverShader struct {
	*ShadowReceiverShader
	BlockerSearchSamples int
	PCFSamples           int
	LightSize            float64
}

// NewPCSShadowReceiverShader creates a new PCSS shadow receiver shader
func NewPCSShadowReceiverShader(matrix, lightMatrix Matrix, lightDirection, cameraPosition Vector, shadowMap *ShadowMap) *PCSShadowReceiverShader {
	baseShader := NewShadowReceiverShader(matrix, lightMatrix, lightDirection, cameraPosition, shadowMap)
	return &PCSShadowReceiverShader{
		ShadowReceiverShader: baseShader,
		BlockerSearchSamples: 8,
		PCFSamples:           8,
		LightSize:            0.02,
	}
}

// calculateShadow computes the shadow factor with PCSS
func (shader *PCSShadowReceiverShader) calculateShadow(v Vertex) float64 {
	if shader.ShadowMap == nil {
		return 0.0
	}

	// Transform fragment position to light space
	lightSpacePos := shader.LightMatrix.MulPositionW(v.Position)
	lightSpacePos = lightSpacePos.DivScalar(lightSpacePos.W)

	x := (lightSpacePos.X*0.5 + 0.5) * float64(shader.ShadowMap.Width)
	y := (lightSpacePos.Y*0.5 + 0.5) * float64(shader.ShadowMap.Height)
	currentDepth := lightSpacePos.Z - shader.ShadowBias

	// Find blocker distance
	blockerDistance := shader.findBlockerDistance(x, y, currentDepth)
	if blockerDistance < 0 {
		return 0.0 // No blockers, fully lit
	}

	// Calculate penumbra size
	penumbraSize := shader.calculatePenumbraSize(blockerDistance, currentDepth)

	// Apply PCF with penumbra size
	return shader.pcfWithFilterSize(x, y, currentDepth, penumbraSize)
}

// findBlockerDistance finds the average distance to blockers
func (shader *PCSShadowReceiverShader) findBlockerDistance(x, y, currentDepth float64) float64 {
	var blockers float64
	var sumBlockerDepth float64
	var samples float64

	searchSize := shader.LightSize * 5.0 // Search area

	for i := 0; i < shader.BlockerSearchSamples; i++ {
		angle := float64(i) / float64(shader.BlockerSearchSamples) * 2.0 * math.Pi
		radius := searchSize * (float64(i) / float64(shader.BlockerSearchSamples))

		sampleX := x + math.Cos(angle)*radius
		sampleY := y + math.Sin(angle)*radius

		sampleDepth := shader.ShadowMap.GetDepth(int(sampleX), int(sampleY))
		if sampleDepth < currentDepth {
			sumBlockerDepth += sampleDepth
			blockers += 1.0
		}
		samples += 1.0
	}

	if blockers == 0 {
		return -1 // No blockers found
	}

	return sumBlockerDepth / blockers
}

// calculatePenumbraSize calculates the penumbra size based on blocker distance
func (shader *PCSShadowReceiverShader) calculatePenumbraSize(blockerDistance, currentDepth float64) float64 {
	distanceToLight := currentDepth - blockerDistance
	return shader.LightSize * distanceToLight / blockerDistance
}

// pcfWithFilterSize applies PCF with a specific filter size
func (shader *PCSShadowReceiverShader) pcfWithFilterSize(x, y, currentDepth, filterSize float64) float64 {
	var shadow float64
	var samples float64

	for i := 0; i < shader.PCFSamples; i++ {
		angle := float64(i) / float64(shader.PCFSamples) * 2.0 * math.Pi
		radius := filterSize * (float64(i) / float64(shader.PCFSamples))

		sampleX := x + math.Cos(angle)*radius
		sampleY := y + math.Sin(angle)*radius

		sampleDepth := shader.ShadowMap.GetDepth(int(sampleX), int(sampleY))
		if currentDepth > sampleDepth {
			shadow += 1.0
		}
		samples += 1.0
	}

	return shadow / samples
}

// ShadowRenderer handles shadow map generation and rendering
type ShadowRenderer struct {
	context     *Context
	shadowMap   *ShadowMap
	light       Light
	lightMatrix Matrix
}

// NewShadowRenderer creates a new shadow renderer
func NewShadowRenderer(context *Context, shadowMapSize int, light Light) *ShadowRenderer {
	return &ShadowRenderer{
		context:   context,
		shadowMap: NewShadowMap(shadowMapSize, shadowMapSize),
		light:     light,
	}
}

// GenerateShadowMap generates a shadow map from the light's perspective
func (sr *ShadowRenderer) GenerateShadowMap(scene *Scene) *ShadowMap {
	// Create orthographic projection for shadow mapping
	// In a real implementation, you would calculate tight bounds
	lightProjection := Orthographic(-10, 10, -10, 10, 0.1, 50)

	// Create view matrix from light direction
	lightView := LookAt(
		sr.light.Direction.MulScalar(10), // Light position
		Vector{0, 0, 0},                  // Look at origin
		Vector{0, 1, 0},                  // Up vector
	)

	sr.lightMatrix = lightProjection.Mul(lightView)

	// Clear shadow map
	sr.shadowMap.Clear(math.MaxFloat64)

	// Create shadow map shader
	shadowShader := NewShadowMapShader(sr.lightMatrix)

	// Save original context state
	originalShader := sr.context.Shader
	originalColorBuffer := sr.context.ColorBuffer
	originalDepthBuffer := sr.context.DepthBuffer
	originalWriteColor := sr.context.WriteColor
	originalWriteDepth := sr.context.WriteDepth

	// Set up context for shadow map rendering
	sr.context.Shader = shadowShader
	sr.context.ColorBuffer = image.NewNRGBA(image.Rect(0, 0, sr.shadowMap.Width, sr.shadowMap.Height))
	sr.context.DepthBuffer = make([]float64, sr.shadowMap.Width*sr.shadowMap.Height)
	sr.context.WriteColor = false // We only care about depth
	sr.context.WriteDepth = true

	// Render scene from light's perspective
	renderables := scene.RootNode.GetRenderableNodes()
	for _, node := range renderables {
		if node.Mesh != nil {
			sr.context.DrawMesh(node.Mesh)
		}
	}

	// Copy depth buffer to shadow map
	sr.ExtractDepthFromBuffer()

	// Restore original context state
	sr.context.Shader = originalShader
	sr.context.ColorBuffer = originalColorBuffer
	sr.context.DepthBuffer = originalDepthBuffer
	sr.context.WriteColor = originalWriteColor
	sr.context.WriteDepth = originalWriteDepth

	return sr.shadowMap
}

// ExtractDepthFromBuffer extracts depth values from the depth buffer to the shadow map
func (sr *ShadowRenderer) ExtractDepthFromBuffer() {
	// 获取深度缓冲区数据
	depthBuffer := sr.context.DepthBuffer
	if depthBuffer == nil {
		// 如果没有深度缓冲区，使用默认值填充
		sr.shadowMap.Clear(1.0)
		return
	}

	bounds := sr.context.ColorBuffer.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	// 确保shadow map尺寸匹配
	if sr.shadowMap.Width != width || sr.shadowMap.Height != height {
		// 重新创建合适尺寸的shadow map
		sr.shadowMap = NewShadowMap(width, height)
	}

	// 从深度缓冲区复制数据到阴影贴图
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// 获取深度值
			depthValue := depthBuffer[y*width+x]
			// 存储到阴影贴图中
			sr.shadowMap.SetDepth(x, y, depthValue)
		}
	}
}

// GetLightMatrix returns the light's view-projection matrix
func (sr *ShadowRenderer) GetLightMatrix() Matrix {
	return sr.lightMatrix
}

// ShadowMapRenderer handles advanced shadow mapping techniques
type ShadowMapRenderer struct {
	context     *Context
	shadowMap   *ShadowMap
	light       Light
	lightMatrix Matrix
	technique   ShadowTechnique
}

// ShadowTechnique represents the type of shadow mapping technique to use
type ShadowTechnique int

const (
	// SimpleShadow uses basic shadow mapping
	SimpleShadow ShadowTechnique = iota
	// PCFShadow uses Percentage Closer Filtering
	PCFShadow
	// PCSSShadow uses Percentage Closer Soft Shadows
	PCSSShadow
	// VSMShadow uses Variance Shadow Maps
	VSMShadow
)

// NewShadowMapRenderer creates a new shadow map renderer with the specified technique
func NewShadowMapRenderer(context *Context, shadowMapSize int, light Light, technique ShadowTechnique) *ShadowMapRenderer {
	return &ShadowMapRenderer{
		context:   context,
		shadowMap: NewShadowMap(shadowMapSize, shadowMapSize),
		light:     light,
		technique: technique,
	}
}

// GenerateShadowMap generates a shadow map using the specified technique
func (sr *ShadowMapRenderer) GenerateShadowMap(scene *Scene) *ShadowMap {
	// Calculate tight bounds for the light's view frustum
	bounds := sr.calculateLightBounds(scene)

	// Create orthographic projection for shadow mapping
	lightProjection := Orthographic(bounds.Min.X, bounds.Max.X, bounds.Min.Y, bounds.Max.Y, bounds.Min.Z, bounds.Max.Z)

	// Create view matrix from light direction
	lightView := LookAt(
		sr.light.Direction.MulScalar(-bounds.Max.Z*0.5), // Light position
		Vector{0, 0, 0}, // Look at origin
		Vector{0, 1, 0}, // Up vector
	)

	sr.lightMatrix = lightProjection.Mul(lightView)

	// Clear shadow map
	sr.shadowMap.Clear(math.MaxFloat64)

	// Create shadow map shader
	shadowShader := NewShadowMapShader(sr.lightMatrix)

	// Save original context state
	originalShader := sr.context.Shader
	originalColorBuffer := sr.context.ColorBuffer
	originalDepthBuffer := sr.context.DepthBuffer
	originalWriteColor := sr.context.WriteColor
	originalWriteDepth := sr.context.WriteDepth

	// Set up context for shadow map rendering
	sr.context.Shader = shadowShader
	sr.context.ColorBuffer = image.NewNRGBA(image.Rect(0, 0, sr.shadowMap.Width, sr.shadowMap.Height))
	sr.context.DepthBuffer = make([]float64, sr.shadowMap.Width*sr.shadowMap.Height)
	sr.context.WriteColor = false // We only care about depth
	sr.context.WriteDepth = true

	// Render scene from light's perspective
	renderables := scene.RootNode.GetRenderableNodes()
	for _, node := range renderables {
		if node.Mesh != nil && node.CastShadows {
			sr.context.DrawMesh(node.Mesh)
		}
	}

	// Copy depth values to shadow map
	sr.extractDepthFromBuffer()

	// Restore original context state
	sr.context.Shader = originalShader
	sr.context.ColorBuffer = originalColorBuffer
	sr.context.DepthBuffer = originalDepthBuffer
	sr.context.WriteColor = originalWriteColor
	sr.context.WriteDepth = originalWriteDepth

	return sr.shadowMap
}

// calculateLightBounds calculates tight bounds for the light's view frustum
func (sr *ShadowMapRenderer) calculateLightBounds(scene *Scene) Box {
	// Get scene bounds
	sceneBounds := scene.GetBounds()

	// Transform scene bounds to light space
	lightView := LookAt(
		sr.light.Direction.MulScalar(10), // Light position
		Vector{0, 0, 0},                  // Look at origin
		Vector{0, 1, 0},                  // Up vector
	)

	transformedMin := lightView.MulPosition(sceneBounds.Min)
	transformedMax := lightView.MulPosition(sceneBounds.Max)

	// Calculate bounds in light space
	minX := math.Min(transformedMin.X, transformedMax.X)
	maxX := math.Max(transformedMin.X, transformedMax.X)
	minY := math.Min(transformedMin.Y, transformedMax.Y)
	maxY := math.Max(transformedMin.Y, transformedMax.Y)
	minZ := math.Min(transformedMin.Z, transformedMax.Z) - 5 // Add some padding
	maxZ := math.Max(transformedMin.Z, transformedMax.Z) + 5 // Add some padding

	return Box{Vector{minX, minY, minZ}, Vector{maxX, maxY, maxZ}}
}

// extractDepthFromBuffer extracts depth values from the depth buffer to the shadow map
func (sr *ShadowMapRenderer) extractDepthFromBuffer() {
	// 获取深度缓冲区数据
	depthBuffer := sr.context.DepthBuffer
	if depthBuffer == nil {
		// 如果没有深度缓冲区，使用默认值填充
		sr.shadowMap.Clear(1.0)
		return
	}

	bounds := sr.context.ColorBuffer.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	// 确保shadow map尺寸匹配
	if sr.shadowMap.Width != width || sr.shadowMap.Height != height {
		// 重新创建合适尺寸的shadow map
		sr.shadowMap = NewShadowMap(width, height)
		sr.shadowMap.Clear(math.MaxFloat64) // 初始化为最大深度值
	}

	// 从深度缓冲区复制数据到阴影贴图
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// 获取深度值
			depthValue := depthBuffer[y*width+x]
			// 存储到阴影贴图中
			sr.shadowMap.SetDepth(x, y, depthValue)
		}
	}
}

// GetLightMatrix returns the light's view-projection matrix
func (sr *ShadowMapRenderer) GetLightMatrix() Matrix {
	return sr.lightMatrix
}

// GetTechnique returns the shadow mapping technique being used
func (sr *ShadowMapRenderer) GetTechnique() ShadowTechnique {
	return sr.technique
}

// CascadeShadowMap implements cascaded shadow maps for large scenes
type CascadeShadowMap struct {
	ShadowMaps     []*ShadowMap
	LightMatrices  []Matrix
	SplitDistances []float64
}

// NewCascadeShadowMap creates a new cascaded shadow map system
func NewCascadeShadowMap(numCascades int, shadowMapSize int) *CascadeShadowMap {
	csm := &CascadeShadowMap{
		ShadowMaps:     make([]*ShadowMap, numCascades),
		LightMatrices:  make([]Matrix, numCascades),
		SplitDistances: make([]float64, numCascades),
	}

	for i := 0; i < numCascades; i++ {
		csm.ShadowMaps[i] = NewShadowMap(shadowMapSize, shadowMapSize)
	}

	// Set split distances (logarithmic splitting)
	near := 0.1
	far := 100.0
	for i := 0; i < numCascades; i++ {
		f := float64(i+1) / float64(numCascades)
		csm.SplitDistances[i] = near * math.Pow(far/near, f)
	}

	return csm
}

// OmniShadowMap implements omnidirectional shadow mapping for point lights
type OmniShadowMap struct {
	ShadowMaps    []*ShadowMap // 6 shadow maps for cube faces
	LightPosition Vector
	LightMatrices []Matrix
}

// NewOmniShadowMap creates a new omnidirectional shadow map for point lights
func NewOmniShadowMap(shadowMapSize int, lightPosition Vector) *OmniShadowMap {
	osm := &OmniShadowMap{
		ShadowMaps:    make([]*ShadowMap, 6),
		LightPosition: lightPosition,
		LightMatrices: make([]Matrix, 6),
	}

	for i := 0; i < 6; i++ {
		osm.ShadowMaps[i] = NewShadowMap(shadowMapSize, shadowMapSize)
	}

	// Create view matrices for each cube face
	osm.LightMatrices[0] = LookAt(lightPosition, lightPosition.Add(Vector{1, 0, 0}), Vector{0, -1, 0})  // +X
	osm.LightMatrices[1] = LookAt(lightPosition, lightPosition.Add(Vector{-1, 0, 0}), Vector{0, -1, 0}) // -X
	osm.LightMatrices[2] = LookAt(lightPosition, lightPosition.Add(Vector{0, 1, 0}), Vector{0, 0, 1})   // +Y
	osm.LightMatrices[3] = LookAt(lightPosition, lightPosition.Add(Vector{0, -1, 0}), Vector{0, 0, -1}) // -Y
	osm.LightMatrices[4] = LookAt(lightPosition, lightPosition.Add(Vector{0, 0, 1}), Vector{0, -1, 0})  // +Z
	osm.LightMatrices[5] = LookAt(lightPosition, lightPosition.Add(Vector{0, 0, -1}), Vector{0, -1, 0}) // -Z

	return osm
}

// SoftShadowReceiverShader implements advanced soft shadow techniques
type SoftShadowReceiverShader struct {
	Matrix              Matrix
	LightDirection      Vector
	CameraPosition      Vector
	ObjectColor         Color
	Texture             Texture
	ShadowMap           *ShadowMap
	LightMatrix         Matrix
	ShadowBias          float64
	ShadowStrength      float64
	SoftShadowTechnique ShadowTechnique
	FilterSize          int
}

// NewSoftShadowReceiverShader creates a new soft shadow receiver shader
func NewSoftShadowReceiverShader(matrix, lightMatrix Matrix, lightDirection, cameraPosition Vector, shadowMap *ShadowMap, technique ShadowTechnique) *SoftShadowReceiverShader {
	return &SoftShadowReceiverShader{
		Matrix:              matrix,
		LightDirection:      lightDirection,
		CameraPosition:      cameraPosition,
		ObjectColor:         Color{0.8, 0.8, 0.8, 1},
		ShadowMap:           shadowMap,
		LightMatrix:         lightMatrix,
		ShadowBias:          0.005,
		ShadowStrength:      0.7,
		SoftShadowTechnique: technique,
		FilterSize:          2,
	}
}

// Vertex processes a vertex
func (shader *SoftShadowReceiverShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

// Fragment performs shading with advanced shadow calculation
func (shader *SoftShadowReceiverShader) Fragment(v Vertex) Color {
	// Sample base color
	color := shader.ObjectColor
	if shader.Texture != nil {
		color = shader.Texture.BilinearSample(v.Texture.X, v.Texture.Y)
	}

	// Calculate lighting without shadows first
	light := Color{0.2, 0.2, 0.2, 1} // Ambient
	diffuse := math.Max(v.Normal.Dot(shader.LightDirection), 0)
	light = light.Add(Color{0.8, 0.8, 0.8, 1}.MulScalar(diffuse))

	// Calculate shadow based on technique
	var shadowFactor float64
	switch shader.SoftShadowTechnique {
	case PCFShadow:
		shadowFactor = shader.calculatePCFShadow(v)
	case PCSSShadow:
		shadowFactor = shader.calculatePCSSShadow(v)
	default:
		shadowFactor = shader.calculateSimpleShadow(v)
	}

	// Apply shadow to lighting
	shadowedLight := light.MulScalar(1.0 - shadowFactor*shader.ShadowStrength)

	return color.Mul(shadowedLight).Min(White).Alpha(color.A)
}

// calculateSimpleShadow computes basic shadow factor
func (shader *SoftShadowReceiverShader) calculateSimpleShadow(v Vertex) float64 {
	if shader.ShadowMap == nil {
		return 0.0
	}

	// Transform fragment position to light space
	lightSpacePos := shader.LightMatrix.MulPositionW(v.Position)
	lightSpacePos = lightSpacePos.DivScalar(lightSpacePos.W)

	// Transform to shadow map coordinates
	x := int((lightSpacePos.X*0.5 + 0.5) * float64(shader.ShadowMap.Width))
	y := int((lightSpacePos.Y*0.5 + 0.5) * float64(shader.ShadowMap.Height))

	// Get depth from shadow map
	shadowDepth := shader.ShadowMap.GetDepth(x, y)

	// Current fragment depth
	currentDepth := lightSpacePos.Z

	// Add bias to reduce shadow acne
	currentDepth -= shader.ShadowBias

	// Simple shadow test
	if currentDepth > shadowDepth {
		return 1.0
	}
	return 0.0
}

// calculatePCFShadow computes shadow factor with Percentage Closer Filtering
func (shader *SoftShadowReceiverShader) calculatePCFShadow(v Vertex) float64 {
	if shader.ShadowMap == nil {
		return 0.0
	}

	// Transform fragment position to light space
	lightSpacePos := shader.LightMatrix.MulPositionW(v.Position)
	lightSpacePos = lightSpacePos.DivScalar(lightSpacePos.W)

	// Transform to shadow map coordinates
	x := (lightSpacePos.X*0.5 + 0.5) * float64(shader.ShadowMap.Width)
	y := (lightSpacePos.Y*0.5 + 0.5) * float64(shader.ShadowMap.Height)

	// Current fragment depth
	currentDepth := lightSpacePos.Z - shader.ShadowBias

	// Apply PCF filtering
	var shadow float64
	var samples float64

	for dx := -shader.FilterSize; dx <= shader.FilterSize; dx++ {
		for dy := -shader.FilterSize; dy <= shader.FilterSize; dy++ {
			sampleX := int(x) + dx
			sampleY := int(y) + dy
			sampleDepth := shader.ShadowMap.GetDepth(sampleX, sampleY)
			if currentDepth > sampleDepth {
				shadow += 1.0
			}
			samples += 1.0
		}
	}

	return shadow / samples
}

// calculatePCSSShadow computes shadow factor with Percentage Closer Soft Shadows
func (shader *SoftShadowReceiverShader) calculatePCSSShadow(v Vertex) float64 {
	// Simplified PCSS implementation
	// In a full implementation, this would include blocker search and penumbra size calculation

	// For now, we'll use a weighted average of PCF with different filter sizes
	pcf1 := shader.calculatePCFShadowWithSize(v, 1)
	pcf2 := shader.calculatePCFShadowWithSize(v, 2)
	pcf3 := shader.calculatePCFShadowWithSize(v, 3)

	// Weighted average (closer to light = softer shadows)
	return (pcf1*0.2 + pcf2*0.3 + pcf3*0.5)
}

// calculatePCFShadowWithSize computes PCF shadow with a specific filter size
func (shader *SoftShadowReceiverShader) calculatePCFShadowWithSize(v Vertex, filterSize int) float64 {
	if shader.ShadowMap == nil {
		return 0.0
	}

	// Transform fragment position to light space
	lightSpacePos := shader.LightMatrix.MulPositionW(v.Position)
	lightSpacePos = lightSpacePos.DivScalar(lightSpacePos.W)

	// Transform to shadow map coordinates
	x := (lightSpacePos.X*0.5 + 0.5) * float64(shader.ShadowMap.Width)
	y := (lightSpacePos.Y*0.5 + 0.5) * float64(shader.ShadowMap.Height)

	// Current fragment depth
	currentDepth := lightSpacePos.Z - shader.ShadowBias

	// Apply PCF filtering
	var shadow float64
	var samples float64

	for dx := -filterSize; dx <= filterSize; dx++ {
		for dy := -filterSize; dy <= filterSize; dy++ {
			sampleX := int(x) + dx
			sampleY := int(y) + dy
			sampleDepth := shader.ShadowMap.GetDepth(sampleX, sampleY)
			if currentDepth > sampleDepth {
				shadow += 1.0
			}
			samples += 1.0
		}
	}

	return shadow / samples
}
