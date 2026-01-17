package dto

type FlavorAccessType string

const (
	FlavorPublicAccess  FlavorAccessType = "true"
	FlavorPrivateAccess FlavorAccessType = "false"
	FlavorAllAccess     FlavorAccessType = "None"
)

type Flavor struct {
	ID          string            `json:"id"`
	Disk        int               `json:"disk"`
	RAM         int               `json:"ram"`
	Name        string            `json:"name"`
	RxTxFactor  float64           `json:"rxtx_factor"`
	Swap        int               `json:"-"`
	VCPUs       int               `json:"vcpus"`
	IsPublic    bool              `json:"os-flavor-access:is_public"`
	Ephemeral   int               `json:"OS-FLV-EXT-DATA:ephemeral"`
	Description string            `json:"description"`
	ExtraSpecs  map[string]string `json:"extra_specs"`
}

type GetFlavorResponse struct {
	Flavor Flavor `json:"flavor"`
}

type ListFlavorsResponse struct {
	Flavors []Flavor `json:"flavors"`
}

type ListFlavorParams struct {
	ChangesSince string           `q:"changes-since"`
	MinDisk      int              `q:"minDisk"`
	MinRAM       int              `q:"minRam"`
	SortDir      string           `q:"sort_dir"`
	SortKey      string           `q:"sort_key"`
	Marker       string           `q:"marker"`
	Limit        int              `q:"limit"`
	AccessType   FlavorAccessType `q:"is_public"`
}
