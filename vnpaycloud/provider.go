package vnpaycloud

import (
	"context"
	applicationcredentials "terraform-provider-vnpaycloud/vnpaycloud/application-credential"
	"terraform-provider-vnpaycloud/vnpaycloud/flavor"
	"terraform-provider-vnpaycloud/vnpaycloud/floatingip"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/mutexkv"
	"terraform-provider-vnpaycloud/vnpaycloud/keypair"
	"terraform-provider-vnpaycloud/vnpaycloud/network"
	lbListener "terraform-provider-vnpaycloud/vnpaycloud/octavia/listener"
	lb "terraform-provider-vnpaycloud/vnpaycloud/octavia/loadbalancer"
	lbMember "terraform-provider-vnpaycloud/vnpaycloud/octavia/member"
	lbMembers "terraform-provider-vnpaycloud/vnpaycloud/octavia/members"
	lbMonitor "terraform-provider-vnpaycloud/vnpaycloud/octavia/monitor"
	lbPool "terraform-provider-vnpaycloud/vnpaycloud/octavia/pool"
	peeringconnection "terraform-provider-vnpaycloud/vnpaycloud/peering_connection"
	"terraform-provider-vnpaycloud/vnpaycloud/port"
	route "terraform-provider-vnpaycloud/vnpaycloud/route"
	securityGroup "terraform-provider-vnpaycloud/vnpaycloud/security-group"
	securityGroupRule "terraform-provider-vnpaycloud/vnpaycloud/security-group-rule"
	"terraform-provider-vnpaycloud/vnpaycloud/server"
	"terraform-provider-vnpaycloud/vnpaycloud/subnet"
	"terraform-provider-vnpaycloud/vnpaycloud/vpc"

	serverGroup "terraform-provider-vnpaycloud/vnpaycloud/server-group"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/volume"
)

// Provider returns a schema.Provider for VNPAY Cloud.
func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_BASE_URL", ""),
				Description: Descriptions["base_url"],
			},
			"application_credential_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_APPLICATION_CREDENTIAL_ID", ""),
				Description: Descriptions["application_credential_id"],
			},
			"application_credential_name": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_APPLICATION_CREDENTIAL_NAME", ""),
				Description: Descriptions["application_credential_name"],
			},
			"application_credential_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_APPLICATION_CREDENTIAL_SECRET", ""),
				Description: Descriptions["application_credential_secret"],
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"vnpaycloud_blockstorage_volume":   volume.DataSourceBlockStorageVolume(),
			"vnpaycloud_networking_vpc":        vpc.DataSourceVpc(),
			"vnpaycloud_compute_server":        server.DataSourceComputeInstance(),
			"vnpaycloud_compute_flavor":        flavor.DataSourceComputeFlavor(),
			"vnpaycloud_compute_keypair":       keypair.DataSourceComputeKeypair(),
			"vnpaycloud_networking_subnet":     subnet.DataSourceNetworkingSubnet(),
			"vnpaycloud_networking_secgroup":   securityGroup.DataSourceNetworkingSecGroup(),
			"vnpaycloud_networking_floatingip": floatingip.DataSourceNetworkingFloatingIP(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"vnpaycloud_blockstorage_volume":             volume.ResourceBlockStorageVolume(),
			"vnpaycloud_networking_vpc":                  vpc.ResourceVpc(),
			"vnpaycloud_networking_peering_connection":   peeringconnection.ResourcePeeringConnection(),
			"vnpaycloud_networking_route":                route.ResourceRoute(),
			"vnpaycloud_compute_server":                  server.ResourceComputeInstance(),
			"vnpaycloud_compute_keypair":                 keypair.ResourceComputeKeypair(),
			"vnpaycloud_compute_server_group":            serverGroup.ResourceComputeServerGroup(),
			"vnpaycloud_identity_application_credential": applicationcredentials.ResourceIdentityApplicationCredentialV3(),
			"vnpaycloud_lb_loadbalancer":                 lb.ResourceLoadBalancer(),
			"vnpaycloud_lb_listener":                     lbListener.ResourceListener(),
			"vnpaycloud_lb_pool":                         lbPool.ResourcePool(),
			"vnpaycloud_lb_member":                       lbMember.ResourceMember(),
			"vnpaycloud_lb_members":                      lbMembers.ResourceMembers(),
			"vnpaycloud_lb_monitor":                      lbMonitor.ResourceMonitor(),
			"vnpaycloud_networking_floatingip":           floatingip.ResourceNetworkingFloatingIP(),
			"vnpaycloud_networking_floatingip_associate": floatingip.ResourceNetworkingFloatingIPAssociate(),
			"vnpaycloud_networking_network":              network.ResourceNetworkingNetwork(),
			"vnpaycloud_networking_port":                 port.ResourceNetworkingPort(),
			"vnpaycloud_networking_secgroup":             securityGroup.ResourceNetworkingSecGroup(),
			"vnpaycloud_networking_secgroup_rule":        securityGroupRule.ResourceNetworkingSecGroupRule(),
			"vnpaycloud_networking_subnet":               subnet.ResourceNetworkingSubnet(),
		},
	}

	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		terraformVersion := provider.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return configureProvider(ctx, d, terraformVersion)
	}

	return provider
}

var Descriptions map[string]string

func init() {
	Descriptions = map[string]string{
		"base_url":                      "The base URL.",
		"application_credential_id":     "Application Credential ID to login with.",
		"application_credential_name":   "Application Credential name to login with.",
		"application_credential_secret": "Application Credential secret to login with.",
	}
}

func configureProvider(ctx context.Context, d *schema.ResourceData, terraformVersion string) (interface{}, diag.Diagnostics) {
	config := config.Config{
		MutexKV: mutexkv.NewMutexKV(),
		ConsoleClientConfig: &client.ClientConfig{
			AppCredID:     d.Get("application_credential_id").(string),
			AppCredSecret: d.Get("application_credential_secret").(string),
			BaseURL:       d.Get("base_url").(string),
		},
	}

	return &config, nil
}
