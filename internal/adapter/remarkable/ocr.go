package remarkable

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
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

//go:embed ocr_words.txt
var ocrWordsFile string

func OCRCustomWords() []string {
	var words []string
	for _, line := range strings.Split(ocrWordsFile, "\n") {
		w := strings.TrimSpace(line)
		if w != "" && !strings.HasPrefix(w, "#") {
			words = append(words, w)
		}
	}
	return words
}

func writeOCRCustomWordsFile() (string, func(), error) {
	f, err := os.CreateTemp("", "ocr-words-*.txt")
	if err != nil {
		return "", nil, err
	}
	words := strings.Join(OCRCustomWords(), "\n")
	if _, err := f.WriteString(words); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return "", nil, err
	}
	_ = f.Close()
	path := f.Name()
	return path, func() { _ = os.Remove(path) }, nil
}

func ParseOCRResults(data []byte) ([]OCRResult, error) {
	var results []OCRResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func RunOCR(ocrToolPath string, pngPath string) ([]OCRResult, error) {
	wordsPath, cleanup, err := writeOCRCustomWordsFile()
	if err != nil {
		return nil, fmt.Errorf("failed to write custom words: %w", err)
	}
	defer cleanup()

	cmd := exec.Command(ocrToolPath, pngPath, "--custom-words", wordsPath)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("OCR failed: %w", err)
	}
	return ParseOCRResults(out)
}
