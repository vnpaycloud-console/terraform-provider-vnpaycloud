package loadbalancer

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/shared"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLoadBalancerCreate,
		ReadContext:   resourceLoadBalancerRead,
		UpdateContext: resourceLoadBalancerUpdate,
		DeleteContext: resourceLoadBalancerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"vip_network_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				AtLeastOneOf: []string{"vip_network_id",
					"vip_subnet_id", "vip_port_id"},
			},

			"vip_subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				AtLeastOneOf: []string{"vip_network_id",
					"vip_subnet_id", "vip_port_id"},
			},

			"vip_port_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				AtLeastOneOf: []string{"vip_network_id",
					"vip_subnet_id", "vip_port_id"},
			},

			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"vip_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"admin_state_up": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},

			"flavor_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"loadbalancer_provider": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"vip_qos_policy_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func resourceLoadBalancerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Terrform Client: %s", err)
	}

	var (
		lbID string
		//vipPortID  string
		//lbProvider string
	)

	//if v, ok := d.GetOk("loadbalancer_provider"); ok {
	//	lbProvider = v.(string)
	//}
	//adminStateUp := d.Get("admin_state_up").(bool)

	createOpts := dto.CreateLoadBalancerOpts{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		VipSubnetID: d.Get("vip_subnet_id").(string),
		FlavorID:    d.Get("flavor_id").(string),
	}

	// availability_zone requires octavia minor version 2.14. Only set when specified.
	//if v, ok := d.GetOk("availability_zone"); ok {
	//	aZ := v.(string)
	//	createOpts.AvailabilityZone = aZ
	//}
	//
	//if v, ok := d.GetOk("tags"); ok {
	//	tags := v.(*schema.Set).List()
	//	createOpts.Tags = util.ExpandToStringSlice(tags)
	//}

	tflog.Info(ctx, "Creating vnpaycloud_lb_loadbalancer create options: %+v", map[string]interface{}{"create_opts": createOpts})

	lbResp := &dto.CreateLoadBalancerResponse{}
	_, err = tfClient.Post(ctx, client.ApiPath.LbaasLoadBalancer, dto.CreateLoadBalancerRequest{LoadBalancer: createOpts}, lbResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_networking_network: %s", err)
	}

	lbID = lbResp.LoadBalancer.ID
	//vipPortID = lbResp.LoadBalancer.VipPortID

	// Store the ID now
	d.SetId(lbID)

	// Wait for load-balancer to become active before continuing.
	timeout := d.Timeout(schema.TimeoutCreate)
	err = shared.WaitForLBLoadBalancer(ctx, tfClient, lbID, "ACTIVE", shared.GetLbPendingStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	// Once the load-balancer has been created, apply any requested security groups
	// to the port that was created behind the scenes.
	//tfClient, err = client.NewClient(ctx, config.ConsoleClientConfig)
	//if err != nil {
	//	return diag.Errorf("Error creating VNPAYCLOUD networking client: %s", err)
	//}
	//if err := shared.ResourceLoadBalancerSetSecurityGroups(ctx, tfClient, vipPortID, d); err != nil {
	//	return diag.Errorf("Error setting vnpaycloud_lb_loadbalancer security groups: %s", err)
	//}

	return resourceLoadBalancerRead(ctx, d, meta)
}

func resourceLoadBalancerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Terrform Client: %s", err)
	}

	lbResp := &dto.GetLoadBalancerResponse{}
	_, err = tfClient.Get(ctx, client.ApiPath.LbaasLoadBalancerWithId(d.Id()), lbResp, &client.RequestOpts{})
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_network"))
	}
	if lbResp == nil || len(lbResp.LoadBalancer.ID) == 0 {
		d.SetId("")
		return diag.FromErr(fmt.Errorf("Error retrieving vnpaycloud_loadbalancer"))
	}
	lb := &lbResp.LoadBalancer

	var vipPortID string

	tflog.Info(ctx, "Retrieving vnpaycloud_lb_loadbalancer %s: %+v",
		map[string]interface{}{"id": d.Id()},
		map[string]interface{}{"lb": lb})

	d.Set("name", lb.Name)
	d.Set("description", lb.Description)
	d.Set("vip_subnet_id", lb.VipSubnetID)
	d.Set("vip_network_id", lb.VipNetworkID)
	d.Set("tenant_id", lb.ProjectID)
	d.Set("vip_address", lb.VipAddress)
	d.Set("vip_port_id", lb.VipPortID)
	d.Set("admin_state_up", lb.AdminStateUp)
	d.Set("flavor_id", lb.FlavorID)
	d.Set("provider", lb.Provider)
	d.Set("availability_zone", lb.AvailabilityZone)
	d.Set("region", util.GetRegion(d, config))
	d.Set("tags", lb.Tags)
	d.Set("vip_qos_policy_id", lb.VipQosPolicyID)

	vipPortID = lb.VipPortID

	// Get any security groups on the VIP Port.
	if vipPortID != "" {
		tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
		if err != nil {
			return diag.Errorf("Error creating VNPAY Cloud Terrform Client: %s", err)
		}
		if err := shared.ResourceLoadBalancerGetSecurityGroups(ctx, tfClient, vipPortID, d); err != nil {
			return diag.Errorf("Error getting port security groups for vnpaycloud_lb_loadbalancer: %s", err)
		}
	}

	return nil
}

func resourceLoadBalancerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceLoadBalancerRead(ctx, d, meta)
}

func resourceLoadBalancerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Terrform Client: %s", err)
	}

	tflog.Info(ctx, "Deleting vnpaycloud_lb_loadbalancer %s",
		map[string]interface{}{"id": d.Id()})
	timeout := d.Timeout(schema.TimeoutDelete)
	err = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		_, err = tfClient.Delete(ctx, client.ApiPath.LbaasLoadBalancerWithId(d.Id()), nil)
		if err != nil {
			return util.CheckForRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_lb_loadbalancer"))
	}

	// Wait for load-balancer to become deleted.
	tfClient, err = client.NewClient(ctx, config.ConsoleClientConfig)
	err = shared.WaitForLBLoadBalancer(ctx, tfClient, d.Id(), "DELETED", shared.GetLbPendingDeleteStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
