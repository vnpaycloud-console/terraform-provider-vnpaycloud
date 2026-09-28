package dto

type BackupServer struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	ServerID       string   `json:"serverId"`
	VolumeIDs      []string `json:"volumeIds"`
	BackupPolicyID string   `json:"backupPolicyId"`
	Purpose        string   `json:"purpose"`
	ZoneID         string   `json:"zoneId"`
	IsCompliant    string   `json:"isCompliant"`
	ComplianceMsg  string   `json:"complianceMsg"`
	Status         string   `json:"status"`
	CreatedAt      string   `json:"createdAt"`
	ProjectID      string   `json:"projectId"`
}

type BackupServerResponse struct {
	BackupServer BackupServer `json:"backupServer"`
}

type ListBackupServersResponse struct {
	BackupServers []BackupServer `json:"backupServers"`
}

type CreateBackupServerRequest struct {
	ServerID       string   `json:"serverId"`
	VolumeIDs      []string `json:"volumeIds"`
	BackupPolicyID string   `json:"backupPolicyId"`
}

type UpdateBackupServerRequest struct {
	VolumeIDs      []string `json:"volumeIds"`
	BackupPolicyID string   `json:"backupPolicyId"`
}
