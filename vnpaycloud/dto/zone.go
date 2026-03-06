package dto

// ResolveProjectByZoneResponse matches the iac-proxy-v2 GetProjectByZoneResponse.
type ResolveProjectByZoneResponse struct {
	ProjectID string `json:"projectId"`
	ZoneID    string `json:"zoneId"`
}
