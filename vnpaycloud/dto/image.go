package dto

// Image matches the iac-proxy-v2 Image proto message.
type Image struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	OsType    string `json:"osType"`
	OsVersion string `json:"osVersion"`
	MinDiskGB int32  `json:"minDiskGb"`
	Status    string `json:"status"`
}

// ImageResponse matches the iac-proxy-v2 ImageResponse proto message.
type ImageResponse struct {
	Image Image `json:"image"`
}

// ListImagesResponse matches the iac-proxy-v2 ListImagesResponse proto message.
type ListImagesResponse struct {
	Images []Image `json:"images"`
}
