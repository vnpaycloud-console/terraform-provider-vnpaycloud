package dto

import (
	"regexp"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"time"
)

type Attachment struct {
	AttachedAt   time.Time `json:"-"`
	AttachmentID string    `json:"attachment_id"`
	Device       string    `json:"device"`
	HostName     string    `json:"host_name"`
	ID           string    `json:"id"`
	ServerID     string    `json:"server_id"`
	VolumeID     string    `json:"volume_id"`
}

type Volume struct {
	ID                  string            `json:"id"`
	Status              string            `json:"status"`
	Size                int               `json:"size"`
	AvailabilityZone    string            `json:"availability_zone"`
	CreatedAt           time.Time         `json:"-"`
	UpdatedAt           time.Time         `json:"-"`
	Attachments         []Attachment      `json:"attachments"`
	Name                string            `json:"name"`
	Description         string            `json:"description"`
	VolumeType          string            `json:"volume_type"`
	SnapshotID          string            `json:"snapshot_id"`
	SourceVolID         string            `json:"source_volid"`
	BackupID            *string           `json:"backup_id"`
	Metadata            map[string]string `json:"metadata"`
	UserID              string            `json:"user_id"`
	Bootable            string            `json:"bootable"`
	Encrypted           bool              `json:"encrypted"`
	ReplicationStatus   string            `json:"replication_status"`
	ConsistencyGroupID  string            `json:"consistencygroup_id"`
	Multiattach         bool              `json:"multiattach"`
	VolumeImageMetadata map[string]string `json:"volume_image_metadata"`
	Host                string            `json:"os-vol-host-attr:host"`
	TenantID            string            `json:"os-vol-tenant-attr:tenant_id"`
}

type GetVolumeResponse struct {
	Volume Volume `json:"volume"`
}

type CreateVolumeOpts struct {
	Size               int               `json:"size,omitempty"`
	AvailabilityZone   string            `json:"availability_zone,omitempty"`
	ConsistencyGroupID string            `json:"consistencygroup_id,omitempty"`
	Description        string            `json:"description,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	Name               string            `json:"name,omitempty"`
	SnapshotID         string            `json:"snapshot_id,omitempty"`
	SourceReplica      string            `json:"source_replica,omitempty"`
	SourceVolID        string            `json:"source_volid,omitempty"`
	ImageID            string            `json:"imageRef,omitempty"`
	BackupID           string            `json:"backup_id,omitempty"`
	VolumeType         string            `json:"volume_type,omitempty"`
	SchedulerHints     map[string]any    `json:"OS-SCH-HNT:scheduler_hints,omitempty"`
}

type CreateVolumeRequest struct {
	Volume CreateVolumeOpts `json:"volume"`
}

type CreateVolumeResponse struct {
	Volume Volume `json:"volume"`
}

type UpdateVolumeOpts struct {
	Name        *string           `json:"name,omitempty"`
	Description *string           `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type UpdateVolumeRequest struct {
	Volume UpdateVolumeOpts `json:"volume"`
}

type ExtendVolumeSizeOpts struct {
	NewSize int `json:"new_size" required:"true"`
}

type ExtendVolumeRequest struct {
	ExtendSize ExtendVolumeSizeOpts `json:"os-extend"`
}

type VolumeMigrationPolicy string

const (
	VolumeMigrationPolicyNever    VolumeMigrationPolicy = "never"
	VolumeMigrationPolicyOnDemand VolumeMigrationPolicy = "on-demand"
)

type ChangeVolumeTypeOpts struct {
	NewType         string                `json:"new_type" required:"true"`
	MigrationPolicy VolumeMigrationPolicy `json:"migration_policy,omitempty"`
}

type ChangeTypeRequest struct {
	ChangeType ChangeVolumeTypeOpts `json:"os-retype"`
}

type SchedulerVolumeHintOpts struct {
	DifferentHost        []string
	SameHost             []string
	LocalToInstance      string
	Query                string
	AdditionalProperties map[string]any
}

func (opts SchedulerVolumeHintOpts) ToSchedulerHintsMap() (map[string]any, error) {
	sh := make(map[string]any)

	uuidRegex, _ := regexp.Compile("^[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}$")

	if len(opts.DifferentHost) > 0 {
		for _, diffHost := range opts.DifferentHost {
			if !uuidRegex.MatchString(diffHost) {
				err := client.ErrInvalidInput{}
				err.Argument = "volumes.SchedulerHintOpts.DifferentHost"
				err.Value = opts.DifferentHost
				err.Info = "The hosts must be in UUID format."
				return nil, err
			}
		}
		sh["different_host"] = opts.DifferentHost
	}

	if len(opts.SameHost) > 0 {
		for _, sameHost := range opts.SameHost {
			if !uuidRegex.MatchString(sameHost) {
				err := client.ErrInvalidInput{}
				err.Argument = "volumes.SchedulerHintOpts.SameHost"
				err.Value = opts.SameHost
				err.Info = "The hosts must be in UUID format."
				return nil, err
			}
		}
		sh["same_host"] = opts.SameHost
	}

	if opts.LocalToInstance != "" {
		if !uuidRegex.MatchString(opts.LocalToInstance) {
			err := client.ErrInvalidInput{}
			err.Argument = "volumes.SchedulerHintOpts.LocalToInstance"
			err.Value = opts.LocalToInstance
			err.Info = "The instance must be in UUID format."
			return nil, err
		}
		sh["local_to_instance"] = opts.LocalToInstance
	}

	if opts.Query != "" {
		sh["query"] = opts.Query
	}

	if opts.AdditionalProperties != nil {
		for k, v := range opts.AdditionalProperties {
			sh[k] = v
		}
	}

	if len(sh) == 0 {
		return sh, nil
	}

	return map[string]any{"OS-SCH-HNT:scheduler_hints": sh}, nil
}

type ListVolumeParams struct {
	AllTenants bool              `q:"all_tenants"`
	Metadata   map[string]string `q:"metadata"`
	Name       string            `q:"name"`
	Status     string            `q:"status"`
	TenantID   string            `q:"project_id"`
	Sort       string            `q:"sort"`
	Limit      int               `q:"limit"`
	Offset     int               `q:"offset"`
	Marker     string            `q:"marker"`
	Bootable   *bool             `q:"bootable,omitempty"`
}

type ListVolumeResponse struct {
	Volumes []Volume `json:"volumes"`
}
