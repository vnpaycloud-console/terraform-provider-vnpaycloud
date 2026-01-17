package floatingip

import (
	"context"
	"log"
	"regexp"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/shared"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
)

func ResourceNetworkingFloatingIP() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkFloatingIPCreate,
		ReadContext:   resourceNetworkFloatingIPRead,
		UpdateContext: resourceNetworkFloatingIPUpdate,
		DeleteContext: resourceNetworkFloatingIPDelete,
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

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"pool": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_POOL_NAME", nil),
			},

			"port_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"fixed_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"subnet_ids": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"subnet_id"},
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

			"all_tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"dns_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"dns_domain": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^$|\.$`), "fully-qualified (unambiguous) DNS domain names must have a dot at the end"),
			},
		},
	}
}

func resourceNetworkFloatingIPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	networkingClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCloud network client: %s", err)
	}

	poolName := d.Get("pool").(string)
	poolID, err := shared.NetworkingNetworkID(ctx, d, meta, poolName)
	if err != nil {
		return diag.Errorf("Error retrieving ID for vnpaycloud_networking_floatingip pool name %s: %s", poolName, err)
	}
	if len(poolID) == 0 {
		return diag.Errorf("No network found with name: %s", poolName)
	}

	subnetID := d.Get("subnet_id").(string)
	var subnetIDs []string
	if v, ok := d.Get("subnet_ids").([]interface{}); ok {
		subnetIDs = make([]string, len(v))
		for i, v := range v {
			subnetIDs[i] = v.(string)
		}
	}

	if subnetID == "" && len(subnetIDs) > 0 {
		subnetID = subnetIDs[0]
	}

	createOpts := &dto.CreateFloatingIPOpts{
		FloatingNetworkID: poolID,
		Description:       d.Get("description").(string),
		FloatingIP:        d.Get("address").(string),
		PortID:            d.Get("port_id").(string),
		TenantID:          d.Get("tenant_id").(string),
		FixedIP:           d.Get("fixed_ip").(string),
		SubnetID:          subnetID,
		ValueSpecs:        util.MapValueSpecs(d),
	}

	dnsName := d.Get("dns_name").(string)
	dnsDomain := d.Get("dns_domain").(string)
	if dnsName != "" || dnsDomain != "" {
		createOpts.DNSName = dnsName
		createOpts.DNSDomain = dnsDomain
	}

	var fipResp dto.CreateFloatingIPResponse

	log.Printf("[DEBUG] vnpaycloud_networking_floatingip create options: %#v", createOpts)

	createReq := dto.CreateFloatingIPRequest{
		FloatingIP: *createOpts,
	}

	if len(subnetIDs) == 0 {
		_, err := networkingClient.Post(ctx, client.ApiPath.FloatingIP, createReq, &fipResp, nil)
		if err != nil {
			return diag.Errorf("Error creating vnpaycloud_networking_floatingip: %s", err)
		}
	} else {
		for i, subnetID := range subnetIDs {
			createReq.FloatingIP.SubnetID = subnetID

			log.Printf("[DEBUG] vnpaycloud_networking_floatingip create options (try %d): %#v", i+1, createReq.FloatingIP)

			_, err := networkingClient.Post(ctx, client.ApiPath.FloatingIP, createReq, &fipResp, nil)
			if err != nil {
				if shared.RetryOn409(err) {
					continue
				}
				return diag.Errorf("Error creating vnpaycloud_networking_floatingip: %s", err)
			}
			break
		}
		if err != nil {
			return diag.Errorf("Error creating vnpaycloud_networking_floatingip: %d subnets exhausted: %s", len(subnetIDs), err)
		}
	}

	log.Printf("[DEBUG] Waiting for vnpaycloud_networking_floatingip %s to become available.", fipResp.FloatingIP.ID)

	stateConf := &retry.StateChangeConf{
		Target:     []string{"ACTIVE", "DOWN"},
		Refresh:    networkingFloatingIPStateRefreshFunc(ctx, networkingClient, fipResp.FloatingIP.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_networking_floatingip %s to become available: %s", fipResp.FloatingIP.ID, err)
	}

	d.SetId(fipResp.FloatingIP.ID)

	if createOpts.SubnetID != "" {
		// resourceNetworkFloatingIPRead doesn't handle this, since FIP GET request doesn't provide this info.
		d.Set("subnet_id", createOpts.SubnetID)
	}

	// tags := shared.NetworkingAttributesTags(d)
	// if len(tags) > 0 {
	// 	tagOpts := attributestags.ReplaceAllOpts{Tags: tags}
	// 	tags, err := attributestags.ReplaceAll(ctx, networkingClient, "floatingips", fip.ID, tagOpts).Extract()
	// 	if err != nil {
	// 		return diag.Errorf("Error setting tags on vnpaycloud_networking_floatingip %s: %s", fip.ID, err)
	// 	}
	// 	log.Printf("[DEBUG] Set tags %s on vnpaycloud_networking_floatingip %s", tags, fip.ID)
	// }

	log.Printf("[DEBUG] Created vnpaycloud_networking_floatingip %s: %#v", fipResp.FloatingIP.ID, fipResp.FloatingIP)
	return resourceNetworkFloatingIPRead(ctx, d, meta)
}

func resourceNetworkFloatingIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	networkingClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCloud network client: %s", err)
	}

	var fipResp dto.GetFloatingIPResponse

	_, err = networkingClient.Get(ctx, client.ApiPath.FloatingIPWithId(d.Id()), &fipResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error getting vnpaycloud_networking_floatingip"))
	}

	log.Printf("[DEBUG] Retrieved vnpaycloud_networking_floatingip %s: %#v", d.Id(), fipResp.FloatingIP)

	d.Set("description", fipResp.FloatingIP.Description)
	d.Set("address", fipResp.FloatingIP.FloatingIP)
	d.Set("port_id", fipResp.FloatingIP.PortID)
	d.Set("fixed_ip", fipResp.FloatingIP.FixedIP)
	d.Set("tenant_id", fipResp.FloatingIP.TenantID)
	d.Set("dns_name", fipResp.FloatingIP.DNSName)
	d.Set("dns_domain", fipResp.FloatingIP.DNSDomain)
	d.Set("region", util.GetRegion(d, config))

	shared.NetworkingReadAttributesTags(d, fipResp.FloatingIP.Tags)

	poolName, err := shared.NetworkingNetworkName(ctx, d, meta, fipResp.FloatingIP.FloatingNetworkID)
	if err != nil {
		return diag.Errorf("Error retrieving pool name for vnpaycloud_networking_floatingip %s: %s", d.Id(), err)
	}
	d.Set("pool", poolName)

	return nil
}

func resourceNetworkFloatingIPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceNetworkFloatingIPRead(ctx, d, meta)
}

func resourceNetworkFloatingIPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	configMeta := meta.(*config.Config)
	networkingClient, err := client.NewClient(ctx, configMeta.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCloud network client: %s", err)
	}

	_, err = networkingClient.Delete(ctx, client.ApiPath.FloatingIPWithId(d.Id()), nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error deleting vnpaycloud_networking_floatingip"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"ACTIVE", "DOWN"},
		Target:     []string{"DELETED"},
		Refresh:    networkingFloatingIPStateRefreshFunc(ctx, networkingClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_networking_floatingip %s to Delete:  %s", d.Id(), err)
	}

	d.SetId("")
	return nil
}
