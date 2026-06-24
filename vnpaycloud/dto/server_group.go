package dto

// ServerGroup matches the backend ServerGroup proto message.
type ServerGroup struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Policy    string   `json:"policy"`
	MemberIDs []string `json:"memberIds"`
	CreatedAt string   `json:"createdAt"`
	ProjectID string   `json:"projectId"`
}

// CreateServerGroupRequest matches the backend CreateServerGroupRequest proto message.
// project_id is passed via URL path, not in the body.
type CreateServerGroupRequest struct {
	Name   string `json:"name"`
	Policy string `json:"policy"`
}

// ServerGroupResponse matches the backend ServerGroupResponse proto message.
type ServerGroupResponse struct {
	ServerGroup ServerGroup `json:"serverGroup"`
}

// ListServerGroupsResponse matches the backend ListServerGroupsResponse proto message.
type ListServerGroupsResponse struct {
	ServerGroups []ServerGroup `json:"serverGroups"`
}
