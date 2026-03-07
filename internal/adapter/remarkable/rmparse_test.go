package remarkable

import (
	"encoding/binary"
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

func TestReadBlock_LengthLessThanFour(t *testing.T) {
	block := make([]byte, 0, 8)
	block = binary.LittleEndian.AppendUint32(block, 3)
	block = append(block, 0x00, 0x01, 0x01, 0x05)
	r := newRMReader(block)
	_, _, _, err := r.readBlock()
	assert.ErrorContains(t, err, "invalid block length")
}

func TestReadBlock(t *testing.T) {
	content := []byte{0xAA, 0xBB}
	block := make([]byte, 0, 8+len(content))
	block = binary.LittleEndian.AppendUint32(block, uint32(len(content)+4))
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
