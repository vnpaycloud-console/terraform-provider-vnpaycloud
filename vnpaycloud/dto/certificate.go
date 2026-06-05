package dto

type Certificate struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	CertType        string   `json:"certType"`
	DomainName      string   `json:"domainName"`
	Description     string   `json:"description"`
	Expiration      string   `json:"expiration"`
	Status          string   `json:"status"`
	ZoneID          string   `json:"zoneId"`
	LoadBalancerIDs []string `json:"loadBalancerIds"`
}

type ListCertificatesResponse struct {
	Certificates []Certificate `json:"certificates"`
}
