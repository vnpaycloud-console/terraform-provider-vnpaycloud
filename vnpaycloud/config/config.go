package config

import (
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/vnpaycloud-console/gophercloud-utils/v2/terraform/auth"
)

type Config struct {
	auth.Config
	ConsoleClientConfig *client.ClientConfig
}
