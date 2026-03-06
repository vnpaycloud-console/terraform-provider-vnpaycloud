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

func ResourceSubnet() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSubnetCreate,
		ReadContext:   resourceSubnetRead,
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
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cidr": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"gateway_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"enable_dhcp": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"used_by_k8s": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
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

	createOpts := dto.CreateSubnetRequest{
		Name:       d.Get("name").(string),
		VpcID:      d.Get("vpc_id").(string),
		CIDR:       d.Get("cidr").(string),
		GatewayIP:  d.Get("gateway_ip").(string),
		EnableDHCP: d.Get("enable_dhcp").(bool),
		UsedByK8S:  d.Get("used_by_k8s").(bool),
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
	d.Set("status", subnetResp.Subnet.Status)
	d.Set("created_at", subnetResp.Subnet.CreatedAt)

	return nil
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
