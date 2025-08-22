package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// GLTF结构体定义（简化版，用于分析）
type GLTFAnalysis struct {
	Asset       Asset        `json:"asset"`
	Scene       int          `json:"scene"`
	Scenes      []Scene      `json:"scenes"`
	Nodes       []Node       `json:"nodes"`
	Meshes      []Mesh       `json:"meshes"`
	Accessors   []Accessor   `json:"accessors"`
	BufferViews []BufferView `json:"bufferViews"`
	Buffers     []Buffer     `json:"buffers"`
	Materials   []Material   `json:"materials"`
	Textures    []Texture    `json:"textures"`
	Images      []Image      `json:"images"`
	Samplers    []Sampler    `json:"samplers"`
}

type Asset struct {
	Version   string `json:"version"`
	Generator string `json:"generator"`
}

type Scene struct {
	Nodes []int  `json:"nodes"`
	Name  string `json:"name,omitempty"`
}

type Node struct {
	Mesh        *int      `json:"mesh,omitempty"`
	Translation []float64 `json:"translation,omitempty"`
	Rotation    []float64 `json:"rotation,omitempty"`
	Scale       []float64 `json:"scale,omitempty"`
	Name        string    `json:"name,omitempty"`
	Children    []int     `json:"children,omitempty"`
}

type Mesh struct {
	Primitives []Primitive `json:"primitives"`
	Name       string      `json:"name,omitempty"`
}

type Primitive struct {
	Attributes map[string]int `json:"attributes"`
	Indices    *int           `json:"indices,omitempty"`
	Material   *int           `json:"material,omitempty"`
}

type Accessor struct {
	BufferView    *int      `json:"bufferView,omitempty"`
	ByteOffset    int       `json:"byteOffset,omitempty"`
	ComponentType int       `json:"componentType"`
	Count         int       `json:"count"`
	Type          string    `json:"type"`
	Max           []float64 `json:"max,omitempty"`
	Min           []float64 `json:"min,omitempty"`
	Name          string    `json:"name,omitempty"`
}

type BufferView struct {
	Buffer     int    `json:"buffer"`
	ByteOffset int    `json:"byteOffset,omitempty"`
	ByteLength int    `json:"byteLength"`
	ByteStride *int   `json:"byteStride,omitempty"`
	Name       string `json:"name,omitempty"`
}

type Buffer struct {
	URI        string `json:"uri"`
	ByteLength int    `json:"byteLength"`
}

type Material struct {
	PBRMetallicRoughness *PBRMetallicRoughness `json:"pbrMetallicRoughness,omitempty"`
	DoubleSided          bool                  `json:"doubleSided,omitempty"`
	Name                 string                `json:"name,omitempty"`
}

type PBRMetallicRoughness struct {
	BaseColorFactor  []float64    `json:"baseColorFactor,omitempty"`
	BaseColorTexture *TextureInfo `json:"baseColorTexture,omitempty"`
	MetallicFactor   *float64     `json:"metallicFactor,omitempty"`
	RoughnessFactor  *float64     `json:"roughnessFactor,omitempty"`
}

type TextureInfo struct {
	Index int `json:"index"`
}

type Texture struct {
	Sampler *int   `json:"sampler,omitempty"`
	Source  int    `json:"source"`
	Name    string `json:"name,omitempty"`
}

type Image struct {
	URI  string `json:"uri"`
	Name string `json:"name,omitempty"`
}

type Sampler struct {
	MagFilter *int `json:"magFilter,omitempty"`
	MinFilter *int `json:"minFilter,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run gltf_analyzer.go <path_to_gltf_file>")
		fmt.Println("Example: go run gltf_analyzer.go examples/mug.gltf")
		return
	}

	gltfPath := os.Args[1]

	// 读取GLTF文件
	data, err := ioutil.ReadFile(gltfPath)
	if err != nil {
		fmt.Printf("Error reading GLTF file: %v\n", err)
		return
	}

	// 解析JSON
	var gltf GLTFAnalysis
	err = json.Unmarshal(data, &gltf)
	if err != nil {
		fmt.Printf("Error parsing GLTF JSON: %v\n", err)
		return
	}

	// 开始分析
	fmt.Printf("=== GLTF File Analysis: %s ===\n\n", gltfPath)

	analyzeAsset(gltf.Asset)
	analyzeScene(gltf, gltf.Scene)
	analyzeNodes(gltf.Nodes)
	analyzeMeshes(gltf.Meshes)
	analyzeMaterials(gltf.Materials)
	analyzeTextures(gltf.Textures, gltf.Images, gltf.Samplers)
	analyzeBuffers(gltf.Buffers, gltf.BufferViews, gltf.Accessors)

	// 生成对应的Go代码
	generateGoCode(gltf, gltfPath)

	fmt.Println("Analysis completed!")
}

func analyzeAsset(asset Asset) {
	fmt.Printf("Asset Information:\n")
	fmt.Printf("  GLTF Version: %s\n", asset.Version)
	fmt.Printf("  Generator: %s\n", asset.Generator)
	fmt.Printf("\n")
}

func analyzeScene(gltf GLTFAnalysis, sceneIndex int) {
	if sceneIndex >= len(gltf.Scenes) {
		return
	}

	scene := gltf.Scenes[sceneIndex]
	fmt.Printf("Active Scene (index %d):\n", sceneIndex)
	fmt.Printf("  Root Nodes: %v\n", scene.Nodes)
	fmt.Printf("\n")
}

func analyzeNodes(nodes []Node) {
	fmt.Printf("Nodes (%d total):\n", len(nodes))
	for i, node := range nodes {
		fmt.Printf("  Node %d:\n", i)
		fmt.Printf("    Name: %s\n", node.Name)
		if node.Mesh != nil {
			fmt.Printf("    Mesh: %d\n", *node.Mesh)
		}
		if len(node.Translation) > 0 {
			fmt.Printf("    Translation: %v\n", node.Translation)
		}
		if len(node.Rotation) > 0 {
			fmt.Printf("    Rotation: %v\n", node.Rotation)
		}
		if len(node.Scale) > 0 {
			fmt.Printf("    Scale: %v\n", node.Scale)
		}
		if len(node.Children) > 0 {
			fmt.Printf("    Children: %v\n", node.Children)
		}
	}
	fmt.Printf("\n")
}

func analyzeMeshes(meshes []Mesh) {
	fmt.Printf("Meshes (%d total):\n", len(meshes))
	for i, mesh := range meshes {
		fmt.Printf("  Mesh %d (%s):\n", i, mesh.Name)
		fmt.Printf("    Primitives: %d\n", len(mesh.Primitives))

		for j, primitive := range mesh.Primitives {
			fmt.Printf("    Primitive %d:\n", j)
			fmt.Printf("      Attributes: %v\n", primitive.Attributes)
			if primitive.Indices != nil {
				fmt.Printf("      Indices: accessor %d\n", *primitive.Indices)
			}
			if primitive.Material != nil {
				fmt.Printf("      Material: %d\n", *primitive.Material)
			}
		}
	}
	fmt.Printf("\n")
}

func analyzeMaterials(materials []Material) {
	fmt.Printf("Materials (%d total):\n", len(materials))
	for i, material := range materials {
		fmt.Printf("  Material %d (%s):\n", i, material.Name)
		fmt.Printf("    Double Sided: %t\n", material.DoubleSided)

		if material.PBRMetallicRoughness != nil {
			pbr := material.PBRMetallicRoughness
			fmt.Printf("    PBR Properties:\n")

			if len(pbr.BaseColorFactor) > 0 {
				fmt.Printf("      Base Color: %v\n", pbr.BaseColorFactor)
			}
			if pbr.MetallicFactor != nil {
				fmt.Printf("      Metallic Factor: %.3f\n", *pbr.MetallicFactor)
			}
			if pbr.RoughnessFactor != nil {
				fmt.Printf("      Roughness Factor: %.3f\n", *pbr.RoughnessFactor)
			}
			if pbr.BaseColorTexture != nil {
				fmt.Printf("      Base Color Texture: %d\n", pbr.BaseColorTexture.Index)
			}
		}
	}
	fmt.Printf("\n")
}

func analyzeTextures(textures []Texture, images []Image, samplers []Sampler) {
	fmt.Printf("Textures (%d total):\n", len(textures))
	for i, texture := range textures {
		fmt.Printf("  Texture %d (%s):\n", i, texture.Name)
		fmt.Printf("    Source Image: %d", texture.Source)
		if texture.Source < len(images) {
			fmt.Printf(" (%s)", images[texture.Source].URI)
		}
		fmt.Printf("\n")
		if texture.Sampler != nil {
			fmt.Printf("    Sampler: %d\n", *texture.Sampler)
		}
	}

	if len(images) > 0 {
		fmt.Printf("\nImages (%d total):\n", len(images))
		for i, image := range images {
			fmt.Printf("  Image %d: %s\n", i, image.URI)
		}
	}
	fmt.Printf("\n")
}

func analyzeBuffers(buffers []Buffer, bufferViews []BufferView, accessors []Accessor) {
	fmt.Printf("Buffers (%d total):\n", len(buffers))
	for i, buffer := range buffers {
		fmt.Printf("  Buffer %d:\n", i)
		fmt.Printf("    URI: %s\n", buffer.URI)
		fmt.Printf("    Size: %d bytes (%.2f KB)\n", buffer.ByteLength, float64(buffer.ByteLength)/1024.0)
	}

	fmt.Printf("\nBuffer Views (%d total):\n", len(bufferViews))
	for i, view := range bufferViews {
		fmt.Printf("  BufferView %d (%s):\n", i, view.Name)
		fmt.Printf("    Buffer: %d\n", view.Buffer)
		fmt.Printf("    Offset: %d bytes\n", view.ByteOffset)
		fmt.Printf("    Length: %d bytes\n", view.ByteLength)
		if view.ByteStride != nil {
			fmt.Printf("    Stride: %d bytes\n", *view.ByteStride)
		}
	}

	fmt.Printf("\nAccessors (%d total):\n", len(accessors))
	vertexCount := 0
	indexCount := 0

	for i, accessor := range accessors {
		fmt.Printf("  Accessor %d (%s):\n", i, accessor.Name)
		fmt.Printf("    Type: %s, Component: %d\n", accessor.Type, accessor.ComponentType)
		fmt.Printf("    Count: %d\n", accessor.Count)

		if len(accessor.Max) > 0 {
			fmt.Printf("    Range: %v to %v\n", accessor.Min, accessor.Max)
		}

		// 统计顶点和索引数量
		if accessor.Type == "VEC3" && (accessor.Name == "accessorPositions" ||
			(len(accessor.Name) == 0 && accessor.ComponentType == 5126)) {
			vertexCount += accessor.Count
		}
		if accessor.Type == "SCALAR" && (accessor.Name == "accessorIndices" ||
			(len(accessor.Name) == 0 && accessor.ComponentType == 5123)) {
			indexCount += accessor.Count
		}
	}

	fmt.Printf("\nGeometry Summary:\n")
	fmt.Printf("  Total Vertices: ~%d\n", vertexCount)
	fmt.Printf("  Total Indices: %d\n", indexCount)
	fmt.Printf("  Approximate Triangles: %d\n", indexCount/3)
	fmt.Printf("\n")
}

func generateGoCode(gltf GLTFAnalysis, originalPath string) {
	fmt.Printf("=== Generated Go Code Structure ===\n")

	// 生成材质创建代码
	fmt.Printf("// Material Setup Code:\n")
	for i, material := range gltf.Materials {
		fmt.Printf("func createMaterial%d() *fauxgl.PBRMaterial {\n", i)
		fmt.Printf("    material := fauxgl.NewPBRMaterial()\n")
		fmt.Printf("    // %s\n", material.Name)

		if material.PBRMetallicRoughness != nil {
			pbr := material.PBRMetallicRoughness
			if len(pbr.BaseColorFactor) >= 4 {
				fmt.Printf("    material.BaseColorFactor = fauxgl.Color{%.3f, %.3f, %.3f, %.3f}\n",
					pbr.BaseColorFactor[0], pbr.BaseColorFactor[1],
					pbr.BaseColorFactor[2], pbr.BaseColorFactor[3])
			}
			if pbr.MetallicFactor != nil {
				fmt.Printf("    material.MetallicFactor = %.3f\n", *pbr.MetallicFactor)
			}
			if pbr.RoughnessFactor != nil {
				fmt.Printf("    material.RoughnessFactor = %.3f\n", *pbr.RoughnessFactor)
			}
		}

		fmt.Printf("    material.DoubleSided = %t\n", material.DoubleSided)
		fmt.Printf("    return material\n}\n\n")
	}

	// 生成场景设置代码
	fmt.Printf("// Scene Setup Code:\n")
	fmt.Printf("func setupMugScene() *fauxgl.Scene {\n")
	fmt.Printf("    scene := fauxgl.NewScene(\"Mug Scene\")\n")
	fmt.Printf("    \n")

	// 添加材质
	for i := range gltf.Materials {
		fmt.Printf("    scene.AddMaterial(\"material_%d\", createMaterial%d())\n", i, i)
	}

	fmt.Printf("    \n")
	fmt.Printf("    // Load GLTF mesh data\n")
	fmt.Printf("    mesh, err := fauxgl.LoadGLTF(\"%s\")\n", originalPath)
	fmt.Printf("    if err != nil {\n")
	fmt.Printf("        panic(err)\n")
	fmt.Printf("    }\n")
	fmt.Printf("    scene.AddMesh(\"mug_mesh\", mesh)\n")
	fmt.Printf("    \n")
	fmt.Printf("    // Create scene node\n")
	fmt.Printf("    node := scene.CreateMeshNode(\"mug_node\", \"mug_mesh\", \"material_1\")\n")

	// 添加变换
	if len(gltf.Nodes) > 0 && len(gltf.Nodes[0].Translation) >= 3 {
		trans := gltf.Nodes[0].Translation
		fmt.Printf("    node.SetTransform(fauxgl.Translate(fauxgl.V(%.6f, %.6f, %.6f)))\n",
			trans[0], trans[1], trans[2])
	}

	fmt.Printf("    scene.RootNode.AddChild(node)\n")
	fmt.Printf("    \n")
	fmt.Printf("    return scene\n")
	fmt.Printf("}\n\n")

	fmt.Printf("=================================\n\n")
}
