package dto

// NOTE: retentions / day_of_week / day_of_month are proto int64 on
// iac-proxy-v2 → grpc-gateway encodes them as quoted strings, so `,string`
// (start_hour / run_priority are int32 → plain numbers, no `,string`).

type BackupPolicyKubernetesDaily struct {
	Retentions int `json:"retentions,string"`
}

type BackupPolicyKubernetesWeekly struct {
	Retentions int `json:"retentions,string"`
	DayOfWeek  int `json:"dayOfWeek,string"`
}

type BackupPolicyKubernetesMonthly struct {
	Retentions int    `json:"retentions,string"`
	DayType    string `json:"dayType"`
	DayOfMonth int    `json:"dayOfMonth,omitempty,string"`
	DayOfWeek  int    `json:"dayOfWeek,omitempty,string"`
}

type BackupPolicyKubernetes struct {
	ID                 string                         `json:"id"`
	Name               string                         `json:"name"`
	Description        string                         `json:"description"`
	AutoApplyNewVolume bool                           `json:"autoApplyNewVolume"`
	IsAuto             bool                           `json:"isAuto"`
	StartHour          int                            `json:"startHour"`
	RunPriority        int                            `json:"runPriority"`
	Purpose            string                         `json:"purpose"`
	BackupVaultIDs     []string                       `json:"backupVaultIds"`
	Daily              *BackupPolicyKubernetesDaily   `json:"daily"`
	Weekly             *BackupPolicyKubernetesWeekly  `json:"weekly"`
	Monthly            *BackupPolicyKubernetesMonthly `json:"monthly"`
	Status             string                         `json:"status"`
	CreatedAt          string                         `json:"createdAt"`
	ProjectID          string                         `json:"projectId"`
}

type BackupPolicyKubernetesResponse struct {
	BackupPolicyKubernetes BackupPolicyKubernetes `json:"backupPolicyKubernetes"`
}

type ListBackupPolicyKubernetesResponse struct {
	BackupPolicyKubernetes []BackupPolicyKubernetes `json:"backupPolicyKubernetes"`
}

type CreateBackupPolicyKubernetesRequest struct {
	Name               string                         `json:"name"`
	Description        string                         `json:"description,omitempty"`
	AutoApplyNewVolume bool                           `json:"autoApplyNewVolume"`
	IsAuto             bool                           `json:"isAuto"`
	StartHour          int                            `json:"startHour"`
	RunPriority        int                            `json:"runPriority,omitempty"`
	BackupVaultIDs     []string                       `json:"backupVaultIds"`
	Daily              *BackupPolicyKubernetesDaily   `json:"daily,omitempty"`
	Weekly             *BackupPolicyKubernetesWeekly  `json:"weekly,omitempty"`
	Monthly            *BackupPolicyKubernetesMonthly `json:"monthly,omitempty"`
}

type UpdateBackupPolicyKubernetesRequest struct {
	Description        string                         `json:"description,omitempty"`
	AutoApplyNewVolume bool                           `json:"autoApplyNewVolume"`
	IsAuto             bool                           `json:"isAuto"`
	StartHour          int                            `json:"startHour"`
	RunPriority        int                            `json:"runPriority,omitempty"`
	Daily              *BackupPolicyKubernetesDaily   `json:"daily,omitempty"`
	Weekly             *BackupPolicyKubernetesWeekly  `json:"weekly,omitempty"`
	Monthly            *BackupPolicyKubernetesMonthly `json:"monthly,omitempty"`
}
