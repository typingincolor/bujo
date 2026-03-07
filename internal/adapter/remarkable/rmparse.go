package remarkable

import (
	"encoding/binary"
	"fmt"
	"math"
)

const rmHeaderV6 = "reMarkable .lines file, version=6          "

const (
	blockTypeSceneLineItem = 0x05
)

type rmPoint struct {
	X float32
	Y float32
}

type rmStroke struct {
	Points []rmPoint
}

type rmReader struct {
	data []byte
	pos  int
}

func newRMReader(data []byte) *rmReader {
	return &rmReader{data: data, pos: 0}
}

func (r *rmReader) parseHeader() error {
	if len(r.data) < len(rmHeaderV6) {
		return fmt.Errorf("invalid .rm header: file too short")
	}
	header := string(r.data[:len(rmHeaderV6)])
	if header != rmHeaderV6 {
		if len(header) > 32 && header[:32] == "reMarkable .lines file, version=" {
			return fmt.Errorf("unsupported .rm version: %s", header[32:])
		}
		return fmt.Errorf("invalid .rm header")
	}
	r.pos = len(rmHeaderV6)
	return nil
}

func (r *rmReader) remaining() int {
	return len(r.data) - r.pos
}

func (r *rmReader) readUint8() (uint8, error) {
	if r.remaining() < 1 {
		return 0, fmt.Errorf("unexpected EOF at offset %d", r.pos)
	}
	v := r.data[r.pos]
	r.pos++
	return v, nil
}

func (r *rmReader) readUint16() (uint16, error) {
	if r.remaining() < 2 {
		return 0, fmt.Errorf("unexpected EOF at offset %d", r.pos)
	}
	v := binary.LittleEndian.Uint16(r.data[r.pos:])
	r.pos += 2
	return v, nil
}

func (r *rmReader) readUint32() (uint32, error) {
	if r.remaining() < 4 {
		return 0, fmt.Errorf("unexpected EOF at offset %d", r.pos)
	}
	v := binary.LittleEndian.Uint32(r.data[r.pos:])
	r.pos += 4
	return v, nil
}

func (r *rmReader) readFloat32() (float32, error) {
	bits, err := r.readUint32()
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(bits), nil
}

func (r *rmReader) readFloat64() (float64, error) {
	if r.remaining() < 8 {
		return 0, fmt.Errorf("unexpected EOF at offset %d", r.pos)
	}
	bits := binary.LittleEndian.Uint64(r.data[r.pos:])
	r.pos += 8
	return math.Float64frombits(bits), nil
}

func (r *rmReader) readVaruint() (uint64, error) {
	var result uint64
	var shift uint
	for {
		if r.remaining() < 1 {
			return 0, fmt.Errorf("unexpected EOF in varuint at offset %d", r.pos)
		}
		b := r.data[r.pos]
		r.pos++
		result |= uint64(b&0x7F) << shift
		if b&0x80 == 0 {
			return result, nil
		}
		shift += 7
	}
}

func (r *rmReader) readBlock() (blockType uint8, version uint8, content []byte, err error) {
	length, err := r.readUint32()
	if err != nil {
		return 0, 0, nil, fmt.Errorf("reading block length: %w", err)
	}

	_, err = r.readUint8() // unknown byte
	if err != nil {
		return 0, 0, nil, err
	}

	_, err = r.readUint8() // min_version
	if err != nil {
		return 0, 0, nil, err
	}

	version, err = r.readUint8() // current_version
	if err != nil {
		return 0, 0, nil, err
	}

	blockType, err = r.readUint8()
	if err != nil {
		return 0, 0, nil, err
	}

	contentLen := int(length) - 4
	if contentLen < 0 || r.remaining() < contentLen {
		return 0, 0, nil, fmt.Errorf("invalid block length %d at offset %d", length, r.pos)
	}

	content = r.data[r.pos : r.pos+contentLen]
	r.pos += contentLen
	return blockType, version, content, nil
}

func (r *rmReader) skip(n int) error {
	if r.remaining() < n {
		return fmt.Errorf("unexpected EOF: cannot skip %d bytes at offset %d", n, r.pos)
	}
	r.pos += n
	return nil
}
