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
}

// CreateRegistryProjectRequest matches the iac-proxy-v2 CreateRegistryProjectRequest proto message.
// project_id is passed via URL path.
type CreateRegistryProjectRequest struct {
	Name         string `json:"name"`
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

// RobotAccount matches the iac-proxy-v2 RobotAccount proto message.
type RobotAccount struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	RegistryID  string   `json:"registryId"`
	Permissions []string `json:"permissions"`
	ExpiresAt   string   `json:"expiresAt"`
	Enabled     bool     `json:"enabled"`
	CreatedAt   string   `json:"createdAt"`
}

// CreateRobotAccountRequest matches the iac-proxy-v2 CreateRobotAccountRequest proto message.
// project_id and registry_id are passed via URL path.
type CreateRobotAccountRequest struct {
	Name          string   `json:"name"`
	Permissions   []string `json:"permissions"`
	ExpiresInDays int      `json:"expiresInDays,omitempty"`
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
