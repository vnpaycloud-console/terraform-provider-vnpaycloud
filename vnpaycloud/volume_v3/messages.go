package volumeV3

var datasourceDescriptions = map[string]string{
	"datasource":       "Use this data source to get information about an existing volume.",
	"name":             "The name of the volume.",
	"description":      "The description of the volume.",
	"size":             "The size of the volume (in gigabytes).",
	"volume_type":      "The type of the volume.",
	"status":           "The status of the volume.",
	"bootable":         "Indicates if the volume is bootable.",
	"source_volume_id": "The ID of the volume from which the current volume was created.",
}

var resourceDescriptions = map[string]string{
	"resource":         "Manages a volume resource within VNPAY Cloud.",
	"name":             "A unique name for the volume. Changing this updates the volume's name.",
	"description":      "A description of the volume. Changing this updates the volume's description.",
	"size":             "The size of the volume to create (in gigabytes).",
	"volume_type":      "The type of volume to create. Changing this creates a new volume.",
	"status":           "The status of the volume.",
	"bootable":         "Indicates if the volume is bootable.",
	"source_volume_id": "The ID of the volume from which the current volume was created.",
}
