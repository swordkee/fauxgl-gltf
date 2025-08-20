package fauxgl

import (
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

func LoadGLTF(path string) (*Mesh, error) {
	// 使用qmuntal/gltf库加载GLTF文件
	doc, err := gltf.Open(path)
	if err != nil {
		return nil, err
	}

	var triangles []*Triangle

	// 遍历所有网格
	for _, mesh := range doc.Meshes {
		// 遍历网格中的所有图元
		for _, primitive := range mesh.Primitives {
			// 只处理三角形图元

			// 获取顶点位置数据
			positionAccessor := doc.Accessors[primitive.Attributes[gltf.POSITION]]
			posBuffer := [][3]float32{}
			positionBuffer, err := modeler.ReadPosition(doc, positionAccessor, posBuffer)
			if err != nil {
				return nil, err
			}

			// 获取法线数据（如果存在）
			var normalBuffer [][3]float32
			if normalAccessorIndex, ok := primitive.Attributes[gltf.NORMAL]; ok {
				normalBuffer1 := [][3]float32{}
				normalAccessor := doc.Accessors[normalAccessorIndex]
				normalBuffer, err = modeler.ReadNormal(doc, normalAccessor, normalBuffer1)
				if err != nil {
					return nil, err
				}
			}

			// 获取纹理坐标数据（如果存在）
			var texCoordBuffer [][2]float32
			if texCoordAccessorIndex, ok := primitive.Attributes[gltf.TEXCOORD_0]; ok {
				uvBuffer := [][2]float32{}
				texCoordAccessor := doc.Accessors[texCoordAccessorIndex]
				texCoordBuffer, err = modeler.ReadTextureCoord(doc, texCoordAccessor, uvBuffer)
				if err != nil {
					return nil, err
				}
			}

			// 获取索引数据
			var indices []uint32
			if primitive.Indices != nil {
				indexAccessor := doc.Accessors[*primitive.Indices]
				indexBuffer := []uint32{}
				indices, err = modeler.ReadIndices(doc, indexAccessor, indexBuffer)
				if err != nil {
					return nil, err
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
	}

	return NewTriangleMesh(triangles), nil
}
