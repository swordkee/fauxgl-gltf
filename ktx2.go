package fauxgl

import (
	"encoding/binary"
)

// KTX2格式解码器
// 基于KTX 2.0规范实现的纹理容器格式解析器
// 支持异步读取、解析、验证、数据格式描述和键值对数据

// KTX2魔数标识符，出现在文件开头
var KTX2_MAGIC = [12]byte{0xAB, 0x4B, 0x54, 0x58, 0x20, 0x32, 0x30, 0xBB, 0x0D, 0x0A, 0x1A, 0x0A}

// ParseError KTX2解析错误类型
type ParseError int

const (
	UnexpectedEnd ParseError = iota
	BadMagic
	ZeroWidth
	ZeroFaceCount
	InvalidSampleBitLength
)

func (e ParseError) Error() string {
	switch e {
	case UnexpectedEnd:
		return "unexpected end of data"
	case BadMagic:
		return "invalid KTX2 magic number"
	case ZeroWidth:
		return "zero pixel width"
	case ZeroFaceCount:
		return "zero face count"
	case InvalidSampleBitLength:
		return "invalid sample bit length"
	default:
		return "unknown parse error"
	}
}

// Format KTX2格式枚举
type Format uint32

// 常见的KTX2格式常量
const (
	FormatUndefined Format = 0
	// 可以根据需要添加更多格式
)

func NewFormat(value uint32) *Format {
	if value == 0 {
		return nil
	}
	f := Format(value)
	return &f
}

func (f Format) Value() uint32 {
	return uint32(f)
}

// SupercompressionScheme 超级压缩方案
type SupercompressionScheme uint32

const (
	SupercompressionNone    SupercompressionScheme = 0
	SupercompressionBasisLZ SupercompressionScheme = 1
	SupercompressionZstd    SupercompressionScheme = 2
	SupercompressionZLIB    SupercompressionScheme = 3
)

func NewSupercompressionScheme(value uint32) *SupercompressionScheme {
	s := SupercompressionScheme(value)
	switch s {
	case SupercompressionNone, SupercompressionBasisLZ, SupercompressionZstd, SupercompressionZLIB:
		return &s
	default:
		return nil
	}
}

func (s SupercompressionScheme) Value() uint32 {
	return uint32(s)
}

// ColorModel 颜色模型
type ColorModel uint8

const (
	ColorModelUnspecified ColorModel = 0
	ColorModelRGBSDA      ColorModel = 1
	ColorModelYUVSDA      ColorModel = 2
	ColorModelYIQSDA      ColorModel = 3
	ColorModelLabSDA      ColorModel = 4
	ColorModelCMYKA       ColorModel = 5
	ColorModelXYZSDA      ColorModel = 6
	ColorModelHSVSDA      ColorModel = 7
	ColorModelHSLSDA      ColorModel = 8
	ColorModelBC7M6       ColorModel = 9
)

func NewColorModel(value uint8) *ColorModel {
	if value == 0 {
		return nil
	}
	c := ColorModel(value)
	return &c
}

func (c ColorModel) Value() uint8 {
	return uint8(c)
}

// ColorPrimaries 颜色原色
type ColorPrimaries uint8

const (
	ColorPrimariesUnspecified ColorPrimaries = 0
	ColorPrimariesBT709       ColorPrimaries = 1
	ColorPrimariesBT2020      ColorPrimaries = 2
	ColorPrimariesACES        ColorPrimaries = 3
)

func NewColorPrimaries(value uint8) *ColorPrimaries {
	if value == 0 {
		return nil
	}
	c := ColorPrimaries(value)
	return &c
}

func (c ColorPrimaries) Value() uint8 {
	return uint8(c)
}

// TransferFunction 传输函数
type TransferFunction uint8

const (
	TransferFunctionUnspecified TransferFunction = 0
	TransferFunctionLinear      TransferFunction = 1
	TransferFunctionSRGB        TransferFunction = 2
	TransferFunctionITU         TransferFunction = 3
)

func NewTransferFunction(value uint8) *TransferFunction {
	if value == 0 {
		return nil
	}
	t := TransferFunction(value)
	return &t
}

func (t TransferFunction) Value() uint8 {
	return uint8(t)
}

// ChannelTypeQualifiers 通道类型限定符
type ChannelTypeQualifiers uint8

const (
	QualifierLinear   ChannelTypeQualifiers = 1 << 0
	QualifierExponent ChannelTypeQualifiers = 1 << 1
	QualifierSigned   ChannelTypeQualifiers = 1 << 2
	QualifierFloat    ChannelTypeQualifiers = 1 << 3
)

// DataFormatFlags 数据格式标志
type DataFormatFlags uint8

const (
	StraightAlpha      DataFormatFlags = 0
	AlphaPremultiplied DataFormatFlags = 1 << 0
)

// Index KTX2文件各部分的字节偏移量和大小索引
type Index struct {
	DFDByteOffset uint32 // 数据格式描述符字节偏移量
	DFDByteLength uint32 // 数据格式描述符字节长度
	KVDByteOffset uint32 // 键值数据字节偏移量
	KVDByteLength uint32 // 键值数据字节长度
	SGDByteOffset uint64 // 超级压缩全局数据字节偏移量
	SGDByteLength uint64 // 超级压缩全局数据字节长度
}

// Header KTX2容器级元数据
type Header struct {
	Format                 *Format                 // VkFormat
	TypeSize               uint32                  // 类型大小
	PixelWidth             uint32                  // 像素宽度
	PixelHeight            uint32                  // 像素高度
	PixelDepth             uint32                  // 像素深度
	LayerCount             uint32                  // 层数
	FaceCount              uint32                  // 面数
	LevelCount             uint32                  // 级别数
	SupercompressionScheme *SupercompressionScheme // 超级压缩方案
	Index                  Index                   // 索引信息
}

const HeaderLength = 80

// FromBytes 从字节数组解析头部
func HeaderFromBytes(data []byte) (*Header, error) {
	if len(data) < HeaderLength {
		return nil, UnexpectedEnd
	}

	// 验证魔数
	for i, b := range KTX2_MAGIC {
		if data[i] != b {
			return nil, BadMagic
		}
	}

	header := &Header{
		Format:                 NewFormat(binary.LittleEndian.Uint32(data[12:16])),
		TypeSize:               binary.LittleEndian.Uint32(data[16:20]),
		PixelWidth:             binary.LittleEndian.Uint32(data[20:24]),
		PixelHeight:            binary.LittleEndian.Uint32(data[24:28]),
		PixelDepth:             binary.LittleEndian.Uint32(data[28:32]),
		LayerCount:             binary.LittleEndian.Uint32(data[32:36]),
		FaceCount:              binary.LittleEndian.Uint32(data[36:40]),
		LevelCount:             binary.LittleEndian.Uint32(data[40:44]),
		SupercompressionScheme: NewSupercompressionScheme(binary.LittleEndian.Uint32(data[44:48])),
		Index: Index{
			DFDByteOffset: binary.LittleEndian.Uint32(data[48:52]),
			DFDByteLength: binary.LittleEndian.Uint32(data[52:56]),
			KVDByteOffset: binary.LittleEndian.Uint32(data[56:60]),
			KVDByteLength: binary.LittleEndian.Uint32(data[60:64]),
			SGDByteOffset: binary.LittleEndian.Uint64(data[64:72]),
			SGDByteLength: binary.LittleEndian.Uint64(data[72:80]),
		},
	}

	// 验证必需字段
	if header.PixelWidth == 0 {
		return nil, ZeroWidth
	}
	if header.FaceCount == 0 {
		return nil, ZeroFaceCount
	}

	return header, nil
}

// AsBytes 将头部转换为字节数组
func (h *Header) AsBytes() []byte {
	data := make([]byte, HeaderLength)

	// 写入魔数
	copy(data[0:12], KTX2_MAGIC[:])

	// 写入格式
	var format uint32
	if h.Format != nil {
		format = h.Format.Value()
	}
	binary.LittleEndian.PutUint32(data[12:16], format)

	// 写入其他字段
	binary.LittleEndian.PutUint32(data[16:20], h.TypeSize)
	binary.LittleEndian.PutUint32(data[20:24], h.PixelWidth)
	binary.LittleEndian.PutUint32(data[24:28], h.PixelHeight)
	binary.LittleEndian.PutUint32(data[28:32], h.PixelDepth)
	binary.LittleEndian.PutUint32(data[32:36], h.LayerCount)
	binary.LittleEndian.PutUint32(data[36:40], h.FaceCount)
	binary.LittleEndian.PutUint32(data[40:44], h.LevelCount)

	// 写入超级压缩方案
	var supercompression uint32
	if h.SupercompressionScheme != nil {
		supercompression = h.SupercompressionScheme.Value()
	}
	binary.LittleEndian.PutUint32(data[44:48], supercompression)

	// 写入索引信息
	binary.LittleEndian.PutUint32(data[48:52], h.Index.DFDByteOffset)
	binary.LittleEndian.PutUint32(data[52:56], h.Index.DFDByteLength)
	binary.LittleEndian.PutUint32(data[56:60], h.Index.KVDByteOffset)
	binary.LittleEndian.PutUint32(data[60:64], h.Index.KVDByteLength)
	binary.LittleEndian.PutUint64(data[64:72], h.Index.SGDByteOffset)
	binary.LittleEndian.PutUint64(data[72:80], h.Index.SGDByteLength)

	return data
}

// LevelIndex 级别索引信息
type LevelIndex struct {
	ByteOffset             uint64 // 字节偏移量
	ByteLength             uint64 // 字节长度
	UncompressedByteLength uint64 // 未压缩字节长度
}

const LevelIndexLength = 24

// FromBytes 从字节数组解析级别索引
func LevelIndexFromBytes(data []byte) (*LevelIndex, error) {
	if len(data) < LevelIndexLength {
		return nil, UnexpectedEnd
	}

	return &LevelIndex{
		ByteOffset:             binary.LittleEndian.Uint64(data[0:8]),
		ByteLength:             binary.LittleEndian.Uint64(data[8:16]),
		UncompressedByteLength: binary.LittleEndian.Uint64(data[16:24]),
	}, nil
}

// AsBytes 将级别索引转换为字节数组
func (l *LevelIndex) AsBytes() []byte {
	data := make([]byte, LevelIndexLength)
	binary.LittleEndian.PutUint64(data[0:8], l.ByteOffset)
	binary.LittleEndian.PutUint64(data[8:16], l.ByteLength)
	binary.LittleEndian.PutUint64(data[16:24], l.UncompressedByteLength)
	return data
}

// Level 纹理层级数据
type Level struct {
	Data                   []byte // 层级数据
	UncompressedByteLength uint64 // 未压缩字节长度
}

// DFDHeader 数据格式描述符头部
type DFDHeader struct {
	VendorID       uint32 // 供应商ID (17位)
	DescriptorType uint32 // 描述符类型 (15位)
	VersionNumber  uint16 // 版本号 (16位)
}

const DFDHeaderLength = 8

var DFDHeaderBasic = DFDHeader{
	VendorID:       0,
	DescriptorType: 0,
	VersionNumber:  2,
}

// AsBytes 将DFD头部转换为字节数组
func (d *DFDHeader) AsBytes(descriptorBlockSize uint16) []byte {
	data := make([]byte, DFDHeaderLength)

	firstWord := (d.VendorID & ((1 << 17) - 1)) | (d.DescriptorType << 17)
	binary.LittleEndian.PutUint32(data[0:4], firstWord)
	binary.LittleEndian.PutUint16(data[4:6], d.VersionNumber)
	binary.LittleEndian.PutUint16(data[6:8], descriptorBlockSize)

	return data
}

// ParseDFDHeader 解析DFD头部
func ParseDFDHeader(data []byte) (*DFDHeader, int, error) {
	if len(data) < DFDHeaderLength {
		return nil, 0, UnexpectedEnd
	}

	firstWord := binary.LittleEndian.Uint32(data[0:4])
	vendorID := firstWord & ((1 << 17) - 1)
	descriptorType := firstWord >> 17

	versionNumber := binary.LittleEndian.Uint16(data[4:6])
	descriptorBlockSize := binary.LittleEndian.Uint16(data[6:8])

	header := &DFDHeader{
		VendorID:       vendorID,
		DescriptorType: descriptorType,
		VersionNumber:  versionNumber,
	}

	return header, int(descriptorBlockSize), nil
}

// DFDBlock 数据格式描述符块
type DFDBlock struct {
	Header *DFDHeader
	Data   []byte
}

// DFDBlockHeaderBasic 基本数据格式描述符块头部
type DFDBlockHeaderBasic struct {
	ColorModel           *ColorModel       // 颜色模型
	ColorPrimaries       *ColorPrimaries   // 颜色原色
	TransferFunction     *TransferFunction // 传输函数
	Flags                DataFormatFlags   // 标志
	TexelBlockDimensions [4]uint8          // 纹素块维度 (每个值+1)
	BytesPlanes          [8]uint8          // 字节平面
}

const DFDBlockHeaderBasicLength = 16

// AsBytes 将基本DFD块头部转换为字节数组
func (d *DFDBlockHeaderBasic) AsBytes() []byte {
	data := make([]byte, DFDBlockHeaderBasicLength)

	var colorModel, colorPrimaries, transferFunction uint8
	if d.ColorModel != nil {
		colorModel = d.ColorModel.Value()
	}
	if d.ColorPrimaries != nil {
		colorPrimaries = d.ColorPrimaries.Value()
	}
	if d.TransferFunction != nil {
		transferFunction = d.TransferFunction.Value()
	}

	// 纹素块维度需要减1存储
	texelBlockDimensions := d.TexelBlockDimensions
	for i := range texelBlockDimensions {
		if texelBlockDimensions[i] > 0 {
			texelBlockDimensions[i]--
		}
	}

	data[0] = colorModel
	data[1] = colorPrimaries
	data[2] = transferFunction
	data[3] = uint8(d.Flags)
	copy(data[4:8], texelBlockDimensions[:])
	copy(data[8:16], d.BytesPlanes[:])

	return data
}

// FromBytes 从字节数组解析基本DFD块头部
func DFDBlockHeaderBasicFromBytes(data []byte) (*DFDBlockHeaderBasic, error) {
	if len(data) < DFDBlockHeaderBasicLength {
		return nil, UnexpectedEnd
	}

	// 纹素块维度需要加1恢复
	var texelBlockDimensions [4]uint8
	copy(texelBlockDimensions[:], data[4:8])
	for i := range texelBlockDimensions {
		texelBlockDimensions[i]++
	}

	var bytesPlanes [8]uint8
	copy(bytesPlanes[:], data[8:16])

	header := &DFDBlockHeaderBasic{
		ColorModel:           NewColorModel(data[0]),
		ColorPrimaries:       NewColorPrimaries(data[1]),
		TransferFunction:     NewTransferFunction(data[2]),
		Flags:                DataFormatFlags(data[3]),
		TexelBlockDimensions: texelBlockDimensions,
		BytesPlanes:          bytesPlanes,
	}

	return header, nil
}

// SampleInformation 采样信息
type SampleInformation struct {
	BitOffset             uint16                // 位偏移量
	BitLength             uint8                 // 位长度 (实际值+1)
	ChannelType           uint8                 // 通道类型
	ChannelTypeQualifiers ChannelTypeQualifiers // 通道类型限定符
	SamplePositions       [4]uint8              // 采样位置
	Lower                 uint32                // 下限
	Upper                 uint32                // 上限
}

const SampleInformationLength = 16

// AsBytes 将采样信息转换为字节数组
func (s *SampleInformation) AsBytes() []byte {
	data := make([]byte, SampleInformationLength)

	channelInfo := s.ChannelType | (uint8(s.ChannelTypeQualifiers) << 4)
	bitLength := s.BitLength
	if bitLength > 0 {
		bitLength-- // 存储时减1
	}

	binary.LittleEndian.PutUint16(data[0:2], s.BitOffset)
	data[2] = bitLength
	data[3] = channelInfo
	copy(data[4:8], s.SamplePositions[:])
	binary.LittleEndian.PutUint32(data[8:12], s.Lower)
	binary.LittleEndian.PutUint32(data[12:16], s.Upper)

	return data
}

// FromBytes 从字节数组解析采样信息
func SampleInformationFromBytes(data []byte) (*SampleInformation, error) {
	if len(data) < SampleInformationLength {
		return nil, UnexpectedEnd
	}

	firstWord := binary.LittleEndian.Uint32(data[0:4])
	bitOffset := uint16(firstWord & 0xFFFF)
	bitLength := uint8((firstWord >> 16) & 0xFF)
	channelType := uint8((firstWord >> 24) & 0xF)
	channelTypeQualifiers := ChannelTypeQualifiers((firstWord >> 28) & 0xF)

	// 位长度需要加1恢复
	if bitLength == 255 {
		return nil, InvalidSampleBitLength
	}
	bitLength++

	var samplePositions [4]uint8
	copy(samplePositions[:], data[4:8])

	lower := binary.LittleEndian.Uint32(data[8:12])
	upper := binary.LittleEndian.Uint32(data[12:16])

	return &SampleInformation{
		BitOffset:             bitOffset,
		BitLength:             bitLength,
		ChannelType:           channelType,
		ChannelTypeQualifiers: channelTypeQualifiers,
		SamplePositions:       samplePositions,
		Lower:                 lower,
		Upper:                 upper,
	}, nil
}

// KeyValuePair 键值对
type KeyValuePair struct {
	Key   string
	Value []byte
}

// Reader KTX2解码器
type Reader struct {
	input  []byte
	header *Header
}

// NewReader 创建新的KTX2解码器
func NewKTX2Reader(input []byte) (*Reader, error) {
	if len(input) < HeaderLength {
		return nil, UnexpectedEnd
	}

	header, err := HeaderFromBytes(input[0:HeaderLength])
	if err != nil {
		return nil, err
	}

	reader := &Reader{
		input:  input,
		header: header,
	}

	// 验证各部分边界
	if err := reader.validateBounds(); err != nil {
		return nil, err
	}

	// 验证级别索引完整性
	if _, err := reader.getLevelIndices(); err != nil {
		return nil, err
	}

	return reader, nil
}

// validateBounds 验证文件边界
func (r *Reader) validateBounds() error {
	inputLen := uint64(len(r.input))

	// 检查DFD边界
	dfdStart := uint64(r.header.Index.DFDByteOffset) + 4
	dfdEnd := uint64(r.header.Index.DFDByteOffset) + uint64(r.header.Index.DFDByteLength)
	if dfdEnd < dfdStart || dfdEnd > inputLen {
		return UnexpectedEnd
	}

	// 检查SGD边界
	sgdEnd := r.header.Index.SGDByteOffset + r.header.Index.SGDByteLength
	if sgdEnd > inputLen {
		return UnexpectedEnd
	}

	// 检查KVD边界
	kvdEnd := uint64(r.header.Index.KVDByteOffset) + uint64(r.header.Index.KVDByteLength)
	if kvdEnd > inputLen {
		return UnexpectedEnd
	}

	return nil
}

// Data 返回底层原始字节
func (r *Reader) Data() []byte {
	return r.input
}

// Header 返回容器级元数据
func (r *Reader) Header() *Header {
	return r.header
}

// getLevelIndices 获取级别索引
func (r *Reader) getLevelIndices() ([]*LevelIndex, error) {
	levelCount := int(r.header.LevelCount)
	if levelCount == 0 {
		levelCount = 1
	}

	levelIndexEndByte := HeaderLength + levelCount*LevelIndexLength
	if levelIndexEndByte > len(r.input) {
		return nil, UnexpectedEnd
	}

	indices := make([]*LevelIndex, levelCount)
	for i := 0; i < levelCount; i++ {
		start := HeaderLength + i*LevelIndexLength
		end := start + LevelIndexLength
		levelIndex, err := LevelIndexFromBytes(r.input[start:end])
		if err != nil {
			return nil, err
		}

		// 验证级别数据边界
		if levelIndex.ByteOffset+levelIndex.ByteLength > uint64(len(r.input)) {
			return nil, UnexpectedEnd
		}

		indices[i] = levelIndex
	}

	return indices, nil
}

// Levels 返回纹理层级迭代器
func (r *Reader) Levels() ([]*Level, error) {
	indices, err := r.getLevelIndices()
	if err != nil {
		return nil, err
	}

	levels := make([]*Level, len(indices))
	for i, index := range indices {
		start := int(index.ByteOffset)
		end := int(index.ByteOffset + index.ByteLength)
		levels[i] = &Level{
			Data:                   r.input[start:end],
			UncompressedByteLength: index.UncompressedByteLength,
		}
	}

	return levels, nil
}

// SupercompressionGlobalData 返回超级压缩全局数据
func (r *Reader) SupercompressionGlobalData() []byte {
	start := int(r.header.Index.SGDByteOffset)
	end := int(r.header.Index.SGDByteOffset + r.header.Index.SGDByteLength)
	return r.input[start:end]
}

// DFDBlocks 返回数据格式描述符块
func (r *Reader) DFDBlocks() ([]*DFDBlock, error) {
	start := int(r.header.Index.DFDByteOffset)
	end := int(r.header.Index.DFDByteOffset + r.header.Index.DFDByteLength)
	data := r.input[start+4 : end] // 跳过前4字节的总长度

	var blocks []*DFDBlock
	offset := 0

	for offset < len(data) {
		if offset+DFDHeaderLength > len(data) {
			break
		}

		header, blockSize, err := ParseDFDHeader(data[offset:])
		if err != nil {
			return nil, err
		}

		if blockSize == 0 || offset+blockSize > len(data) {
			break
		}

		blockData := data[offset+DFDHeaderLength : offset+blockSize]
		blocks = append(blocks, &DFDBlock{
			Header: header,
			Data:   blockData,
		})

		offset += blockSize
	}

	return blocks, nil
}

// KeyValueData 返回键值对数据迭代器
func (r *Reader) KeyValueData() ([]*KeyValuePair, error) {
	start := int(r.header.Index.KVDByteOffset)
	end := int(r.header.Index.KVDByteOffset + r.header.Index.KVDByteLength)
	data := r.input[start:end]

	var pairs []*KeyValuePair
	offset := 0

	for offset < len(data) {
		// 读取长度
		if offset+4 > len(data) {
			break
		}

		length := binary.LittleEndian.Uint32(data[offset : offset+4])
		offset += 4

		startOffset := offset
		endOffset := offset + int(length)

		if endOffset > len(data) {
			break
		}

		// 确保4字节对齐
		if endOffset%4 != 0 {
			endOffset += 4 - (endOffset % 4)
		}

		keyAndValue := data[startOffset : startOffset+int(length)]

		// 找到键的结束位置（NUL字符）
		nullIndex := -1
		for i, b := range keyAndValue {
			if b == 0 {
				nullIndex = i
				break
			}
		}

		if nullIndex == -1 {
			offset = endOffset
			continue
		}

		key := string(keyAndValue[:nullIndex])
		value := keyAndValue[nullIndex+1:]

		pairs = append(pairs, &KeyValuePair{
			Key:   key,
			Value: value,
		})

		offset = endOffset
	}

	return pairs, nil
}
