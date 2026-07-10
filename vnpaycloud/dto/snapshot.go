package dto

// Snapshot matches the backend Snapshot proto message.
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

// CreateSnapshotRequest matches the backend CreateSnapshotRequest proto message.
// project_id is passed via URL path, not in the body. description is read-only
// (set by the backend), so it is not part of the create request.
type CreateSnapshotRequest struct {
	Name     string `json:"name"`
	VolumeID string `json:"volumeId"`
}

// SnapshotResponse matches the backend SnapshotResponse proto message.
type SnapshotResponse struct {
	Snapshot Snapshot `json:"snapshot"`
}

// ListSnapshotsResponse matches the backend ListSnapshotsResponse proto message.
type ListSnapshotsResponse struct {
	Snapshots []Snapshot `json:"snapshots"`
}
