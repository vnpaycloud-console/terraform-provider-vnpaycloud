package dto

import "time"

type Member struct {
	Name               string    `json:"name"`
	Weight             int       `json:"weight"`
	AdminStateUp       bool      `json:"admin_state_up"`
	ProjectID          string    `json:"project_id"`
	SubnetID           string    `json:"subnet_id"`
	PoolID             string    `json:"pool_id"`
	Address            string    `json:"address"`
	ProtocolPort       int       `json:"protocol_port"`
	ID                 string    `json:"id"`
	ProvisioningStatus string    `json:"provisioning_status"`
	CreatedAt          time.Time `json:"-"`
	UpdatedAt          time.Time `json:"-"`
	OperatingStatus    string    `json:"operating_status"`
	Backup             bool      `json:"backup"`
	MonitorAddress     string    `json:"monitor_address"`
	MonitorPort        int       `json:"monitor_port"`
	Tags               []string  `json:"tags"`
}

type GetMemberResponse struct {
	Member Member `json:"member"`
}

type ListMembersResponse struct {
	Members []Member `json:"members"`
}

type CreateMemberRequest struct {
	Member CreateMemberOpts `json:"member"`
}

type CreateMemberOpts struct {
	Address      string `json:"address"`       // IP backend server
	ProtocolPort int    `json:"protocol_port"` // Port backend server
	SubnetID     string `json:"subnet_id,omitempty"`
	Weight       int    `json:"weight,omitempty"`
	Name         string `json:"name,omitempty"`
	PoolID       string `json:"pool_id,omitempty"`
}

type CreateMemberResponse struct {
	Member Member `json:"member"`
}

type UpdatePoolMemberRequest struct {
	Member UpdatePoolMemberOpts `json:"member"`
}

type UpdatePoolMemberOpts struct {
	Name         string `json:"name,omitempty"`
	Weight       int    `json:"weight,omitempty"`
	AdminStateUp bool   `json:"admin_state_up,omitempty"`
	Backup       bool   `json:"backup,omitempty"`
}

type UpdatePoolMemberResponse struct {
	Member Member `json:"member"`
}

type BatchUpdateMemberOpts struct {
	Address        string   `json:"address" required:"true"`
	ProtocolPort   int      `json:"protocol_port" required:"true"`
	Name           *string  `json:"name,omitempty"`
	ProjectID      string   `json:"project_id,omitempty"`
	Weight         *int     `json:"weight,omitempty"`
	SubnetID       *string  `json:"subnet_id,omitempty"`
	AdminStateUp   *bool    `json:"admin_state_up,omitempty"`
	Backup         *bool    `json:"backup,omitempty"`
	MonitorAddress *string  `json:"monitor_address,omitempty"`
	MonitorPort    *int     `json:"monitor_port,omitempty"`
	Tags           []string `json:"tags,omitempty"`
}

type BatchUpdateMemberRequest struct {
	Members []BatchUpdateMemberOpts `json:"members"`
}
