package dto

// VolumeType matches the iac-proxy-v2 VolumeType proto message.
type VolumeType struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	IOPS          int32  `json:"iops"`
	IsEncrypted   bool   `json:"isEncrypted"`
	IsMultiattach bool   `json:"isMultiattach"`
	Zone          string `json:"zone"`
}

// VolumeTypeResponse matches the iac-proxy-v2 VolumeTypeResponse proto message.
type VolumeTypeResponse struct {
	VolumeType VolumeType `json:"volumeType"`
}

// ListVolumeTypesResponse matches the iac-proxy-v2 ListVolumeTypesResponse proto message.
type ListVolumeTypesResponse struct {
	VolumeTypes []VolumeType `json:"volumeTypes"`
}
