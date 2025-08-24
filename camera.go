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

// OrbitCamera implements an orbiting camera controller
type OrbitCamera struct {
	*Camera
	Target          Vector
	Distance        float64
	HorizontalAngle float64
	VerticalAngle   float64
}

// NewOrbitCamera creates a new orbit camera
func NewOrbitCamera(name string, target Vector, distance, fov, aspectRatio, near, far float64) *OrbitCamera {
	camera := NewPerspectiveCamera(name, Vector{0, 0, distance}, target, Vector{0, 1, 0}, fov, aspectRatio, near, far)

	return &OrbitCamera{
		Camera:          camera,
		Target:          target,
		Distance:        distance,
		HorizontalAngle: 0,
		VerticalAngle:   0,
	}
}

// Update updates the orbit camera position based on angles
func (oc *OrbitCamera) Update() {
	// Convert spherical coordinates to Cartesian
	x := oc.Distance * math.Sin(oc.VerticalAngle) * math.Cos(oc.HorizontalAngle)
	y := oc.Distance * math.Cos(oc.VerticalAngle)
	z := oc.Distance * math.Sin(oc.VerticalAngle) * math.Sin(oc.HorizontalAngle)

	oc.Position = oc.Target.Add(Vector{x, y, z})

	// Update camera to look at target
	oc.LookAt(oc.Position, oc.Target, Vector{0, 1, 0})
}

// Rotate rotates the camera around the target
func (oc *OrbitCamera) Rotate(horizontalDelta, verticalDelta float64) {
	oc.HorizontalAngle += horizontalDelta
	oc.VerticalAngle += verticalDelta

	// Clamp vertical angle to avoid flipping
	oc.VerticalAngle = math.Max(0.1, math.Min(math.Pi-0.1, oc.VerticalAngle))

	oc.Update()
}

// Zoom zooms the camera in or out
func (oc *OrbitCamera) Zoom(delta float64) {
	oc.Distance = math.Max(0.1, oc.Distance+delta)
	oc.Update()
}

// FirstPersonCamera implements a first-person camera controller
type FirstPersonCamera struct {
	*Camera
	Yaw   float64
	Pitch float64
	Speed float64
}

// NewFirstPersonCamera creates a new first-person camera
func NewFirstPersonCamera(name string, position Vector, fov, aspectRatio, near, far float64) *FirstPersonCamera {
	camera := NewPerspectiveCamera(name, position, position.Add(Vector{0, 0, -1}), Vector{0, 1, 0}, fov, aspectRatio, near, far)

	return &FirstPersonCamera{
		Camera: camera,
		Yaw:    -math.Pi / 2, // Look along negative Z axis
		Pitch:  0,
		Speed:  1.0,
	}
}

// Update updates the first-person camera orientation
func (fpc *FirstPersonCamera) Update() {
	// Calculate forward vector based on yaw and pitch
	forward := Vector{
		math.Cos(fpc.Yaw) * math.Cos(fpc.Pitch),
		math.Sin(fpc.Pitch),
		math.Sin(fpc.Yaw) * math.Cos(fpc.Pitch),
	}.Normalize()

	// Calculate right and up vectors
	right := forward.Cross(Vector{0, 1, 0}).Normalize()
	up := right.Cross(forward).Normalize()

	fpc.Target = fpc.Position.Add(forward)
	fpc.Up = up
}

// Rotate rotates the camera view
func (fpc *FirstPersonCamera) Rotate(yawDelta, pitchDelta float64) {
	fpc.Yaw += yawDelta
	fpc.Pitch += pitchDelta

	// Clamp pitch to avoid flipping
	fpc.Pitch = math.Max(-math.Pi/2+0.1, math.Min(math.Pi/2-0.1, fpc.Pitch))

	fpc.Update()
}

// Move moves the camera in the specified direction
func (fpc *FirstPersonCamera) Move(direction Vector) {
	// Calculate forward and right vectors
	forward := Vector{
		math.Cos(fpc.Yaw) * math.Cos(fpc.Pitch),
		math.Sin(fpc.Pitch),
		math.Sin(fpc.Yaw) * math.Cos(fpc.Pitch),
	}.Normalize()

	right := forward.Cross(Vector{0, 1, 0}).Normalize()
	up := Vector{0, 1, 0}

	// Calculate movement vector
	move := Vector{0, 0, 0}
	move = move.Add(forward.MulScalar(direction.Z * fpc.Speed))
	move = move.Add(right.MulScalar(direction.X * fpc.Speed))
	move = move.Add(up.MulScalar(direction.Y * fpc.Speed))

	fpc.Position = fpc.Position.Add(move)
	fpc.Update()
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

// ViewFrustum represents a camera viewing frustum for culling
type ViewFrustum struct {
	Planes [6]Plane
}

// Plane represents a plane in 3D space
type Plane struct {
	Normal   Vector
	Distance float64
}

// NewViewFrustumFromMatrix creates a frustum from a projection-view matrix
func NewViewFrustumFromMatrix(matrix Matrix) *ViewFrustum {
	frustum := &ViewFrustum{}

	// Extract frustum planes from projection-view matrix
	// Left plane
	frustum.Planes[0] = Plane{
		Normal:   Vector{matrix.X03 + matrix.X00, matrix.X13 + matrix.X10, matrix.X23 + matrix.X20},
		Distance: matrix.X33 + matrix.X30,
	}

	// Right plane
	frustum.Planes[1] = Plane{
		Normal:   Vector{matrix.X03 - matrix.X00, matrix.X13 - matrix.X10, matrix.X23 - matrix.X20},
		Distance: matrix.X33 - matrix.X30,
	}

	// Bottom plane
	frustum.Planes[2] = Plane{
		Normal:   Vector{matrix.X03 + matrix.X01, matrix.X13 + matrix.X11, matrix.X23 + matrix.X21},
		Distance: matrix.X33 + matrix.X31,
	}

	// Top plane
	frustum.Planes[3] = Plane{
		Normal:   Vector{matrix.X03 - matrix.X01, matrix.X13 - matrix.X11, matrix.X23 - matrix.X21},
		Distance: matrix.X33 - matrix.X31,
	}

	// Near plane
	frustum.Planes[4] = Plane{
		Normal:   Vector{matrix.X02, matrix.X12, matrix.X22},
		Distance: matrix.X32,
	}

	// Far plane
	frustum.Planes[5] = Plane{
		Normal:   Vector{matrix.X03 - matrix.X02, matrix.X13 - matrix.X12, matrix.X23 - matrix.X22},
		Distance: matrix.X33 - matrix.X32,
	}

	// Normalize plane normals and distances
	for i := range frustum.Planes {
		length := frustum.Planes[i].Normal.Length()
		frustum.Planes[i].Normal = frustum.Planes[i].Normal.DivScalar(length)
		frustum.Planes[i].Distance /= length
	}

	return frustum
}

// IntersectsBox checks if a box intersects with the frustum
func (f *ViewFrustum) IntersectsBox(box Box) bool {
	for _, plane := range f.Planes {
		// Find the positive and negative vertices relative to the plane
		var positive, negative Vector

		if plane.Normal.X >= 0 {
			positive.X = box.Max.X
			negative.X = box.Min.X
		} else {
			positive.X = box.Min.X
			negative.X = box.Max.X
		}

		if plane.Normal.Y >= 0 {
			positive.Y = box.Max.Y
			negative.Y = box.Min.Y
		} else {
			positive.Y = box.Min.Y
			negative.Y = box.Max.Y
		}

		if plane.Normal.Z >= 0 {
			positive.Z = box.Max.Z
			negative.Z = box.Min.Z
		} else {
			positive.Z = box.Min.Z
			negative.Z = box.Max.Z
		}

		// Check if the positive vertex is outside (behind) the plane
		if positive.Dot(plane.Normal)+plane.Distance < 0 {
			return false
		}

		// Check if the negative vertex is outside (in front of) the plane
		// If it is, the box intersects the plane, so we continue
		// If both vertices are on the same side, we need to check further
	}

	return true
}

// CullingSceneRenderer extends SceneRenderer with frustum culling
type CullingSceneRenderer struct {
	*SceneRenderer
}

// NewCullingSceneRenderer creates a new culling scene renderer
func NewCullingSceneRenderer(context *Context) *CullingSceneRenderer {
	return &CullingSceneRenderer{
		SceneRenderer: NewSceneRenderer(context),
	}
}

// RenderScene renders a complete scene with frustum culling
func (csr *CullingSceneRenderer) RenderScene(scene *Scene) {
	if scene.ActiveCamera == nil {
		return
	}

	// Get camera matrices
	viewMatrix := scene.ActiveCamera.GetViewMatrix()
	projectionMatrix := scene.ActiveCamera.GetProjectionMatrix()
	cameraMatrix := projectionMatrix.Mul(viewMatrix)

	// Create frustum for culling
	frustum := NewViewFrustumFromMatrix(cameraMatrix)

	// Get all renderable nodes
	renderables := scene.RootNode.GetRenderableNodes()

	// Render each node with culling
	for _, node := range renderables {
		csr.RenderNodeWithCulling(node, cameraMatrix, scene.Lights, frustum)
	}
}

// RenderNodeWithCulling renders a single scene node with frustum culling
func (csr *CullingSceneRenderer) RenderNodeWithCulling(node *SceneNode, cameraMatrix Matrix, lights []Light, frustum *ViewFrustum) {
	if node.Mesh == nil || node.Material == nil {
		return
	}

	// Transform mesh bounds to world space
	meshBounds := node.Mesh.BoundingBox()
	worldBounds := node.WorldTransform.MulBox(meshBounds)

	// Check if the node is within the view frustum
	if !frustum.IntersectsBox(worldBounds) {
		return // Skip rendering this node
	}

	// Calculate final transform matrix
	modelMatrix := node.WorldTransform
	finalMatrix := cameraMatrix.Mul(modelMatrix)

	// Create PBR shader
	pbrShader := NewPBRShader(finalMatrix, node.Material, lights, Vector{0, 0, 5})

	// Set shader and render
	csr.context.Shader = pbrShader
	csr.context.DrawMesh(node.Mesh)
}
