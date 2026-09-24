package backuppolicykubernetes

import (
	"fmt"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var monthlyDayTypes = []string{
	"fixed_day",
	"first_week",
	"second_week",
	"third_week",
	"fourth_week",
	"fifth_week",
}

func validateMonthlyDayType(d *schema.ResourceDiff) error {
	m, ok := firstBlock(d.Get("monthly"))
	if !ok {
		return nil
	}
	dayType, _ := m["day_type"].(string)
	dayOfMonth, _ := m["day_of_month"].(int)
	dayOfWeek, _ := m["day_of_week"].(int)

	if dayType == "fixed_day" {
		if dayOfMonth < 1 || dayOfMonth > 31 {
			return fmt.Errorf(`monthly.day_of_month is required (1-31) when day_type is "fixed_day"`)
		}
		if dayOfWeek != 0 {
			return fmt.Errorf(`monthly.day_of_week must not be set when day_type is "fixed_day"`)
		}
		return nil
	}

	if dayOfMonth != 0 {
		return fmt.Errorf(`monthly.day_of_month must not be set when day_type is %q (use day_of_week)`, dayType)
	}
	return nil
}

func expandDaily(v interface{}) *dto.BackupPolicyKubernetesDaily {
	m, ok := firstBlock(v)
	if !ok {
		return nil
	}
	return &dto.BackupPolicyKubernetesDaily{Retentions: m["retentions"].(int)}
}

func expandWeekly(v interface{}) *dto.BackupPolicyKubernetesWeekly {
	m, ok := firstBlock(v)
	if !ok {
		return nil
	}
	return &dto.BackupPolicyKubernetesWeekly{
		Retentions: m["retentions"].(int),
		DayOfWeek:  m["day_of_week"].(int),
	}
}

func expandMonthly(v interface{}) *dto.BackupPolicyKubernetesMonthly {
	m, ok := firstBlock(v)
	if !ok {
		return nil
	}
	return &dto.BackupPolicyKubernetesMonthly{
		Retentions: m["retentions"].(int),
		DayType:    m["day_type"].(string),
		DayOfMonth: m["day_of_month"].(int),
		DayOfWeek:  m["day_of_week"].(int),
	}
}

func firstBlock(v interface{}) (map[string]interface{}, bool) {
	l, ok := v.([]interface{})
	if !ok || len(l) == 0 || l[0] == nil {
		return nil, false
	}
	m, ok := l[0].(map[string]interface{})
	return m, ok
}

func flattenDaily(d *dto.BackupPolicyKubernetesDaily) []interface{} {
	if d == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{"retentions": d.Retentions}}
}

func flattenWeekly(w *dto.BackupPolicyKubernetesWeekly) []interface{} {
	if w == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		"retentions":  w.Retentions,
		"day_of_week": w.DayOfWeek,
	}}
}

func flattenMonthly(m *dto.BackupPolicyKubernetesMonthly) []interface{} {
	if m == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		"retentions":   m.Retentions,
		"day_type":     m.DayType,
		"day_of_month": m.DayOfMonth,
		"day_of_week":  m.DayOfWeek,
	}}
}

func setBackupPolicyKubernetesData(d *schema.ResourceData, p *dto.BackupPolicyKubernetes) {
	d.Set("name", p.Name)
	d.Set("description", p.Description)
	d.Set("auto_apply_new_volume", p.AutoApplyNewVolume)
	d.Set("is_auto", p.IsAuto)
	d.Set("start_hour", p.StartHour)
	d.Set("run_priority", p.RunPriority)
	d.Set("purpose", p.Purpose)
	d.Set("backup_vault_ids", p.BackupVaultIDs)
	d.Set("daily", flattenDaily(p.Daily))
	d.Set("weekly", flattenWeekly(p.Weekly))
	d.Set("monthly", flattenMonthly(p.Monthly))
	d.Set("status", p.Status)
	d.Set("created_at", p.CreatedAt)
}

func backupPolicyKubernetesAttrs(p *dto.BackupPolicyKubernetes) map[string]interface{} {
	return map[string]interface{}{
		"id":                    p.ID,
		"name":                  p.Name,
		"description":           p.Description,
		"auto_apply_new_volume": p.AutoApplyNewVolume,
		"is_auto":               p.IsAuto,
		"start_hour":            p.StartHour,
		"run_priority":          p.RunPriority,
		"purpose":               p.Purpose,
		"backup_vault_ids":      p.BackupVaultIDs,
		"daily":                 flattenDaily(p.Daily),
		"weekly":                flattenWeekly(p.Weekly),
		"monthly":               flattenMonthly(p.Monthly),
		"status":                p.Status,
		"created_at":            p.CreatedAt,
	}
}
