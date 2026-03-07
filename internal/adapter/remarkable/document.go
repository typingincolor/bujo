package remarkable

type Document struct {
	ID          string `json:"ID"`
	Version     int    `json:"Version"`
	VisibleName string `json:"VissibleName"`
	Type        string `json:"Type"`
	Parent      string `json:"Parent"`
	ModifiedAt  string `json:"ModifiedClient"`
	BlobURLGet  string `json:"BlobURLGet"`
}
