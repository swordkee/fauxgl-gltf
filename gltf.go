package fauxgl

import (
	"fmt"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

// LoadGLTFScene loads a complete GLTF scene with materials, cameras, lights, etc.
func LoadGLTFScene(path string) (*Scene, error) {
	doc, err := gltf.Open(path)
	if err != nil {
		return nil, err
	}

	scene := NewScene("GLTF Scene")
	loader := &GLTFLoader{doc: doc, scene: scene}

	// Load textures
	err = loader.loadTextures()
	if err != nil {
		return nil, err
	}

	// Load materials
	err = loader.loadMaterials()
	if err != nil {
		return nil, err
	}

	// Load meshes
	err = loader.loadMeshes()
	if err != nil {
		return nil, err
	}

	// Load cameras
	err = loader.loadCameras()
	if err != nil {
		return nil, err
	}

	// Load lights
	err = loader.loadLights()
	if err != nil {
		return nil, err
	}

	// Load scene nodes
	if len(doc.Scenes) > 0 {
		err = loader.loadSceneNodes(doc.Scenes[0])
		if err != nil {
			return nil, err
		}
	}

	return scene, nil
}

// GLTFLoader handles loading of GLTF files
type GLTFLoader struct {
	doc   *gltf.Document
	scene *Scene
}

// LoadGLTF loads a GLTF file and returns only the mesh (legacy function)
func LoadGLTF(path string) (*Mesh, error) {
	scene, err := LoadGLTFScene(path)
	if err != nil {
		return nil, err
	}

	// Combine all meshes into one for backward compatibility
	var allTriangles []*Triangle
	scene.RootNode.VisitNodes(func(node *SceneNode) {
		if node.Mesh != nil {
			allTriangles = append(allTriangles, node.Mesh.Triangles...)
		}
	})

	return NewTriangleMesh(allTriangles), nil
}

// loadTextures loads all textures from the GLTF document
func (loader *GLTFLoader) loadTextures() error {
	for i, texture := range loader.doc.Textures {
		if texture.Source == nil {
			continue
		}

		sourceIndex := int(*texture.Source)
		if sourceIndex >= len(loader.doc.Images) {
			continue
		}

		image := loader.doc.Images[sourceIndex]
		if image.URI == "" {
			continue // Skip embedded images for now
		}

		// Load texture from URI
		textureName := fmt.Sprintf("texture_%d", i)
		advTexture, err := LoadAdvancedTexture(image.URI, BaseColorTexture)
		if err != nil {
			continue // Skip failed textures
		}

		loader.scene.AddTexture(textureName, advTexture)
	}

	return nil
}

// loadMaterials loads all materials from the GLTF document
func (loader *GLTFLoader) loadMaterials() error {
	for i, gltfMat := range loader.doc.Materials {
		material := NewPBRMaterial()

		// Base color
		if gltfMat.PBRMetallicRoughness != nil {
			pbr := gltfMat.PBRMetallicRoughness

			material.BaseColorFactor = Color{
				float64(pbr.BaseColorFactor[0]),
				float64(pbr.BaseColorFactor[1]),
				float64(pbr.BaseColorFactor[2]),
				float64(pbr.BaseColorFactor[3]),
			}

			if pbr.MetallicFactor != nil {
				material.MetallicFactor = float64(*pbr.MetallicFactor)
			}

			if pbr.RoughnessFactor != nil {
				material.RoughnessFactor = float64(*pbr.RoughnessFactor)
			}

			// Base color texture
			if pbr.BaseColorTexture != nil {
				textureName := fmt.Sprintf("texture_%d", pbr.BaseColorTexture.Index)
				if texture := loader.scene.GetTexture(textureName); texture != nil {
					material.BaseColorTexture = texture
				}
			}

			// Metallic roughness texture
			if pbr.MetallicRoughnessTexture != nil {
				textureName := fmt.Sprintf("texture_%d", pbr.MetallicRoughnessTexture.Index)
				if texture := loader.scene.GetTexture(textureName); texture != nil {
					material.MetallicRoughnessTexture = texture
				}
			}
		}

		// Normal texture
		if gltfMat.NormalTexture != nil {
			textureName := fmt.Sprintf("texture_%d", gltfMat.NormalTexture.Index)
			if texture := loader.scene.GetTexture(textureName); texture != nil {
				material.NormalTexture = texture
				if gltfMat.NormalTexture.Scale != nil {
					material.NormalScale = float64(*gltfMat.NormalTexture.Scale)
				}
			}
		}

		// Occlusion texture
		if gltfMat.OcclusionTexture != nil {
			textureName := fmt.Sprintf("texture_%d", gltfMat.OcclusionTexture.Index)
			if texture := loader.scene.GetTexture(textureName); texture != nil {
				material.OcclusionTexture = texture
				if gltfMat.OcclusionTexture.Strength != nil {
					material.OcclusionStrength = float64(*gltfMat.OcclusionTexture.Strength)
				}
			}
		}

		// Emissive
		material.EmissiveFactor = Color{
			float64(gltfMat.EmissiveFactor[0]),
			float64(gltfMat.EmissiveFactor[1]),
			float64(gltfMat.EmissiveFactor[2]),
			1.0,
		}

		if gltfMat.EmissiveTexture != nil {
			textureName := fmt.Sprintf("texture_%d", gltfMat.EmissiveTexture.Index)
			if texture := loader.scene.GetTexture(textureName); texture != nil {
				material.EmissiveTexture = texture
			}
		}

		// Alpha mode
		switch gltfMat.AlphaMode {
		case gltf.AlphaOpaque:
			material.AlphaMode = AlphaOpaque
		case gltf.AlphaMask:
			material.AlphaMode = AlphaMask
			if gltfMat.AlphaCutoff != nil {
				material.AlphaCutoff = float64(*gltfMat.AlphaCutoff)
			}
		case gltf.AlphaBlend:
			material.AlphaMode = AlphaBlend
		}

		material.DoubleSided = gltfMat.DoubleSided

		materialName := fmt.Sprintf("material_%d", i)
		loader.scene.AddMaterial(materialName, material)
	}

	return nil
}

// loadCameras loads all cameras from the GLTF document
func (loader *GLTFLoader) loadCameras() error {
	for i, gltfCamera := range loader.doc.Cameras {
		cameraName := fmt.Sprintf("camera_%d", i)

		var camera *Camera
		if gltfCamera.Perspective != nil {
			p := gltfCamera.Perspective
			camera = NewPerspectiveCamera(
				cameraName,
				Vector{0, 0, 0},  // Position will be set by node
				Vector{0, 0, -1}, // Default target
				Vector{0, 1, 0},  // Up
				float64(p.Yfov),
				float64(1.0), // Default aspect ratio if not specified
				float64(p.Znear),
				float64(1000.0), // Default far if not specified
			)
			if p.AspectRatio != nil {
				camera.AspectRatio = float64(*p.AspectRatio)
			}
			if p.Zfar != nil {
				camera.FarPlane = float64(*p.Zfar)
			}
		} else if gltfCamera.Orthographic != nil {
			o := gltfCamera.Orthographic
			camera = NewOrthographicCamera(
				cameraName,
				Vector{0, 0, 0},                 // Position will be set by node
				Vector{0, 0, -1},                // Default target
				Vector{0, 1, 0},                 // Up
				float64(o.Ymag)*2,               // OrthoSize
				float64(o.Xmag)/float64(o.Ymag), // AspectRatio
				float64(o.Znear),
				float64(o.Zfar),
			)
		}

		if camera != nil {
			loader.scene.AddCamera(camera)
		}
	}

	return nil
}

// loadLights loads lights from GLTF extensions
func (loader *GLTFLoader) loadLights() error {
	// GLTF lights are typically in extensions
	// For now, add a default light
	defaultLight := Light{
		Type:      DirectionalLight,
		Direction: Vector{-1, -1, -1}.Normalize(),
		Color:     White,
		Intensity: 1.0,
	}

	loader.scene.AddLight(defaultLight)
	return nil
}

// loadSceneNodes loads the scene hierarchy
func (loader *GLTFLoader) loadSceneNodes(gltfScene *gltf.Scene) error {
	// Load nodes recursively
	for _, nodeIndex := range gltfScene.Nodes {
		if int(nodeIndex) < len(loader.doc.Nodes) {
			childNode, err := loader.loadNode(int(nodeIndex), nil)
			if err != nil {
				return err
			}
			loader.scene.RootNode.AddChild(childNode)
		}
	}

	return nil
}

// loadNode loads a single node and its children
func (loader *GLTFLoader) loadNode(nodeIndex int, parent *SceneNode) (*SceneNode, error) {
	gltfNode := loader.doc.Nodes[nodeIndex]

	nodeName := gltfNode.Name
	if nodeName == "" {
		nodeName = fmt.Sprintf("node_%d", nodeIndex)
	}

	node := NewSceneNode(nodeName)

	// Set transform
	var hasMatrix bool
	for _, v := range gltfNode.Matrix {
		if v != 0 {
			hasMatrix = true
			break
		}
	}

	if hasMatrix {
		// Matrix transform
		m := gltfNode.Matrix
		node.SetTransform(Matrix{
			float64(m[0]), float64(m[4]), float64(m[8]), float64(m[12]),
			float64(m[1]), float64(m[5]), float64(m[9]), float64(m[13]),
			float64(m[2]), float64(m[6]), float64(m[10]), float64(m[14]),
			float64(m[3]), float64(m[7]), float64(m[11]), float64(m[15]),
		})
	} else {
		// TRS transform
		transform := Identity()

		// Translation
		var hasTranslation bool
		for _, v := range gltfNode.Translation {
			if v != 0 {
				hasTranslation = true
				break
			}
		}
		if hasTranslation {
			t := gltfNode.Translation
			transform = transform.Translate(Vector{float64(t[0]), float64(t[1]), float64(t[2])})
		}

		// Rotation (quaternion)
		var hasRotation bool
		for _, v := range gltfNode.Rotation {
			if v != 0 {
				hasRotation = true
				break
			}
		}
		if hasRotation {
			// Convert quaternion to rotation matrix
			// For simplicity, we'll skip this for now
		}

		// Scale
		var hasScale bool
		for i, v := range gltfNode.Scale {
			if i < 3 && v != 1.0 { // Scale default is 1.0
				hasScale = true
				break
			}
		}
		if hasScale {
			s := gltfNode.Scale
			transform = transform.Scale(Vector{float64(s[0]), float64(s[1]), float64(s[2])})
		}

		node.SetTransform(transform)
	}

	// Assign mesh and material
	if gltfNode.Mesh != nil {
		meshName := fmt.Sprintf("mesh_%d", *gltfNode.Mesh)
		node.Mesh = loader.scene.GetMesh(meshName)

		// For now, use the first material
		if len(loader.scene.Materials) > 0 {
			for _, material := range loader.scene.Materials {
				node.Material = material
				break
			}
		}
	}

	// Load children
	for _, childIndex := range gltfNode.Children {
		if int(childIndex) < len(loader.doc.Nodes) {
			childNode, err := loader.loadNode(int(childIndex), node)
			if err != nil {
				return nil, err
			}
			node.AddChild(childNode)
		}
	}

	return node, nil
}

// loadMeshes loads all meshes from the GLTF document
func (loader *GLTFLoader) loadMeshes() error {
	for i, gltfMesh := range loader.doc.Meshes {
		var triangles []*Triangle

		// 遍历网格中的所有图元
		for _, primitive := range gltfMesh.Primitives {
			// 只处理三角形图元

			// 获取顶点位置数据
			positionAccessor := loader.doc.Accessors[primitive.Attributes[gltf.POSITION]]
			posBuffer := [][3]float32{}
			positionBuffer, err := modeler.ReadPosition(loader.doc, positionAccessor, posBuffer)
			if err != nil {
				return err
			}

			// 获取法线数据（如果存在）
			var normalBuffer [][3]float32
			if normalAccessorIndex, ok := primitive.Attributes[gltf.NORMAL]; ok {
				normalBuffer1 := [][3]float32{}
				normalAccessor := loader.doc.Accessors[normalAccessorIndex]
				normalBuffer, err = modeler.ReadNormal(loader.doc, normalAccessor, normalBuffer1)
				if err != nil {
					return err
				}
			}

			// 获取纹理坐标数据（如果存在）
			var texCoordBuffer [][2]float32
			if texCoordAccessorIndex, ok := primitive.Attributes[gltf.TEXCOORD_0]; ok {
				uvBuffer := [][2]float32{}
				texCoordAccessor := loader.doc.Accessors[texCoordAccessorIndex]
				texCoordBuffer, err = modeler.ReadTextureCoord(loader.doc, texCoordAccessor, uvBuffer)
				if err != nil {
					return err
				}
			}

			// 获取索引数据
			var indices []uint32
			if primitive.Indices != nil {
				indexAccessor := loader.doc.Accessors[*primitive.Indices]
				indexBuffer := []uint32{}
				indices, err = modeler.ReadIndices(loader.doc, indexAccessor, indexBuffer)
				if err != nil {
					return err
				}
			} else {
				// 如果没有索引，则按顺序生成
				indices = make([]uint32, len(positionBuffer))
				for i := range indices {
					indices[i] = uint32(i)
				}
			}

			// 将顶点数据转换为三角形
			for i := 0; i < len(indices); i += 3 {
				t := &Triangle{}

				// 第一个顶点
				i1 := indices[i]
				t.V1.Position = Vector{
					float64(positionBuffer[i1][0]),
					float64(positionBuffer[i1][1]),
					float64(positionBuffer[i1][2]),
				}
				if len(normalBuffer) > 0 {
					t.V1.Normal = Vector{
						float64(normalBuffer[i1][0]),
						float64(normalBuffer[i1][1]),
						float64(normalBuffer[i1][2]),
					}
				}
				if len(texCoordBuffer) > 0 {
					t.V1.Texture = Vector{
						float64(texCoordBuffer[i1][0]),
						float64(texCoordBuffer[i1][1]),
						0,
					}
				}

				// 第二个顶点
				i2 := indices[i+1]
				t.V2.Position = Vector{
					float64(positionBuffer[i2][0]),
					float64(positionBuffer[i2][1]),
					float64(positionBuffer[i2][2]),
				}
				if len(normalBuffer) > 0 {
					t.V2.Normal = Vector{
						float64(normalBuffer[i2][0]),
						float64(normalBuffer[i2][1]),
						float64(normalBuffer[i2][2]),
					}
				}
				if len(texCoordBuffer) > 0 {
					t.V2.Texture = Vector{
						float64(texCoordBuffer[i2][0]),
						float64(texCoordBuffer[i2][1]),
						0,
					}
				}

				// 第三个顶点
				i3 := indices[i+2]
				t.V3.Position = Vector{
					float64(positionBuffer[i3][0]),
					float64(positionBuffer[i3][1]),
					float64(positionBuffer[i3][2]),
				}
				if len(normalBuffer) > 0 {
					t.V3.Normal = Vector{
						float64(normalBuffer[i3][0]),
						float64(normalBuffer[i3][1]),
						float64(normalBuffer[i3][2]),
					}
				}
				if len(texCoordBuffer) > 0 {
					t.V3.Texture = Vector{
						float64(texCoordBuffer[i3][0]),
						float64(texCoordBuffer[i3][1]),
						0,
					}
				}

				// 如果没有法线数据，则自动计算
				if len(normalBuffer) == 0 {
					t.FixNormals()
				}

				triangles = append(triangles, t)
			}
		}

		mesh := NewTriangleMesh(triangles)
		meshName := fmt.Sprintf("mesh_%d", i)
		loader.scene.AddMesh(meshName, mesh)
	}

	return nil
}

// LoadGLTFWithMaterials loads a GLTF file and creates objects with PBR materials
func LoadGLTFWithMaterials(path string) ([]*SceneNode, error) {
	scene, err := LoadGLTFScene(path)
	if err != nil {
		return nil, err
	}

	return scene.RootNode.GetRenderableNodes(), nil
}
