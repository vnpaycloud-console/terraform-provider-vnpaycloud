package dto

// VolumeType matches the backend VolumeType proto message.
type VolumeType struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	IOPS          int32  `json:"iops"`
	IsEncrypted   bool   `json:"isEncrypted"`
	IsMultiattach bool   `json:"isMultiattach"`
	Zone          string `json:"zone"`
}

// VolumeTypeResponse matches the backend VolumeTypeResponse proto message.
type VolumeTypeResponse struct {
	VolumeType VolumeType `json:"volumeType"`
}

// ListVolumeTypesResponse matches the backend ListVolumeTypesResponse proto message.
type ListVolumeTypesResponse struct {
	VolumeTypes []VolumeType `json:"volumeTypes"`
}
