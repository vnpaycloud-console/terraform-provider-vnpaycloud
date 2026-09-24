package backupserver

import (
	"context"
	"net/url"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func computedBackupServerAttrs() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"server_id":        {Type: schema.TypeString, Computed: true},
		"volume_ids":       {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
		"backup_policy_id": {Type: schema.TypeString, Computed: true},
		"purpose":          {Type: schema.TypeString, Computed: true},
		"zone_id":          {Type: schema.TypeString, Computed: true},
		"is_compliant":     {Type: schema.TypeString, Computed: true},
		"compliance_msg":   {Type: schema.TypeString, Computed: true},
		"status":           {Type: schema.TypeString, Computed: true},
		"created_at":       {Type: schema.TypeString, Computed: true},
	}
}

func DataSourceBackupServer() *schema.Resource {
	s := computedBackupServerAttrs()
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
		ReadContext: dataSourceBackupServerRead,
		Schema:      s,
	}
}

func dataSourceBackupServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok && id.(string) != "" {
		resp := &dto.BackupServerResponse{}
		if _, err := cfg.Client.Get(ctx, client.ApiPath.BackupServerWithID(cfg.ProjectID, id.(string)), resp, nil); err != nil {
			return diag.Errorf("Error fetching vnpaycloud_backup_server %s: %s", id, err)
		}
		if v, ok := d.GetOk("name"); ok && resp.BackupServer.Name != v.(string) {
			return diag.Errorf("vnpaycloud_backup_server %q does not match name %q", id.(string), v.(string))
		}
		d.SetId(resp.BackupServer.ID)
		setBackupServerData(d, &resp.BackupServer)
		return nil
	}

	name := d.Get("name").(string)
	listResp := &dto.ListBackupServersResponse{}
	path := client.ApiPath.BackupServers(cfg.ProjectID) + "?name=" + url.QueryEscape(name)
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_backup_server: %s", err)
	}

	var matched []dto.BackupServer
	for _, s := range listResp.BackupServers {
		if s.Name == name {
			matched = append(matched, s)
		}
	}

	if len(matched) == 0 {
		return diag.Errorf("No vnpaycloud_backup_server found with name %q", name)
	}
	if len(matched) > 1 {
		return diag.Errorf("Found %d vnpaycloud_backup_server matching name %q; use id to select one", len(matched), name)
	}

	d.SetId(matched[0].ID)
	setBackupServerData(d, &matched[0])
	return nil
}

func DataSourceBackupServers() *schema.Resource {
	elem := computedBackupServerAttrs()
	elem["id"] = &schema.Schema{Type: schema.TypeString, Computed: true}
	elem["name"] = &schema.Schema{Type: schema.TypeString, Computed: true}

	return &schema.Resource{
		ReadContext: dataSourceBackupServersRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"server_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"purpose": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"on_demand"}, false),
			},
			"backup_servers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Resource{Schema: elem},
			},
		},
	}
}

func dataSourceBackupServersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	path := client.ApiPath.BackupServers(cfg.ProjectID)
	q := url.Values{}
	if v, ok := d.GetOk("name"); ok {
		q.Set("name", v.(string))
	}
	if v, ok := d.GetOk("purpose"); ok {
		q.Set("purpose", v.(string))
	}
	for _, id := range expandStringList(d.Get("server_ids")) {
		q.Add("serverIds", id)
	}
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	listResp := &dto.ListBackupServersResponse{}
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Unable to query vnpaycloud_backup_servers: %s", err)
	}

	servers := make([]map[string]interface{}, 0, len(listResp.BackupServers))
	for i := range listResp.BackupServers {
		servers = append(servers, backupServerAttrs(&listResp.BackupServers[i]))
	}

	d.SetId(cfg.ProjectID)
	d.Set("backup_servers", servers)

	return nil
}
