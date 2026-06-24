package dto

// VolumeAttachment matches the backend VolumeAttachment proto message.
type VolumeAttachment struct {
	ID         string `json:"id"`
	VolumeID   string `json:"volumeId"`
	ServerID   string `json:"serverId"`
	Device     string `json:"device"`
	Status     string `json:"status"`
	AttachedAt string `json:"attachedAt"`
	ProjectID  string `json:"projectId"`
	ZoneID     string `json:"zoneId"`
}

// AttachVolumeRequest matches the backend AttachVolumeRequest proto message.
// project_id and volume_id are passed via URL path.
type AttachVolumeRequest struct {
	ServerID string `json:"serverId"`
}

// DetachVolumeRequest matches the backend DetachVolumeRequest proto message.
// project_id and volume_id are passed via URL path.
type DetachVolumeRequest struct {
	ServerID string `json:"serverId"`
}

// VolumeAttachmentResponse matches the backend VolumeAttachmentResponse proto message.
type VolumeAttachmentResponse struct {
	Attachment VolumeAttachment `json:"attachment"`
}

// ListVolumeAttachmentsResponse matches the backend ListVolumeAttachmentsResponse proto message.
type ListVolumeAttachmentsResponse struct {
	Attachments []VolumeAttachment `json:"attachments"`
}
