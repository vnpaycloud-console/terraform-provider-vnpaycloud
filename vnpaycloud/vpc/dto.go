package vpc

import "time"

type VpcDto struct {
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

type GetVpcDtoResponse struct {
	VPC VpcDto `json:"vpc"`
}

type CreateVpcDto struct {
	VPC struct {
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
		CIDR        string `json:"cidr,omitempty"`
	} `json:"vpc,omitempty"`
}

type CreateVpcDtoResponse struct {
	VPC VpcDto `json:"vpc"`
}
