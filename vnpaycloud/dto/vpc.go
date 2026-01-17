package dto

import "time"

type Vpc struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CIDR        string    `json:"cidr"`
	SNATAddress string    `json:"snat_address"`
	EnableSNAT  bool      `json:"enable_snat"`
	Region      string    `json:"region"`
	UpdatedAt   time.Time `json:"-"`
	CreatedAt   time.Time `json:"-"`
	ProjectID   string    `json:"project_id"`
	Status      string    `json:"status"`
}

type GetVpcResponse struct {
	VPC Vpc `json:"vpc"`
}

type CreateVpcRequest struct {
	VPC struct {
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
		CIDR        string `json:"cidr,omitempty"`
	} `json:"vpc,omitempty"`
}

type CreateVpcResponse struct {
	VPC Vpc `json:"vpc"`
}

type ListVpcParams struct {
	Name      string `q:"name"`
	CIDR      string `q:"cidr"`
	ID        string `q:"id"`
	ProjectID string `q:"project_id"`
	Limit     int    `q:"limit"`
	Status    string `q:"status"`
}

type ListVpcResponse struct {
	VPCs []Vpc `json:"vpcs"`
}
