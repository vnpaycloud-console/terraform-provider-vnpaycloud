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
		DeleteContext: resourceRobotAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"registry_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"permissions": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"expires_in_days": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
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

	registryID := d.Get("registry_id").(string)

	createOpts := dto.CreateRobotAccountRequest{
		Name: d.Get("name").(string),
	}

	if v, ok := d.GetOk("permissions"); ok {
		for _, p := range v.([]interface{}) {
			createOpts.Permissions = append(createOpts.Permissions, p.(string))
		}
	}

	if v, ok := d.GetOk("expires_in_days"); ok {
		createOpts.ExpiresInDays = v.(int)
	}

	tflog.Debug(ctx, "vnpaycloud_registry_robot_account create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.RobotAccountResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.RobotAccounts(cfg.ProjectID, registryID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_registry_robot_account: %s", err)
	}

	d.SetId(createResp.RobotAccount.ID)

	// Secret is only available at creation time — store it immediately.
	d.Set("secret", createResp.Secret)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "unknown"},
		Target:     []string{"active"},
		Refresh:    robotAccountStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, registryID, createResp.RobotAccount.ID),
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

	registryID := d.Get("registry_id").(string)

	resp := &dto.RobotAccountResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.RobotAccountWithID(cfg.ProjectID, registryID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_registry_robot_account"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_registry_robot_account "+d.Id(), map[string]interface{}{"robot_account": resp.RobotAccount})

	// Note: registry_id and name are NOT updated from response because the server
	// transforms them (registry_id: name→Vertix ID, name: simple→full harbor bot name).
	// Keeping user-provided values prevents perpetual drift.
	d.Set("permissions", resp.RobotAccount.Permissions)
	d.Set("expires_at", resp.RobotAccount.ExpiresAt)
	d.Set("enabled", resp.RobotAccount.Enabled)
	d.Set("created_at", resp.RobotAccount.CreatedAt)
	// Note: secret is NOT returned on read — preserve existing state value.

	return nil
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

	registryID := d.Get("registry_id").(string)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.RobotAccountWithID(cfg.ProjectID, registryID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_registry_robot_account"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "disabled", "unknown"},
		Target:     []string{"deleted"},
		Refresh:    robotAccountStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, registryID, d.Id()),
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
