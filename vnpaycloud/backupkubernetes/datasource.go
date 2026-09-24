package backupkubernetes

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceBackupKubernetes() *schema.Resource {
	s := computedBackupKubernetesAttrs()
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
		ReadContext: dataSourceBackupKubernetesRead,
		Schema:      s,
	}
}

func dataSourceBackupKubernetesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok && id.(string) != "" {
		resp := &dto.BackupKubernetesResponse{}
		if _, err := cfg.Client.Get(ctx, client.ApiPath.BackupKubernetesWithID(cfg.ProjectID, id.(string)), resp, nil); err != nil {
			return diag.Errorf("Error fetching vnpaycloud_backup_kubernetes %s: %s", id, err)
		}
		if v, ok := d.GetOk("name"); ok && resp.BackupKubernetes.Name != v.(string) {
			return diag.Errorf("vnpaycloud_backup_kubernetes %q does not match name %q", id.(string), v.(string))
		}
		d.SetId(resp.BackupKubernetes.ID)
		setBackupKubernetesData(d, &resp.BackupKubernetes)
		return nil
	}

	name := d.Get("name").(string)
	listResp := &dto.ListBackupKubernetesResponse{}
	path := client.ApiPath.BackupKuberneteses(cfg.ProjectID) + "?name=" + url.QueryEscape(name)
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_backup_kubernetes: %s", err)
	}

	var matched []dto.BackupKubernetes
	for _, bk := range listResp.BackupKubernetes {
		if bk.Name == name {
			matched = append(matched, bk)
		}
	}

	if len(matched) == 0 {
		return diag.Errorf("No vnpaycloud_backup_kubernetes found with name %q", name)
	}
	if len(matched) > 1 {
		return diag.Errorf("Found %d vnpaycloud_backup_kubernetes matching name %q; use id to select one", len(matched), name)
	}

	d.SetId(matched[0].ID)
	setBackupKubernetesData(d, &matched[0])
	return nil
}

func computedBackupKubernetesAttrs() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id":                   {Type: schema.TypeString, Computed: true},
		"name":                 {Type: schema.TypeString, Computed: true},
		"description":          {Type: schema.TypeString, Computed: true},
		"protected_cluster_id": {Type: schema.TypeString, Computed: true},
		"pvc_backup_policy_id": {Type: schema.TypeString, Computed: true},
		"cluster_backup_type":  {Type: schema.TypeString, Computed: true},
		"zone_id":              {Type: schema.TypeString, Computed: true},
		"project_id":           {Type: schema.TypeString, Computed: true},
		"customer_username":    {Type: schema.TypeString, Computed: true},
		"status":               {Type: schema.TypeString, Computed: true},
		"created_at":           {Type: schema.TypeString, Computed: true},
	}
}

func DataSourceBackupKubernetesList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBackupKubernetesListRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"protected_cluster_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"pvc_backup_policy_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"backup_kubernetes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Resource{Schema: computedBackupKubernetesAttrs()},
			},
		},
	}
}

func dataSourceBackupKubernetesListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	path := client.ApiPath.BackupKuberneteses(cfg.ProjectID)
	q := url.Values{}
	for _, key := range []string{"name", "protected_cluster_id", "pvc_backup_policy_id"} {
		if v, ok := d.GetOk(key); ok {
			q.Set(toCamel(key), v.(string))
		}
	}
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	listResp := &dto.ListBackupKubernetesResponse{}
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_backup_kubernetes_list: %s", err)
	}

	items := make([]map[string]interface{}, 0, len(listResp.BackupKubernetes))
	for _, bk := range listResp.BackupKubernetes {
		items = append(items, map[string]interface{}{
			"id":                   bk.ID,
			"name":                 bk.Name,
			"description":          bk.Description,
			"protected_cluster_id": bk.ProtectedClusterID,
			"pvc_backup_policy_id": bk.PVCBackupPolicyID,
			"cluster_backup_type":  bk.ClusterBackupType,
			"zone_id":              bk.ZoneID,
			"project_id":           bk.ProjectID,
			"customer_username":    bk.CustomerUsername,
			"status":               bk.Status,
			"created_at":           bk.CreatedAt,
		})
	}

	if err := d.Set("backup_kubernetes", items); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.Itoa(len(items)) + "-backup-kubernetes")
	return nil
}

// toCamel maps a snake_case filter name to the camelCase query parameter the
// gateway expects.
func toCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, "")
}
