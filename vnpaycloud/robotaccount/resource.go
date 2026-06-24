package robotaccount

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

func ResourceRobotAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRobotAccountCreate,
		ReadContext:   resourceRobotAccountRead,
		UpdateContext: resourceRobotAccountUpdate,
		DeleteContext: resourceRobotAccountDelete,
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
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Free-form description / label. Editable in-place — changes do not recreate the robot account.",
			},
			"permission": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"registry_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"actions": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"expires_in_days": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Days until expiry. Set -1 for never-expire, or a positive integer. Editable in-place — backend recomputes expires_at.",
			},
			"username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Full registry principal name (format: 'bot$<YYMMDD>-<random>-<name>'). Use together with 'secret' for docker login.",
			},
			"secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"expires_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
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

func resourceRobotAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateRobotAccountRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Permissions: expandPermissions(d.Get("permission").([]interface{})),
	}

	if v, ok := d.GetOk("expires_in_days"); ok {
		createOpts.ExpiresInDays = v.(int)
	}

	tflog.Debug(ctx, "vnpaycloud_registry_robot_account create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.RobotAccountResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.RobotAccounts(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_registry_robot_account: %s", err)
	}

	d.SetId(createResp.RobotAccount.ID)

	// Secret is only available at creation time — store it immediately.
	d.Set("secret", createResp.Secret)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "unknown"},
		Target:     []string{"active"},
		Refresh:    robotAccountStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.RobotAccount.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      3 * time.Second,
		MinTimeout: 2 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_registry_robot_account %s to become ready: %s", createResp.RobotAccount.ID, err)
	}

	// Read remaining fields but preserve secret from create response.
	return resourceRobotAccountReadPreserveSecret(ctx, d, meta)
}

func resourceRobotAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.RobotAccountResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.RobotAccountWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_registry_robot_account"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_registry_robot_account "+d.Id(), map[string]interface{}{"robot_account": resp.RobotAccount})

	d.Set("name", resp.RobotAccount.Name)
	d.Set("username", resp.RobotAccount.Username)
	d.Set("description", resp.RobotAccount.Description)
	d.Set("permission", flattenPermissions(resp.RobotAccount.Permissions))
	d.Set("expires_at", resp.RobotAccount.ExpiresAt)
	d.Set("expires_in_days", resp.RobotAccount.ExpiresInDays)
	d.Set("enabled", resp.RobotAccount.Enabled)
	d.Set("created_at", resp.RobotAccount.CreatedAt)
	// Note: secret is NOT returned on read — preserve existing state value.

	return nil
}

func resourceRobotAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if !d.HasChanges("description", "expires_in_days", "permission") {
		return resourceRobotAccountReadPreserveSecret(ctx, d, meta)
	}

	updateOpts := dto.UpdateRobotAccountRequest{
		Description: d.Get("description").(string),
	}
	if d.HasChange("expires_in_days") {
		updateOpts.ExpiresInDays = d.Get("expires_in_days").(int)
	}
	if d.HasChange("permission") {
		updateOpts.Permissions = expandPermissions(d.Get("permission").([]interface{}))
	}

	tflog.Debug(ctx, "vnpaycloud_registry_robot_account update options", map[string]interface{}{"update_opts": updateOpts})

	if _, err := cfg.Client.Put(ctx, client.ApiPath.RobotAccountWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil); err != nil {
		return diag.Errorf("Error updating vnpaycloud_registry_robot_account %s: %s", d.Id(), err)
	}

	return resourceRobotAccountReadPreserveSecret(ctx, d, meta)
}

// resourceRobotAccountReadPreserveSecret reads robot account but preserves
// the secret that was set during create (since the API doesn't return it on read).
func resourceRobotAccountReadPreserveSecret(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	secret := d.Get("secret").(string)
	diags := resourceRobotAccountRead(ctx, d, meta)
	if diags.HasError() {
		return diags
	}
	d.Set("secret", secret)
	return nil
}

func resourceRobotAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.RobotAccountWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_registry_robot_account"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "disabled", "unknown"},
		Target:     []string{"deleted"},
		Refresh:    robotAccountStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      3 * time.Second,
		MinTimeout: 2 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_registry_robot_account %s to delete: %s", d.Id(), err)
	}

	return nil
}

// expandPermissions converts Terraform permission blocks to DTO.
func expandPermissions(raw []interface{}) []dto.RobotAccountPermission {
	perms := make([]dto.RobotAccountPermission, 0, len(raw))
	for _, v := range raw {
		m := v.(map[string]interface{})
		actions := make([]string, 0)
		for _, a := range m["actions"].([]interface{}) {
			actions = append(actions, a.(string))
		}
		perms = append(perms, dto.RobotAccountPermission{
			RegistryID: m["registry_id"].(string),
			Actions:    actions,
		})
	}
	return perms
}

// flattenPermissions converts DTO permissions to Terraform state.
func flattenPermissions(perms []dto.RobotAccountPermission) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(perms))
	for _, p := range perms {
		result = append(result, map[string]interface{}{
			"registry_id": p.RegistryID,
			"actions":     p.Actions,
		})
	}
	return result
}
