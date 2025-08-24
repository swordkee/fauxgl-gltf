# FauxGL-GLTF - 专业GLTF渲染引擎

FauxGL-GLTF 是一个专门针对GLTF格式优化的纯Go语言3D渲染引擎，支持完整的物理基础渲染(PBR)、高级材质系统、场景管理和动画播放，代码大部分由[Goder](https://qoder.com)编写，基于[FauxGL](https://github.com/fogleman/fauxgl)开发。

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

### 🎨 动态UV映射系统 🆕

- **智能UV修改器**: 支持实时UV坐标变换
- **多层变换效果**: 缩放、旋转、偏移、剪切变换
- **区域化映射**: 矩形、圆形、渐变遮罩系统
- **多种混合模式**: 替换、加法、乘法、叠加模式
- **部分区域贴图**: 支持杯子等模型的局部纹理应用
- **动画UV效果**: 时间管理和动态变换
- **智能纹理选择**: 自动回退机制和内容验证

#### UV修改器使用示例

**基础UV变换**:
```
// 创建UV修改器
modifier := fauxgl.NewUVModifier()
// 设置全局变换
globalTransform := fauxgl.NewUVTransform()
globalTransform.ScaleU = 2.0
globalTransform.ScaleV = 1.5
globalTransform.Rotation = math.Pi / 4 // 45度旋转
modifier.SetGlobalTransform(globalTransform)
// 应用到纹理
texture.UVModifier = modifier
```

**部分区域贴图**:
```
// 创建前面板标志区域
frontLogoMapping := &fauxgl.UVMapping{
    Name:    "front_logo_area",
    Enabled: true,
    Region: fauxgl.UVRegion{
        MinU: 0.25, MaxU: 0.75,
        MinV: 0.35, MaxV: 0.65,
        MaskType: fauxgl.UVMaskRectangle,
    },
    Transform: &fauxgl.UVTransform{
        ScaleU: 0.8, ScaleV: 0.6,
    },
    BlendMode: fauxgl.UVBlendReplace,
    Priority:  2,
}
modifier.AddMapping(frontLogoMapping)
```

**智能纹理加载**:
```
// 程序会按优先级自动加载纹理文件
// 1. your_texture.jpg (用户自定义)
// 2. custom_texture.jpg 
// 3. logo_texture.png
// 4. texture.png (回退选项)

// 配置自定义纹理
const CUSTOM_TEXTURE_FILE = "my_logo.jpg"
```

### 🧩 UV编辑器系统 🆕

- **平面展开算法**: 基于Blender的UV展开逻辑实现
- **自动松弛算法**: 实现Seam Relaxation算法优化UV分布
- **UV岛屿处理**: 支持多UV岛屿提取和独立处理
- **坐标映射**: 2D画布坐标与3D模型UV坐标双向转换
- **UV可视化**: 生成UV展开图用于调试和分析
- **保持约束**: 支持面积保持和角度保持约束

#### UV编辑器使用示例

**UV松弛处理**:
```
// 创建UV松弛设置
settings := fauxgl.NewUVRelaxationSettings()
settings.Iterations = 20     // 松弛迭代次数
settings.StepSize = 0.3      // 松弛步长
settings.PinBoundary = true  // 固定边界顶点

// 应用UV松弛算法到网格
err := fauxgl.ApplyUVRelaxation(mesh, settings)
if err != nil {
    log.Fatal(err)
}
```

**坐标映射**:
```
// UV坐标转画布坐标
uv := fauxgl.Vector2{0.5, 0.5}
x, y := fauxgl.UVToCanvas(uv, 1024, 1024)
fmt.Printf("UV(0.5, 0.5) -> 画布坐标(%d, %d)\n", x, y)

// 画布坐标转UV坐标
uv2 := fauxgl.CanvasToUV(512, 512, 1024, 1024)
fmt.Printf("画布坐标(512, 512) -> UV(%.2f, %.2f)\n", uv2.X, uv2.Y)
```

### 🖼️ 高级纹理系统

- **多种纹理类型**: 基础颜色、法线、金属度、粗糙度等
- **纹理过滤**: 最近邻、双线性插值
- **环绕模式**: 重复、夹取、镜像
- **Mipmap支持**: 自动生成多级细节
- **集成UV修改器**: 无缝集成到纹理采样流程 🆕

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

```
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
```
// 室内散射光
scene.AddAmbientLight(fauxgl.Color{0.25, 0.27, 0.3, 1.0}, 0.4)
// 窗户光
scene.AddDirectionalLight(fauxgl.V(-0.5, -0.8, -0.3), fauxgl.Color{0.95, 0.9, 0.8, 1.0}, 3.0)
```

**户外场景**:
```
// 天空散射光
scene.AddAmbientLight(fauxgl.Color{0.4, 0.45, 0.6, 1.0}, 0.6)
// 太阳光
scene.AddDirectionalLight(fauxgl.V(-0.3, -0.8, -0.5), fauxgl.Color{1.0, 0.95, 0.85, 1.0}, 4.0)
```

**夜晚场景**:
```
// 月光环境光
scene.AddAmbientLight(fauxgl.Color{0.1, 0.15, 0.25, 1.0}, 0.2)
// 月光主光源
scene.AddDirectionalLight(fauxgl.V(-0.2, -0.9, -0.4), fauxgl.Color{0.7, 0.8, 1.0, 1.0}, 1.5)
```





## 快速开始

### 安装

```
go get github.com/swordkee/fauxgl-gltf
```

### 基本使用

```
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

### 自定义PBR材质和UV映射

```
// 创建自定义PBR材质
material := fauxgl.NewPBRMaterial()
material.BaseColorFactor = fauxgl.Color{0.8, 0.2, 0.2, 1.0} // 红色
material.MetallicFactor = 0.8   // 高金属度
material.RoughnessFactor = 0.2  // 低粗糙度

// 加载纹理
baseColorTexture, _ := fauxgl.LoadAdvancedTexture("base_color.jpg", fauxgl.BaseColorTexture)

// 添加UV修改器
modifier := fauxgl.NewUVModifier()
globalTransform := fauxgl.NewUVTransform()
globalTransform.ScaleU = 1.5
globalTransform.ScaleV = 1.2
modifier.SetGlobalTransform(globalTransform)
baseColorTexture.UVModifier = modifier

// 应用到场景节点
node.Material = material
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
go run mug.go
```

### PBR材质演示
```bash
cd examples
go run pbr_demo.go
```

### 精确GLTF渲染
```bash
cd examples
go run mug_uv.go
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

### UV调试工具 🆕
```bash
cd examples
go run debug_uv.go
```

### 部分区域贴图演示 🆕
```bash
cd examples
go run mug_uv_final.go
```

### UV编辑器演示 🆕
```bash
cd examples
go run uv_editor.go
```

### 完整功能演示（UV编辑器+多光源） 🆕
```bash
cd examples
go run complete_demo.go
```

## 支持的GLTF特性

✅ **完全支持**:
- GLTF 2.0格式
- PBR材质 (Metallic-Roughness workflow)
- 纹理映射 (Base Color, Normal, Metallic-Roughness)
- **动态UV修改器**: 实时UV坐标变换和部分贴图 🆕
- 场景层次结构
- 网格几何体
- 关键帧动画
- **蒙皮动画 (Skinned Animation)** 🆕
- **变形目标 (Morph Targets)** 🆕
- 相机定义
- 光源设置
- **环境光功能 (AmbientLight)**: 支持均匀全局照明 🆕

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

## 🚀 性能优化

### SIMD向量计算优化 🆕
FauxGL-GLTF现在支持SIMD（单指令多数据）优化的向量计算，显著提升渲染性能：

- **向量运算加速**: 向量加法、减法、点积、叉积等运算性能提升2-4倍
- **矩阵变换优化**: 批量矩阵变换操作利用SIMD指令优化
- **几何处理加速**: 网格变换、法线计算等几何处理性能显著提升
- **自动切换机制**: 根据网格大小自动选择传统算法或SIMD优化算法

```
// SIMD优化的向量运算示例
v1 := fauxgl.Vector{1, 2, 3}
v2 := fauxgl.Vector{4, 5, 6}

// 传统向量加法
result1 := v1.Add(v2)

// SIMD优化的批量向量运算
vectors1 := []fauxgl.Vector{v1, v1, v1}
vectors2 := []fauxgl.Vector{v2, v2, v2}
result2 := fauxgl.SIMDAddVectors(vectors1, vectors2)
```

### 高质量渲染优化
- **4K超分辨率渲染**: 支持高达8K的渲染输出
- **智能超采样**: 自适应超采样抗锯齿技术
- **并行渲染**: 充分利用多核CPU进行并行渲染
- **内存优化**: 优化的内存管理和垃圾回收

## 🎯 使用示例

### 基础渲染
```
# 基础GLTF渲染
go run examples/gltf_demo.go

# 高质量4K渲染（带SIMD优化）
go run examples/mug_uv_improved.go

# SIMD性能测试
go run examples/simd_demo.go
```

### 高级功能演示
```
# UV修改器完整演示
go run examples/mug_uv_final.go

# 多光源PBR渲染
go run examples/pbr_demo.go

# 环境光效果演示
go run examples/ambient_light_demo.go

# KTX2纹理格式支持
go run examples/ktx2_texture_demo.go
```

## 📚 API参考

### 核心类型

```
// SIMD优化向量
type SIMDVector4 [4]float64

// SIMD优化矩阵
type SIMDMat4 [16]float64

// SIMD优化顶点
type SIMDVertex struct {
    Position SIMDVector4
    Normal   SIMDVector4
    Color    SIMDVector4
    TexCoord SIMDVector4
}

// 批量SIMD操作
func SIMDAddVectors(a, b []Vector) []Vector
func SIMDMulScalarVectors(vectors []Vector, scalar float64) []Vector
func SIMDNormalizeVectors(vectors []Vector) []Vector

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

// UV修改器 🆕
type UVModifier struct {
    GlobalTransform *UVTransform
    Mappings        []*UVMapping
}

type UVTransform struct {
    OffsetU, OffsetV float64 // UV偏移
    ScaleU, ScaleV   float64 // UV缩放
    Rotation         float64 // 旋转角度
    SkewU, SkewV     float64 // UV剪切
    PivotU, PivotV   float64 // 旋转中心点
}

type UVMapping struct {
    Name      string
    Enabled   bool
    Region    UVRegion
    Transform *UVTransform
    BlendMode UVBlendMode
    Priority  int
}

// UV编辑器系统 🆕
type UVRelaxationSettings struct {
    Iterations     int     // 松弛迭代次数
    StepSize       float64 // 松弛步长
    PinBoundary    bool    // 是否固定边界
    PreserveArea   bool    // 是否保持面积
    PreserveAngles bool    // 是否保持角度
    EnableSeams    bool    // 是否启用接缝
}

type UVIsland struct {
    Vertices      []Vector2 // UV坐标
    Indices       []int     // 三角形索引
    OriginalUVs   []Vector2 // 原始UV坐标
    Seams         []UVSeam  // 接缝列表
    BoundaryVerts []int     // 边界顶点
    PinnedVerts   []int     // 固定顶点
}

type UVSeam struct {
    Edge     [2]Vector // 接缝边的两个顶点
    Strength float64   // 接缝强度 (0-1)
    Fixed    bool      // 是否固定（不参与松弛）
}

type Vector2 struct {
    X, Y float64
}
```

### 主要函数

```
// GLTF加载
func LoadGLTFScene(path string) (*Scene, error)

// 场景渲染
func NewSceneRenderer(context *Context) *SceneRenderer
func (r *SceneRenderer) RenderScene(scene *Scene)

// 材质和纹理
func NewPBRMaterial() *PBRMaterial
func LoadAdvancedTexture(path string, textureType TextureType) (*AdvancedTexture, error)

// UV修改器 🆕
func NewUVModifier() *UVModifier
func NewUVTransform() *UVTransform
func (modifier *UVModifier) AddMapping(mapping *UVMapping)
func (modifier *UVModifier) SetGlobalTransform(transform *UVTransform)
func (modifier *UVModifier) TransformUV(u, v float64) (float64, float64)
func ApplyUVModifierToMesh(mesh *Mesh, modifier *UVModifier)

// UV编辑器系统 🆕
func NewUVRelaxationSettings() *UVRelaxationSettings
func ApplyUVRelaxation(mesh *Mesh, settings *UVRelaxationSettings) error
func ExtractUVIslands(mesh *Mesh) []*UVIsland
func RelaxUVs(island *UVIsland, settings *UVRelaxationSettings)
func UVToCanvas(uv Vector2, canvasWidth, canvasHeight int) (int, int)
func CanvasToUV(x, y, canvasWidth, canvasHeight int) Vector2

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

### v1.3.0 (UV编辑器和多光源系统版) 🆕
- 🧩 **UV编辑器系统**: 基于Blender的UV展开逻辑实现
- 🔄 **自动松弛算法**: 实现Seam Relaxation算法优化UV分布
- 🗺️ **UV坐标映射**: 2D画布坐标与3D模型UV坐标双向转换
- 🎨 **UV可视化**: 生成UV展开图用于调试和分析
- 💡 **多光源系统**: 支持方向光、环境光等多种光源组合
- 🌈 **PBR渲染**: 基于物理的渲染支持多光源效果
- 📊 **保持约束**: 支持面积保持和角度保持约束

### v1.2.0 (UV映射系统版) 🆕
- 🎨 **动态UV修改器**: 实时UV坐标变换系统
- 🎯 **部分区域贴图**: 支持模型局部纹理应用
- 🔄 **多层变换效果**: 缩放、旋转、偏移、剪切
- 🔳 **区域化映射**: 矩形、圆形、渐变遮罩系统
- 🌈 **多种混合模式**: 替换、加法、乘法、叠加
- 🤖 **智能纹理选择**: 自动回退和内容验证
- 🎥 **动画UV效果**: 时间管理和动态变换
- 🛠️ **高质量渲染**: 支持300KB+大文件输出
- 📝 **完整UV文档**: 附带详细使用指南

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

```
require (
    github.com/qmuntal/gltf v0.28.0
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

- 📧 邮箱: swordkee.zhu@gmail.com
- 🐛 问题报告: [GitHub Issues](https://github.com/swordkee/fauxgl-gltf/issues)
- 📖 文档: [Wiki](https://github.com/swordkee/fauxgl-gltf/wiki)

---

**FauxGL-GLTF** - 让GLTF渲染变得简单高效！