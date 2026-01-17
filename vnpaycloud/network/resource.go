package network

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceNetworkingNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkingNetworkCreate,
		ReadContext:   resourceNetworkingNetworkRead,
		UpdateContext: resourceNetworkingNetworkUpdate,
		DeleteContext: resourceNetworkingNetworkDelete,
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
				ForceNew: true,
				Computed: true,
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

			"admin_state_up": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Computed: true,
			},

			"shared": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Computed: true,
			},

			"external": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Computed: true,
			},

			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"segments": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"physical_network": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"network_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"segmentation_id": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
					},
				},
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

			"availability_zone_hints": {
				Type:     schema.TypeSet,
				Computed: true,
				ForceNew: true,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"transparent_vlan": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"port_security_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"mtu": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"dns_domain": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^$|\.$`), "fully-qualified (unambiguous) DNS domain names must have a dot at the end"),
			},

			"qos_policy_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Computed: true,
			},
		},
	}
}

func resourceNetworkingNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	networkingClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud networking client: %s", err)
	}

	createOpts := dto.CreateNetworkOpts{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	if v, ok := util.GetOkExists(d, "admin_state_up"); ok {
		asu := v.(bool)
		createOpts.AdminStateUp = &asu
	}

	log.Printf("[DEBUG] vnpaycloud_networking_network create options: %#v", createOpts)
	networkResp := &dto.CreateNetworkResponse{}
	_, err = networkingClient.Post(ctx, client.ApiPath.Network, dto.CreateNetworkRequest{Network: createOpts}, networkResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_networking_network: %s", err)
	}

	log.Printf("[DEBUG] Waiting for vnpaycloud_networking_network %s to become available.", networkResp.Network.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"BUILD"},
		Target:     []string{"ACTIVE", "DOWN"},
		Refresh:    resourceNetworkingNetworkStateRefreshFunc(ctx, networkingClient, networkResp.Network.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_networking_network %s to become available: %s",
			networkResp.Network.ID, err)
	}

	d.SetId(networkResp.Network.ID)

	log.Printf("[DEBUG] Created vnpaycloud_networking_network %s: %#v",
		networkResp.Network.ID, networkResp.Network)
	return resourceNetworkingNetworkRead(ctx, d, meta)
}

func resourceNetworkingNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	networkingClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud networking client: %s", err)
	}

	networkResp := &dto.GetNetworkResponse{}
	otps := &client.RequestOpts{}
	_, err = networkingClient.Get(ctx, client.ApiPath.NetworkWithId(d.Id()), networkResp, otps)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_network"))
	}
	if networkResp == nil || len(networkResp.Network.ID) == 0 {
		d.SetId("")
		return diag.FromErr(fmt.Errorf("Error retrieving vnpaycloud_network"))
	}
	network := networkResp.Network

	log.Printf("[DEBUG] Retrieved vnpaycloud_networking_network %s: %#v", d.Id(), network)

	d.Set("name", network.Name)
	d.Set("description", network.Description)
	d.Set("admin_state_up", network.AdminStateUp)
	d.Set("shared", network.Shared)
	d.Set("external", network.External)
	d.Set("tenant_id", network.TenantID)
	d.Set("segments", flattenNetworkingNetworkSegments(network))
	d.Set("transparent_vlan", network.VLANTransparent)
	d.Set("port_security_enabled", network.PortSecurityEnabled)
	d.Set("mtu", network.MTU)
	d.Set("dns_domain", network.DNSDomain)
	d.Set("qos_policy_id", network.QoSPolicyID)
	d.Set("region", util.GetRegion(d, config))

	if err := d.Set("availability_zone_hints", network.AvailabilityZoneHints); err != nil {
		log.Printf("[DEBUG] Unable to set vnpaycloud_networking_network %s availability_zone_hints: %s", d.Id(), err)
	}

	return nil
}

func resourceNetworkingNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceNetworkingNetworkRead(ctx, d, meta)
}

func resourceNetworkingNetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	networkingClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud networking client: %s", err)
	}

	if _, err := networkingClient.Delete(ctx, client.ApiPath.NetworkWithId(d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_networking_network"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"ACTIVE"},
		Target:     []string{"DELETED"},
		Refresh:    resourceNetworkingNetworkStateRefreshFunc(ctx, networkingClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_networking_network %s to Delete:  %s", d.Id(), err)
	}

	d.SetId("")
	return nil
}
