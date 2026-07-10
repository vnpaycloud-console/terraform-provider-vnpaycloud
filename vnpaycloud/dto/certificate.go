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

type CertificateResponse struct {
	Certificate Certificate `json:"certificate"`
}

type CreateSelfSignedCertificateRequest struct {
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	Country          string `json:"country"`
	Province         string `json:"province"`
	City             string `json:"city,omitempty"`
	Organization     string `json:"organization"`
	OrganizationUnit string `json:"organizationUnit"`
	DomainName       string `json:"domainName"`
	Email            string `json:"email,omitempty"`
	KeyType          string `json:"keyType"`
	Expiration       int    `json:"expiration"`
	DigestAlgorithm  string `json:"digestAlgorithm"`
}

type UploadCertificateRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	PrivateKey     string `json:"privateKey"`
	Cert           string `json:"cert"`
	IntermediateCA string `json:"intermediateCa,omitempty"`
}

type UploadCACertificateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Cert        string `json:"cert"`
}
