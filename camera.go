package fauxgl

import "math"

// Camera represents a camera in the scene
type Camera struct {
	Name           string
	Position       Vector
	Target         Vector
	Up             Vector
	FOV            float64 // Field of view in radians
	AspectRatio    float64
	NearPlane      float64
	FarPlane       float64
	ProjectionType ProjectionType
	OrthoSize      float64 // For orthographic projection
}

// ProjectionType represents the type of camera projection
type ProjectionType int

const (
	// PerspectiveProjection uses perspective projection
	PerspectiveProjection ProjectionType = iota
	// OrthographicProjection uses orthographic projection
	OrthographicProjection
)

// NewPerspectiveCamera creates a new perspective camera
func NewPerspectiveCamera(name string, position, target, up Vector, fov, aspectRatio, near, far float64) *Camera {
	return &Camera{
		Name:           name,
		Position:       position,
		Target:         target,
		Up:             up,
		FOV:            fov,
		AspectRatio:    aspectRatio,
		NearPlane:      near,
		FarPlane:       far,
		ProjectionType: PerspectiveProjection,
	}
}

// NewOrthographicCamera creates a new orthographic camera
func NewOrthographicCamera(name string, position, target, up Vector, orthoSize, aspectRatio, near, far float64) *Camera {
	return &Camera{
		Name:           name,
		Position:       position,
		Target:         target,
		Up:             up,
		AspectRatio:    aspectRatio,
		NearPlane:      near,
		FarPlane:       far,
		ProjectionType: OrthographicProjection,
		OrthoSize:      orthoSize,
	}
}

// GetViewMatrix returns the view matrix for the camera
func (camera *Camera) GetViewMatrix() Matrix {
	return LookAt(camera.Position, camera.Target, camera.Up)
}

// GetProjectionMatrix returns the projection matrix for the camera
func (camera *Camera) GetProjectionMatrix() Matrix {
	switch camera.ProjectionType {
	case PerspectiveProjection:
		return Perspective(camera.FOV, camera.AspectRatio, camera.NearPlane, camera.FarPlane)
	case OrthographicProjection:
		width := camera.OrthoSize * camera.AspectRatio
		height := camera.OrthoSize
		return Orthographic(-width/2, width/2, -height/2, height/2, camera.NearPlane, camera.FarPlane)
	default:
		return Identity()
	}
}

// GetCameraMatrix returns the combined view-projection matrix
func (camera *Camera) GetCameraMatrix() Matrix {
	return camera.GetProjectionMatrix().Mul(camera.GetViewMatrix())
}

// LookAt sets the camera to look at a specific target
func (camera *Camera) LookAt(position, target, up Vector) {
	camera.Position = position
	camera.Target = target
	camera.Up = up
}

// OrbitAroundTarget orbits the camera around its target
func (camera *Camera) OrbitAroundTarget(horizontalAngle, verticalAngle float64) {
	// Calculate current distance from target
	direction := camera.Position.Sub(camera.Target)
	distance := direction.Length()

	// Convert to spherical coordinates
	phi := math.Atan2(direction.Z, direction.X) + horizontalAngle
	theta := math.Acos(direction.Y/distance) + verticalAngle

	// Clamp theta to avoid flipping
	theta = math.Max(0.1, math.Min(math.Pi-0.1, theta))

	// Convert back to Cartesian
	newDirection := Vector{
		distance * math.Sin(theta) * math.Cos(phi),
		distance * math.Cos(theta),
		distance * math.Sin(theta) * math.Sin(phi),
	}

	camera.Position = camera.Target.Add(newDirection)
}

// SceneRenderer handles rendering of scenes
type SceneRenderer struct {
	context *Context
}

// NewSceneRenderer creates a new scene renderer
func NewSceneRenderer(context *Context) *SceneRenderer {
	return &SceneRenderer{
		context: context,
	}
}

// RenderScene renders a complete scene
func (renderer *SceneRenderer) RenderScene(scene *Scene) {
	if scene.ActiveCamera == nil {
		return
	}

	// Get camera matrices
	viewMatrix := scene.ActiveCamera.GetViewMatrix()
	projectionMatrix := scene.ActiveCamera.GetProjectionMatrix()
	cameraMatrix := projectionMatrix.Mul(viewMatrix)

	// Get all renderable nodes
	renderables := scene.RootNode.GetRenderableNodes()

	// Render each node
	for _, node := range renderables {
		renderer.RenderNode(node, cameraMatrix, scene.Lights)
	}
}

// RenderNode renders a single scene node
func (renderer *SceneRenderer) RenderNode(node *SceneNode, cameraMatrix Matrix, lights []Light) {
	if node.Mesh == nil || node.Material == nil {
		return
	}

	// Calculate final transform matrix
	modelMatrix := node.WorldTransform
	finalMatrix := cameraMatrix.Mul(modelMatrix)

	// Create PBR shader
	pbrShader := NewPBRShader(finalMatrix, node.Material, lights, Vector{0, 0, 5})

	// Set shader and render
	renderer.context.Shader = pbrShader
	renderer.context.DrawMesh(node.Mesh)
}
