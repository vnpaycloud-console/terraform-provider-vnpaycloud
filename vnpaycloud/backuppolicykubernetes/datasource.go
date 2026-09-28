package backuppolicykubernetes

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
		"description":           {Type: schema.TypeString, Computed: true},
		"auto_apply_new_volume": {Type: schema.TypeBool, Computed: true},
		"is_auto":               {Type: schema.TypeBool, Computed: true},
		"start_hour":            {Type: schema.TypeInt, Computed: true},
		"run_priority":          {Type: schema.TypeInt, Computed: true},
		"purpose":               {Type: schema.TypeString, Computed: true},
		"backup_vault_ids":      {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
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
		"status":     {Type: schema.TypeString, Computed: true},
		"created_at": {Type: schema.TypeString, Computed: true},
	}
}

func DataSourceBackupPolicyKubernetes() *schema.Resource {
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
		ReadContext: dataSourceBackupPolicyKubernetesRead,
		Schema:      s,
	}
}

func dataSourceBackupPolicyKubernetesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok && id.(string) != "" {
		resp := &dto.BackupPolicyKubernetesResponse{}
		if _, err := cfg.Client.Get(ctx, client.ApiPath.BackupPolicyKubernetesWithID(cfg.ProjectID, id.(string)), resp, nil); err != nil {
			return diag.Errorf("Error fetching vnpaycloud_backup_policy_kubernetes %s: %s", id, err)
		}
		if v, ok := d.GetOk("name"); ok && resp.BackupPolicyKubernetes.Name != v.(string) {
			return diag.Errorf("vnpaycloud_backup_policy_kubernetes %q does not match name %q", id.(string), v.(string))
		}
		d.SetId(resp.BackupPolicyKubernetes.ID)
		d.Set("name", resp.BackupPolicyKubernetes.Name)
		setBackupPolicyKubernetesData(d, &resp.BackupPolicyKubernetes)
		return nil
	}

	name := d.Get("name").(string)
	listResp := &dto.ListBackupPolicyKubernetesResponse{}
	path := client.ApiPath.BackupPolicyKuberneteses(cfg.ProjectID) + "?name=" + url.QueryEscape(name)
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_backup_policy_kubernetes: %s", err)
	}

	var matched []dto.BackupPolicyKubernetes
	for _, p := range listResp.BackupPolicyKubernetes {
		if p.Name == name {
			matched = append(matched, p)
		}
	}

	if len(matched) == 0 {
		return diag.Errorf("No vnpaycloud_backup_policy_kubernetes found with name %q", name)
	}
	if len(matched) > 1 {
		return diag.Errorf("Found %d vnpaycloud_backup_policy_kubernetes matching name %q; use id to select one", len(matched), name)
	}

	d.SetId(matched[0].ID)
	setBackupPolicyKubernetesData(d, &matched[0])
	return nil
}

func DataSourceBackupPolicyKubernetesList() *schema.Resource {
	elem := computedPolicyAttrs()
	elem["id"] = &schema.Schema{Type: schema.TypeString, Computed: true}
	elem["name"] = &schema.Schema{Type: schema.TypeString, Computed: true}

	return &schema.Resource{
		ReadContext: dataSourceBackupPolicyKubernetesListRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"backup_policy_kubernetes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Resource{Schema: elem},
			},
		},
	}
}

func dataSourceBackupPolicyKubernetesListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	path := client.ApiPath.BackupPolicyKuberneteses(cfg.ProjectID)
	if v, ok := d.GetOk("name"); ok {
		path += "?name=" + url.QueryEscape(v.(string))
	}

	listResp := &dto.ListBackupPolicyKubernetesResponse{}
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Unable to query vnpaycloud_backup_policy_kubernetes_list: %s", err)
	}

	policies := make([]map[string]interface{}, 0, len(listResp.BackupPolicyKubernetes))
	for i := range listResp.BackupPolicyKubernetes {
		policies = append(policies, backupPolicyKubernetesAttrs(&listResp.BackupPolicyKubernetes[i]))
	}

	d.SetId(cfg.ProjectID)
	d.Set("backup_policy_kubernetes", policies)

	return nil
}
