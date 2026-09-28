package dto

// NOTE: every numeric field below is backed by a proto int64 on iac-proxy-v2.
// grpc-gateway (proto3 JSON) encodes int64 as a quoted string, so the DTO uses
// the `,string` json option — same convention as dto/volume.go `sizeGb,string`.

type BackupPolicyServerDaily struct {
	Retentions int `json:"retentions,string"`
}

type BackupPolicyServerWeekly struct {
	Retentions int `json:"retentions,string"`
	DayOfWeek  int `json:"dayOfWeek,string"`
}

type BackupPolicyServerMonthly struct {
	Retentions int    `json:"retentions,string"`
	DayType    string `json:"dayType"`
	DayOfMonth int    `json:"dayOfMonth,omitempty,string"`
	DayOfWeek  int    `json:"dayOfWeek,omitempty,string"`
}

type BackupPolicyServerSecondTierWeekly struct {
	Retentions int `json:"retentions,string"`
	DayOfWeek  int `json:"dayOfWeek,string"`
}

type BackupPolicyServerSecondTierMonthly struct {
	Retentions int    `json:"retentions,string"`
	DayType    string `json:"dayType"`
	DayOfMonth int    `json:"dayOfMonth,omitempty,string"`
	DayOfWeek  int    `json:"dayOfWeek,omitempty,string"`
}

type BackupPolicyServerSecondTierYearly struct {
	Retentions int `json:"retentions,string"`
	Month      int `json:"month,string"`
	DayOfMonth int `json:"dayOfMonth,string"`
}

type BackupPolicyServer struct {
	ID                   string                               `json:"id"`
	Name                 string                               `json:"name"`
	Description          string                               `json:"description"`
	StartHour            int                                  `json:"startHour,string"`
	ResourceType         string                               `json:"resourceType"`
	Purpose              string                               `json:"purpose"`
	IsAuto               bool                                 `json:"isAuto"`
	IsAutoApplyForVolume bool                                 `json:"isAutoApplyForVolume"`
	BackupVaultIDs       []string                             `json:"backupVaultIds"`
	Daily                *BackupPolicyServerDaily             `json:"daily"`
	Weekly               *BackupPolicyServerWeekly            `json:"weekly"`
	Monthly              *BackupPolicyServerMonthly           `json:"monthly"`
	SecondTierEnabled    bool                                 `json:"secondTierEnabled"`
	SecondTierWeekly     *BackupPolicyServerSecondTierWeekly  `json:"secondTierWeekly"`
	SecondTierMonthly    *BackupPolicyServerSecondTierMonthly `json:"secondTierMonthly"`
	SecondTierYearly     *BackupPolicyServerSecondTierYearly  `json:"secondTierYearly"`
	ProtectedServers     int                                  `json:"protectedServers,string"`
	ProtectedVolumes     int                                  `json:"protectedVolumes,string"`
	ProtectedVolumeSizes int                                  `json:"protectedVolumeSizes,string"`
	ComplianceState      string                               `json:"complianceState"`
	ComplianceMsg        string                               `json:"complianceMsg"`
	Status               string                               `json:"status"`
	CreatedAt            string                               `json:"createdAt"`
	ProjectID            string                               `json:"projectId"`
}

type BackupPolicyServerResponse struct {
	BackupPolicyServer BackupPolicyServer `json:"backupPolicyServer"`
}

type ListBackupPolicyServersResponse struct {
	BackupPolicyServers []BackupPolicyServer `json:"backupPolicyServers"`
}

type CreateBackupPolicyServerRequest struct {
	Name                 string                               `json:"name"`
	Description          string                               `json:"description,omitempty"`
	StartHour            int                                  `json:"startHour,string"`
	ResourceType         string                               `json:"resourceType"`
	Purpose              string                               `json:"purpose"`
	IsAuto               bool                                 `json:"isAuto"`
	IsAutoApplyForVolume bool                                 `json:"isAutoApplyForVolume"`
	BackupVaultIDs       []string                             `json:"backupVaultIds,omitempty"`
	Daily                *BackupPolicyServerDaily             `json:"daily"`
	Weekly               *BackupPolicyServerWeekly            `json:"weekly,omitempty"`
	Monthly              *BackupPolicyServerMonthly           `json:"monthly,omitempty"`
	SecondTierEnabled    bool                                 `json:"secondTierEnabled"`
	SecondTierWeekly     *BackupPolicyServerSecondTierWeekly  `json:"secondTierWeekly,omitempty"`
	SecondTierMonthly    *BackupPolicyServerSecondTierMonthly `json:"secondTierMonthly,omitempty"`
	SecondTierYearly     *BackupPolicyServerSecondTierYearly  `json:"secondTierYearly,omitempty"`
}

type UpdateBackupPolicyServerRequest struct {
	Description          string                               `json:"description,omitempty"`
	StartHour            int                                  `json:"startHour,string"`
	IsAuto               bool                                 `json:"isAuto"`
	IsAutoApplyForVolume bool                                 `json:"isAutoApplyForVolume"`
	BackupVaultIDs       []string                             `json:"backupVaultIds,omitempty"`
	Daily                *BackupPolicyServerDaily             `json:"daily"`
	Weekly               *BackupPolicyServerWeekly            `json:"weekly,omitempty"`
	Monthly              *BackupPolicyServerMonthly           `json:"monthly,omitempty"`
	SecondTierEnabled    bool                                 `json:"secondTierEnabled"`
	SecondTierWeekly     *BackupPolicyServerSecondTierWeekly  `json:"secondTierWeekly,omitempty"`
	SecondTierMonthly    *BackupPolicyServerSecondTierMonthly `json:"secondTierMonthly,omitempty"`
	SecondTierYearly     *BackupPolicyServerSecondTierYearly  `json:"secondTierYearly,omitempty"`
}
