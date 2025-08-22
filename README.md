# FauxGL-GLTF - 专业GLTF渲染引擎

FauxGL-GLTF 是一个专门针对GLTF格式优化的纯Go语言3D渲染引擎，支持完整的物理基础渲染(PBR)、高级材质系统、场景管理和动画播放。

## 特色功能

### 🎯 专业GLTF支持

- **完整GLTF 2.0解析**: 支持所有标准GLTF特性
- **场景层次加载**: 完整的节点树结构
- **材质系统**: 自动转换GLTF材质为PBR材质
- **动画播放**: 支持关键帧动画和变换动画
- **纹理加载**: 支持所有GLTF纹理类型
- **相机和光源**: 从GLTF文件加载场景设置

### 🎨 物理基础渲染 (PBR)

- **完整的PBR材质系统**: 支持金属工作流
- **基于物理的光照**: Cook-Torrance BRDF模型
- **材质属性**: 
  - 基础颜色 (Base Color)
  - 金属度 (Metallic) 
  - 粗糙度 (Roughness)
  - 法线贴图 (Normal Maps)
  - 遮蔽贴图 (Occlusion Maps)
  - 自发光 (Emissive)

### 🖼️ 高级纹理系统

- **多种纹理类型**: 基础颜色、法线、金属度、粗糙度等
- **纹理过滤**: 最近邻、双线性插值
- **环绕模式**: 重复、夹取、镜像
- **Mipmap支持**: 自动生成多级细节

### 🎬 场景管理

- **层次化场景图**: 支持复杂的节点关系
- **变换系统**: 本地和世界坐标系统
- **多相机支持**: 透视和正交投影
- **动画播放器**: 关键帧动画和插值

### 💡 环境光系统

- **AmbientLight支持**: 提供均匀的全局照明
- **PBR兼容**: 与物理基础渲染完全兼容
- **多光源组合**: 支持方向光、点光源、聚光灯和环境光
- **高效渲染**: 直接应用到材质基色，考虑遮蔽贴图
- **智能兼容性**: 优先使用AmbientLight光源，保持向后兼容

#### 环境光使用示例

```go
// 基本环境光使用
scene.AddAmbientLight(fauxgl.Color{0.2, 0.3, 0.4, 1.0}, 0.8)

// 多光源组合
scene.AddAmbientLight(fauxgl.Color{0.3, 0.3, 0.4, 1.0}, 0.5)  // 环境光
scene.AddDirectionalLight(                                      // 主光源
    fauxgl.V(-1, -1, -1),
    fauxgl.Color{1.0, 0.9, 0.8, 1.0},
    2.0,
)
scene.AddPointLight(                                            // 点光源
    fauxgl.V(2, 3, 1),
    fauxgl.Color{1.0, 0.8, 0.6, 1.0},
    5.0, 10.0,
)
```

#### 场景照明最佳实践

**室内场景**:
```go
// 室内散射光
scene.AddAmbientLight(fauxgl.Color{0.25, 0.27, 0.3, 1.0}, 0.4)
// 窗户光
scene.AddDirectionalLight(fauxgl.V(-0.5, -0.8, -0.3), fauxgl.Color{0.95, 0.9, 0.8, 1.0}, 3.0)
// 室内灯光
scene.AddPointLight(fauxgl.V(0, 2.5, 0), fauxgl.Color{1.0, 0.95, 0.8, 1.0}, 8.0, 15.0)
```

**户外场景**:
```go
// 天空散射光
scene.AddAmbientLight(fauxgl.Color{0.4, 0.45, 0.6, 1.0}, 0.6)
// 太阳光
scene.AddDirectionalLight(fauxgl.V(-0.3, -0.8, -0.5), fauxgl.Color{1.0, 0.95, 0.85, 1.0}, 4.0)
```

**夜晚场景**:
```go
// 月光环境光
scene.AddAmbientLight(fauxgl.Color{0.1, 0.15, 0.25, 1.0}, 0.2)
// 月光主光源
scene.AddDirectionalLight(fauxgl.V(-0.2, -0.9, -0.4), fauxgl.Color{0.7, 0.8, 1.0, 1.0}, 1.5)
```

## 快速开始

### 安装

```bash
go get github.com/swordkee/fauxgl-gltf
```

### 基本使用

```go
package main

import (
    "log"
    "github.com/swordkee/fauxgl-gltf"
)

func main() {
    // 加载GLTF场景
    scene, err := fauxgl.LoadGLTFScene("model.gltf")
    if err != nil {
        log.Fatal(err)
    }

    // 创建渲染上下文
    context := fauxgl.NewContext(1920, 1080)
    context.ClearColor = fauxgl.Color{0.2, 0.2, 0.3, 1.0}
    context.ClearColorBuffer()

    // 渲染场景
    renderer := fauxgl.NewSceneRenderer(context)
    renderer.RenderScene(scene)

    // 保存结果
    err = fauxgl.SavePNG("output.png", context.Image())
    if err != nil {
        log.Fatal(err)
    }
}
```

### 自定义PBR材质

```go
// 创建自定义PBR材质
material := fauxgl.NewPBRMaterial()
material.BaseColorFactor = fauxgl.Color{0.8, 0.2, 0.2, 1.0} // 红色
material.MetallicFactor = 0.8   // 高金属度
material.RoughnessFactor = 0.2  // 低粗糙度（光滑）

// 加载纹理
baseColorTexture, _ := fauxgl.LoadAdvancedTexture("base_color.jpg", fauxgl.BaseColorTexture)
normalTexture, _ := fauxgl.LoadAdvancedTexture("normal.jpg", fauxgl.NormalTexture)

material.BaseColorTexture = baseColorTexture
material.NormalTexture = normalTexture

// 应用到场景节点
node.Material = material
```

### UV贴图和自定义渲染

```go
// 定义UV区域
type UVRegion struct {
    Name        string
    MinU, MaxU  float64
    MinV, MaxV  float64
    Color       fauxgl.Color
    TexturePath string
    Enabled     bool
}

// 创建UV映射区域
regions := []UVRegion{
    {
        Name:    "杯身主体",
        MinU:    0.0, MaxU: 0.6,
        MinV:    0.2, MaxV: 0.8,
        Color:   fauxgl.Color{0.8, 0.3, 0.3, 1.0},
        Enabled: true,
    },
    {
        Name:    "杯口装饰",
        MinU:    0.0, MaxU: 1.0,
        MinV:    0.8, MaxV: 1.0,
        Color:   fauxgl.Color{1.0, 0.8, 0.2, 1.0},
        Enabled: true,
    },
}

// 应用UV映射
applyUVMapping(mesh, regions)
```

## 运行示例

项目包含了多个完整的示例程序：

### GLTF基础渲染
```bash
cd examples
go run gltf_demo.go
```

### Mug模型渲染
```bash
cd examples
go run mug_gltf_enhanced.go
```

### PBR材质演示
```bash
cd examples
go run pbr_demo.go
```

### 精确GLTF渲染
```bash
cd examples
go run mug_gltf_precise.go
```

### 环境光功能演示
```bash
cd examples
go run ambient_light_demo.go
```

### 高级功能演示 (蒙皮动画 + 变形目标 + GLTF扩展)
```bash
cd examples
go run advanced_features_simple.go
```

### KTX2纹理格式演示 🆕
```bash
cd examples
go run ktx2_texture_demo.go
```

### 扩展材质功能演示 🆕
```bash
cd examples
go run extended_materials_demo.go
```

### GLTF扩展综合展示 🆕
```bash
cd examples
go run gltf_extensions_showcase.go
```

## 支持的GLTF特性

✅ **完全支持**:
- GLTF 2.0格式
- PBR材质 (Metallic-Roughness workflow)
- 纹理映射 (Base Color, Normal, Metallic-Roughness)
- 场景层次结构
- 网格几何体
- 关键帧动画
- **蒙皮动画 (Skinned Animation)** 🆕
- **变形目标 (Morph Targets)** 🆕
- 相机定义
- 光源设置

🚧 **部分支持**:
- **GLTF扩展系统** 🆕 (21个扩展):
  - **材质扩展** (13个):
    - KHR_materials_emissive_strength (增强自发光强度)
    - KHR_materials_ior (折射率)
    - KHR_materials_specular (镜面反射)
    - KHR_materials_transmission (透射)
    - KHR_materials_volume (体积材质)
    - KHR_materials_anisotropy (各向异性) 🆕
    - KHR_materials_sheen (光泽效果) 🆕
    - KHR_materials_iridescence (彩虹色) 🆕
    - KHR_materials_dispersion (色散) 🆕
    - KHR_materials_clearcoat (清漆) 🆕
    - KHR_materials_unlit (无光照材质)
    - KHR_materials_variants (材质变体) 🆕
    - KHR_materials_pbrSpecularGlossiness (镜面光泽工作流)
  - **纹理扩展** (3个):
    - KHR_texture_basisu (KTX2/Basis Universal纹理) 🆕
    - KHR_texture_transform (纹理坐标变换)
    - EXT_texture_webp (WebP纹理) 🆕
  - **光照扩展** (1个):
    - KHR_lights_punctual (标准光源)
  - **动画扩展** (1个):
    - KHR_animation_pointer (动画指针) 🆕
  - **网格扩展** (2个):
    - KHR_mesh_quantization (网格量化) 🆕
    - EXT_mesh_gpu_instancing (GPU实例化) 🆕
  - **元数据扩展** (1个):
    - KHR_xmp_json_ld (XMP元数据) 🆕
- **高级动画功能**:
  - 骨骼系统和关节矩阵
  - 形状插值和面部动画
  - 四元数旋转插值
- **KTX2纹理格式** 🆕:
  - KTX2容器格式解析
  - 多级mipmap支持
  - 数据格式描述符(DFD)解析
  - 键值对元数据提取
  - 超级压缩检测

⚠️ **计划支持** (高难度):
- KTX2纹理解压缩 (Basis Universal, Zstd等)
- Draco几何压缩 (需要CGO集成)
- 某些高级扩展 (依赖外部库)

## 性能特点

- **CPU渲染**: 纯软件实现，无需GPU
- **高质量输出**: 支持超采样抗锯齿
- **内存效率**: 优化的内存使用
- **并行处理**: 利用多核CPU加速

**适用场景**:
- 离线渲染和批处理
- 高质量静态图像生成
- 无GPU环境的渲染
- GLTF模型预览和转换
- 教学和原型开发

## API参考

### 核心类型

```go
// 场景加载
type Scene struct {
    RootNode  *SceneNode
    Cameras   map[string]*Camera
    Lights    []Light
    Materials map[string]*PBRMaterial
    Meshes    map[string]*Mesh
}

// PBR材质
type PBRMaterial struct {
    BaseColorFactor   Color
    MetallicFactor    float64
    RoughnessFactor   float64
    BaseColorTexture  *AdvancedTexture
    NormalTexture     *AdvancedTexture
    // ... 更多属性
}

// 场景节点
type SceneNode struct {
    Name        string
    Transform   Matrix
    Children    []*SceneNode
    Mesh        *Mesh
    Material    *PBRMaterial
}
```

### 主要函数

```go
// GLTF加载
func LoadGLTFScene(path string) (*Scene, error)
func LoadGLTF(path string) (*Mesh, error)

// 场景渲染
func NewSceneRenderer(context *Context) *SceneRenderer
func (r *SceneRenderer) RenderScene(scene *Scene)

// 材质和纹理
func NewPBRMaterial() *PBRMaterial
func LoadAdvancedTexture(path string, textureType TextureType) (*AdvancedTexture, error)

// 动画
func NewAnimationPlayer() *AnimationPlayer
func (p *AnimationPlayer) Play(name string)

// 光源管理
func (scene *Scene) AddDirectionalLight(direction Vector, color Color, intensity float64)
func (scene *Scene) AddPointLight(position Vector, color Color, intensity, range_ float64)
func (scene *Scene) AddSpotLight(position, direction Vector, color Color, intensity, range_, innerCone, outerCone float64)
func (scene *Scene) AddAmbientLight(color Color, intensity float64)
func (scene *Scene) ClearLights()
func (scene *Scene) GetLightsByType(lightType LightType) []Light
```

## 版本历史

### v1.1.0 (高级特性版)
- 🦾 **蒙皮动画支持**: 骨骼系统和关节矩阵
- 🎦 **变形目标支持**: 形状插值和面部动画
- 🔌 **GLTF扩展系统**: 支持多种KHR扩展
- 🎯 **增强场景管理**: 蒙皮和变形目标集成
- 🔧 **动画系统扩展**: 四元数插值和高级动画

### v1.0.0 (GLTF专版)
- 🎯 专注GLTF格式支持
- 🗑️ 移除非GLTF格式解析器 (STL, OBJ, PLY等)
- 🆕 完整PBR材质系统
- 🆕 高级纹理支持
- 🆕 场景管理和动画
- 💡 **环境光功能 (AmbientLight)**: 支持均匀全局照明
- 🔧 优化GLTF加载性能
- 📚 完整中文文档

## 依赖项

```go
require (
    github.com/qmuntal/gltf v0.28.0
    github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
    github.com/fogleman/simplify v0.0.0-20170216171241-d32f302d5046
)
```

## 许可证

本项目采用MIT许可证。详见 [LICENSE.md](LICENSE.md) 文件。

## 贡献指南

欢迎提交问题报告和功能请求！

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

## 支持

- 📧 邮箱: support@example.com
- 🐛 问题报告: [GitHub Issues](https://github.com/swordkee/fauxgl-gltf/issues)
- 📖 文档: [Wiki](https://github.com/swordkee/fauxgl-gltf/wiki)

---

**FauxGL-GLTF** - 让GLTF渲染变得简单高效！