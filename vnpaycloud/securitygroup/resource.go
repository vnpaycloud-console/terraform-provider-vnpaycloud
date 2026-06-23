package securitygroup

import (
	"context"
	"fmt"
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

func ResourceSecurityGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecurityGroupCreate,
		ReadContext:   resourceSecurityGroupRead,
		UpdateContext: resourceSecurityGroupUpdate,
		DeleteContext: resourceSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			if d.Id() == "" {
				return nil
			}
			if d.HasChange("enable_log") && !d.Get("can_enable_log").(bool) {
				return fmt.Errorf("enable_log cannot be changed: network logging is not supported for this security group in this zone (can_enable_log = false)")
			}
			return nil
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rules": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":                {Type: schema.TypeString, Computed: true},
						"security_group_id": {Type: schema.TypeString, Computed: true},
						"direction":         {Type: schema.TypeString, Computed: true},
						"protocol":          {Type: schema.TypeString, Computed: true},
						"ethertype":         {Type: schema.TypeString, Computed: true},
						"port_range_min":    {Type: schema.TypeInt, Computed: true},
						"port_range_max":    {Type: schema.TypeInt, Computed: true},
						"remote_ip_prefix":  {Type: schema.TypeString, Computed: true},
					},
				},
			},
			"enable_log": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"can_enable_log": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSecurityGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateSecurityGroupRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_security_group create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.SecurityGroupResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.SecurityGroups(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_security_group: %s", err)
	}

	d.SetId(createResp.SecurityGroup.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating", "build", "building", "unknown"},
		Target:     []string{"active"},
		Refresh:    securityGroupStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.SecurityGroup.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_security_group %s to become active: %s", createResp.SecurityGroup.ID, err)
	}

	if d.Get("enable_log").(bool) {
		sg, err := readSecurityGroup(ctx, cfg, d.Id())
		if err != nil {
			return diag.Errorf("Error reading vnpaycloud_security_group %s after create: %s", d.Id(), err)
		}
		if !sg.CanEnableLog {
			if _, delErr := cfg.Client.Delete(ctx, client.ApiPath.SecurityGroupWithID(cfg.ProjectID, d.Id()), nil); delErr != nil {
				tflog.Warn(ctx, "failed to roll back vnpaycloud_security_group after enable_log validation error", map[string]interface{}{"id": d.Id(), "error": delErr.Error()})
			}
			id := d.Id()
			d.SetId("")
			return diag.Errorf("Error creating vnpaycloud_security_group %s: enable_log = true but network logging is not supported for this security group in this zone (can_enable_log = false)", id)
		}
		if err := setSecurityGroupLog(ctx, cfg, d.Id(), true); err != nil {
			return diag.Errorf("Error enabling network logging for vnpaycloud_security_group %s after create: %s", d.Id(), err)
		}
	}

	return resourceSecurityGroupRead(ctx, d, meta)
}

func resourceSecurityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	sgResp := &dto.SecurityGroupResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.SecurityGroupWithID(cfg.ProjectID, d.Id()), sgResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_security_group"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_security_group "+d.Id(), map[string]interface{}{"security_group": sgResp.SecurityGroup})

	d.Set("name", sgResp.SecurityGroup.Name)
	d.Set("description", sgResp.SecurityGroup.Description)
	d.Set("enable_log", sgResp.SecurityGroup.EnableLog)
	d.Set("can_enable_log", sgResp.SecurityGroup.CanEnableLog)
	d.Set("created_at", sgResp.SecurityGroup.CreatedAt)

	rules := make([]map[string]interface{}, len(sgResp.SecurityGroup.Rules))
	for i, r := range sgResp.SecurityGroup.Rules {
		rules[i] = map[string]interface{}{
			"id":                r.ID,
			"security_group_id": r.SecurityGroupID,
			"direction":         r.Direction,
			"protocol":          r.Protocol,
			"ethertype":         r.EtherType,
			"port_range_min":    r.PortRangeMin,
			"port_range_max":    r.PortRangeMax,
			"remote_ip_prefix":  r.RemoteIPPrefix,
		}
	}
	d.Set("rules", rules)

	return nil
}

func resourceSecurityGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChanges("name", "description") {
		if err := updateSecurityGroup(ctx, d, cfg); err != nil {
			return diag.Errorf("Error updating vnpaycloud_security_group %s: %s", d.Id(), err)
		}
	}

	if d.HasChange("enable_log") && d.Get("can_enable_log").(bool) {
		if err := setSecurityGroupLog(ctx, cfg, d.Id(), d.Get("enable_log").(bool)); err != nil {
			return diag.Errorf("Error updating network logging for vnpaycloud_security_group %s: %s", d.Id(), err)
		}
	}

	return resourceSecurityGroupRead(ctx, d, meta)
}

func updateSecurityGroup(ctx context.Context, d *schema.ResourceData, cfg *config.Config) error {
	updateOpts := dto.UpdateSecurityGroupRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_security_group update options", map[string]interface{}{"update_opts": updateOpts})

	_, err := cfg.Client.Put(ctx, client.ApiPath.SecurityGroupWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
	return err
}

func readSecurityGroup(ctx context.Context, cfg *config.Config, id string) (*dto.SecurityGroup, error) {
	sgResp := &dto.SecurityGroupResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.SecurityGroupWithID(cfg.ProjectID, id), sgResp, nil)
	if err != nil {
		return nil, err
	}
	return &sgResp.SecurityGroup, nil
}

func setSecurityGroupLog(ctx context.Context, cfg *config.Config, id string, enable bool) error {
	logOpts := dto.UpdateSecurityGroupLogRequest{EnableLog: enable}

	tflog.Debug(ctx, "vnpaycloud_security_group log options", map[string]interface{}{"log_opts": logOpts})

	_, err := cfg.Client.Put(ctx, client.ApiPath.SecurityGroupLog(cfg.ProjectID, id), logOpts, nil, nil)
	return err
}

func resourceSecurityGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.SecurityGroupWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_security_group"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active"},
		Target:     []string{"deleted"},
		Refresh:    securityGroupStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_security_group %s to delete: %s", d.Id(), err)
	}

	return nil
}
