package dto

// Image matches the backend Image proto message.
type Image struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	OsType    string `json:"osType"`
	OsVersion string `json:"osVersion"`
	MinDiskGB int32  `json:"minDiskGb"`
	Status    string `json:"status"`
	Zone      string `json:"zone"`
}

// ImageResponse matches the backend ImageResponse proto message.
type ImageResponse struct {
	Image Image `json:"image"`
}

// ListImagesResponse matches the backend ListImagesResponse proto message.
type ListImagesResponse struct {
	Images []Image `json:"images"`
}
