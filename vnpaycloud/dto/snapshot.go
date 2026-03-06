package dto

// Snapshot matches the iac-proxy-v2 Snapshot proto message.
type Snapshot struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	VolumeID    string `json:"volumeId"`
	SizeGB      int64  `json:"sizeGb,string"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	ProjectID   string `json:"projectId"`
	ZoneID      string `json:"zoneId"`
}

// CreateSnapshotRequest matches the iac-proxy-v2 CreateSnapshotRequest proto message.
// project_id is passed via URL path, not in the body.
type CreateSnapshotRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	VolumeID    string `json:"volumeId"`
}

// SnapshotResponse matches the iac-proxy-v2 SnapshotResponse proto message.
type SnapshotResponse struct {
	Snapshot Snapshot `json:"snapshot"`
}

// ListSnapshotsResponse matches the iac-proxy-v2 ListSnapshotsResponse proto message.
type ListSnapshotsResponse struct {
	Snapshots []Snapshot `json:"snapshots"`
}
