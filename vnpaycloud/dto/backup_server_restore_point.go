package dto

type BackupServerRestorePoint struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	BackupServerID    string `json:"backupServerId"`
	BackupPolicyID    string `json:"backupPolicyId"`
	Type              string `json:"type"`
	Purpose           string `json:"purpose"`
	ZoneID            string `json:"zoneId"`
	IsVisible         bool   `json:"isVisible"`
	CreateByBackupNow bool   `json:"createByBackupNow"`
	BackupPoint       string `json:"backupPoint"`
	Status            string `json:"status"`
	CreatedAt         string `json:"createdAt"`
}

type ListBackupServerRestorePointsResponse struct {
	RestorePoints []BackupServerRestorePoint `json:"restorePoints"`
}

type CreateBackupServerRestorePointRequest struct {
	BackupServerID string `json:"backupServerId"`
}

type BackupServerRestorePointResponse struct {
	RestorePoint BackupServerRestorePoint `json:"restorePoint"`
}

type BackupServerDisasterRestorePoint struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	BackupServerID string `json:"backupServerId"`
	BackupPolicyID string `json:"backupPolicyId"`
	Type           string `json:"type"`
	Purpose        string `json:"purpose"`
	ZoneID         string `json:"zoneId"`
	BackupPoint    string `json:"backupPoint"`
	Status         string `json:"status"`
	CreatedAt      string `json:"createdAt"`
	ServerID       string `json:"serverId"`
	ServerName     string `json:"serverName"`
	ServerZoneID   string `json:"serverZoneId"`
}

type ListBackupServerDisasterRestorePointsResponse struct {
	RestorePoints []BackupServerDisasterRestorePoint `json:"restorePoints"`
}
