package backuppolicyserver

import (
	"fmt"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var monthlyDayTypes = []string{
	"day_of_month",
	"day_of_1st_week",
	"day_of_2nd_week",
	"day_of_3rd_week",
	"day_of_4th_week",
	"day_of_5th_week",
}

func validateMonthlyBlock(field string, v interface{}, weekdayZeroIsValid bool) error {
	m, ok := firstBlock(v)
	if !ok {
		return nil
	}
	dayType, _ := m["day_type"].(string)
	dayOfMonth, _ := m["day_of_month"].(int)
	dayOfWeek, _ := m["day_of_week"].(int)

	if dayType == "day_of_month" {
		if dayOfMonth < 1 || dayOfMonth > 31 {
			return fmt.Errorf(`%s.day_of_month is required (1-31) when day_type is "day_of_month"`, field)
		}
		if dayOfWeek != 0 {
			return fmt.Errorf(`%s.day_of_week must not be set when day_type is "day_of_month"`, field)
		}
		return nil
	}

	if dayOfMonth != 0 {
		return fmt.Errorf(`%s.day_of_month must not be set when day_type is %q (use day_of_week)`, field, dayType)
	}
	if !weekdayZeroIsValid && (dayOfWeek < 1 || dayOfWeek > 7) {
		return fmt.Errorf(`%s.day_of_week is required (1-7) when day_type is %q`, field, dayType)
	}
	return nil
}

func expandDaily(v interface{}) *dto.BackupPolicyServerDaily {
	m, ok := firstBlock(v)
	if !ok {
		return nil
	}
	return &dto.BackupPolicyServerDaily{Retentions: m["retentions"].(int)}
}

func expandWeekly(v interface{}) *dto.BackupPolicyServerWeekly {
	m, ok := firstBlock(v)
	if !ok {
		return nil
	}
	return &dto.BackupPolicyServerWeekly{
		Retentions: m["retentions"].(int),
		DayOfWeek:  m["day_of_week"].(int),
	}
}

func expandMonthly(v interface{}) *dto.BackupPolicyServerMonthly {
	m, ok := firstBlock(v)
	if !ok {
		return nil
	}
	return &dto.BackupPolicyServerMonthly{
		Retentions: m["retentions"].(int),
		DayType:    m["day_type"].(string),
		DayOfMonth: m["day_of_month"].(int),
		DayOfWeek:  m["day_of_week"].(int),
	}
}

func expandSecondTierWeekly(v interface{}) *dto.BackupPolicyServerSecondTierWeekly {
	m, ok := firstBlock(v)
	if !ok {
		return nil
	}
	return &dto.BackupPolicyServerSecondTierWeekly{
		Retentions: m["retentions"].(int),
		DayOfWeek:  m["day_of_week"].(int),
	}
}

func expandSecondTierMonthly(v interface{}) *dto.BackupPolicyServerSecondTierMonthly {
	m, ok := firstBlock(v)
	if !ok {
		return nil
	}
	return &dto.BackupPolicyServerSecondTierMonthly{
		Retentions: m["retentions"].(int),
		DayType:    m["day_type"].(string),
		DayOfMonth: m["day_of_month"].(int),
		DayOfWeek:  m["day_of_week"].(int),
	}
}

func expandSecondTierYearly(v interface{}) *dto.BackupPolicyServerSecondTierYearly {
	m, ok := firstBlock(v)
	if !ok {
		return nil
	}
	return &dto.BackupPolicyServerSecondTierYearly{
		Retentions: m["retentions"].(int),
		Month:      m["month"].(int),
		DayOfMonth: m["day_of_month"].(int),
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

func flattenDaily(d *dto.BackupPolicyServerDaily) []interface{} {
	if d == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{"retentions": d.Retentions}}
}

func flattenWeekly(w *dto.BackupPolicyServerWeekly) []interface{} {
	if w == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		"retentions":  w.Retentions,
		"day_of_week": w.DayOfWeek,
	}}
}

func flattenMonthly(m *dto.BackupPolicyServerMonthly) []interface{} {
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

func flattenSecondTierWeekly(w *dto.BackupPolicyServerSecondTierWeekly) []interface{} {
	if w == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		"retentions":  w.Retentions,
		"day_of_week": w.DayOfWeek,
	}}
}

func flattenSecondTierMonthly(m *dto.BackupPolicyServerSecondTierMonthly) []interface{} {
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

func flattenSecondTierYearly(y *dto.BackupPolicyServerSecondTierYearly) []interface{} {
	if y == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		"retentions":   y.Retentions,
		"month":        y.Month,
		"day_of_month": y.DayOfMonth,
	}}
}

func setBackupPolicyServerData(d *schema.ResourceData, p *dto.BackupPolicyServer) {
	d.Set("name", p.Name)
	d.Set("description", p.Description)
	d.Set("start_hour", p.StartHour)
	d.Set("resource_type", p.ResourceType)
	d.Set("purpose", p.Purpose)
	d.Set("is_auto", p.IsAuto)
	d.Set("is_auto_apply_for_volume", p.IsAutoApplyForVolume)
	d.Set("backup_vault_ids", p.BackupVaultIDs)
	d.Set("daily", flattenDaily(p.Daily))
	d.Set("weekly", flattenWeekly(p.Weekly))
	d.Set("monthly", flattenMonthly(p.Monthly))
	d.Set("second_tier_enabled", p.SecondTierEnabled)
	d.Set("second_tier_weekly", flattenSecondTierWeekly(p.SecondTierWeekly))
	d.Set("second_tier_monthly", flattenSecondTierMonthly(p.SecondTierMonthly))
	d.Set("second_tier_yearly", flattenSecondTierYearly(p.SecondTierYearly))
	d.Set("protected_servers", p.ProtectedServers)
	d.Set("protected_volumes", p.ProtectedVolumes)
	d.Set("protected_volume_sizes", p.ProtectedVolumeSizes)
	d.Set("compliance_state", p.ComplianceState)
	d.Set("compliance_msg", p.ComplianceMsg)
	d.Set("status", p.Status)
	d.Set("created_at", p.CreatedAt)
}

func backupPolicyServerAttrs(p *dto.BackupPolicyServer) map[string]interface{} {
	return map[string]interface{}{
		"id":                       p.ID,
		"name":                     p.Name,
		"description":              p.Description,
		"start_hour":               p.StartHour,
		"resource_type":            p.ResourceType,
		"purpose":                  p.Purpose,
		"is_auto":                  p.IsAuto,
		"is_auto_apply_for_volume": p.IsAutoApplyForVolume,
		"backup_vault_ids":         p.BackupVaultIDs,
		"daily":                    flattenDaily(p.Daily),
		"weekly":                   flattenWeekly(p.Weekly),
		"monthly":                  flattenMonthly(p.Monthly),
		"second_tier_enabled":      p.SecondTierEnabled,
		"second_tier_weekly":       flattenSecondTierWeekly(p.SecondTierWeekly),
		"second_tier_monthly":      flattenSecondTierMonthly(p.SecondTierMonthly),
		"second_tier_yearly":       flattenSecondTierYearly(p.SecondTierYearly),
		"protected_servers":        p.ProtectedServers,
		"protected_volumes":        p.ProtectedVolumes,
		"protected_volume_sizes":   p.ProtectedVolumeSizes,
		"compliance_state":         p.ComplianceState,
		"compliance_msg":           p.ComplianceMsg,
		"status":                   p.Status,
		"created_at":               p.CreatedAt,
	}
}
