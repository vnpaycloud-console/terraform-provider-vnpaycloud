package vnpaycloud

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/bucket"
	"terraform-provider-vnpaycloud/vnpaycloud/certificate"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/customergateway"
	"terraform-provider-vnpaycloud/vnpaycloud/databaseflavor"
	"terraform-provider-vnpaycloud/vnpaycloud/databasepostgres"
	"terraform-provider-vnpaycloud/vnpaycloud/databasepostgresaccount"
	"terraform-provider-vnpaycloud/vnpaycloud/databasepostgresdatabase"
	"terraform-provider-vnpaycloud/vnpaycloud/databaseredis"
	"terraform-provider-vnpaycloud/vnpaycloud/databaseredisaccount"
	"terraform-provider-vnpaycloud/vnpaycloud/databaseredissentinel"
	"terraform-provider-vnpaycloud/vnpaycloud/databaseredissentinelaccount"
	"terraform-provider-vnpaycloud/vnpaycloud/databaseversion"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/flavor"
	"terraform-provider-vnpaycloud/vnpaycloud/floatingip"
	"terraform-provider-vnpaycloud/vnpaycloud/healthmonitor"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/mutexkv"
	"terraform-provider-vnpaycloud/vnpaycloud/image"
	"terraform-provider-vnpaycloud/vnpaycloud/instance"
	"terraform-provider-vnpaycloud/vnpaycloud/internetgateway"
	"terraform-provider-vnpaycloud/vnpaycloud/keypair"
	"terraform-provider-vnpaycloud/vnpaycloud/kubernetescluster"
	"terraform-provider-vnpaycloud/vnpaycloud/l7policy"
	"terraform-provider-vnpaycloud/vnpaycloud/l7rule"
	"terraform-provider-vnpaycloud/vnpaycloud/lbflavor"
	"terraform-provider-vnpaycloud/vnpaycloud/listener"
	"terraform-provider-vnpaycloud/vnpaycloud/loadbalancer"
	"terraform-provider-vnpaycloud/vnpaycloud/networkacl"
	"terraform-provider-vnpaycloud/vnpaycloud/networkaclrule"
	"terraform-provider-vnpaycloud/vnpaycloud/networkinterface"
	"terraform-provider-vnpaycloud/vnpaycloud/networkinterfaceattachment"
	"terraform-provider-vnpaycloud/vnpaycloud/pool"
	"terraform-provider-vnpaycloud/vnpaycloud/privategateway"
	"terraform-provider-vnpaycloud/vnpaycloud/registrypermission"
	"terraform-provider-vnpaycloud/vnpaycloud/registryproject"
	"terraform-provider-vnpaycloud/vnpaycloud/robotaccount"
	"terraform-provider-vnpaycloud/vnpaycloud/routetable"
	"terraform-provider-vnpaycloud/vnpaycloud/securitygroup"
	"terraform-provider-vnpaycloud/vnpaycloud/securitygrouprule"
	"terraform-provider-vnpaycloud/vnpaycloud/servergroup"
	"terraform-provider-vnpaycloud/vnpaycloud/serviceendpoint"
	"terraform-provider-vnpaycloud/vnpaycloud/servicegateway"
	"terraform-provider-vnpaycloud/vnpaycloud/snapshot"
	"terraform-provider-vnpaycloud/vnpaycloud/subnet"
	"terraform-provider-vnpaycloud/vnpaycloud/subnetsnat"
	"terraform-provider-vnpaycloud/vnpaycloud/volume"
	"terraform-provider-vnpaycloud/vnpaycloud/volumeattachment"
	"terraform-provider-vnpaycloud/vnpaycloud/volumetype"
	"terraform-provider-vnpaycloud/vnpaycloud/vpc"
	"terraform-provider-vnpaycloud/vnpaycloud/vpcpeering"
	"terraform-provider-vnpaycloud/vnpaycloud/vpnconnection"
	"terraform-provider-vnpaycloud/vnpaycloud/vpngateway"
	"terraform-provider-vnpaycloud/vnpaycloud/vpnpublicip"
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
			"vnpaycloud_server_group":                      servergroup.DataSourceServerGroup(),
			"vnpaycloud_server_groups":                     servergroup.DataSourceServerGroups(),
			"vnpaycloud_vpc":                               vpc.DataSourceVpc(),
			"vnpaycloud_vpcs":                              vpc.DataSourceVpcs(),
			"vnpaycloud_subnet":                            subnet.DataSourceSubnet(),
			"vnpaycloud_subnets":                           subnet.DataSourceSubnets(),
			"vnpaycloud_security_group":                    securitygroup.DataSourceSecurityGroup(),
			"vnpaycloud_security_groups":                   securitygroup.DataSourceSecurityGroups(),
			"vnpaycloud_floating_ip":                       floatingip.DataSourceFloatingIP(),
			"vnpaycloud_floating_ips":                      floatingip.DataSourceFloatingIPs(),
			"vnpaycloud_network_interface":                 networkinterface.DataSourceNetworkInterface(),
			"vnpaycloud_volume":                            volume.DataSourceVolume(),
			"vnpaycloud_volumes":                           volume.DataSourceVolumes(),
			"vnpaycloud_instance":                          instance.DataSourceInstance(),
			"vnpaycloud_instances":                         instance.DataSourceInstances(),
			"vnpaycloud_keypair":                           keypair.DataSourceKeyPair(),
			"vnpaycloud_keypairs":                          keypair.DataSourceKeyPairs(),
			"vnpaycloud_snapshot":                          snapshot.DataSourceSnapshot(),
			"vnpaycloud_snapshots":                         snapshot.DataSourceSnapshots(),
			"vnpaycloud_internet_gateway":                  internetgateway.DataSourceInternetGateway(),
			"vnpaycloud_internet_gateways":                 internetgateway.DataSourceInternetGateways(),
			"vnpaycloud_service_gateway":                   servicegateway.DataSourceServiceGateway(),
			"vnpaycloud_service_gateways":                  servicegateway.DataSourceServiceGateways(),
			"vnpaycloud_service_gateway_flavors":           servicegateway.DataSourceServiceGatewayFlavors(),
			"vnpaycloud_service_endpoint":                  serviceendpoint.DataSourceServiceEndpoint(),
			"vnpaycloud_service_endpoints":                 serviceendpoint.DataSourceServiceEndpoints(),
			"vnpaycloud_service_providers":                 serviceendpoint.DataSourceServiceProviders(),
			"vnpaycloud_services":                          serviceendpoint.DataSourceServices(),
			"vnpaycloud_route_table":                       routetable.DataSourceRouteTable(),
			"vnpaycloud_route_tables":                      routetable.DataSourceRouteTables(),
			"vnpaycloud_network_acl":                       networkacl.DataSourceNetworkACL(),
			"vnpaycloud_network_acl_rule":                  networkaclrule.DataSourceNetworkACLRule(),
			"vnpaycloud_lb_loadbalancer":                   loadbalancer.DataSourceLoadBalancer(),
			"vnpaycloud_lb_loadbalancers":                  loadbalancer.DataSourceLoadBalancers(),
			"vnpaycloud_lb_listener":                       listener.DataSourceListener(),
			"vnpaycloud_lb_listeners":                      listener.DataSourceListeners(),
			"vnpaycloud_lb_pool":                           pool.DataSourcePool(),
			"vnpaycloud_lb_pools":                          pool.DataSourcePools(),
			"vnpaycloud_lb_health_monitor":                 healthmonitor.DataSourceHealthMonitor(),
			"vnpaycloud_lb_health_monitors":                healthmonitor.DataSourceHealthMonitors(),
			"vnpaycloud_lb_l7policy":                       l7policy.DataSourceL7Policy(),
			"vnpaycloud_lb_l7policies":                     l7policy.DataSourceL7Policies(),
			"vnpaycloud_lb_l7rule":                         l7rule.DataSourceL7Rule(),
			"vnpaycloud_lb_l7rules":                        l7rule.DataSourceL7Rules(),
			"vnpaycloud_lb_flavors":                        lbflavor.DataSourceLBFlavors(),
			"vnpaycloud_certificates":                      certificate.DataSourceCertificates(),
			"vnpaycloud_registry_project":                  registryproject.DataSourceRegistryProject(),
			"vnpaycloud_registry_projects":                 registryproject.DataSourceRegistryProjects(),
			"vnpaycloud_registry_permissions":              registrypermission.DataSourceRegistryPermissions(),
			"vnpaycloud_registry_robot_account":            robotaccount.DataSourceRobotAccount(),
			"vnpaycloud_kubernetes_cluster":                kubernetescluster.DataSourceKubernetesCluster(),
			"vnpaycloud_kubernetes_clusters":               kubernetescluster.DataSourceKubernetesClusters(),
			"vnpaycloud_kubernetes_kubeconfig":             kubernetescluster.DataSourceKubernetesKubeconfig(),
			"vnpaycloud_kubernetes_worker_group":           workergroup.DataSourceWorkerGroup(),
			"vnpaycloud_kubernetes_worker_groups":          workergroup.DataSourceWorkerGroups(),
			"vnpaycloud_bucket":                            bucket.DataSourceBucket(),
			"vnpaycloud_buckets":                           bucket.DataSourceBuckets(),
			"vnpaycloud_database_postgres_instance":        databasepostgres.DataSourceDatabasePostgresInstance(),
			"vnpaycloud_database_postgres_instances":       databasepostgres.DataSourceDatabasePostgresInstances(),
			"vnpaycloud_database_postgres_account":         databasepostgresaccount.DataSourceDatabasePostgresAccount(),
			"vnpaycloud_database_postgres_accounts":        databasepostgresaccount.DataSourceDatabasePostgresAccounts(),
			"vnpaycloud_database_postgres_database":        databasepostgresdatabase.DataSourceDatabasePostgresDatabase(),
			"vnpaycloud_database_postgres_databases":       databasepostgresdatabase.DataSourceDatabasePostgresDatabases(),
			"vnpaycloud_database_redis_account":            databaseredisaccount.DataSourceDatabaseRedisAccount(),
			"vnpaycloud_database_redis_accounts":           databaseredisaccount.DataSourceDatabaseRedisAccounts(),
			"vnpaycloud_database_redis_sentinel_account":   databaseredissentinelaccount.DataSourceDatabaseRedisSentinelAccount(),
			"vnpaycloud_database_redis_sentinel_accounts":  databaseredissentinelaccount.DataSourceDatabaseRedisSentinelAccounts(),
			"vnpaycloud_database_redis_instance":           databaseredis.DataSourceDatabaseRedisInstance(),
			"vnpaycloud_database_redis_instances":          databaseredis.DataSourceDatabaseRedisInstances(),
			"vnpaycloud_database_redis_sentinel_instance":  databaseredissentinel.DataSourceDatabaseRedisSentinelInstance(),
			"vnpaycloud_database_redis_sentinel_instances": databaseredissentinel.DataSourceDatabaseRedisSentinelInstances(),
			"vnpaycloud_database_flavor":                   databaseflavor.DataSourceDatabaseFlavor(),
			"vnpaycloud_database_flavors":                  databaseflavor.DataSourceDatabaseFlavors(),
			"vnpaycloud_database_postgres_versions":        databaseversion.DataSourceDatabasePostgresVersions(),
			"vnpaycloud_database_redis_versions":           databaseversion.DataSourceDatabaseRedisVersions(),
			"vnpaycloud_vpc_peering":                       vpcpeering.DataSourceVPCPeering(),
			"vnpaycloud_vpc_peerings":                      vpcpeering.DataSourceVPCPeerings(),
			"vnpaycloud_vpn_gateway":                       vpngateway.DataSourceVPNGateway(),
			"vnpaycloud_vpn_gateways":                      vpngateway.DataSourceVPNGateways(),
			"vnpaycloud_vpn_connection":                    vpnconnection.DataSourceVPNConnection(),
			"vnpaycloud_vpn_connections":                   vpnconnection.DataSourceVPNConnections(),
			"vnpaycloud_customer_gateway":                  customergateway.DataSourceCustomerGateway(),
			"vnpaycloud_customer_gateways":                 customergateway.DataSourceCustomerGateways(),
			"vnpaycloud_vpn_public_ip":                     vpnpublicip.DataSourceVPNPublicIP(),
			"vnpaycloud_vpn_public_ips":                    vpnpublicip.DataSourceVPNPublicIPs(),
			"vnpaycloud_flavor":                            flavor.DataSourceFlavor(),
			"vnpaycloud_flavors":                           flavor.DataSourceFlavors(),
			"vnpaycloud_image":                             image.DataSourceImage(),
			"vnpaycloud_images":                            image.DataSourceImages(),
			"vnpaycloud_volume_type":                       volumetype.DataSourceVolumeType(),
			"vnpaycloud_volume_types":                      volumetype.DataSourceVolumeTypes(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"vnpaycloud_server_group":                     servergroup.ResourceServerGroup(),
			"vnpaycloud_vpc":                              vpc.ResourceVpc(),
			"vnpaycloud_subnet":                           subnet.ResourceSubnet(),
			"vnpaycloud_subnet_snat":                      subnetsnat.ResourceSubnetSNAT(),
			"vnpaycloud_security_group":                   securitygroup.ResourceSecurityGroup(),
			"vnpaycloud_security_group_rule":              securitygrouprule.ResourceSecurityGroupRule(),
			"vnpaycloud_floating_ip":                      floatingip.ResourceFloatingIP(),
			"vnpaycloud_network_interface":                networkinterface.ResourceNetworkInterface(),
			"vnpaycloud_network_interface_attachment":     networkinterfaceattachment.ResourceNetworkInterfaceAttachment(),
			"vnpaycloud_volume":                           volume.ResourceVolume(),
			"vnpaycloud_volume_attachment":                volumeattachment.ResourceVolumeAttachment(),
			"vnpaycloud_instance":                         instance.ResourceInstance(),
			"vnpaycloud_keypair":                          keypair.ResourceKeyPair(),
			"vnpaycloud_snapshot":                         snapshot.ResourceSnapshot(),
			"vnpaycloud_internet_gateway":                 internetgateway.ResourceInternetGateway(),
			"vnpaycloud_service_gateway":                  servicegateway.ResourceServiceGateway(),
			"vnpaycloud_service_endpoint":                 serviceendpoint.ResourceServiceEndpoint(),
			"vnpaycloud_lb_loadbalancer":                  loadbalancer.ResourceLoadBalancer(),
			"vnpaycloud_lb_listener":                      listener.ResourceListener(),
			"vnpaycloud_lb_pool":                          pool.ResourcePool(),
			"vnpaycloud_lb_health_monitor":                healthmonitor.ResourceHealthMonitor(),
			"vnpaycloud_lb_l7policy":                      l7policy.ResourceL7Policy(),
			"vnpaycloud_lb_l7rule":                        l7rule.ResourceL7Rule(),
			"vnpaycloud_registry_project":                 registryproject.ResourceRegistryProject(),
			"vnpaycloud_registry_robot_account":           robotaccount.ResourceRobotAccount(),
			"vnpaycloud_kubernetes_cluster":               kubernetescluster.ResourceKubernetesCluster(),
			"vnpaycloud_kubernetes_worker_group":          workergroup.ResourceWorkerGroup(),
			"vnpaycloud_route_table":                      routetable.ResourceRouteTable(),
			"vnpaycloud_network_acl":                      networkacl.ResourceNetworkACL(),
			"vnpaycloud_network_acl_rule":                 networkaclrule.ResourceNetworkACLRule(),
			"vnpaycloud_private_gateway":                  privategateway.ResourcePrivateGateway(),
			"vnpaycloud_vpn_gateway":                      vpngateway.ResourceVPNGateway(),
			"vnpaycloud_vpn_gateway_vpc_attachment":       vpngateway.ResourceVPNGatewayVPCAttachment(),
			"vnpaycloud_vpn_connection":                   vpnconnection.ResourceVPNConnection(),
			"vnpaycloud_customer_gateway":                 customergateway.ResourceCustomerGateway(),
			"vnpaycloud_vpn_public_ip":                    vpnpublicip.ResourceVPNPublicIP(),
			"vnpaycloud_bucket":                           bucket.ResourceBucket(),
			"vnpaycloud_vpc_peering":                      vpcpeering.ResourceVPCPeering(),
			"vnpaycloud_database_postgres_instance":       databasepostgres.ResourceDatabasePostgresInstance(),
			"vnpaycloud_database_postgres_account":        databasepostgresaccount.ResourceDatabasePostgresAccount(),
			"vnpaycloud_database_postgres_database":       databasepostgresdatabase.ResourceDatabasePostgresDatabase(),
			"vnpaycloud_database_redis_account":           databaseredisaccount.ResourceDatabaseRedisAccount(),
			"vnpaycloud_database_redis_sentinel_account":  databaseredissentinelaccount.ResourceDatabaseRedisSentinelAccount(),
			"vnpaycloud_database_redis_instance":          databaseredis.ResourceDatabaseRedisInstance(),
			"vnpaycloud_database_redis_sentinel_instance": databaseredissentinel.ResourceDatabaseRedisSentinelInstance(),
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
