package remarkable

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHeader_Valid(t *testing.T) {
	header := []byte("reMarkable .lines file, version=6          ")
	r := newRMReader(header)
	err := r.parseHeader()
	require.NoError(t, err)
}

func TestParseHeader_InvalidMagic(t *testing.T) {
	header := []byte("not a remarkable file")
	r := newRMReader(header)
	err := r.parseHeader()
	assert.ErrorContains(t, err, "invalid .rm header")
}

func TestParseHeader_WrongVersion(t *testing.T) {
	header := []byte("reMarkable .lines file, version=5          ")
	r := newRMReader(header)
	err := r.parseHeader()
	assert.ErrorContains(t, err, "unsupported .rm version")
}

func TestReadVaruint(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint64
	}{
		{"single byte", []byte{0x7F}, 127},
		{"two bytes", []byte{0x80, 0x01}, 128},
		{"example from spec", []byte{0x8C, 0x01}, 140},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRMReader(tt.data)
			v, err := r.readVaruint()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, v)
		})
	}
}

func TestReadVaruint_EOF(t *testing.T) {
	r := newRMReader([]byte{0x80})
	_, err := r.readVaruint()
	assert.ErrorContains(t, err, "unexpected EOF in varuint")
}

func TestReadVaruint_Overflow(t *testing.T) {
	data := make([]byte, 11)
	for i := range data {
		data[i] = 0x80
	}
	r := newRMReader(data)
	_, err := r.readVaruint()
	assert.ErrorContains(t, err, "varuint overflow")
}

func TestReadBlock_TruncatedData(t *testing.T) {
	r := newRMReader([]byte{0x01, 0x02})
	_, _, _, err := r.readBlock()
	assert.ErrorContains(t, err, "reading block length")
}

func TestReadBlock_ContentExceedsFile(t *testing.T) {
	block := make([]byte, 0, 8)
	block = binary.LittleEndian.AppendUint32(block, 100)
	block = append(block, 0x00, 0x01, 0x01, 0x05)
	r := newRMReader(block)
	_, _, _, err := r.readBlock()
	assert.ErrorContains(t, err, "invalid block length")
}

func TestReadBlock(t *testing.T) {
	content := []byte{0xAA, 0xBB}
	block := make([]byte, 0, 8+len(content))
	block = binary.LittleEndian.AppendUint32(block, uint32(len(content)))
	block = append(block, 0x00)       // unknown
	block = append(block, 0x01)       // min_version
	block = append(block, 0x01)       // current_version
	block = append(block, 0x05)       // block_type = SceneLineItem
	block = append(block, content...)

	r := newRMReader(block)
	blockType, blockVersion, blockContent, err := r.readBlock()
	require.NoError(t, err)
	assert.Equal(t, uint8(0x05), blockType)
	assert.Equal(t, uint8(0x01), blockVersion)
	assert.Equal(t, content, blockContent)
}

func TestReadTag(t *testing.T) {
	// Tag varuint = (index << 4) | tagType
	// index=1, tagType=4 (Byte4) => varuint = (1 << 4) | 4 = 0x14
	data := []byte{0x14}
	data = binary.LittleEndian.AppendUint32(data, 42)
	r := newRMReader(data)

	index, tagType, err := r.readTag()
	require.NoError(t, err)
	assert.Equal(t, uint64(1), index)
	assert.Equal(t, uint8(0x4), tagType)
}

func buildTestRM(t *testing.T, points []rmPoint) []byte {
	t.Helper()
	return buildTestRMMultiStroke(t, [][]rmPoint{points})
}

func buildSceneLineItemBlock(t *testing.T, points []rmPoint) []byte {
	t.Helper()

	pointData := make([]byte, 0, len(points)*24)
	for _, p := range points {
		pointData = binary.LittleEndian.AppendUint32(pointData, math.Float32bits(p.X))
		pointData = binary.LittleEndian.AppendUint32(pointData, math.Float32bits(p.Y))
		pointData = binary.LittleEndian.AppendUint32(pointData, 0) // speed
		pointData = binary.LittleEndian.AppendUint32(pointData, 0) // direction
		pointData = binary.LittleEndian.AppendUint32(pointData, 0) // width
		pointData = binary.LittleEndian.AppendUint32(pointData, 0) // pressure
	}

	var valueSub []byte
	valueSub = append(valueSub, 0x03) // item_type = line
	valueSub = append(valueSub, 0x14)
	valueSub = binary.LittleEndian.AppendUint32(valueSub, 0)
	valueSub = append(valueSub, 0x24)
	valueSub = binary.LittleEndian.AppendUint32(valueSub, 0)
	valueSub = append(valueSub, 0x38)
	valueSub = binary.LittleEndian.AppendUint64(valueSub, math.Float64bits(1.0))
	valueSub = append(valueSub, 0x44)
	valueSub = binary.LittleEndian.AppendUint32(valueSub, 0)
	valueSub = append(valueSub, 0x5C)
	valueSub = binary.LittleEndian.AppendUint32(valueSub, uint32(len(pointData)))
	valueSub = append(valueSub, pointData...)

	var blockContent []byte
	blockContent = append(blockContent, 0x1F, 0x01, 0x01)
	blockContent = append(blockContent, 0x2F, 0x01, 0x01)
	blockContent = append(blockContent, 0x3F, 0x01, 0x01)
	blockContent = append(blockContent, 0x4F, 0x01, 0x01)
	blockContent = append(blockContent, 0x54)
	blockContent = binary.LittleEndian.AppendUint32(blockContent, 0)
	blockContent = append(blockContent, 0x6C)
	blockContent = binary.LittleEndian.AppendUint32(blockContent, uint32(len(valueSub)))
	blockContent = append(blockContent, valueSub...)

	var block []byte
	block = binary.LittleEndian.AppendUint32(block, uint32(len(blockContent)))
	block = append(block, 0x00)
	block = append(block, 0x01)
	block = append(block, 0x01)
	block = append(block, blockTypeSceneLineItem)
	block = append(block, blockContent...)
	return block
}

func buildTestRMMultiStroke(t *testing.T, strokePoints [][]rmPoint) []byte {
	t.Helper()
	var rm []byte
	rm = append(rm, []byte(rmHeaderV6)...)
	for _, points := range strokePoints {
		rm = append(rm, buildSceneLineItemBlock(t, points)...)
	}
	return rm
}

func TestParseRM_SyntheticStrokes(t *testing.T) {
	rmData := buildTestRM(t, []rmPoint{
		{X: 100.0, Y: 200.0},
		{X: 150.0, Y: 250.0},
	})

	strokes, err := ParseRM(rmData)
	require.NoError(t, err)
	require.Len(t, strokes, 1)
	require.Len(t, strokes[0].Points, 2)
	assert.InDelta(t, 100.0, strokes[0].Points[0].X, 0.01)
	assert.InDelta(t, 200.0, strokes[0].Points[0].Y, 0.01)
	assert.InDelta(t, 150.0, strokes[0].Points[1].X, 0.01)
	assert.InDelta(t, 250.0, strokes[0].Points[1].Y, 0.01)
}

func TestParseRM_EmptyFile(t *testing.T) {
	rmData := []byte(rmHeaderV6)
	strokes, err := ParseRM(rmData)
	require.NoError(t, err)
	assert.Empty(t, strokes)
}
