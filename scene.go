package fauxgl

// Scene represents a 3D scene with a hierarchy of nodes
type Scene struct {
	RootNode     *SceneNode
	Cameras      []*Camera
	Lights       []Light
	Materials    map[string]*PBRMaterial
	Textures     map[string]*AdvancedTexture
	Meshes       map[string]*Mesh
	Animations   map[string]*Animation
	Skins        map[string]*Skin         // Skinned animation support
	MorphTargets map[string]*MorphTargets // Morph targets support
	Extensions   *ExtensionRegistry       // GLTF extensions support
	ActiveCamera *Camera
	Name         string
}

// NewScene creates a new empty scene
func NewScene(name string) *Scene {
	return &Scene{
		RootNode:     NewSceneNode("root"),
		Cameras:      make([]*Camera, 0),
		Lights:       make([]Light, 0),
		Materials:    make(map[string]*PBRMaterial),
		Textures:     make(map[string]*AdvancedTexture),
		Meshes:       make(map[string]*Mesh),
		Animations:   make(map[string]*Animation),
		Skins:        make(map[string]*Skin),
		MorphTargets: make(map[string]*MorphTargets),
		Extensions:   NewExtensionRegistry(),
		Name:         name,
	}
}

// SceneNode represents a node in the scene hierarchy
type SceneNode struct {
	Name           string
	LocalTransform Matrix
	WorldTransform Matrix
	Parent         *SceneNode
	Children       []*SceneNode
	Mesh           *Mesh
	Material       *PBRMaterial
	Skin           *Skin         // Skinned mesh support
	MorphTargets   *MorphTargets // Morph target support
	Visible        bool
	CastShadows    bool
	ReceiveShadows bool
}

// NewSceneNode creates a new scene node
func NewSceneNode(name string) *SceneNode {
	return &SceneNode{
		Name:           name,
		LocalTransform: Identity(),
		WorldTransform: Identity(),
		Children:       make([]*SceneNode, 0),
		Visible:        true,
		CastShadows:    true,
		ReceiveShadows: true,
	}
}

// AddChild adds a child node to this node
func (node *SceneNode) AddChild(child *SceneNode) {
	if child == nil {
		return
	}

	// Remove child from previous parent
	if child.Parent != nil {
		child.Parent.RemoveChild(child)
	}

	// Add to this node
	child.Parent = node
	node.Children = append(node.Children, child)

	// Update transforms
	child.UpdateWorldTransform()
}

// RemoveChild removes a child node from this node
func (node *SceneNode) RemoveChild(child *SceneNode) {
	for i, c := range node.Children {
		if c == child {
			node.Children = append(node.Children[:i], node.Children[i+1:]...)
			child.Parent = nil
			break
		}
	}
}

// UpdateWorldTransform updates the world transform based on parent transforms
func (node *SceneNode) UpdateWorldTransform() {
	if node.Parent != nil {
		node.WorldTransform = node.Parent.WorldTransform.Mul(node.LocalTransform)
	} else {
		node.WorldTransform = node.LocalTransform
	}

	// Update all children
	for _, child := range node.Children {
		child.UpdateWorldTransform()
	}
}

// SetTransform sets the local transform of the node
func (node *SceneNode) SetTransform(transform Matrix) {
	node.LocalTransform = transform
	node.UpdateWorldTransform()
}

// Translate translates the node by the given vector
func (node *SceneNode) Translate(translation Vector) {
	node.LocalTransform = node.LocalTransform.Translate(translation)
	node.UpdateWorldTransform()
}

// Rotate rotates the node around the given axis by the given angle
func (node *SceneNode) Rotate(axis Vector, angle float64) {
	node.LocalTransform = node.LocalTransform.Rotate(axis, angle)
	node.UpdateWorldTransform()
}

// Scale scales the node by the given factors
func (node *SceneNode) Scale(scale Vector) {
	node.LocalTransform = node.LocalTransform.Scale(scale)
	node.UpdateWorldTransform()
}

// GetWorldPosition returns the world position of the node
func (node *SceneNode) GetWorldPosition() Vector {
	return node.WorldTransform.MulPosition(Vector{0, 0, 0})
}

// VisitNodes visits all nodes in the hierarchy with the given function
func (node *SceneNode) VisitNodes(visitor func(*SceneNode)) {
	visitor(node)
	for _, child := range node.Children {
		child.VisitNodes(visitor)
	}
}

// GetRenderableNodes returns all nodes that have both mesh and material
func (node *SceneNode) GetRenderableNodes() []*SceneNode {
	var renderables []*SceneNode

	node.VisitNodes(func(n *SceneNode) {
		if n.Visible && n.Mesh != nil && n.Material != nil {
			renderables = append(renderables, n)
		}
	})

	return renderables
}

// FindChild finds a child node by name (recursive)
func (node *SceneNode) FindChild(name string) *SceneNode {
	if node.Name == name {
		return node
	}

	for _, child := range node.Children {
		if result := child.FindChild(name); result != nil {
			return result
		}
	}

	return nil
}

// SceneBounds calculates the bounding box of the entire scene
func (scene *Scene) GetBounds() Box {
	bounds := EmptyBox

	scene.RootNode.VisitNodes(func(node *SceneNode) {
		if node.Mesh != nil {
			// Transform mesh bounds to world space
			meshBounds := node.Mesh.BoundingBox()
			worldBounds := node.WorldTransform.MulBox(meshBounds)
			bounds = bounds.Extend(worldBounds)
		}
	})

	return bounds
}

// AddCamera adds a camera to the scene
func (scene *Scene) AddCamera(camera *Camera) {
	scene.Cameras = append(scene.Cameras, camera)
	if scene.ActiveCamera == nil {
		scene.ActiveCamera = camera
	}
}

// SetActiveCamera sets the active camera by name
func (scene *Scene) SetActiveCamera(name string) bool {
	for _, camera := range scene.Cameras {
		if camera.Name == name {
			scene.ActiveCamera = camera
			return true
		}
	}
	return false
}

// AddLight adds a light to the scene
func (scene *Scene) AddLight(light Light) {
	scene.Lights = append(scene.Lights, light)
}

// AddMaterial adds a material to the scene material library
func (scene *Scene) AddMaterial(name string, material *PBRMaterial) {
	scene.Materials[name] = material
}

// GetMaterial gets a material from the scene material library
func (scene *Scene) GetMaterial(name string) *PBRMaterial {
	return scene.Materials[name]
}

// AddTexture adds a texture to the scene texture library
func (scene *Scene) AddTexture(name string, texture *AdvancedTexture) {
	scene.Textures[name] = texture
}

// GetTexture gets a texture from the scene texture library
func (scene *Scene) GetTexture(name string) *AdvancedTexture {
	return scene.Textures[name]
}

// AddMesh adds a mesh to the scene mesh library
func (scene *Scene) AddMesh(name string, mesh *Mesh) {
	scene.Meshes[name] = mesh
}

// GetMesh gets a mesh from the scene mesh library
func (scene *Scene) GetMesh(name string) *Mesh {
	return scene.Meshes[name]
}

// CreateMeshNode creates a new scene node with a mesh and material
func (scene *Scene) CreateMeshNode(name, meshName, materialName string) *SceneNode {
	node := NewSceneNode(name)
	node.Mesh = scene.GetMesh(meshName)
	node.Material = scene.GetMaterial(materialName)
	return node
}

// AddDirectionalLight adds a directional light to the scene
func (scene *Scene) AddDirectionalLight(direction Vector, color Color, intensity float64) {
	light := Light{
		Type:      DirectionalLight,
		Direction: direction.Normalize(),
		Color:     color,
		Intensity: intensity,
	}
	scene.AddLight(light)
}

// AddPointLight adds a point light to the scene
func (scene *Scene) AddPointLight(position Vector, color Color, intensity, range_ float64) {
	light := Light{
		Type:      PointLight,
		Position:  position,
		Color:     color,
		Intensity: intensity,
		Range:     range_,
	}
	scene.AddLight(light)
}

// AddSpotLight adds a spot light to the scene
func (scene *Scene) AddSpotLight(position, direction Vector, color Color, intensity, range_, innerCone, outerCone float64) {
	light := Light{
		Type:      SpotLight,
		Position:  position,
		Direction: direction.Normalize(),
		Color:     color,
		Intensity: intensity,
		Range:     range_,
		InnerCone: innerCone,
		OuterCone: outerCone,
	}
	scene.AddLight(light)
}

// AddAmbientLight adds an ambient light to the scene
// Ambient light provides uniform illumination to all surfaces
func (scene *Scene) AddAmbientLight(color Color, intensity float64) {
	light := Light{
		Type:      AmbientLight,
		Color:     color,
		Intensity: intensity,
	}
	scene.AddLight(light)
}

// ClearLights removes all lights from the scene
func (scene *Scene) ClearLights() {
	scene.Lights = make([]Light, 0)
}

// GetLightsByType returns all lights of a specific type
func (scene *Scene) GetLightsByType(lightType LightType) []Light {
	var lights []Light
	for _, light := range scene.Lights {
		if light.Type == lightType {
			lights = append(lights, light)
		}
	}
	return lights
}

// AddSkin adds a skin to the scene skin library
func (scene *Scene) AddSkin(name string, skin *Skin) {
	scene.Skins[name] = skin
}

// GetSkin gets a skin from the scene skin library
func (scene *Scene) GetSkin(name string) *Skin {
	return scene.Skins[name]
}

// AddMorphTargets adds morph targets to the scene morph target library
func (scene *Scene) AddMorphTargets(name string, targets *MorphTargets) {
	scene.MorphTargets[name] = targets
}

// GetMorphTargets gets morph targets from the scene morph target library
func (scene *Scene) GetMorphTargets(name string) *MorphTargets {
	return scene.MorphTargets[name]
}

// AddAnimation adds an animation to the scene animation library
func (scene *Scene) AddAnimation(name string, animation *Animation) {
	scene.Animations[name] = animation
}

// GetAnimation gets an animation from the scene animation library
func (scene *Scene) GetAnimation(name string) *Animation {
	return scene.Animations[name]
}

// ProcessGLTFExtensions processes GLTF extensions for the scene
func (scene *Scene) ProcessGLTFExtensions(extensions map[string]interface{}) error {
	return scene.Extensions.ProcessExtensions(extensions, scene)
}

// GetSupportedGLTFExtensions returns supported GLTF extensions
func (scene *Scene) GetSupportedGLTFExtensions() []string {
	return scene.Extensions.GetSupportedExtensions()
}

// CreateSkinnedMeshNode creates a scene node with mesh, material, and skin
func (scene *Scene) CreateSkinnedMeshNode(name, meshName, materialName, skinName string) *SceneNode {
	node := NewSceneNode(name)
	node.Mesh = scene.GetMesh(meshName)
	node.Material = scene.GetMaterial(materialName)
	node.Skin = scene.GetSkin(skinName)
	return node
}

// CreateMorphTargetMeshNode creates a scene node with mesh, material, and morph targets
func (scene *Scene) CreateMorphTargetMeshNode(name, meshName, materialName, morphTargetName string) *SceneNode {
	node := NewSceneNode(name)
	node.Mesh = scene.GetMesh(meshName)
	node.Material = scene.GetMaterial(materialName)
	node.MorphTargets = scene.GetMorphTargets(morphTargetName)
	return node
}

// UpdateSkinnedMeshes updates all skinned meshes in the scene
func (scene *Scene) UpdateSkinnedMeshes() {
	scene.RootNode.VisitNodes(func(node *SceneNode) {
		if node.Skin != nil && node.Mesh != nil {
			// Update joint matrices
			node.Skin.UpdateJointMatrices()

			// Apply skinning (would need joint matrices array)
			// This is a placeholder for the actual skinning implementation
		}
	})
}

// ApplyMorphTargetsToMeshes applies morph targets to all relevant meshes in the scene
func (scene *Scene) ApplyMorphTargetsToMeshes() {
	scene.RootNode.VisitNodes(func(node *SceneNode) {
		if node.MorphTargets != nil && node.Mesh != nil {
			// Apply morph target deformation
			deformedMesh := ApplyMorphTargets(node.Mesh, node.MorphTargets)
			node.Mesh = deformedMesh
		}
	})
}
