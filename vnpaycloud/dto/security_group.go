package dto

import (
	"time"
)

type SecGroup struct {
	// The UUID for the security group.
	ID string

	// Human-readable name for the security group. Might not be unique.
	// Cannot be named "default" as that is automatically created for a tenant.
	Name string

	// The security group description.
	Description string

	// A slice of security group rules that dictate the permitted behaviour for
	// traffic entering and leaving the group.
	Rules []SecGroupRule `json:"security_group_rules"`

	// Indicates if the security group is stateful or stateless.
	Stateful bool `json:"stateful"`

	// TenantID is the project owner of the security group.
	TenantID string `json:"tenant_id"`

	// UpdatedAt and CreatedAt contain ISO-8601 timestamps of when the state of the
	// security group last changed, and when it was created.
	UpdatedAt time.Time `json:"-"`
	CreatedAt time.Time `json:"-"`

	// ProjectID is the project owner of the security group.
	ProjectID string `json:"project_id"`

	// Tags optionally set via extensions/attributestags
	Tags []string `json:"tags"`
}

type CreateSecurityGroupOpts struct {
	// Human-readable name for the Security Group. Does not have to be unique.
	Name string `json:"name" required:"true"`

	// Describes the security group.
	Description string `json:"description,omitempty"`
}

type CreateSecurityGroupRequest struct {
	SecurityGroup CreateSecurityGroupOpts `json:"security_group,omitempty"`
}

type CreateSecurityGroupResponse struct {
	SecurityGroup SecGroup `json:"security_group"`
}

type GetSecurityGroupResponse struct {
	SecurityGroup SecGroup `json:"security_group"`
}

type ListSecurityGroupParams struct {
	ID          string `q:"id"`
	Name        string `q:"name"`
	Description string `q:"description"`
	Stateful    *bool  `q:"stateful"`
	TenantID    string `q:"tenant_id"`
	ProjectID   string `q:"project_id"`
	Limit       int    `q:"limit"`
	Marker      string `q:"marker"`
	SortKey     string `q:"sort_key"`
	SortDir     string `q:"sort_dir"`
	Tags        string `q:"tags"`
	TagsAny     string `q:"tags-any"`
	NotTags     string `q:"not-tags"`
	NotTagsAny  string `q:"not-tags-any"`
}

type ListSecurityGroupResponse struct {
	SecurityGroups []SecGroup `json:"security_groups"`
}
