package subnet

import (
	"context"
	"log"
	"net"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/shared"
	"terraform-provider-vnpaycloud/vnpaycloud/types"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceNetworkingSubnet() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkingSubnetCreate,
		ReadContext:   resourceNetworkingSubnetRead,
		UpdateContext: resourceNetworkingSubnetUpdate,
		DeleteContext: resourceNetworkingSubnetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"cidr": {
				Type:          schema.TypeString,
				ConflictsWith: []string{"prefix_length"},
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsCIDR,
				},
			},
			"prefix_length": {
				Type:          schema.TypeInt,
				ConflictsWith: []string{"cidr"},
				Optional:      true,
				ForceNew:      true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"allocation_pool": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start": {
							Type:     schema.TypeString,
							Required: true,
						},
						"end": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"gateway_ip": {
				Type:          schema.TypeString,
				ConflictsWith: []string{"no_gateway"},
				Optional:      true,
				ForceNew:      false,
				Computed:      true,
			},
			"no_gateway": {
				Type:          schema.TypeBool,
				ConflictsWith: []string{"gateway_ip"},
				Optional:      true,
				Default:       false,
				ForceNew:      false,
			},
			"ip_version": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      4,
				ForceNew:     true,
				ValidateFunc: validation.IntInSlice([]int{4, 6}),
			},
			"enable_dhcp": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  true,
			},
			"dns_nameservers": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"dns_publish_fixed_ip": {
				Type:     schema.TypeBool,
				Optional: true,
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
				Optional: true,
				ForceNew: true,
			},
			"value_specs": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"service_types": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Computed: true,
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

func resourceNetworkingSubnetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	subnetClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD Subnet client: %s", err)
	}

	// Check nameservers.
	if err := networkingSubnetDNSNameserverAreUnique(d.Get("dns_nameservers").([]interface{})); err != nil {
		return diag.Errorf("vnpaycloud_networking_subnet dns_nameservers argument is invalid: %s", err)
	}

	// Get raw allocation pool value.
	allocationPool := d.Get("allocation_pool").(*schema.Set).List()

	// Set basic options.
	createOpts := dto.CreateSubnetRequest{
		Subnet: dto.SubnetCreateOpts{
			NetworkID:       d.Get("network_id").(string),
			Name:            d.Get("name").(string),
			Description:     d.Get("description").(string),
			TenantID:        d.Get("tenant_id").(string),
			IPv6AddressMode: d.Get("ipv6_address_mode").(string),
			IPv6RAMode:      d.Get("ipv6_ra_mode").(string),
			AllocationPools: expandNetworkingSubnetAllocationPools(allocationPool),
			DNSNameservers:  util.ExpandToStringSlice(d.Get("dns_nameservers").([]interface{})),
			ServiceTypes:    util.ExpandToStringSlice(d.Get("service_types").([]interface{})),
			SubnetPoolID:    d.Get("subnetpool_id").(string),
			IPVersion:       types.IPVersion(d.Get("ip_version").(int)),
			VPCID:           d.Get("vpc_id").(string),
			ValueSpecs:      util.MapValueSpecs(d),
		},
	}

	if v, ok := d.GetOk("dns_publish_fixed_ip"); ok {
		v := v.(bool)
		createOpts.Subnet.DNSPublishFixedIP = &v
	}

	// Set CIDR if provided. Check if inferred subnet would match the provided cidr.
	if v, ok := d.GetOk("cidr"); ok {
		cidr := v.(string)
		_, netAddr, err := net.ParseCIDR(cidr)
		if err != nil {
			return diag.Errorf("Invalid CIDR %s: %s", cidr, err)
		}
		if netAddr.String() != cidr {
			return diag.Errorf("cidr %s doesn't match subnet address %s for vnpaycloud_networking_subnet", cidr, netAddr.String())
		}
		createOpts.Subnet.CIDR = cidr
	}

	// Set gateway options if provided.
	if v, ok := d.GetOk("gateway_ip"); ok {
		gatewayIP := v.(string)
		createOpts.Subnet.GatewayIP = &gatewayIP
	}

	noGateway := d.Get("no_gateway").(bool)
	if noGateway {
		gatewayIP := ""
		createOpts.Subnet.GatewayIP = &gatewayIP
	}

	// Validate and set prefix options.
	if v, ok := d.GetOk("prefix_length"); ok {
		if d.Get("subnetpool_id").(string) == "" {
			return diag.Errorf("'prefix_length' is only valid if 'subnetpool_id' is set for vnpaycloud_networking_subnet")
		}
		prefixLength := v.(int)
		createOpts.Subnet.Prefixlen = prefixLength
	}

	// Set DHCP options if provided.
	enableDHCP := d.Get("enable_dhcp").(bool)
	createOpts.Subnet.EnableDHCP = &enableDHCP

	log.Printf("[DEBUG] vnpaycloud_networking_subnet create options: %#v", createOpts)
	createResp := &dto.CreateSubnetResponse{}
	_, err = subnetClient.Post(ctx, client.ApiPath.Subnet, createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_networking_subnet: %s", err)
	}

	log.Printf("[DEBUG] Waiting for vnpaycloud_networking_subnet %s to become available", createResp.Subnet.ID)
	stateConf := &retry.StateChangeConf{
		Target:     []string{"ACTIVE"},
		Refresh:    networkingSubnetStateRefreshFunc(ctx, subnetClient, createResp.Subnet.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_networking_subnet %s to become available: %s", createResp.Subnet.ID, err)
	}

	d.SetId(createResp.Subnet.ID)

	// tags := shared.NetworkingAttributesTags(d)
	// if len(tags) > 0 {
	// 	tagOpts := attributestags.ReplaceAllOpts{Tags: tags}
	// 	tags, err := attributestags.ReplaceAll(ctx, subnetClient, "subnets", createResp.Subnet.ID, tagOpts).Extract()
	// 	if err != nil {
	// 		return diag.Errorf("Error creating tags on vnpaycloud_networking_subnet %s: %s", createResp.Subnet.ID, err)
	// 	}
	// 	log.Printf("[DEBUG] Set tags %s on vnpaycloud_networking_subnet %s", tags, createResp.Subnet.ID)
	// }

	log.Printf("[DEBUG] Created vnpaycloud_networking_subnet %s: %#v", createResp.Subnet.ID, createResp.Subnet)
	return resourceNetworkingSubnetRead(ctx, d, meta)
}

func resourceNetworkingSubnetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	subnetClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD Subnet client: %s", err)
	}

	subnetResp := &dto.GetSubnetResponse{}
	_, err = subnetClient.Get(ctx, client.ApiPath.SubnetWithId(d.Id()), subnetResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error getting vnpaycloud_networking_subnet"))
	}

	log.Printf("[DEBUG] Retrieved vnpaycloud_networking_subnet %s: %#v", d.Id(), subnetResp.Subnet)

	d.Set("network_id", subnetResp.Subnet.NetworkID)
	d.Set("cidr", subnetResp.Subnet.CIDR)
	d.Set("ip_version", subnetResp.Subnet.IPVersion)
	d.Set("name", subnetResp.Subnet.Name)
	d.Set("description", subnetResp.Subnet.Description)
	d.Set("tenant_id", subnetResp.Subnet.TenantID)
	d.Set("dns_nameservers", subnetResp.Subnet.DNSNameservers)
	d.Set("service_types", subnetResp.Subnet.ServiceTypes)
	d.Set("enable_dhcp", subnetResp.Subnet.EnableDHCP)
	d.Set("network_id", subnetResp.Subnet.NetworkID)
	d.Set("ipv6_address_mode", subnetResp.Subnet.IPv6AddressMode)
	d.Set("ipv6_ra_mode", subnetResp.Subnet.IPv6RAMode)
	d.Set("subnetpool_id", subnetResp.Subnet.SubnetPoolID)
	d.Set("dns_publish_fixed_ip", subnetResp.Subnet.DNSPublishFixedIP)
	d.Set("vpc_id", subnetResp.Subnet.VPCID)

	shared.NetworkingReadAttributesTags(d, subnetResp.Subnet.Tags)

	// Set the allocation_pool attribute
	allocationPools := flattenNetworkingSubnetAllocationPools(subnetResp.Subnet.AllocationPools)
	if err := d.Set("allocation_pool", allocationPools); err != nil {
		log.Printf("[DEBUG] Unable to set vnpaycloud_networking_subnet allocation_pool: %s", err)
	}

	// Set the subnet's "gateway_ip" and "no_gateway" attributes.
	d.Set("gateway_ip", subnetResp.Subnet.GatewayIP)
	d.Set("no_gateway", false)
	if subnetResp.Subnet.GatewayIP != "" {
		d.Set("no_gateway", false)
	} else {
		d.Set("no_gateway", true)
	}

	d.Set("region", util.GetRegion(d, config))

	return nil
}

func resourceNetworkingSubnetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceNetworkingSubnetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	subnetClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD Subnet client: %s", err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"ACTIVE"},
		Target:     []string{"DELETED"},
		Refresh:    networkingSubnetStateRefreshFuncDelete(ctx, subnetClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_networking_subnet %s to become deleted: %s", d.Id(), err)
	}

	return nil
}
