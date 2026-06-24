package dto

// ResolveProjectByZoneResponse matches the backend GetProjectByZoneResponse.
type ResolveProjectByZoneResponse struct {
	ProjectID string `json:"projectId"`
	ZoneID    string `json:"zoneId"`
}
