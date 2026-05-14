package dto

// RegistryProject matches the iac-proxy-v2 RegistryProject proto message.
type RegistryProject struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	IsPublic     bool   `json:"isPublic"`
	StorageLimit string `json:"storageLimit"`
	StorageUsed  int64  `json:"storageUsed,string"`
	RepoCount    int32  `json:"repoCount"`
	Status       string `json:"status"`
	CreatedAt    string `json:"createdAt"`
	Namespace    string `json:"namespace"`
}

// CreateRegistryProjectRequest matches the iac-proxy-v2 CreateRegistryProjectRequest proto message.
// project_id is passed via URL path.
type CreateRegistryProjectRequest struct {
	Name         string `json:"name"`
	IsPublic     bool   `json:"isPublic"`
	StorageLimit string `json:"storageLimit,omitempty"`
}

// UpdateRegistryProjectRequest matches the iac-proxy-v2 UpdateRegistryProjectRequest proto message.
// Editable fields: is_public, storage_limit.
type UpdateRegistryProjectRequest struct {
	IsPublic     bool   `json:"isPublic"`
	StorageLimit string `json:"storageLimit,omitempty"`
}

// RegistryProjectResponse matches the iac-proxy-v2 RegistryProjectResponse proto message.
type RegistryProjectResponse struct {
	Registry RegistryProject `json:"registry"`
}

// ListRegistryProjectsResponse matches the iac-proxy-v2 ListRegistryProjectsResponse proto message.
type ListRegistryProjectsResponse struct {
	Registries []RegistryProject `json:"registries"`
}

// RegistryPermission represents a (resource, action) pair the registry accepts.
type RegistryPermission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// ListRegistryPermissionsResponse matches the iac-proxy-v2 ListRegistryPermissionsResponse proto message.
type ListRegistryPermissionsResponse struct {
	Permissions []RegistryPermission `json:"permissions"`
}

// RobotAccountPermission represents per-project permissions for a robot account.
type RobotAccountPermission struct {
	RegistryID string   `json:"registryId"`
	Actions    []string `json:"actions"`
}

// RobotAccount matches the iac-proxy-v2 RobotAccount proto message.
type RobotAccount struct {
	ID            string                   `json:"id"`
	Name          string                   `json:"name"`
	Username      string                   `json:"username"`
	Description   string                   `json:"description"`
	Permissions   []RobotAccountPermission `json:"permissions"`
	ExpiresAt     string                   `json:"expiresAt"`
	ExpiresInDays int                      `json:"expiresInDays"`
	Enabled       bool                     `json:"enabled"`
	CreatedAt     string                   `json:"createdAt"`
}

// CreateRobotAccountRequest matches the iac-proxy-v2 CreateRobotAccountRequest proto message.
type CreateRobotAccountRequest struct {
	Name          string                   `json:"name"`
	Description   string                   `json:"description,omitempty"`
	Permissions   []RobotAccountPermission `json:"permissions"`
	ExpiresInDays int                      `json:"expiresInDays,omitempty"`
}

// UpdateRobotAccountRequest matches the iac-proxy-v2 UpdateRobotAccountRequest proto message.
// Editable fields: description, expires_in_days, permissions.
type UpdateRobotAccountRequest struct {
	Description   string                   `json:"description"`
	ExpiresInDays int                      `json:"expiresInDays,omitempty"`
	Permissions   []RobotAccountPermission `json:"permissions,omitempty"`
}

// RobotAccountResponse matches the iac-proxy-v2 RobotAccountResponse proto message.
// Secret is only returned on create.
type RobotAccountResponse struct {
	RobotAccount RobotAccount `json:"robotAccount"`
	Secret       string       `json:"secret"`
}

// ListRobotAccountsResponse matches the iac-proxy-v2 ListRobotAccountsResponse proto message.
type ListRobotAccountsResponse struct {
	RobotAccounts []RobotAccount `json:"robotAccounts"`
}
