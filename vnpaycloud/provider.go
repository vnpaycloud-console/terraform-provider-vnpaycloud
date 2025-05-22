package vnpaycloud

import (
	"context"
	"os"
	"runtime/debug"
	applicationcredentials "terraform-provider-vnpaycloud/vnpaycloud/application-credential"
	"terraform-provider-vnpaycloud/vnpaycloud/flavor"
	"terraform-provider-vnpaycloud/vnpaycloud/floatingip"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
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
	routetable "terraform-provider-vnpaycloud/vnpaycloud/route_table"
	securityGroup "terraform-provider-vnpaycloud/vnpaycloud/security-group"
	securityGroupRule "terraform-provider-vnpaycloud/vnpaycloud/security-group-rule"
	"terraform-provider-vnpaycloud/vnpaycloud/server"
	"terraform-provider-vnpaycloud/vnpaycloud/subnet"
	"terraform-provider-vnpaycloud/vnpaycloud/vpc"

	serverGroup "terraform-provider-vnpaycloud/vnpaycloud/server-group"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/volume"

	"github.com/vnpaycloud-console/gophercloud-utils/v2/terraform/auth"
	"github.com/vnpaycloud-console/gophercloud-utils/v2/terraform/mutexkv"
	"github.com/vnpaycloud-console/gophercloud/v2"
)

var version = "dev"

// Use vnpaycloudbase.Config as the base/foundation of this provider's
// Config struct.

// Provider returns a schema.Provider for VNPAY Cloud.
func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"auth_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_AUTH_URL", ""),
				Description: Descriptions["auth_url"],
			},

			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_BASE_URL", ""),
				Description: Descriptions["base_url"],
			},

			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: Descriptions["region"],
				DefaultFunc: schema.EnvDefaultFunc("OS_REGION_NAME", ""),
			},

			"user_name": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_USERNAME", ""),
				Description: Descriptions["user_name"],
			},

			"user_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_USER_ID", ""),
				Description: Descriptions["user_id"],
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

			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"OS_TENANT_ID",
					"OS_PROJECT_ID",
				}, ""),
				Description: Descriptions["tenant_id"],
			},

			"tenant_name": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"OS_TENANT_NAME",
					"OS_PROJECT_NAME",
				}, ""),
				Description: Descriptions["tenant_name"],
			},

			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("OS_PASSWORD", ""),
				Description: Descriptions["password"],
			},

			"token": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"OS_TOKEN",
					"OS_AUTH_TOKEN",
				}, ""),
				Description: Descriptions["token"],
			},

			"user_domain_name": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_USER_DOMAIN_NAME", ""),
				Description: Descriptions["user_domain_name"],
			},

			"user_domain_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_USER_DOMAIN_ID", ""),
				Description: Descriptions["user_domain_id"],
			},

			"project_domain_name": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_PROJECT_DOMAIN_NAME", ""),
				Description: Descriptions["project_domain_name"],
			},

			"project_domain_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_PROJECT_DOMAIN_ID", ""),
				Description: Descriptions["project_domain_id"],
			},

			"domain_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_DOMAIN_ID", ""),
				Description: Descriptions["domain_id"],
			},

			"domain_name": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_DOMAIN_NAME", ""),
				Description: Descriptions["domain_name"],
			},

			"default_domain": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_DEFAULT_DOMAIN", "default"),
				Description: Descriptions["default_domain"],
			},

			"system_scope": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_SYSTEM_SCOPE", false),
				Description: Descriptions["system_scope"],
			},

			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_INSECURE", nil),
				Description: Descriptions["insecure"],
			},

			"endpoint_type": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_ENDPOINT_TYPE", ""),
			},

			"cacert_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_CACERT", ""),
				Description: Descriptions["cacert_file"],
			},

			"cert": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_CERT", ""),
				Description: Descriptions["cert"],
			},

			"key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_KEY", ""),
				Description: Descriptions["key"],
			},

			"swauth": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_SWAUTH", false),
				Description: Descriptions["swauth"],
			},

			"delayed_auth": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_DELAYED_AUTH", true),
				Description: Descriptions["delayed_auth"],
			},

			"allow_reauth": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_ALLOW_REAUTH", true),
				Description: Descriptions["allow_reauth"],
			},

			"cloud": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_CLOUD", ""),
				Description: Descriptions["cloud"],
			},

			"max_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: Descriptions["max_retries"],
			},

			"endpoint_overrides": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: Descriptions["endpoint_overrides"],
			},

			"disable_no_cache_header": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: Descriptions["disable_no_cache_header"],
			},

			"enable_logging": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: Descriptions["enable_logging"],
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			// "vnpaycloud_blockstorage_availability_zones_v3":       dataSourceBlockStorageAvailabilityZonesV3(),
			// "vnpaycloud_blockstorage_snapshot_v3":                 dataSourceBlockStorageSnapshotV3(),
			"vnpaycloud_blockstorage_volume": volume.DataSourceBlockStorageVolume(),
			"vnpaycloud_vpc":                 vpc.DataSourceVpc(),
			// "vnpaycloud_blockstorage_quotaset_v3":                 dataSourceBlockStorageQuotasetV3(),
			// "vnpaycloud_compute_aggregate_v2":                     dataSourceComputeAggregateV2(),
			// "vnpaycloud_compute_availability_zones_v2":            dataSourceComputeAvailabilityZonesV2(),
			"vnpaycloud_server": server.DataSourceComputeInstanceV2(),
			"vnpaycloud_flavor": flavor.DataSourceComputeFlavorV2(),
			// "vnpaycloud_compute_hypervisor_v2":                    dataSourceComputeHypervisorV2(),
			"vnpaycloud_keypair": keypair.DataSourceComputeKeypairV2(),
			// "vnpaycloud_compute_quotaset_v2":                      dataSourceComputeQuotasetV2(),
			// "vnpaycloud_compute_limits_v2":                        dataSourceComputeLimitsV2(),
			// "vnpaycloud_containerinfra_nodegroup_v1":              dataSourceContainerInfraNodeGroupV1(),
			// "vnpaycloud_containerinfra_clustertemplate_v1":        dataSourceContainerInfraClusterTemplateV1(),
			// "vnpaycloud_containerinfra_cluster_v1":                dataSourceContainerInfraCluster(),
			// "vnpaycloud_dns_zone_v2":                              dataSourceDNSZoneV2(),
			// "vnpaycloud_fw_group_v2":                              dataSourceFWGroupV2(),
			// "vnpaycloud_fw_policy_v2":                             dataSourceFWPolicyV2(),
			// "vnpaycloud_fw_rule_v2":                               dataSourceFWRuleV2(),
			// "vnpaycloud_identity_role_v3":                         dataSourceIdentityRoleV3(),
			// "vnpaycloud_identity_project_v3":                      dataSourceIdentityProjectV3(),
			// "vnpaycloud_identity_project_ids_v3":                  dataSourceIdentityProjectIdsV3(),
			// "vnpaycloud_identity_user_v3":                         dataSourceIdentityUserV3(),
			// "vnpaycloud_identity_auth_scope_v3":                   dataSourceIdentityAuthScopeV3(),
			// "vnpaycloud_identity_endpoint_v3":                     dataSourceIdentityEndpointV3(),
			// "vnpaycloud_identity_service_v3":                      dataSourceIdentityServiceV3(),
			// "vnpaycloud_identity_group_v3":                        dataSourceIdentityGroupV3(),
			// "vnpaycloud_images_image_v2":                          dataSourceImagesImageV2(),
			// "vnpaycloud_images_image_ids_v2":                      dataSourceImagesImageIDsV2(),
			// "vnpaycloud_networking_addressscope_v2":               dataSourceNetworkingAddressScopeV2(),
			// "vnpaycloud_networking_network_v2":                    dataSourceNetworkingNetworkV2(),
			// "vnpaycloud_networking_qos_bandwidth_limit_rule_v2":   dataSourceNetworkingQoSBandwidthLimitRuleV2(),
			// "vnpaycloud_networking_qos_dscp_marking_rule_v2":      dataSourceNetworkingQoSDSCPMarkingRuleV2(),
			// "vnpaycloud_networking_qos_minimum_bandwidth_rule_v2": dataSourceNetworkingQoSMinimumBandwidthRuleV2(),
			// "vnpaycloud_networking_qos_policy_v2":                 dataSourceNetworkingQoSPolicyV2(),
			// "vnpaycloud_networking_quota_v2":                      dataSourceNetworkingQuotaV2(),
			"vnpaycloud_networking_subnet": subnet.DataSourceNetworkingSubnetV2(),
			// "vnpaycloud_networking_subnet_ids_v2":                 dataSourceNetworkingSubnetIDsV2(),
			"vnpaycloud_networking_secgroup": securityGroup.DataSourceNetworkingSecGroupV2(),
			// "vnpaycloud_networking_subnetpool_v2":                 dataSourceNetworkingSubnetPoolV2(),
			"vnpaycloud_networking_floatingip": floatingip.DataSourceNetworkingFloatingIPV2(),
			// "vnpaycloud_networking_router_v2":                     dataSourceNetworkingRouterV2(),
			// "vnpaycloud_networking_port_v2":                       dataSourceNetworkingPortV2(),
			// "vnpaycloud_networking_port_ids_v2":                   dataSourceNetworkingPortIDsV2(),
			// "vnpaycloud_networking_trunk_v2":                      dataSourceNetworkingTrunkV2(),
			// "vnpaycloud_sharedfilesystem_availability_zones_v2":   dataSourceSharedFilesystemAvailabilityZonesV2(),
			// "vnpaycloud_sharedfilesystem_sharenetwork_v2":         dataSourceSharedFilesystemShareNetworkV2(),
			// "vnpaycloud_sharedfilesystem_share_v2":                dataSourceSharedFilesystemShareV2(),
			// "vnpaycloud_sharedfilesystem_snapshot_v2":             dataSourceSharedFilesystemSnapshotV2(),
			// "vnpaycloud_keymanager_secret_v1":                     dataSourceKeyManagerSecretV1(),
			// "vnpaycloud_keymanager_container_v1":                  dataSourceKeyManagerContainerV1(),
			// "vnpaycloud_loadbalancer_flavor_v2":                   dataSourceLBFlavorV2(),
			// "vnpaycloud_workflow_workflow_v2":                     dataSourceWorkflowWorkflowV2(),
		},

		ResourcesMap: map[string]*schema.Resource{
			// "vnpaycloud_blockstorage_qos_association_v3":          resourceBlockStorageQosAssociationV3(),
			// "vnpaycloud_blockstorage_qos_v3":                      resourceBlockStorageQosV3(),
			// "vnpaycloud_blockstorage_quotaset_v3":                 resourceBlockStorageQuotasetV3(),
			"vnpaycloud_blockstorage_volume": volume.ResourceBlockStorageVolume(),
			"vnpaycloud_vpc":                 vpc.ResourceVpc(),
			"vnpaycloud_peering_connection":  peeringconnection.ResourcePeeringConnection(),
			"vnpaycloud_route_table":         routetable.ResourceRouteTable(),
			// "vnpaycloud_blockstorage_volume_attach_v3":            resourceBlockStorageVolumeAttachV3(),
			// "vnpaycloud_blockstorage_volume_type_access_v3":       resourceBlockstorageVolumeTypeAccessV3(),
			// "vnpaycloud_blockstorage_volume_type_v3":              resourceBlockStorageVolumeTypeV3(),
			// "vnpaycloud_compute_aggregate_v2":                     resourceComputeAggregateV2(),
			// "vnpaycloud_compute_flavor_v2":                        resourceComputeFlavorV2(),
			// "vnpaycloud_compute_flavor_access_v2":                 resourceComputeFlavorAccessV2(),
			"vnpaycloud_server": server.ResourceComputeInstanceV2(),
			// "vnpaycloud_compute_interface_attach_v2":              resourceComputeInterfaceAttachV2(),
			"vnpaycloud_keypair":      keypair.ResourceComputeKeypairV2(),
			"vnpaycloud_server_group": serverGroup.ResourceComputeServerGroupV2(),
			// "vnpaycloud_compute_quotaset_v2":                      resourceComputeQuotasetV2(),
			// "vnpaycloud_compute_volume_attach_v2":                 resourceComputeVolumeAttachV2(),
			// "vnpaycloud_containerinfra_nodegroup_v1":              resourceContainerInfraNodeGroupV1(),
			// "vnpaycloud_containerinfra_clustertemplate_v1":        resourceContainerInfraClusterTemplateV1(),
			// "vnpaycloud_containerinfra_cluster_v1":                resourceContainerInfraClusterV1(),
			// "vnpaycloud_db_instance_v1":                           resourceDatabaseInstanceV1(),
			// "vnpaycloud_db_user_v1":                               resourceDatabaseUserV1(),
			// "vnpaycloud_db_configuration_v1":                      resourceDatabaseConfigurationV1(),
			// "vnpaycloud_db_database_v1":                           resourceDatabaseDatabaseV1(),
			// "vnpaycloud_dns_recordset_v2":                         resourceDNSRecordSetV2(),
			// "vnpaycloud_dns_zone_v2":                              resourceDNSZoneV2(),
			// "vnpaycloud_dns_transfer_request_v2":                  resourceDNSTransferRequestV2(),
			// "vnpaycloud_dns_transfer_accept_v2":                   resourceDNSTransferAcceptV2(),
			// "vnpaycloud_fw_group_v2":                              resourceFWGroupV2(),
			// "vnpaycloud_fw_policy_v2":                             resourceFWPolicyV2(),
			// "vnpaycloud_fw_rule_v2":                               resourceFWRuleV2(),
			// "vnpaycloud_identity_endpoint_v3":                     resourceIdentityEndpointV3(),
			// "vnpaycloud_identity_project_v3":                      resourceIdentityProjectV3(),
			// "vnpaycloud_identity_role_v3":                         resourceIdentityRoleV3(),
			// "vnpaycloud_identity_role_assignment_v3":              resourceIdentityRoleAssignmentV3(),
			// "vnpaycloud_identity_inherit_role_assignment_v3":      resourceIdentityInheritRoleAssignmentV3(),
			// "vnpaycloud_identity_service_v3":                      resourceIdentityServiceV3(),
			// "vnpaycloud_identity_user_v3":                         resourceIdentityUserV3(),
			// "vnpaycloud_identity_user_membership_v3":              resourceIdentityUserMembershipV3(),
			// "vnpaycloud_identity_group_v3":                        resourceIdentityGroupV3(),
			"vnpaycloud_identity_application_credential": applicationcredentials.ResourceIdentityApplicationCredentialV3(),
			// "vnpaycloud_identity_ec2_credential_v3":               resourceIdentityEc2CredentialV3(),
			// "vnpaycloud_images_image_v2":                          resourceImagesImageV2(),
			// "vnpaycloud_images_image_access_v2":                   resourceImagesImageAccessV2(),
			// "vnpaycloud_images_image_access_accept_v2":            resourceImagesImageAccessAcceptV2(),
			// "vnpaycloud_lb_flavorprofile_v2":                      resourceLoadBalancerFlavorProfileV2(),
			"vnpaycloud_lb_loadbalancer": lb.ResourceLoadBalancerV2(),
			"vnpaycloud_lb_listener":     lbListener.ResourceListenerV2(),
			"vnpaycloud_lb_pool":         lbPool.ResourcePoolV2(),
			"vnpaycloud_lb_member":       lbMember.ResourceMemberV2(),
			"vnpaycloud_lb_members":      lbMembers.ResourceMembersV2(),
			"vnpaycloud_lb_monitor":      lbMonitor.ResourceMonitorV2(),
			// "vnpaycloud_lb_l7policy_v2":                           resourceL7PolicyV2(),
			// "vnpaycloud_lb_l7rule_v2":                             resourceL7RuleV2(),
			// "vnpaycloud_lb_quota_v2":                              resourceLoadBalancerQuotaV2(),
			"vnpaycloud_networking_floatingip":           floatingip.ResourceNetworkingFloatingIPV2(),
			"vnpaycloud_networking_floatingip_associate": floatingip.ResourceNetworkingFloatingIPAssociateV2(),
			"vnpaycloud_networking_network":              network.ResourceNetworkingNetwork(),
			"vnpaycloud_port":                            port.ResourceNetworkingPortV2(),
			// "vnpaycloud_networking_rbac_policy_v2":                resourceNetworkingRBACPolicyV2(),
			// "vnpaycloud_networking_port_secgroup_associate_v2":    resourceNetworkingPortSecGroupAssociateV2(),
			// "vnpaycloud_networking_qos_bandwidth_limit_rule_v2":   resourceNetworkingQoSBandwidthLimitRuleV2(),
			// "vnpaycloud_networking_qos_dscp_marking_rule_v2":      resourceNetworkingQoSDSCPMarkingRuleV2(),
			// "vnpaycloud_networking_qos_minimum_bandwidth_rule_v2": resourceNetworkingQoSMinimumBandwidthRuleV2(),
			// "vnpaycloud_networking_qos_policy_v2":                 resourceNetworkingQoSPolicyV2(),
			// "vnpaycloud_networking_quota_v2":                      resourceNetworkingQuotaV2(),
			// "vnpaycloud_networking_router_v2":                     resourceNetworkingRouterV2(),
			// "vnpaycloud_networking_router_interface_v2":           resourceNetworkingRouterInterfaceV2(),
			// "vnpaycloud_networking_router_route_v2":               resourceNetworkingRouterRouteV2(),
			"vnpaycloud_networking_secgroup":      securityGroup.ResourceNetworkingSecGroupV2(),
			"vnpaycloud_networking_secgroup_rule": securityGroupRule.ResourceNetworkingSecGroupRuleV2(),
			"vnpaycloud_networking_subnet":        subnet.ResourceNetworkingSubnetV2(),
			// "vnpaycloud_networking_subnet_route_v2":               resourceNetworkingSubnetRouteV2(),
			// "vnpaycloud_networking_subnetpool_v2":                 resourceNetworkingSubnetPoolV2(),
			// "vnpaycloud_networking_addressscope_v2":               resourceNetworkingAddressScopeV2(),
			// "vnpaycloud_networking_trunk_v2":                      resourceNetworkingTrunkV2(),
			// "vnpaycloud_networking_portforwarding_v2":             resourceNetworkingPortForwardingV2(),
			// "vnpaycloud_objectstorage_account_v1":                 resourceObjectStorageAccountV1(),
			// "vnpaycloud_objectstorage_container_v1":               resourceObjectStorageContainerV1(),
			// "vnpaycloud_objectstorage_object_v1":                  resourceObjectStorageObjectV1(),
			// "vnpaycloud_objectstorage_tempurl_v1":                 resourceObjectstorageTempurlV1(),
			// "vnpaycloud_orchestration_stack_v1":                   resourceOrchestrationStackV1(),
			// "vnpaycloud_vpnaas_ipsec_policy_v2":                   resourceIPSecPolicyV2(),
			// "vnpaycloud_vpnaas_service_v2":                        resourceServiceV2(),
			// "vnpaycloud_vpnaas_ike_policy_v2":                     resourceIKEPolicyV2(),
			// "vnpaycloud_vpnaas_endpoint_group_v2":                 resourceEndpointGroupV2(),
			// "vnpaycloud_vpnaas_site_connection_v2":                resourceSiteConnectionV2(),
			// "vnpaycloud_sharedfilesystem_securityservice_v2":      resourceSharedFilesystemSecurityServiceV2(),
			// "vnpaycloud_sharedfilesystem_sharenetwork_v2":         resourceSharedFilesystemShareNetworkV2(),
			// "vnpaycloud_sharedfilesystem_share_v2":                resourceSharedFilesystemShareV2(),
			// "vnpaycloud_sharedfilesystem_share_access_v2":         resourceSharedFilesystemShareAccessV2(),
			// "vnpaycloud_keymanager_secret_v1":                     resourceKeyManagerSecretV1(),
			// "vnpaycloud_keymanager_container_v1":                  resourceKeyManagerContainerV1(),
			// "vnpaycloud_keymanager_order_v1":                      resourceKeyManagerOrderV1(),
			// "vnpaycloud_bgpvpn_v2":                                resourceBGPVPNV2(),
			// "vnpaycloud_bgpvpn_network_associate_v2":              resourceBGPVPNNetworkAssociateV2(),
			// "vnpaycloud_bgpvpn_router_associate_v2":               resourceBGPVPNRouterAssociateV2(),
			// "vnpaycloud_bgpvpn_port_associate_v2":                 resourceBGPVPNPortAssociateV2(),
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
		"auth_url": "The Identity authentication URL.",

		"base_url": "The base URL.",

		"cloud": "An entry in a `clouds.yaml` file to use.",

		"region": "The VNPAY Cloud region to connect to.",

		"user_name": "Username to login with.",

		"user_id": "User ID to login with.",

		"application_credential_id": "Application Credential ID to login with.",

		"application_credential_name": "Application Credential name to login with.",

		"application_credential_secret": "Application Credential secret to login with.",

		"tenant_id": "The ID of the Tenant (Identity v2) or Project (Identity v3)\n" +
			"to login with.",

		"tenant_name": "The name of the Tenant (Identity v2) or Project (Identity v3)\n" +
			"to login with.",

		"password": "Password to login with.",

		"token": "Authentication token to use as an alternative to username/password.",

		"user_domain_name": "The name of the domain where the user resides (Identity v3).",

		"user_domain_id": "The ID of the domain where the user resides (Identity v3).",

		"project_domain_name": "The name of the domain where the project resides (Identity v3).",

		"project_domain_id": "The ID of the domain where the proejct resides (Identity v3).",

		"domain_id": "The ID of the Domain to scope to (Identity v3).",

		"domain_name": "The name of the Domain to scope to (Identity v3).",

		"default_domain": "The name of the Domain ID to scope to if no other domain is specified. Defaults to `default` (Identity v3).",

		"system_scope": "If set to `true`, system scoped authorization will be enabled. Defaults to `false` (Identity v3).",

		"insecure": "Trust self-signed certificates.",

		"cacert_file": "A Custom CA certificate.",

		"cert": "A client certificate to authenticate with.",

		"key": "A client private key to authenticate with.",

		"endpoint_type": "The catalog endpoint type to use.",

		"endpoint_overrides": "A map of services with an endpoint to override what was\n" +
			"from the Keystone catalog",

		"swauth": "Use Swift's authentication system instead of Keystone. Only used for\n" +
			"interaction with Swift.",

		"disable_no_cache_header": "If set to `true`, the HTTP `Cache-Control: no-cache` header will not be added by default to all API requests.",

		"delayed_auth": "If set to `false`, VNPAY Cloud authorization will be perfomed,\n" +
			"every time the service provider client is called. Defaults to `true`.",

		"allow_reauth": "If set to `false`, VNPAY Cloud authorization won't be perfomed\n" +
			"automatically, if the initial auth token get expired. Defaults to `true`",

		"max_retries": "How many times HTTP connection should be retried until giving up.",

		"enable_logging": "Outputs very verbose logs with all calls made to and responses from VNPAY Cloud",
	}
}

func getSDKVersion() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	for _, v := range buildInfo.Deps {
		if v.Path == "github.com/hashicorp/terraform-plugin-sdk/v2" {
			return v.Version
		}
	}

	return ""
}

func configureProvider(ctx context.Context, d *schema.ResourceData, terraformVersion string) (interface{}, diag.Diagnostics) {
	enableLogging := d.Get("enable_logging").(bool)
	if !enableLogging {
		// enforce logging (similar to OS_DEBUG) when TF_LOG is 'DEBUG' or 'TRACE'
		if logLevel := logging.LogLevel(); logLevel != "" && os.Getenv("OS_DEBUG") == "" {
			if logLevel == "DEBUG" || logLevel == "TRACE" {
				enableLogging = true
			}
		}
	}

	authOpts := &gophercloud.AuthOptions{
		Scope: &gophercloud.AuthScope{System: d.Get("system_scope").(bool)},
	}

	config := config.Config{
		Config: auth.Config{
			CACertFile:                  d.Get("cacert_file").(string),
			ClientCertFile:              d.Get("cert").(string),
			ClientKeyFile:               d.Get("key").(string),
			Cloud:                       d.Get("cloud").(string),
			DefaultDomain:               d.Get("default_domain").(string),
			DomainID:                    d.Get("domain_id").(string),
			DomainName:                  d.Get("domain_name").(string),
			EndpointOverrides:           d.Get("endpoint_overrides").(map[string]interface{}),
			EndpointType:                d.Get("endpoint_type").(string),
			IdentityEndpoint:            d.Get("auth_url").(string),
			Password:                    d.Get("password").(string),
			ProjectDomainID:             d.Get("project_domain_id").(string),
			ProjectDomainName:           d.Get("project_domain_name").(string),
			Region:                      d.Get("region").(string),
			Swauth:                      d.Get("swauth").(bool),
			Token:                       d.Get("token").(string),
			TenantID:                    d.Get("tenant_id").(string),
			TenantName:                  d.Get("tenant_name").(string),
			UserDomainID:                d.Get("user_domain_id").(string),
			UserDomainName:              d.Get("user_domain_name").(string),
			Username:                    d.Get("user_name").(string),
			UserID:                      d.Get("user_id").(string),
			UseOctavia:                  true,
			ApplicationCredentialID:     d.Get("application_credential_id").(string),
			ApplicationCredentialName:   d.Get("application_credential_name").(string),
			ApplicationCredentialSecret: d.Get("application_credential_secret").(string),
			DelayedAuth:                 d.Get("delayed_auth").(bool),
			AllowReauth:                 d.Get("allow_reauth").(bool),
			AuthOpts:                    authOpts,
			MaxRetries:                  d.Get("max_retries").(int),
			DisableNoCacheHeader:        d.Get("disable_no_cache_header").(bool),
			TerraformVersion:            terraformVersion,
			SDKVersion:                  getSDKVersion() + " Terraform Provider VNPAY Cloud/" + version,
			MutexKV:                     mutexkv.NewMutexKV(),
			EnableLogger:                enableLogging,
		},
		ConsoleClientConfig: &client.ClientConfig{
			AppCredID:     d.Get("application_credential_id").(string),
			AppCredSecret: d.Get("application_credential_secret").(string),
			BaseURL:       d.Get("base_url").(string),
		},
	}

	v, ok := util.GetOkExists(d, "insecure")
	if ok {
		insecure := v.(bool)
		config.Insecure = &insecure
	}

	if err := config.LoadAndValidate(ctx); err != nil {
		return nil, diag.FromErr(err)
	}

	return &config, nil
}
