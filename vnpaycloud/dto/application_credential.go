package dto

import "time"

type Role struct {
	DomainID string `json:"domain_id,omitempty"`
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
}

type AccessRule struct {
	ID      string `json:"id,omitempty"`
	Path    string `json:"path,omitempty"`
	Method  string `json:"method,omitempty"`
	Service string `json:"service,omitempty"`
}

type ApplicationCredential struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Unrestricted bool           `json:"unrestricted"`
	Secret       string         `json:"secret"`
	ProjectID    string         `json:"project_id"`
	Roles        []Role         `json:"roles"`
	ExpiresAt    time.Time      `json:"-"`
	AccessRules  []AccessRule   `json:"access_rules,omitempty"`
	Links        map[string]any `json:"links"`
}

type CreateApplicationCredentialOpts struct {
	Name         string            `json:"name,omitempty" required:"true"`
	Description  string            `json:"description,omitempty"`
	Unrestricted bool              `json:"unrestricted"`
	Secret       string            `json:"secret,omitempty"`
	Roles        []Role            `json:"roles,omitempty"`
	AccessRules  []AccessRule      `json:"access_rules,omitempty"`
	ExpiresAt    *time.Time        `json:"-"`
	ValueSpecs   map[string]string `json:"value_specs,omitempty"`
}

type CreateApplicationCredentialRequest struct {
	ApplicationCredential CreateApplicationCredentialOpts `json:"application_credential"`
}

type CreateApplicationCredentialResponse struct {
	ApplicationCredential ApplicationCredential `json:"application_credential"`
}

type GetApplicationCredentialResponse struct {
	ApplicationCredential ApplicationCredential `json:"application_credential"`
}
