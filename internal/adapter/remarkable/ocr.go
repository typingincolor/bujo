package remarkable

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type OCRCandidate struct {
	Text       string  `json:"text"`
	Confidence float32 `json:"confidence"`
}

type OCRResult struct {
	Text       string         `json:"text"`
	X          float64        `json:"x"`
	Y          float64        `json:"y"`
	Width      float64        `json:"width"`
	Height     float64        `json:"height"`
	Confidence float32        `json:"confidence"`
	Candidates []OCRCandidate `json:"candidates,omitempty"`
}

func ParseOCRResults(data []byte) ([]OCRResult, error) {
	var results []OCRResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func RunOCR(ocrToolPath string, pngPath string) ([]OCRResult, error) {
	cmd := exec.Command(ocrToolPath, pngPath)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("OCR failed: %w", err)
	}
	return ParseOCRResults(out)
}
