package dto

import "time"

type ImageDateFilter string

const (
	FilterGT  ImageDateFilter = "gt"
	FilterGTE ImageDateFilter = "gte"
	FilterLT  ImageDateFilter = "lt"
	FilterLTE ImageDateFilter = "lte"
	FilterNEQ ImageDateFilter = "neq"
	FilterEQ  ImageDateFilter = "eq"
)

type ImageVisibility string

const (
	ImageVisibilityPublic    ImageVisibility = "public"
	ImageVisibilityPrivate   ImageVisibility = "private"
	ImageVisibilityShared    ImageVisibility = "shared"
	ImageVisibilityCommunity ImageVisibility = "community"
)

type ImageMemberStatus string

const (
	ImageMemberStatusAccepted ImageMemberStatus = "accepted"
	ImageMemberStatusPending  ImageMemberStatus = "pending"
	ImageMemberStatusRejected ImageMemberStatus = "rejected"
	ImageMemberStatusAll      ImageMemberStatus = "all"
)

type ImageStatus string

const (
	ImageStatusQueued        ImageStatus = "queued"
	ImageStatusSaving        ImageStatus = "saving"
	ImageStatusActive        ImageStatus = "active"
	ImageStatusKilled        ImageStatus = "killed"
	ImageStatusDeleted       ImageStatus = "deleted"
	ImageStatusPendingDelete ImageStatus = "pending_delete"
	ImageStatusDeactivated   ImageStatus = "deactivated"
	ImageStatusImporting     ImageStatus = "importing"
)

type ListImagesParams struct {
	ID              string            `q:"id"`
	Limit           int               `q:"limit"`
	Marker          string            `q:"marker"`
	Name            string            `q:"name"`
	Visibility      ImageVisibility   `q:"visibility"`
	Hidden          bool              `q:"os_hidden"`
	MemberStatus    ImageMemberStatus `q:"member_status"`
	Owner           string            `q:"owner"`
	Status          ImageStatus       `q:"status"`
	SizeMin         int64             `q:"size_min"`
	SizeMax         int64             `q:"size_max"`
	Sort            string            `q:"sort"`
	SortKey         string            `q:"sort_key"`
	SortDir         string            `q:"sort_dir"`
	Tags            []string          `q:"tag"`
	CreatedAtQuery  *ImageDateQuery
	UpdatedAtQuery  *ImageDateQuery
	ContainerFormat string `q:"container_format"`
	DiskFormat      string `q:"disk_format"`
}

type ImageDateQuery struct {
	Date   time.Time
	Filter ImageDateFilter
}

type Image struct {
	ID                          string            `json:"id"`
	Name                        string            `json:"name"`
	Status                      ImageStatus       `json:"status"`
	Tags                        []string          `json:"tags"`
	ContainerFormat             string            `json:"container_format"`
	DiskFormat                  string            `json:"disk_format"`
	MinDiskGigabytes            int               `json:"min_disk"`
	MinRAMMegabytes             int               `json:"min_ram"`
	Owner                       string            `json:"owner"`
	Protected                   bool              `json:"protected"`
	Visibility                  ImageVisibility   `json:"visibility"`
	Hidden                      bool              `json:"os_hidden"`
	Checksum                    string            `json:"checksum"`
	SizeBytes                   int64             `json:"-"`
	Metadata                    map[string]string `json:"metadata"`
	Properties                  map[string]any
	CreatedAt                   time.Time `json:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at"`
	File                        string    `json:"file"`
	Schema                      string    `json:"schema"`
	VirtualSize                 int64     `json:"virtual_size"`
	OpenStackImageImportMethods []string  `json:"-"`
	OpenStackImageStoreIDs      []string  `json:"-"`
}

type ListImagesResponse struct {
	Images []Image `json:"images"`
}

type GetImageResponse struct {
	Image Image `json:"image"`
}
