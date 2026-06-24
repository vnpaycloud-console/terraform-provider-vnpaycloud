package subnet

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandSubnetRoutes(raw []interface{}) []dto.HostRoute {
	routes := make([]dto.HostRoute, 0, len(raw))
	for _, r := range raw {
		m := r.(map[string]interface{})
		routes = append(routes, dto.HostRoute{
			Destination: m["destination"].(string),
			Nexthop:     m["nexthop"].(string),
		})
	}
	return routes
}

func flattenSubnetRoutes(routes []dto.HostRoute) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(routes))
	for _, r := range routes {
		result = append(result, map[string]interface{}{
			"destination": r.Destination,
			"nexthop":     r.Nexthop,
		})
	}
	return result
}

func ResourceSubnet() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSubnetCreate,
		ReadContext:   resourceSubnetRead,
		UpdateContext: resourceSubnetUpdate,
		DeleteContext: resourceSubnetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cidr": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"gateway_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_dhcp": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"used_by_k8s": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"used_by_si": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					return oldValue == ""
				},
			},
			"dns_nameservers": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"route": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination": {
							Type:     schema.TypeString,
							Required: true,
						},
						"nexthop": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSubnetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	var dnsNameservers []string
	for _, v := range d.Get("dns_nameservers").([]interface{}) {
		dnsNameservers = append(dnsNameservers, v.(string))
	}

	createOpts := dto.CreateSubnetRequest{
		Name:           d.Get("name").(string),
		VpcID:          d.Get("vpc_id").(string),
		CIDR:           d.Get("cidr").(string),
		GatewayIP:      d.Get("gateway_ip").(string),
		EnableDHCP:     d.Get("enable_dhcp").(bool),
		UsedByK8S:      d.Get("used_by_k8s").(bool),
		UsedBySI:       d.Get("used_by_si").(bool),
		DNSNameservers: dnsNameservers,
	}

	tflog.Debug(ctx, "vnpaycloud_subnet create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.SubnetResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.Subnets(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_subnet: %s", err)
	}

	d.SetId(createResp.Subnet.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating"},
		Target:     []string{"active", "created"},
		Refresh:    subnetStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.Subnet.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_subnet %s to become ready: %s", createResp.Subnet.ID, err)
	}

	if v, ok := d.GetOk("route"); ok {
		routesReq := dto.UpdateSubnetRoutesRequest{Routes: expandSubnetRoutes(v.([]interface{}))}
		if _, err := cfg.Client.Put(ctx, client.ApiPath.SubnetRoutes(cfg.ProjectID, d.Id()), routesReq, nil, nil); err != nil {
			return diag.Errorf("Error setting routes for vnpaycloud_subnet %s: %s", d.Id(), err)
		}
	}

	return resourceSubnetRead(ctx, d, meta)
}

func resourceSubnetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	subnetResp := &dto.SubnetResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.SubnetWithID(cfg.ProjectID, d.Id()), subnetResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_subnet"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_subnet "+d.Id(), map[string]interface{}{"subnet": subnetResp.Subnet})

	d.Set("name", subnetResp.Subnet.Name)
	d.Set("vpc_id", subnetResp.Subnet.VpcID)
	d.Set("cidr", subnetResp.Subnet.CIDR)
	d.Set("gateway_ip", subnetResp.Subnet.GatewayIP)
	d.Set("enable_dhcp", subnetResp.Subnet.EnableDHCP)
	d.Set("used_by_k8s", subnetResp.Subnet.UsedByK8S)
	d.Set("dns_nameservers", subnetResp.Subnet.DNSNameservers)
	d.Set("route", flattenSubnetRoutes(subnetResp.Subnet.Routes))
	d.Set("status", subnetResp.Subnet.Status)
	d.Set("created_at", subnetResp.Subnet.CreatedAt)

	return nil
}

func resourceSubnetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChanges("name", "dns_nameservers") {
		var dnsNameservers []string
		for _, v := range d.Get("dns_nameservers").([]interface{}) {
			dnsNameservers = append(dnsNameservers, v.(string))
		}

		updateOpts := dto.UpdateSubnetRequest{
			Name:           d.Get("name").(string),
			DNSNameservers: dnsNameservers,
		}

		tflog.Debug(ctx, "vnpaycloud_subnet update options", map[string]interface{}{"update_opts": updateOpts})

		if _, err := cfg.Client.Put(ctx, client.ApiPath.SubnetWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil); err != nil {
			return diag.Errorf("Error updating vnpaycloud_subnet %s: %s", d.Id(), err)
		}
	}

	if d.HasChange("route") {
		routesReq := dto.UpdateSubnetRoutesRequest{Routes: expandSubnetRoutes(d.Get("route").([]interface{}))}
		if _, err := cfg.Client.Put(ctx, client.ApiPath.SubnetRoutes(cfg.ProjectID, d.Id()), routesReq, nil, nil); err != nil {
			return diag.Errorf("Error updating routes for vnpaycloud_subnet %s: %s", d.Id(), err)
		}
	}

	return resourceSubnetRead(ctx, d, meta)
}

func resourceSubnetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	subnetResp := &dto.SubnetResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.SubnetWithID(cfg.ProjectID, d.Id()), subnetResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_subnet"))
	}

	if subnetResp.Subnet.Status != "deleting" {
		if _, err := cfg.Client.Delete(ctx, client.ApiPath.SubnetWithID(cfg.ProjectID, d.Id()), nil); err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_subnet"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "created"},
		Target:     []string{"deleted"},
		Refresh:    subnetStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_subnet %s to delete: %s", d.Id(), err)
	}

	return nil
}
