package backupserver

import (
	"terraform-provider-vnpaycloud/vnpaycloud/dto"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandStringList(v interface{}) []string {
	l, ok := v.([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(l))
	for _, item := range l {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func setBackupServerData(d *schema.ResourceData, s *dto.BackupServer) {
	d.Set("name", s.Name)
	d.Set("server_id", s.ServerID)
	d.Set("volume_ids", s.VolumeIDs)
	d.Set("backup_policy_id", s.BackupPolicyID)
	d.Set("purpose", s.Purpose)
	d.Set("zone_id", s.ZoneID)
	d.Set("is_compliant", s.IsCompliant)
	d.Set("compliance_msg", s.ComplianceMsg)
	d.Set("status", s.Status)
	d.Set("created_at", s.CreatedAt)
}

func backupServerAttrs(s *dto.BackupServer) map[string]interface{} {
	return map[string]interface{}{
		"id":               s.ID,
		"name":             s.Name,
		"server_id":        s.ServerID,
		"volume_ids":       s.VolumeIDs,
		"backup_policy_id": s.BackupPolicyID,
		"purpose":          s.Purpose,
		"zone_id":          s.ZoneID,
		"is_compliant":     s.IsCompliant,
		"compliance_msg":   s.ComplianceMsg,
		"status":           s.Status,
		"created_at":       s.CreatedAt,
	}
}
