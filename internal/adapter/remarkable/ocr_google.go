package remarkable

import (
	"context"
	"fmt"
	"os"
	"strings"

	vision "cloud.google.com/go/vision/v2/apiv1"
	visionpb "cloud.google.com/go/vision/v2/apiv1/visionpb"
)

type GoogleVisionOCR struct{}

func (g *GoogleVisionOCR) RecognizeText(ctx context.Context, imagePath string) ([]OCRResult, error) {
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create vision client: %w", err)
	}
	defer func() { _ = client.Close() }()

	imgBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}

	image := &visionpb.Image{Content: imgBytes}
	req := &visionpb.BatchAnnotateImagesRequest{
		Requests: []*visionpb.AnnotateImageRequest{
			{
				Image: image,
				Features: []*visionpb.Feature{
					{Type: visionpb.Feature_DOCUMENT_TEXT_DETECTION},
				},
				ImageContext: &visionpb.ImageContext{
					LanguageHints: []string{"en"},
				},
			},
		},
	}

	resp, err := client.BatchAnnotateImages(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("vision API call failed: %w", err)
	}
	if len(resp.Responses) == 0 {
		return nil, fmt.Errorf("vision API returned no responses")
	}
	result := resp.Responses[0]
	if result.Error != nil {
		return nil, fmt.Errorf("vision API error: %s", result.Error.Message)
	}

	return convertAnnotationToResults(result.FullTextAnnotation), nil
}

func convertAnnotationToResults(annotation *visionpb.TextAnnotation) []OCRResult {
	if annotation == nil {
		return nil
	}

	var results []OCRResult

	for _, page := range annotation.Pages {
		for _, block := range page.Blocks {
			for _, paragraph := range block.Paragraphs {
				for _, word := range paragraph.Words {
					results = append(results, wordToOCRResult(word))
				}
			}
		}
	}

	return results
}

func wordToOCRResult(word *visionpb.Word) OCRResult {
	var symbols []string
	for _, sym := range word.Symbols {
		symbols = append(symbols, sym.Text)
	}
	text := strings.Join(symbols, "")

	var x, y, w, h float64
	if bb := word.BoundingBox; bb != nil && len(bb.Vertices) >= 4 {
		x = float64(bb.Vertices[0].X)
		y = float64(bb.Vertices[0].Y)
		w = float64(bb.Vertices[1].X) - x
		h = float64(bb.Vertices[2].Y) - y
	}

	return OCRResult{
		Text:       text,
		X:          x,
		Y:          y,
		Width:      w,
		Height:     h,
		Confidence: word.Confidence,
	}
}
