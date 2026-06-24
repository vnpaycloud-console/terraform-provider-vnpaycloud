package floatingip

import (
	"context"
	"strings"
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

func ResourceFloatingIP() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFloatingIPCreate,
		ReadContext:   resourceFloatingIPRead,
		UpdateContext: resourceFloatingIPUpdate,
		DeleteContext: resourceFloatingIPDelete,
		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
			raw := d.GetRawConfig()
			if raw.IsNull() {
				return nil
			}
			for _, field := range []string{"vpc_id", "port_id"} {
				cv := raw.GetAttr(field)
				oldRaw, _ := d.GetChange(field)
				oldValue, _ := oldRaw.(string)
				if cv.IsKnown() && !cv.IsNull() && cv.AsString() == "" && oldValue != "" {
					if err := d.SetNew(field, ""); err != nil {
						return err
					}
				}
			}
			return nil
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"vpc_id"},
			},
			"vpc_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"port_id"},
			},
			"instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"fixed_ip": {
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

func resourceFloatingIPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateFloatingIPRequest{}

	tflog.Debug(ctx, "vnpaycloud_floating_ip create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.FloatingIPResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.FloatingIPs(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_floating_ip: %s", err)
	}

	d.SetId(createResp.FloatingIP.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating"},
		Target:     []string{"active", "created"},
		Refresh:    floatingIPStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.FloatingIP.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_floating_ip %s to become ready: %s", createResp.FloatingIP.ID, err)
	}

	assocReq := buildAssociateRequest(d)
	if assocReq != nil {
		if err := associateFloatingIPWithRetry(ctx, cfg, d.Id(), assocReq, d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.Errorf("Error associating vnpaycloud_floating_ip %s: %s", d.Id(), err)
		}
	}

	return resourceFloatingIPRead(ctx, d, meta)
}

func associateFloatingIPWithRetry(ctx context.Context, cfg *config.Config, id string, assocReq *dto.AssociateFloatingIPRequest, timeout time.Duration) error {
	return retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		assocResp := &dto.FloatingIPResponse{}
		_, err := cfg.Client.Post(ctx, client.ApiPath.FloatingIPAssociate(cfg.ProjectID, id), assocReq, assocResp, nil)
		if err == nil {
			return nil
		}

		if strings.Contains(err.Error(), "port has been associated to another floating IP") {
			return retry.RetryableError(err)
		}
		return retry.NonRetryableError(err)
	})
}

func resourceFloatingIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	fipResp := &dto.FloatingIPResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.FloatingIPWithID(cfg.ProjectID, d.Id()), fipResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_floating_ip"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_floating_ip "+d.Id(), map[string]interface{}{"floating_ip": fipResp.FloatingIP})

	d.Set("address", fipResp.FloatingIP.Address)
	d.Set("status", fipResp.FloatingIP.Status)
	d.Set("port_id", fipResp.FloatingIP.PortID)
	d.Set("vpc_id", fipResp.FloatingIP.VpcID)
	d.Set("instance_id", fipResp.FloatingIP.InstanceID)
	d.Set("instance_name", fipResp.FloatingIP.InstanceName)
	d.Set("fixed_ip", fipResp.FloatingIP.FixedIP)
	d.Set("created_at", fipResp.FloatingIP.CreatedAt)

	return nil
}

func resourceFloatingIPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	assocReq, managesAssociation := buildAssociateRequestFromRawConfig(d)
	if !managesAssociation {
		return resourceFloatingIPRead(ctx, d, meta)
	}

	currentResp := &dto.FloatingIPResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.FloatingIPWithID(cfg.ProjectID, d.Id()), currentResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_floating_ip before update"))
	}

	if currentResp.FloatingIP.PortID != "" || currentResp.FloatingIP.VpcID != "" {
		tflog.Debug(ctx, "Disassociating vnpaycloud_floating_ip", map[string]interface{}{"floating_ip_id": d.Id()})
		disassocResp := &dto.FloatingIPResponse{}
		_, err = cfg.Client.Post(ctx, client.ApiPath.FloatingIPDisassociate(cfg.ProjectID, d.Id()), dto.DisassociateFloatingIPRequest{}, disassocResp, nil)
		if err != nil {
			return diag.Errorf("Error disassociating vnpaycloud_floating_ip %s: %s", d.Id(), err)
		}
	}

	if assocReq != nil {
		if err := associateFloatingIPWithRetry(ctx, cfg, d.Id(), assocReq, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("Error associating vnpaycloud_floating_ip %s: %s", d.Id(), err)
		}
	}

	return resourceFloatingIPRead(ctx, d, meta)
}

func buildAssociateRequest(d *schema.ResourceData) *dto.AssociateFloatingIPRequest {
	if portID, ok := d.GetOk("port_id"); ok && portID.(string) != "" {
		return &dto.AssociateFloatingIPRequest{PortID: portID.(string)}
	}
	if vpcID, ok := d.GetOk("vpc_id"); ok && vpcID.(string) != "" {
		return &dto.AssociateFloatingIPRequest{VpcID: vpcID.(string)}
	}
	return nil
}

func buildAssociateRequestFromRawConfig(d *schema.ResourceData) (*dto.AssociateFloatingIPRequest, bool) {
	raw := d.GetRawConfig()
	if raw.IsNull() {
		return nil, false
	}

	portID := raw.GetAttr("port_id")
	if portID.IsKnown() && !portID.IsNull() {
		value := portID.AsString()
		if value == "" {
			return nil, true
		}
		return &dto.AssociateFloatingIPRequest{PortID: value}, true
	}

	vpcID := raw.GetAttr("vpc_id")
	if vpcID.IsKnown() && !vpcID.IsNull() {
		value := vpcID.AsString()
		if value == "" {
			return nil, true
		}
		return &dto.AssociateFloatingIPRequest{VpcID: value}, true
	}

	return nil, false
}

func resourceFloatingIPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	fipResp := &dto.FloatingIPResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.FloatingIPWithID(cfg.ProjectID, d.Id()), fipResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_floating_ip"))
	}

	// Disassociate before deleting if currently associated
	if fipResp.FloatingIP.PortID != "" || fipResp.FloatingIP.VpcID != "" {
		disassocResp := &dto.FloatingIPResponse{}
		_, err := cfg.Client.Post(ctx, client.ApiPath.FloatingIPDisassociate(cfg.ProjectID, d.Id()), dto.DisassociateFloatingIPRequest{}, disassocResp, nil)
		if err != nil {
			return diag.Errorf("Error disassociating vnpaycloud_floating_ip %s before deletion: %s", d.Id(), err)
		}
	}

	if fipResp.FloatingIP.Status != "deleting" {
		if _, err := cfg.Client.Delete(ctx, client.ApiPath.FloatingIPWithID(cfg.ProjectID, d.Id()), nil); err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_floating_ip"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "created"},
		Target:     []string{"deleted"},
		Refresh:    floatingIPStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_floating_ip %s to delete: %s", d.Id(), err)
	}

	return nil
}
