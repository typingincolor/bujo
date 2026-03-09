package remarkable

import (
	"bytes"
	"image/png"
	"math"

	"github.com/fogleman/gg"
)

const remarkableScreenHeight = 1872

func detectXOffset(strokes []rmStroke) float64 {
	halfWidth := float64(remarkableScreenWidth) / 2
	if len(strokes) == 0 {
		return halfWidth
	}

	minX, maxX := math.MaxFloat64, -math.MaxFloat64
	for _, s := range strokes {
		for _, p := range s.Points {
			x := float64(p.X)
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
		}
	}

	centerX := (minX + maxX) / 2
	if math.Abs(centerX) < math.Abs(centerX-halfWidth) {
		return halfWidth
	}
	return 0
}

func detectYOffset(strokes []rmStroke) float64 {
	if len(strokes) == 0 {
		return 0
	}
	minY := math.MaxFloat64
	for _, s := range strokes {
		for _, p := range s.Points {
			y := float64(p.Y)
			if y < minY {
				minY = y
			}
		}
	}
	if minY < 0 {
		return -minY
	}
	return 0
}

func RenderStrokes(strokes []rmStroke) ([]byte, error) {
	xOffset := detectXOffset(strokes)
	yOffset := detectYOffset(strokes)

	dc := gg.NewContext(remarkableScreenWidth, remarkableScreenHeight)

	dc.SetRGB(1, 1, 1)
	dc.Clear()

	dc.SetRGB(0, 0, 0)
	dc.SetLineWidth(1)

	for _, stroke := range strokes {
		if len(stroke.Points) < 2 {
			continue
		}
		dc.MoveTo(float64(stroke.Points[0].X)+xOffset, float64(stroke.Points[0].Y)+yOffset)
		for _, p := range stroke.Points[1:] {
			dc.LineTo(float64(p.X)+xOffset, float64(p.Y)+yOffset)
		}
		dc.Stroke()
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
