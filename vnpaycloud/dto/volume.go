package dto

// Volume matches the iac-proxy-v2 Volume proto message.
type Volume struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	SizeGB             int64  `json:"sizeGb,string"`
	VolumeType         string `json:"volumeType"`
	Zone               string `json:"zone"`
	Status             string `json:"status"`
	IOPS               int32  `json:"iops"`
	IsEncrypted        bool   `json:"isEncrypted"`
	IsMultiattach      bool   `json:"isMultiattach"`
	IsBootable         bool   `json:"isBootable"`
	AttachedServerID   string `json:"attachedServerId"`
	AttachedServerName string `json:"attachedServerName"`
	CreatedAt          string `json:"createdAt"`
	ProjectID          string `json:"projectId"`
	ZoneID             string `json:"zoneId"`
}

// CreateVolumeRequest matches the iac-proxy-v2 CreateVolumeRequest proto message.
// project_id is passed via URL path, not in the body.
// zone is resolved server-side from project_id.
type CreateVolumeRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	SizeGB      int64  `json:"sizeGb,string"`
	VolumeType  string `json:"volumeType"`
	Encrypt     bool   `json:"encrypt,omitempty"`
	Multiattach bool   `json:"multiattach,omitempty"`
	SnapshotID  string `json:"snapshotId,omitempty"`
}

// UpdateVolumeRequest matches the iac-proxy-v2 UpdateVolumeRequest proto message.
// project_id and id are passed via URL path.
type UpdateVolumeRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ResizeVolumeRequest matches the iac-proxy-v2 ResizeVolumeRequest proto message.
// project_id and id are passed via URL path.
type ResizeVolumeRequest struct {
	SizeGB int64 `json:"sizeGb,string"`
}

// VolumeResponse matches the iac-proxy-v2 VolumeResponse proto message.
type VolumeResponse struct {
	Volume Volume `json:"volume"`
}

// ListVolumesResponse matches the iac-proxy-v2 ListVolumesResponse proto message.
type ListVolumesResponse struct {
	Volumes []Volume `json:"volumes"`
}
