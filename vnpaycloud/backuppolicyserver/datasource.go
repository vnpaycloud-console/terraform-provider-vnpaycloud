package backuppolicyserver

import (
	"context"
	"net/url"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func scheduleBlockSchema(fields map[string]*schema.Schema) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem:     &schema.Resource{Schema: fields},
	}
}

func computedPolicyAttrs() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"description":              {Type: schema.TypeString, Computed: true},
		"start_hour":               {Type: schema.TypeInt, Computed: true},
		"resource_type":            {Type: schema.TypeString, Computed: true},
		"purpose":                  {Type: schema.TypeString, Computed: true},
		"is_auto":                  {Type: schema.TypeBool, Computed: true},
		"is_auto_apply_for_volume": {Type: schema.TypeBool, Computed: true},
		"backup_vault_ids":         {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
		"daily": scheduleBlockSchema(map[string]*schema.Schema{
			"retentions": {Type: schema.TypeInt, Computed: true},
		}),
		"weekly": scheduleBlockSchema(map[string]*schema.Schema{
			"retentions":  {Type: schema.TypeInt, Computed: true},
			"day_of_week": {Type: schema.TypeInt, Computed: true},
		}),
		"monthly": scheduleBlockSchema(map[string]*schema.Schema{
			"retentions":   {Type: schema.TypeInt, Computed: true},
			"day_type":     {Type: schema.TypeString, Computed: true},
			"day_of_month": {Type: schema.TypeInt, Computed: true},
			"day_of_week":  {Type: schema.TypeInt, Computed: true},
		}),
		"second_tier_enabled": {Type: schema.TypeBool, Computed: true},
		"second_tier_weekly": scheduleBlockSchema(map[string]*schema.Schema{
			"retentions":  {Type: schema.TypeInt, Computed: true},
			"day_of_week": {Type: schema.TypeInt, Computed: true},
		}),
		"second_tier_monthly": scheduleBlockSchema(map[string]*schema.Schema{
			"retentions":   {Type: schema.TypeInt, Computed: true},
			"day_type":     {Type: schema.TypeString, Computed: true},
			"day_of_month": {Type: schema.TypeInt, Computed: true},
			"day_of_week":  {Type: schema.TypeInt, Computed: true},
		}),
		"second_tier_yearly": scheduleBlockSchema(map[string]*schema.Schema{
			"retentions":   {Type: schema.TypeInt, Computed: true},
			"month":        {Type: schema.TypeInt, Computed: true},
			"day_of_month": {Type: schema.TypeInt, Computed: true},
		}),
		"protected_servers":      {Type: schema.TypeInt, Computed: true},
		"protected_volumes":      {Type: schema.TypeInt, Computed: true},
		"protected_volume_sizes": {Type: schema.TypeInt, Computed: true},
		"compliance_state":       {Type: schema.TypeString, Computed: true},
		"compliance_msg":         {Type: schema.TypeString, Computed: true},
		"status":                 {Type: schema.TypeString, Computed: true},
		"created_at":             {Type: schema.TypeString, Computed: true},
	}
}

func DataSourceBackupPolicyServer() *schema.Resource {
	s := computedPolicyAttrs()
	s["id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		AtLeastOneOf: []string{"id", "name"},
	}
	s["name"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		AtLeastOneOf: []string{"id", "name"},
	}

	return &schema.Resource{
		ReadContext: dataSourceBackupPolicyServerRead,
		Schema:      s,
	}
}

func dataSourceBackupPolicyServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok && id.(string) != "" {
		resp := &dto.BackupPolicyServerResponse{}
		if _, err := cfg.Client.Get(ctx, client.ApiPath.BackupPolicyServerWithID(cfg.ProjectID, id.(string)), resp, nil); err != nil {
			return diag.Errorf("Error fetching vnpaycloud_backup_policy_server %s: %s", id, err)
		}
		if v, ok := d.GetOk("name"); ok && resp.BackupPolicyServer.Name != v.(string) {
			return diag.Errorf("vnpaycloud_backup_policy_server %q does not match name %q", id.(string), v.(string))
		}
		d.SetId(resp.BackupPolicyServer.ID)
		d.Set("name", resp.BackupPolicyServer.Name)
		setBackupPolicyServerData(d, &resp.BackupPolicyServer)
		return nil
	}

	name := d.Get("name").(string)
	listResp := &dto.ListBackupPolicyServersResponse{}
	path := client.ApiPath.BackupPolicyServers(cfg.ProjectID) + "?name=" + url.QueryEscape(name)
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_backup_policy_server: %s", err)
	}

	var matched []dto.BackupPolicyServer
	for _, p := range listResp.BackupPolicyServers {
		if p.Name == name {
			matched = append(matched, p)
		}
	}

	if len(matched) == 0 {
		return diag.Errorf("No vnpaycloud_backup_policy_server found with name %q", name)
	}
	if len(matched) > 1 {
		return diag.Errorf("Found %d vnpaycloud_backup_policy_server matching name %q; use id to select one", len(matched), name)
	}

	d.SetId(matched[0].ID)
	setBackupPolicyServerData(d, &matched[0])
	return nil
}

func DataSourceBackupPolicyServers() *schema.Resource {
	elem := computedPolicyAttrs()
	elem["id"] = &schema.Schema{Type: schema.TypeString, Computed: true}
	elem["name"] = &schema.Schema{Type: schema.TypeString, Computed: true}

	return &schema.Resource{
		ReadContext: dataSourceBackupPolicyServersRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"backup_policy_servers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Resource{Schema: elem},
			},
		},
	}
}

func dataSourceBackupPolicyServersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	path := client.ApiPath.BackupPolicyServers(cfg.ProjectID)
	if v, ok := d.GetOk("name"); ok {
		path += "?name=" + url.QueryEscape(v.(string))
	}

	listResp := &dto.ListBackupPolicyServersResponse{}
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Unable to query vnpaycloud_backup_policy_servers: %s", err)
	}

	policies := make([]map[string]interface{}, 0, len(listResp.BackupPolicyServers))
	for i := range listResp.BackupPolicyServers {
		policies = append(policies, backupPolicyServerAttrs(&listResp.BackupPolicyServers[i]))
	}

	d.SetId(cfg.ProjectID)
	d.Set("backup_policy_servers", policies)

	return nil
}
