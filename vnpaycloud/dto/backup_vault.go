package dto

type BackupVault struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Purpose           string `json:"purpose"`
	Type              string `json:"type"`
	ZoneID            string `json:"zoneId"`
	StorageLocationID string `json:"storageLocationId"`
	Description       string `json:"description"`
	DiskUsed          int    `json:"diskUsed,string"`
	Quota             int    `json:"quota,string"`
	CommittedQuota    int64  `json:"committedQuota,string"`
	ObjectLock        bool   `json:"objectLock"`
	LockTimeUnit      string `json:"lockTimeUnit"`
	LockTimeNumber    int    `json:"lockTimeNumber"`
	Status            string `json:"status"`
	CreatedAt         string `json:"createdAt"`
	ProjectID         string `json:"projectId"`
}

type BackupVaultResponse struct {
	BackupVault BackupVault `json:"backupVault"`
}

type ListBackupVaultsResponse struct {
	BackupVaults []BackupVault `json:"backupVaults"`
}

type CreateBackupVaultRequest struct {
	Name        string `json:"name"`
	Purpose     string `json:"purpose"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type UpdateBackupVaultRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
