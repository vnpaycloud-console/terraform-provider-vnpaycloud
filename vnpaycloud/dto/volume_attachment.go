package dto

// VolumeAttachment matches the iac-proxy-v2 VolumeAttachment proto message.
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

// AttachVolumeRequest matches the iac-proxy-v2 AttachVolumeRequest proto message.
// project_id and volume_id are passed via URL path.
type AttachVolumeRequest struct {
	ServerID string `json:"serverId"`
}

// DetachVolumeRequest matches the iac-proxy-v2 DetachVolumeRequest proto message.
// project_id and volume_id are passed via URL path.
type DetachVolumeRequest struct {
	ServerID string `json:"serverId"`
}

// VolumeAttachmentResponse matches the iac-proxy-v2 VolumeAttachmentResponse proto message.
type VolumeAttachmentResponse struct {
	Attachment VolumeAttachment `json:"attachment"`
}

// ListVolumeAttachmentsResponse matches the iac-proxy-v2 ListVolumeAttachmentsResponse proto message.
type ListVolumeAttachmentsResponse struct {
	Attachments []VolumeAttachment `json:"attachments"`
}
