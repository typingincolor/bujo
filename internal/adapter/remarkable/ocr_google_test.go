package remarkable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	visionpb "cloud.google.com/go/vision/v2/apiv1/visionpb"
)

func TestGoogleVisionOCR_ImplementsOCRProvider(t *testing.T) {
	var _ OCRProvider = &GoogleVisionOCR{}
}

func TestConvertAnnotationToResults_EmitsPerWord(t *testing.T) {
	annotation := &visionpb.TextAnnotation{
		Pages: []*visionpb.Page{
			{
				Width:  1000,
				Height: 1000,
				Blocks: []*visionpb.Block{
					{
						Paragraphs: []*visionpb.Paragraph{
							{
								Words: []*visionpb.Word{
									{
										BoundingBox: &visionpb.BoundingPoly{
											Vertices: []*visionpb.Vertex{
												{X: 50, Y: 100},
												{X: 70, Y: 100},
												{X: 70, Y: 130},
												{X: 50, Y: 130},
											},
										},
										Confidence: 0.92,
										Symbols:    []*visionpb.Symbol{{Text: "."}},
									},
									{
										BoundingBox: &visionpb.BoundingPoly{
											Vertices: []*visionpb.Vertex{
												{X: 80, Y: 100},
												{X: 150, Y: 100},
												{X: 150, Y: 130},
												{X: 80, Y: 130},
											},
										},
										Confidence: 0.88,
										Symbols: []*visionpb.Symbol{
											{Text: "b"},
											{Text: "u"},
											{Text: "y"},
										},
									},
									{
										BoundingBox: &visionpb.BoundingPoly{
											Vertices: []*visionpb.Vertex{
												{X: 160, Y: 100},
												{X: 250, Y: 100},
												{X: 250, Y: 130},
												{X: 160, Y: 130},
											},
										},
										Confidence: 0.91,
										Symbols: []*visionpb.Symbol{
											{Text: "m"},
											{Text: "i"},
											{Text: "l"},
											{Text: "k"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	results := convertAnnotationToResults(annotation)
	require.Len(t, results, 3)
	assert.Equal(t, ".", results[0].Text)
	assert.Equal(t, "buy", results[1].Text)
	assert.Equal(t, "milk", results[2].Text)
	assert.InDelta(t, 50.0, results[0].X, 0.01)
	assert.InDelta(t, 80.0, results[1].X, 0.01)
	assert.InDelta(t, 0.88, float64(results[1].Confidence), 0.01)
}

func TestConvertAnnotationToResults_WordBoundingBox(t *testing.T) {
	annotation := &visionpb.TextAnnotation{
		Pages: []*visionpb.Page{
			{
				Width:  1000,
				Height: 1000,
				Blocks: []*visionpb.Block{
					{
						Paragraphs: []*visionpb.Paragraph{
							{
								Words: []*visionpb.Word{
									{
										BoundingBox: &visionpb.BoundingPoly{
											Vertices: []*visionpb.Vertex{
												{X: 50, Y: 100},
												{X: 250, Y: 100},
												{X: 250, Y: 130},
												{X: 50, Y: 130},
											},
										},
										Confidence: 0.95,
										Symbols: []*visionpb.Symbol{
											{Text: "h"},
											{Text: "e"},
											{Text: "l"},
											{Text: "l"},
											{Text: "o"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	results := convertAnnotationToResults(annotation)
	require.Len(t, results, 1)
	assert.Equal(t, "hello", results[0].Text)
	assert.InDelta(t, 50.0, results[0].X, 0.01)
	assert.InDelta(t, 100.0, results[0].Y, 0.01)
	assert.InDelta(t, 200.0, results[0].Width, 0.01)
	assert.InDelta(t, 30.0, results[0].Height, 0.01)
	assert.InDelta(t, 0.95, float64(results[0].Confidence), 0.01)
}

func TestConvertAnnotationToResults_NilAnnotation(t *testing.T) {
	results := convertAnnotationToResults(nil)
	assert.Nil(t, results)
}

func TestConvertAnnotationToResults_EmptyPage(t *testing.T) {
	annotation := &visionpb.TextAnnotation{
		Pages: []*visionpb.Page{
			{Width: 1000, Height: 1000},
		},
	}
	results := convertAnnotationToResults(annotation)
	assert.Empty(t, results)
}

func TestConvertAnnotationToResults_MultipleLines(t *testing.T) {
	annotation := &visionpb.TextAnnotation{
		Pages: []*visionpb.Page{
			{
				Width:  1000,
				Height: 1000,
				Blocks: []*visionpb.Block{
					{
						Paragraphs: []*visionpb.Paragraph{
							{
								Words: []*visionpb.Word{
									{
										BoundingBox: &visionpb.BoundingPoly{
											Vertices: []*visionpb.Vertex{
												{X: 50, Y: 100},
												{X: 250, Y: 100},
												{X: 250, Y: 130},
												{X: 50, Y: 130},
											},
										},
										Confidence: 0.95,
										Symbols: []*visionpb.Symbol{
											{Text: "f"},
											{Text: "i"},
											{Text: "r"},
											{Text: "s"},
											{Text: "t"},
										},
									},
								},
							},
							{
								Words: []*visionpb.Word{
									{
										BoundingBox: &visionpb.BoundingPoly{
											Vertices: []*visionpb.Vertex{
												{X: 50, Y: 200},
												{X: 300, Y: 200},
												{X: 300, Y: 230},
												{X: 50, Y: 230},
											},
										},
										Confidence: 0.88,
										Symbols: []*visionpb.Symbol{
											{Text: "s"},
											{Text: "e"},
											{Text: "c"},
											{Text: "o"},
											{Text: "n"},
											{Text: "d"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	results := convertAnnotationToResults(annotation)
	require.Len(t, results, 2)
	assert.Equal(t, "first", results[0].Text)
	assert.Equal(t, "second", results[1].Text)
	assert.InDelta(t, 100.0, results[0].Y, 0.01)
	assert.InDelta(t, 200.0, results[1].Y, 0.01)
}

func TestConvertAnnotationToResults_ReconstructionIntegration(t *testing.T) {
	annotation := &visionpb.TextAnnotation{
		Pages: []*visionpb.Page{
			{
				Width:  1000,
				Height: 1400,
				Blocks: []*visionpb.Block{
					{
						Paragraphs: []*visionpb.Paragraph{
							{
								Words: []*visionpb.Word{
									wordAt(50, 100, 70, 130, ".", 0.95),
									wordAt(80, 100, 200, 130, "buy", 0.90),
									wordAt(210, 100, 350, 130, "milk", 0.92),
								},
							},
							{
								Words: []*visionpb.Word{
									wordAt(50, 200, 70, 230, "-", 0.93),
									wordAt(80, 200, 300, 230, "meeting", 0.89),
									wordAt(310, 200, 450, 230, "notes", 0.91),
								},
							},
							{
								Words: []*visionpb.Word{
									wordAt(100, 300, 120, 330, ".", 0.88),
									wordAt(130, 300, 280, 330, "sub", 0.85),
									wordAt(290, 300, 450, 330, "task", 0.87),
								},
							},
						},
					},
				},
			},
		},
	}

	results := convertAnnotationToResults(annotation)
	text := ReconstructText(results)
	assert.Contains(t, text, ". buy milk")
	assert.Contains(t, text, "- meeting notes")
	assert.Contains(t, text, ". sub task")
}

func wordAt(x1, y1, x2, y2 int32, text string, confidence float32) *visionpb.Word {
	var symbols []*visionpb.Symbol
	for _, r := range text {
		symbols = append(symbols, &visionpb.Symbol{Text: string(r)})
	}
	return &visionpb.Word{
		BoundingBox: &visionpb.BoundingPoly{
			Vertices: []*visionpb.Vertex{
				{X: x1, Y: y1},
				{X: x2, Y: y1},
				{X: x2, Y: y2},
				{X: x1, Y: y2},
			},
		},
		Confidence: confidence,
		Symbols:    symbols,
	}
}
