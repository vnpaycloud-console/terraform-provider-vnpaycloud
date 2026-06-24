package dto

// RegistryProject matches the backend RegistryProject proto message.
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

// CreateRegistryProjectRequest matches the backend CreateRegistryProjectRequest proto message.
// project_id is passed via URL path.
type CreateRegistryProjectRequest struct {
	Name         string `json:"name"`
	IsPublic     bool   `json:"isPublic"`
	StorageLimit string `json:"storageLimit,omitempty"`
}

// UpdateRegistryProjectRequest matches the backend UpdateRegistryProjectRequest proto message.
// Editable fields: is_public, storage_limit.
type UpdateRegistryProjectRequest struct {
	IsPublic     bool   `json:"isPublic"`
	StorageLimit string `json:"storageLimit,omitempty"`
}

// RegistryProjectResponse matches the backend RegistryProjectResponse proto message.
type RegistryProjectResponse struct {
	Registry RegistryProject `json:"registry"`
}

// ListRegistryProjectsResponse matches the backend ListRegistryProjectsResponse proto message.
type ListRegistryProjectsResponse struct {
	Registries []RegistryProject `json:"registries"`
}

// RegistryPermission represents a (resource, action) pair the registry accepts.
type RegistryPermission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// ListRegistryPermissionsResponse matches the backend ListRegistryPermissionsResponse proto message.
type ListRegistryPermissionsResponse struct {
	Permissions []RegistryPermission `json:"permissions"`
}

// RobotAccountPermission represents per-project permissions for a robot account.
type RobotAccountPermission struct {
	RegistryID string   `json:"registryId"`
	Actions    []string `json:"actions"`
}

// RobotAccount matches the backend RobotAccount proto message.
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

// CreateRobotAccountRequest matches the backend CreateRobotAccountRequest proto message.
type CreateRobotAccountRequest struct {
	Name          string                   `json:"name"`
	Description   string                   `json:"description,omitempty"`
	Permissions   []RobotAccountPermission `json:"permissions"`
	ExpiresInDays int                      `json:"expiresInDays,omitempty"`
}

// UpdateRobotAccountRequest matches the backend UpdateRobotAccountRequest proto message.
// Editable fields: description, expires_in_days, permissions.
type UpdateRobotAccountRequest struct {
	Description   string                   `json:"description"`
	ExpiresInDays int                      `json:"expiresInDays,omitempty"`
	Permissions   []RobotAccountPermission `json:"permissions,omitempty"`
}

// RobotAccountResponse matches the backend RobotAccountResponse proto message.
// Secret is only returned on create.
type RobotAccountResponse struct {
	RobotAccount RobotAccount `json:"robotAccount"`
	Secret       string       `json:"secret"`
}

// ListRobotAccountsResponse matches the backend ListRobotAccountsResponse proto message.
type ListRobotAccountsResponse struct {
	RobotAccounts []RobotAccount `json:"robotAccounts"`
}
