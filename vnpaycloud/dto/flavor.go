package dto

// Flavor matches the iac-proxy-v2 Flavor proto message.
type Flavor struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	VCPUs    int32  `json:"vcpus"`
	RAMMB    int32  `json:"ramMb"`
	DiskGB   int32  `json:"diskGb"`
	IsPublic bool   `json:"isPublic"`
	Zone     string `json:"zone"`
}

// FlavorResponse matches the iac-proxy-v2 FlavorResponse proto message.
type FlavorResponse struct {
	Flavor Flavor `json:"flavor"`
}

// ListFlavorsResponse matches the iac-proxy-v2 ListFlavorsResponse proto message.
type ListFlavorsResponse struct {
	Flavors []Flavor `json:"flavors"`
}
