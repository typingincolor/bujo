package remarkable

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type Document struct {
	ID          string `json:"ID"`
	Version     int    `json:"Version"`
	VisibleName string `json:"VissibleName"`
	Type        string `json:"Type"`
	Parent      string `json:"Parent"`
	ModifiedAt  string `json:"ModifiedClient"`
	BlobURLGet  string `json:"BlobURLGet"`
}

func ExtractTextFromZIP(data []byte) ([]string, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP: %w", err)
	}

	var texts []string
	for _, f := range r.File {
		ext := strings.ToLower(filepath.Ext(f.Name))
		if ext == ".txt" {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open %s: %w", f.Name, err)
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to read %s: %w", f.Name, err)
			}
			texts = append(texts, string(content))
		}
	}
	return texts, nil
}

func ListZIPContents(data []byte) ([]string, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP: %w", err)
	}
	var names []string
	for _, f := range r.File {
		names = append(names, fmt.Sprintf("%s (%d bytes)", f.Name, f.UncompressedSize64))
	}
	return names, nil
}
