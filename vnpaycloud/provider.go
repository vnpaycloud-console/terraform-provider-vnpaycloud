package vnpaycloud

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/bucket"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/mutexkv"
	"terraform-provider-vnpaycloud/vnpaycloud/floatingip"
	"terraform-provider-vnpaycloud/vnpaycloud/healthmonitor"
	"terraform-provider-vnpaycloud/vnpaycloud/instance"
	"terraform-provider-vnpaycloud/vnpaycloud/internetgateway"
	"terraform-provider-vnpaycloud/vnpaycloud/kubernetescluster"
	"terraform-provider-vnpaycloud/vnpaycloud/listener"
	"terraform-provider-vnpaycloud/vnpaycloud/loadbalancer"
	"terraform-provider-vnpaycloud/vnpaycloud/keypair"
	"terraform-provider-vnpaycloud/vnpaycloud/networkinterface"
	"terraform-provider-vnpaycloud/vnpaycloud/networkinterfaceattachment"
	"terraform-provider-vnpaycloud/vnpaycloud/pool"
	"terraform-provider-vnpaycloud/vnpaycloud/privategateway"
	"terraform-provider-vnpaycloud/vnpaycloud/registryproject"
	"terraform-provider-vnpaycloud/vnpaycloud/robotaccount"
	"terraform-provider-vnpaycloud/vnpaycloud/routetable"
	"terraform-provider-vnpaycloud/vnpaycloud/securitygroup"
	"terraform-provider-vnpaycloud/vnpaycloud/securitygrouprule"
	"terraform-provider-vnpaycloud/vnpaycloud/servergroup"
	"terraform-provider-vnpaycloud/vnpaycloud/snapshot"
	"terraform-provider-vnpaycloud/vnpaycloud/subnet"
	"terraform-provider-vnpaycloud/vnpaycloud/subnetsnat"
	"terraform-provider-vnpaycloud/vnpaycloud/volume"
	"terraform-provider-vnpaycloud/vnpaycloud/volumeattachment"
	"terraform-provider-vnpaycloud/vnpaycloud/vpc"
	"terraform-provider-vnpaycloud/vnpaycloud/vpcpeering"
	"terraform-provider-vnpaycloud/vnpaycloud/workergroup"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a schema.Provider for VNPAY Cloud.
func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"base_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VNPAYCLOUD_BASE_URL", nil),
				Description: "The base URL of the iac-api-gateway.",
			},
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("VNPAYCLOUD_TOKEN", nil),
				Description: "Authentication token (e.g. vtx_pat_xxx).",
			},
			"zone_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VNPAYCLOUD_ZONE_ID", nil),
				Description: "The availability zone ID. The provider resolves the project for your account in this zone.",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"vnpaycloud_server_group":  servergroup.DataSourceServerGroup(),
			"vnpaycloud_server_groups": servergroup.DataSourceServerGroups(),
			"vnpaycloud_vpc":            vpc.DataSourceVpc(),
			"vnpaycloud_vpcs":           vpc.DataSourceVpcs(),
			"vnpaycloud_subnet":         subnet.DataSourceSubnet(),
			"vnpaycloud_subnets":        subnet.DataSourceSubnets(),
			"vnpaycloud_security_group":    securitygroup.DataSourceSecurityGroup(),
			"vnpaycloud_security_groups":   securitygroup.DataSourceSecurityGroups(),
			"vnpaycloud_floating_ip":       floatingip.DataSourceFloatingIP(),
			"vnpaycloud_floating_ips":      floatingip.DataSourceFloatingIPs(),
			"vnpaycloud_network_interface": networkinterface.DataSourceNetworkInterface(),
			"vnpaycloud_volume":            volume.DataSourceVolume(),
			"vnpaycloud_volumes":           volume.DataSourceVolumes(),
			"vnpaycloud_instance":          instance.DataSourceInstance(),
			"vnpaycloud_instances":         instance.DataSourceInstances(),
			"vnpaycloud_keypair":           keypair.DataSourceKeyPair(),
			"vnpaycloud_keypairs":          keypair.DataSourceKeyPairs(),
			"vnpaycloud_snapshot":          snapshot.DataSourceSnapshot(),
			"vnpaycloud_snapshots":         snapshot.DataSourceSnapshots(),
			"vnpaycloud_internet_gateway":  internetgateway.DataSourceInternetGateway(),
			"vnpaycloud_internet_gateways": internetgateway.DataSourceInternetGateways(),
			"vnpaycloud_lb_loadbalancer":   loadbalancer.DataSourceLoadBalancer(),
			"vnpaycloud_lb_loadbalancers":  loadbalancer.DataSourceLoadBalancers(),
			"vnpaycloud_lb_listener":       listener.DataSourceListener(),
			"vnpaycloud_lb_pool":           pool.DataSourcePool(),
			"vnpaycloud_lb_health_monitor": healthmonitor.DataSourceHealthMonitor(),
			"vnpaycloud_registry_project":      registryproject.DataSourceRegistryProject(),
			"vnpaycloud_registry_projects":     registryproject.DataSourceRegistryProjects(),
			"vnpaycloud_registry_robot_account": robotaccount.DataSourceRobotAccount(),
			"vnpaycloud_kubernetes_cluster":     kubernetescluster.DataSourceKubernetesCluster(),
			"vnpaycloud_kubernetes_clusters":    kubernetescluster.DataSourceKubernetesClusters(),
			"vnpaycloud_kubernetes_worker_group":  workergroup.DataSourceWorkerGroup(),
			"vnpaycloud_kubernetes_worker_groups": workergroup.DataSourceWorkerGroups(),
			"vnpaycloud_bucket":                   bucket.DataSourceBucket(),
			"vnpaycloud_buckets":                  bucket.DataSourceBuckets(),
			"vnpaycloud_vpc_peering":              vpcpeering.DataSourceVPCPeering(),
			"vnpaycloud_vpc_peerings":             vpcpeering.DataSourceVPCPeerings(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"vnpaycloud_server_group":        servergroup.ResourceServerGroup(),
			"vnpaycloud_vpc":                 vpc.ResourceVpc(),
			"vnpaycloud_subnet":              subnet.ResourceSubnet(),
			"vnpaycloud_subnet_snat":         subnetsnat.ResourceSubnetSNAT(),
			"vnpaycloud_security_group":      securitygroup.ResourceSecurityGroup(),
			"vnpaycloud_security_group_rule": securitygrouprule.ResourceSecurityGroupRule(),
			"vnpaycloud_floating_ip":         floatingip.ResourceFloatingIP(),
			"vnpaycloud_network_interface":   networkinterface.ResourceNetworkInterface(),
			"vnpaycloud_network_interface_attachment": networkinterfaceattachment.ResourceNetworkInterfaceAttachment(),
			"vnpaycloud_volume":              volume.ResourceVolume(),
			"vnpaycloud_volume_attachment":   volumeattachment.ResourceVolumeAttachment(),
			"vnpaycloud_instance":            instance.ResourceInstance(),
			"vnpaycloud_keypair":             keypair.ResourceKeyPair(),
			"vnpaycloud_snapshot":            snapshot.ResourceSnapshot(),
			"vnpaycloud_internet_gateway":   internetgateway.ResourceInternetGateway(),
			"vnpaycloud_lb_loadbalancer":    loadbalancer.ResourceLoadBalancer(),
			"vnpaycloud_lb_listener":        listener.ResourceListener(),
			"vnpaycloud_lb_pool":            pool.ResourcePool(),
			"vnpaycloud_lb_health_monitor":      healthmonitor.ResourceHealthMonitor(),
			"vnpaycloud_registry_project":       registryproject.ResourceRegistryProject(),
			"vnpaycloud_registry_robot_account": robotaccount.ResourceRobotAccount(),
			"vnpaycloud_kubernetes_cluster":      kubernetescluster.ResourceKubernetesCluster(),
			"vnpaycloud_kubernetes_worker_group": workergroup.ResourceWorkerGroup(),
			"vnpaycloud_route_table":             routetable.ResourceRouteTable(),
			"vnpaycloud_private_gateway":         privategateway.ResourcePrivateGateway(),
			"vnpaycloud_bucket":                  bucket.ResourceBucket(),
			"vnpaycloud_vpc_peering":             vpcpeering.ResourceVPCPeering(),
		},
	}

	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return configureProvider(ctx, d)
	}

	return provider
}

func configureProvider(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	c, err := client.NewClient(ctx, &client.ClientConfig{
		BaseURL: d.Get("base_url").(string),
		Token:   d.Get("token").(string),
	})
	if err != nil {
		return nil, diag.FromErr(err)
	}

	// Resolve project_id from zone_id via backend API
	zoneID := d.Get("zone_id").(string)
	var resolveResp dto.ResolveProjectByZoneResponse
	_, err = c.Get(ctx, client.ApiPath.ResolveProjectByZone(zoneID), &resolveResp, nil)
	if err != nil {
		return nil, diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "Failed to resolve project for zone",
				Detail:   fmt.Sprintf("Could not resolve project_id for zone_id=%s: %s", zoneID, err),
			},
		}
	}

	cfg := &config.Config{
		MutexKV:   mutexkv.NewMutexKV(),
		Client:    c,
		ProjectID: resolveResp.ProjectID,
		ZoneID:    zoneID,
	}

	return cfg, nil
}
