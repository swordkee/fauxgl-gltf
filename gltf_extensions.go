package fauxgl

import (
	"fmt"
)

// GLTFExtension represents a GLTF extension
type GLTFExtension struct {
	Name string
	Data map[string]interface{}
}

// GLTFExtensionHandler handles specific GLTF extensions
type GLTFExtensionHandler interface {
	GetName() string
	Process(data map[string]interface{}, scene *Scene) error
}

// KHRLightsPunctualExtension handles KHR_lights_punctual extension
type KHRLightsPunctualExtension struct{}

func (ext *KHRLightsPunctualExtension) GetName() string {
	return "KHR_lights_punctual"
}

func (ext *KHRLightsPunctualExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process punctual lights extension
	if lights, ok := data["lights"].([]interface{}); ok {
		for i, lightData := range lights {
			if lightMap, ok := lightData.(map[string]interface{}); ok {
				light := Light{
					Color:     Color{1.0, 1.0, 1.0, 1.0},
					Intensity: 1.0,
				}

				if lightType, ok := lightMap["type"].(string); ok {
					switch lightType {
					case "directional":
						light.Type = DirectionalLight
					case "point":
						light.Type = PointLight
					case "spot":
						light.Type = SpotLight
					}
				}

				if intensity, ok := lightMap["intensity"].(float64); ok {
					light.Intensity = intensity
				}

				if color, ok := lightMap["color"].([]interface{}); ok && len(color) >= 3 {
					if r, ok := color[0].(float64); ok {
						light.Color.R = r
					}
					if g, ok := color[1].(float64); ok {
						light.Color.G = g
					}
					if b, ok := color[2].(float64); ok {
						light.Color.B = b
					}
				}

				if spot, ok := lightMap["spot"].(map[string]interface{}); ok {
					if innerCone, ok := spot["innerConeAngle"].(float64); ok {
						light.InnerCone = innerCone
					}
					if outerCone, ok := spot["outerConeAngle"].(float64); ok {
						light.OuterCone = outerCone
					}
				}

				if point, ok := lightMap["point"].(map[string]interface{}); ok {
					if range_, ok := point["range"].(float64); ok {
						light.Range = range_
					}
				}

				scene.AddLight(light)
				_ = i // Use i if needed for naming
			}
		}
	}
	return nil
}

// KHRMaterialsUnlitExtension handles KHR_materials_unlit extension
type KHRMaterialsUnlitExtension struct{}

func (ext *KHRMaterialsUnlitExtension) GetName() string {
	return "KHR_materials_unlit"
}

func (ext *KHRMaterialsUnlitExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process unlit materials - materials that don't use lighting
	// This could modify material properties to disable lighting calculations
	return nil
}

// KHRMaterialsPBRSpecularGlossinessExtension handles PBR specular-glossiness workflow
type KHRMaterialsPBRSpecularGlossinessExtension struct{}

func (ext *KHRMaterialsPBRSpecularGlossinessExtension) GetName() string {
	return "KHR_materials_pbrSpecularGlossiness"
}

func (ext *KHRMaterialsPBRSpecularGlossinessExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process specular-glossiness workflow materials
	// This would convert specular-glossiness to metallic-roughness workflow
	return nil
}

// KHRTextureTransformExtension handles texture coordinate transformations
type KHRTextureTransformExtension struct{}

func (ext *KHRTextureTransformExtension) GetName() string {
	return "KHR_texture_transform"
}

func (ext *KHRTextureTransformExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process texture coordinate transformations (offset, rotation, scale)
	return nil
}

// KHRMaterialsClearcoatExtension handles clearcoat materials
type KHRMaterialsClearcoatExtension struct{}

func (ext *KHRMaterialsClearcoatExtension) GetName() string {
	return "KHR_materials_clearcoat"
}

func (ext *KHRMaterialsClearcoatExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process clearcoat material properties
	return nil
}

// KHRTextureBasisuExtension handles KTX2/Basis Universal textures
type KHRTextureBasisuExtension struct{}

func (ext *KHRTextureBasisuExtension) GetName() string {
	return "KHR_texture_basisu"
}

func (ext *KHRTextureBasisuExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process Basis Universal (KTX2) texture extension
	// This extension indicates that a texture uses the KTX2 container format
	// with Basis Universal supercompression
	return nil
}

// ExtensionRegistry manages GLTF extensions
type ExtensionRegistry struct {
	handlers map[string]GLTFExtensionHandler
}

// NewExtensionRegistry creates a new extension registry
func NewExtensionRegistry() *ExtensionRegistry {
	reg := &ExtensionRegistry{
		handlers: make(map[string]GLTFExtensionHandler),
	}

	// Register built-in extensions
	reg.RegisterHandler(&KHRLightsPunctualExtension{})
	reg.RegisterHandler(&KHRMaterialsUnlitExtension{})
	reg.RegisterHandler(&KHRMaterialsPBRSpecularGlossinessExtension{})
	reg.RegisterHandler(&KHRTextureTransformExtension{})
	reg.RegisterHandler(&KHRMaterialsClearcoatExtension{})
	reg.RegisterHandler(&KHRTextureBasisuExtension{})

	// Register new material extensions
	reg.RegisterHandler(&KHRMaterialsEmissiveStrengthExtension{})
	reg.RegisterHandler(&KHRMaterialsIORExtension{})
	reg.RegisterHandler(&KHRMaterialsSpecularExtension{})
	reg.RegisterHandler(&KHRMaterialsTransmissionExtension{})
	reg.RegisterHandler(&KHRMaterialsVolumeExtension{})
	reg.RegisterHandler(&KHRMaterialsAnisotropyExtension{})
	reg.RegisterHandler(&KHRMaterialsSheenExtension{})
	reg.RegisterHandler(&KHRMaterialsIridescenceExtension{})
	reg.RegisterHandler(&KHRMaterialsDispersionExtension{})
	reg.RegisterHandler(&KHRMaterialsVariantsExtension{})
	reg.RegisterHandler(&KHRAnimationPointerExtension{})
	reg.RegisterHandler(&KHRMeshQuantizationExtension{})
	reg.RegisterHandler(&KHRXMPJsonLdExtension{})
	reg.RegisterHandler(&EXTMeshGPUInstancingExtension{})
	reg.RegisterHandler(&EXTTextureWebPExtension{})

	return reg
}

// RegisterHandler registers an extension handler
func (reg *ExtensionRegistry) RegisterHandler(handler GLTFExtensionHandler) {
	reg.handlers[handler.GetName()] = handler
}

// ProcessExtensions processes GLTF extensions
func (reg *ExtensionRegistry) ProcessExtensions(extensions map[string]interface{}, scene *Scene) error {
	for extName, extData := range extensions {
		if handler, exists := reg.handlers[extName]; exists {
			if dataMap, ok := extData.(map[string]interface{}); ok {
				err := handler.Process(dataMap, scene)
				if err != nil {
					return fmt.Errorf("failed to process extension %s: %w", extName, err)
				}
			}
		}
	}
	return nil
}

// GetSupportedExtensions returns a list of supported extension names
func (reg *ExtensionRegistry) GetSupportedExtensions() []string {
	var extensions []string
	for name := range reg.handlers {
		extensions = append(extensions, name)
	}
	return extensions
}

// ========== Material Extensions Implementation ==========

// KHRMaterialsEmissiveStrengthExtension handles enhanced emissive strength
type KHRMaterialsEmissiveStrengthExtension struct{}

func (ext *KHRMaterialsEmissiveStrengthExtension) GetName() string {
	return "KHR_materials_emissive_strength"
}

func (ext *KHRMaterialsEmissiveStrengthExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process emissive strength multiplier
	if strength, ok := data["emissiveStrength"].(float64); ok {
		// Apply emissive strength to materials in the scene
		// This would typically be applied to specific materials during GLTF loading
		_ = strength // Use the strength value
	}
	return nil
}

// KHRMaterialsIORExtension handles index of refraction
type KHRMaterialsIORExtension struct{}

func (ext *KHRMaterialsIORExtension) GetName() string {
	return "KHR_materials_ior"
}

func (ext *KHRMaterialsIORExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process index of refraction
	if ior, ok := data["ior"].(float64); ok {
		_ = ior // Apply IOR to materials
	}
	return nil
}

// KHRMaterialsSpecularExtension handles enhanced specular reflection
type KHRMaterialsSpecularExtension struct{}

func (ext *KHRMaterialsSpecularExtension) GetName() string {
	return "KHR_materials_specular"
}

func (ext *KHRMaterialsSpecularExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process specular factor and color
	if specularFactor, ok := data["specularFactor"].(float64); ok {
		_ = specularFactor
	}
	if specularColor, ok := data["specularColorFactor"].([]interface{}); ok && len(specularColor) >= 3 {
		_ = specularColor
	}
	return nil
}

// KHRMaterialsTransmissionExtension handles material transmission
type KHRMaterialsTransmissionExtension struct{}

func (ext *KHRMaterialsTransmissionExtension) GetName() string {
	return "KHR_materials_transmission"
}

func (ext *KHRMaterialsTransmissionExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process transmission factor
	if transmission, ok := data["transmissionFactor"].(float64); ok {
		_ = transmission
	}
	return nil
}

// KHRMaterialsVolumeExtension handles volumetric materials
type KHRMaterialsVolumeExtension struct{}

func (ext *KHRMaterialsVolumeExtension) GetName() string {
	return "KHR_materials_volume"
}

func (ext *KHRMaterialsVolumeExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process volume properties
	if thickness, ok := data["thicknessFactor"].(float64); ok {
		_ = thickness
	}
	if attenuationDistance, ok := data["attenuationDistance"].(float64); ok {
		_ = attenuationDistance
	}
	if attenuationColor, ok := data["attenuationColor"].([]interface{}); ok && len(attenuationColor) >= 3 {
		_ = attenuationColor
	}
	return nil
}

// ========== New Advanced Extensions ==========

// KHRMaterialsAnisotropyExtension handles anisotropic materials
type KHRMaterialsAnisotropyExtension struct{}

func (ext *KHRMaterialsAnisotropyExtension) GetName() string {
	return "KHR_materials_anisotropy"
}

func (ext *KHRMaterialsAnisotropyExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process anisotropic reflection properties
	if anisotropyStrength, ok := data["anisotropyStrength"].(float64); ok {
		_ = anisotropyStrength
	}
	if anisotropyRotation, ok := data["anisotropyRotation"].(float64); ok {
		_ = anisotropyRotation
	}
	return nil
}

// KHRMaterialsSheenExtension handles fabric sheen effect
type KHRMaterialsSheenExtension struct{}

func (ext *KHRMaterialsSheenExtension) GetName() string {
	return "KHR_materials_sheen"
}

func (ext *KHRMaterialsSheenExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process sheen properties for fabric materials
	if sheenColorFactor, ok := data["sheenColorFactor"].([]interface{}); ok && len(sheenColorFactor) >= 3 {
		_ = sheenColorFactor
	}
	if sheenRoughnessFactor, ok := data["sheenRoughnessFactor"].(float64); ok {
		_ = sheenRoughnessFactor
	}
	return nil
}

// KHRMaterialsIridescenceExtension handles iridescent materials
type KHRMaterialsIridescenceExtension struct{}

func (ext *KHRMaterialsIridescenceExtension) GetName() string {
	return "KHR_materials_iridescence"
}

func (ext *KHRMaterialsIridescenceExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process iridescence properties (soap bubble, oil slick effects)
	if iridescenceFactor, ok := data["iridescenceFactor"].(float64); ok {
		_ = iridescenceFactor
	}
	if iridescenceIor, ok := data["iridescenceIor"].(float64); ok {
		_ = iridescenceIor
	}
	if iridescenceThicknessMinimum, ok := data["iridescenceThicknessMinimum"].(float64); ok {
		_ = iridescenceThicknessMinimum
	}
	if iridescenceThicknessMaximum, ok := data["iridescenceThicknessMaximum"].(float64); ok {
		_ = iridescenceThicknessMaximum
	}
	return nil
}

// KHRMaterialsDispersionExtension handles chromatic dispersion
type KHRMaterialsDispersionExtension struct{}

func (ext *KHRMaterialsDispersionExtension) GetName() string {
	return "KHR_materials_dispersion"
}

func (ext *KHRMaterialsDispersionExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process chromatic dispersion (rainbow effect in transparent materials)
	if dispersion, ok := data["dispersion"].(float64); ok {
		_ = dispersion
	}
	return nil
}

// KHRMaterialsVariantsExtension handles material variants
type KHRMaterialsVariantsExtension struct{}

func (ext *KHRMaterialsVariantsExtension) GetName() string {
	return "KHR_materials_variants"
}

func (ext *KHRMaterialsVariantsExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process material variants for switching between different material looks
	if variants, ok := data["variants"].([]interface{}); ok {
		for _, variant := range variants {
			if variantMap, ok := variant.(map[string]interface{}); ok {
				if name, ok := variantMap["name"].(string); ok {
					_ = name
				}
			}
		}
	}
	return nil
}

// ========== Animation and Mesh Extensions ==========

// KHRAnimationPointerExtension handles animation pointer targeting
type KHRAnimationPointerExtension struct{}

func (ext *KHRAnimationPointerExtension) GetName() string {
	return "KHR_animation_pointer"
}

func (ext *KHRAnimationPointerExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process animation pointers for targeting arbitrary properties
	if pointer, ok := data["pointer"].(string); ok {
		_ = pointer
	}
	return nil
}

// KHRMeshQuantizationExtension handles mesh quantization
type KHRMeshQuantizationExtension struct{}

func (ext *KHRMeshQuantizationExtension) GetName() string {
	return "KHR_mesh_quantization"
}

func (ext *KHRMeshQuantizationExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process quantized mesh attributes for reduced memory usage
	// This extension allows for more compact vertex data representation
	return nil
}

// ========== Metadata and Instancing Extensions ==========

// KHRXMPJsonLdExtension handles XMP metadata
type KHRXMPJsonLdExtension struct{}

func (ext *KHRXMPJsonLdExtension) GetName() string {
	return "KHR_xmp_json_ld"
}

func (ext *KHRXMPJsonLdExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process XMP metadata in JSON-LD format
	if packet, ok := data["packet"].(string); ok {
		_ = packet // Store XMP metadata
	}
	return nil
}

// EXTMeshGPUInstancingExtension handles GPU instancing
type EXTMeshGPUInstancingExtension struct{}

func (ext *EXTMeshGPUInstancingExtension) GetName() string {
	return "EXT_mesh_gpu_instancing"
}

func (ext *EXTMeshGPUInstancingExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process GPU instancing attributes (can be simulated with CPU)
	if attributes, ok := data["attributes"].(map[string]interface{}); ok {
		_ = attributes // Process instance transforms, colors, etc.
	}
	return nil
}

// ========== Texture Extensions ==========

// EXTTextureWebPExtension handles WebP textures
type EXTTextureWebPExtension struct{}

func (ext *EXTTextureWebPExtension) GetName() string {
	return "EXT_texture_webp"
}

func (ext *EXTTextureWebPExtension) Process(data map[string]interface{}, scene *Scene) error {
	// Process WebP texture format support
	if source, ok := data["source"].(float64); ok {
		_ = source // Reference to WebP image
	}
	return nil
}
