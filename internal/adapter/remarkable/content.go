package remarkable

import "encoding/json"

type ContentFile struct {
	CPages struct {
		Pages []struct {
			ID string `json:"id"`
		} `json:"pages"`
	} `json:"cPages"`
	FileType string `json:"fileType"`
}

func ParsePageOrder(data []byte) ([]string, error) {
	var content ContentFile
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, err
	}

	pages := make([]string, 0, len(content.CPages.Pages))
	for _, p := range content.CPages.Pages {
		pages = append(pages, p.ID)
	}
	return pages, nil
}
