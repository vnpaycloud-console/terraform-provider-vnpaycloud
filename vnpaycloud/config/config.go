package config

import (
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/mutexkv"
)

type Config struct {
	*mutexkv.MutexKV
	Client    *client.Client
	ProjectID string // Resolved from zone_id at provider init
	ZoneID    string // User-provided zone_id
}
