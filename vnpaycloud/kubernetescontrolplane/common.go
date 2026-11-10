package kubernetescontrolplane

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func getBaseline(ctx context.Context, cfg *config.Config, clusterID string) (*dto.ClusterControlPlaneBaselineResponse, error) {
	resp := &dto.ClusterControlPlaneBaselineResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.ClusterControlPlaneBaseline(cfg.ProjectID, clusterID), resp, nil); err != nil {
		return nil, err
	}
	return resp, nil
}

func baselineComputedSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"supported":             {Type: schema.TypeBool, Computed: true},
		"unsupported_reason":    {Type: schema.TypeString, Computed: true},
		"rolling_update_needed": {Type: schema.TypeBool, Computed: true},
		"outdated_items":        {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
		"control_plane_rolling": {Type: schema.TypeBool, Computed: true},
		"audit_log_enabled":     {Type: schema.TypeBool, Computed: true},
		"audit_log_ready":       {Type: schema.TypeBool, Computed: true},
	}
}

func setBaselineData(d *schema.ResourceData, b *dto.ClusterControlPlaneBaselineResponse) {
	d.Set("supported", b.Supported)
	d.Set("unsupported_reason", b.UnsupportedReason)
	d.Set("rolling_update_needed", b.RollingUpdateNeeded)
	d.Set("outdated_items", b.OutdatedItems)
	d.Set("control_plane_rolling", b.ControlPlaneRolling)
	d.Set("audit_log_enabled", b.AuditLogEnabled)
	d.Set("audit_log_ready", b.AuditLogReady)
}
