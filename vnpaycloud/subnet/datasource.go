package subnet

import (
	"context"
	"log"
	"strings"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/shared"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/subnets"
)

var descriptions map[string]string

func DataSourceNetworkingSubnetV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworkingSubnetV2Read,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"dhcp_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"tenant_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: descriptions["tenant_id"],
			},

			"ip_version": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntInSlice([]int{4, 6}),
			},

			"gateway_ip": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"cidr": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"dns_publish_fixed_ip": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			// Computed values
			"allocation_pools": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"end": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"enable_dhcp": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"dns_nameservers": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"service_types": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"host_routes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination_cidr": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"next_hop": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"ipv6_address_mode": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"slaac", "dhcpv6-stateful", "dhcpv6-stateless",
				}, false),
			},

			"ipv6_ra_mode": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"slaac", "dhcpv6-stateful", "dhcpv6-stateless",
				}, false),
			},

			"subnetpool_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"all_tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceNetworkingSubnetV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	networkingClient, err := config.NetworkingV2Client(ctx, util.GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD networking client: %s", err)
	}

	listOpts := subnets.ListOpts{}

	if v, ok := d.GetOk("name"); ok {
		listOpts.Name = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		listOpts.Description = v.(string)
	}

	if v, ok := util.GetOkExists(d, "dhcp_enabled"); ok {
		enableDHCP := v.(bool)
		listOpts.EnableDHCP = &enableDHCP
	}

	if v, ok := d.GetOk("network_id"); ok {
		listOpts.NetworkID = v.(string)
	}

	if v, ok := d.GetOk("tenant_id"); ok {
		listOpts.TenantID = v.(string)
	}

	if v, ok := d.GetOk("ip_version"); ok {
		listOpts.IPVersion = v.(int)
	}

	if v, ok := d.GetOk("gateway_ip"); ok {
		listOpts.GatewayIP = v.(string)
	}

	if v, ok := d.GetOk("cidr"); ok {
		listOpts.CIDR = v.(string)
	}

	if v, ok := d.GetOk("subnet_id"); ok {
		listOpts.ID = v.(string)
	}

	if v, ok := d.GetOk("ipv6_address_mode"); ok {
		listOpts.IPv6AddressMode = v.(string)
	}

	if v, ok := d.GetOk("ipv6_ra_mode"); ok {
		listOpts.IPv6RAMode = v.(string)
	}

	if v, ok := d.GetOk("subnetpool_id"); ok {
		listOpts.SubnetPoolID = v.(string)
	}

	if v, ok := d.GetOk("dns_publish_fixed_ip"); ok {
		v := v.(bool)
		listOpts.DNSPublishFixedIP = &v
	}

	tags := shared.NetworkingAttributesTags(d)
	if len(tags) > 0 {
		listOpts.Tags = strings.Join(tags, ",")
	}

	pages, err := subnets.List(networkingClient, listOpts).AllPages(ctx)
	if err != nil {
		return diag.Errorf("Unable to retrieve vnpaycloud_networking_subnet: %s", err)
	}

	allSubnets, err := subnets.ExtractSubnets(pages)
	if err != nil {
		return diag.Errorf("Unable to extract vnpaycloud_networking_subnet: %s", err)
	}

	if len(allSubnets) < 1 {
		return diag.Errorf("Your query returned no vnpaycloud_networking_subnet. " +
			"Please change your search criteria and try again.")
	}

	if len(allSubnets) > 1 {
		return diag.Errorf("Your query returned more than one vnpaycloud_networking_subnet." +
			" Please try a more specific search criteria")
	}

	subnet := allSubnets[0]

	log.Printf("[DEBUG] Retrieved vnpaycloud_networking_subnet %s: %+v", subnet.ID, subnet)
	d.SetId(subnet.ID)

	d.Set("name", subnet.Name)
	d.Set("description", subnet.Description)
	d.Set("tenant_id", subnet.TenantID)
	d.Set("network_id", subnet.NetworkID)
	d.Set("cidr", subnet.CIDR)
	d.Set("ip_version", subnet.IPVersion)
	d.Set("ipv6_address_mode", subnet.IPv6AddressMode)
	d.Set("ipv6_ra_mode", subnet.IPv6RAMode)
	d.Set("gateway_ip", subnet.GatewayIP)
	d.Set("enable_dhcp", subnet.EnableDHCP)
	d.Set("subnetpool_id", subnet.SubnetPoolID)
	d.Set("dns_publish_fixed_ip", subnet.DNSPublishFixedIP)
	d.Set("all_tags", subnet.Tags)
	d.Set("region", util.GetRegion(d, config))

	if err := d.Set("dns_nameservers", subnet.DNSNameservers); err != nil {
		log.Printf("[DEBUG] Unable to set vnpaycloud_networking_subnet dns_nameservers: %s", err)
	}

	if err := d.Set("service_types", subnet.ServiceTypes); err != nil {
		log.Printf("[DEBUG] Unable to set vnpaycloud_networking_subnet service_types: %s", err)
	}

	hostRoutes := flattenNetworkingSubnetV2HostRoutes(subnet.HostRoutes)
	if err := d.Set("host_routes", hostRoutes); err != nil {
		log.Printf("[DEBUG] Unable to set vnpaycloud_networking_subnet host_routes: %s", err)
	}

	allocationPools := flattenNetworkingSubnetV2AllocationPools(subnet.AllocationPools)
	if err := d.Set("allocation_pools", allocationPools); err != nil {
		log.Printf("[DEBUG] Unable to set vnpaycloud_networking_subnet allocation_pools: %s", err)
	}

	return nil
}
