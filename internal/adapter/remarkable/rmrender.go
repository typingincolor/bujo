package remarkable

import (
	"bytes"
	"image/png"

	"github.com/fogleman/gg"
)

const remarkableScreenHeight = 1872

func RenderStrokes(strokes []rmStroke) ([]byte, error) {
	dc := gg.NewContext(remarkableScreenWidth, remarkableScreenHeight)

	dc.SetRGB(1, 1, 1)
	dc.Clear()

	dc.SetRGB(0, 0, 0)
	dc.SetLineWidth(1)

	for _, stroke := range strokes {
		if len(stroke.Points) < 2 {
			continue
		}
		dc.MoveTo(float64(stroke.Points[0].X), float64(stroke.Points[0].Y))
		for _, p := range stroke.Points[1:] {
			dc.LineTo(float64(p.X), float64(p.Y))
		}
		dc.Stroke()
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
