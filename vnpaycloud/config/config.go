package config

import (
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"terraform-provider-vnpaycloud/vnpaycloud/helper/mutexkv"
)

type Config struct {
	*mutexkv.MutexKV
	ConsoleClientConfig *client.ClientConfig
}
