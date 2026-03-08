package remarkable

import (
	"encoding/binary"
	"fmt"
	"math"
)

const rmHeaderV6 = "reMarkable .lines file, version=6          "

const (
	blockTypeSceneLineItem = 0x05
	fieldIndexValue        = 6
	fieldIndexPoints       = 5
	itemTypeLine           = 0x03
	pointSizeV1            = 24
	pointSizeV2            = 14
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
		if shift >= 64 {
			return 0, fmt.Errorf("varuint overflow at offset %d", r.pos)
		}
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

	contentLen := int(length)
	if r.remaining() < contentLen {
		return 0, 0, nil, fmt.Errorf("invalid block length %d at offset %d", length, r.pos)
	}

	content = r.data[r.pos : r.pos+contentLen]
	r.pos += contentLen
	return blockType, version, content, nil
}

const (
	tagByte1   = 0x1
	tagByte4   = 0x4
	tagByte8   = 0x8
	tagLength4 = 0xC
	tagID      = 0xF
)

func (r *rmReader) readTag() (index uint64, tagType uint8, err error) {
	v, err := r.readVaruint()
	if err != nil {
		return 0, 0, err
	}
	return v >> 4, uint8(v & 0xF), nil
}

func (r *rmReader) skipCrdtID() error {
	if _, err := r.readUint8(); err != nil {
		return err
	}
	if _, err := r.readVaruint(); err != nil {
		return err
	}
	return nil
}

func (r *rmReader) skipTaggedValue(tagType uint8) error {
	switch tagType {
	case tagByte1:
		return r.skip(1)
	case tagByte4:
		return r.skip(4)
	case tagByte8:
		return r.skip(8)
	case tagLength4:
		length, err := r.readUint32()
		if err != nil {
			return err
		}
		return r.skip(int(length))
	case tagID:
		return r.skipCrdtID()
	default:
		return fmt.Errorf("unknown tag type 0x%x at offset %d", tagType, r.pos)
	}
}

func (r *rmReader) skip(n int) error {
	if r.remaining() < n {
		return fmt.Errorf("unexpected EOF: cannot skip %d bytes at offset %d", n, r.pos)
	}
	r.pos += n
	return nil
}

func (r *rmReader) parseLineItemContent(blockVersion uint8, content []byte) (rmStroke, error) {
	cr := newRMReader(content)
	var stroke rmStroke

	for cr.remaining() > 0 {
		index, tagType, err := cr.readTag()
		if err != nil {
			return stroke, err
		}

		if index == fieldIndexValue && tagType == tagLength4 {
			length, err := cr.readUint32()
			if err != nil {
				return stroke, err
			}
			subEnd := cr.pos + int(length)

			itemType, err := cr.readUint8()
			if err != nil {
				return stroke, err
			}
			if itemType != itemTypeLine {
				cr.pos = subEnd
				continue
			}

			for cr.pos < subEnd {
				fi, ft, err := cr.readTag()
				if err != nil {
					break
				}
				if fi == fieldIndexPoints && ft == tagLength4 {
					pointsLen, err := cr.readUint32()
					if err != nil {
						return stroke, err
					}
					stroke.Points, err = cr.parsePoints(blockVersion, int(pointsLen))
					if err != nil {
						return stroke, err
					}
				} else {
					if err := cr.skipTaggedValue(ft); err != nil {
						return stroke, err
					}
				}
			}
			cr.pos = subEnd
		} else {
			if err := cr.skipTaggedValue(tagType); err != nil {
				return stroke, err
			}
		}
	}
	return stroke, nil
}

func (r *rmReader) parsePoints(blockVersion uint8, dataLen int) ([]rmPoint, error) {
	end := r.pos + dataLen
	var points []rmPoint

	if blockVersion >= 2 {
		for r.pos+pointSizeV2 <= end {
			x, err := r.readFloat32()
			if err != nil {
				return points, err
			}
			y, err := r.readFloat32()
			if err != nil {
				return points, err
			}
			if err := r.skip(2 + 2 + 1 + 1); err != nil { // speed, width, direction, pressure
				return points, err
			}
			points = append(points, rmPoint{X: x, Y: y})
		}
	} else {
		for r.pos+pointSizeV1 <= end {
			x, err := r.readFloat32()
			if err != nil {
				return points, err
			}
			y, err := r.readFloat32()
			if err != nil {
				return points, err
			}
			if err := r.skip(4 * 4); err != nil { // speed, direction, width, pressure
				return points, err
			}
			points = append(points, rmPoint{X: x, Y: y})
		}
	}
	r.pos = end
	return points, nil
}

func ParseRM(data []byte) ([]rmStroke, error) {
	r := newRMReader(data)
	if err := r.parseHeader(); err != nil {
		return nil, err
	}

	var strokes []rmStroke
	for r.remaining() > 0 {
		blockType, version, content, err := r.readBlock()
		if err != nil {
			return strokes, err
		}

		if blockType == blockTypeSceneLineItem {
			stroke, err := r.parseLineItemContent(version, content)
			if err != nil {
				continue
			}
			if len(stroke.Points) > 0 {
				strokes = append(strokes, stroke)
			}
		}
	}
	return strokes, nil
}
