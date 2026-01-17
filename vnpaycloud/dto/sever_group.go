package dto

type ServerGroup struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Policies  []string `json:"policies"`
	Members   []string `json:"members"`
	UserID    string   `json:"user_id"`
	ProjectID string   `json:"project_id"`
	Metadata  map[string]any
	Policy    *string `json:"policy"`
	Rules     *Rules  `json:"rules"`
}

type CreateServerGroupOpts struct {
	Name       string            `json:"name" required:"true"`
	Policies   []string          `json:"policies,omitempty"`
	Policy     string            `json:"policy,omitempty"`
	Rules      *Rules            `json:"rules,omitempty"`
	ValueSpecs map[string]string `json:"value_specs,omitempty"`
}

type CreateServerGroupRequest struct {
	ServerGroup CreateServerGroupOpts `json:"server_group"`
}

type CreateServerGroupResponse struct {
	ServerGroup ServerGroup `json:"server_group"`
}

type GetServerGroupResponse struct {
	ServerGroup ServerGroup `json:"server_group"`
}

type Rules struct {
	MaxServerPerHost int `json:"max_server_per_host"`
}
