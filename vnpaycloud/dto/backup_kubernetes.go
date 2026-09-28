package dto

type BackupKubernetes struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	ProtectedClusterID string `json:"protectedClusterId"`
	PVCBackupPolicyID  string `json:"pvcBackupPolicyId"`
	ClusterBackupType  string `json:"clusterBackupType"`
	ZoneID             string `json:"zoneId"`
	ProjectID          string `json:"projectId"`
	CustomerUsername   string `json:"customerUsername"`
	Status             string `json:"status"`
	CreatedAt          string `json:"createdAt"`
}

type BackupKubernetesResponse struct {
	BackupKubernetes BackupKubernetes `json:"backupKubernetes"`
}

type ListBackupKubernetesResponse struct {
	BackupKubernetes []BackupKubernetes `json:"backupKubernetes"`
}

type CreateBackupKubernetesRequest struct {
	ProtectedClusterID string `json:"protectedClusterId"`
	PVCBackupPolicyID  string `json:"pvcBackupPolicyId"`
	Description        string `json:"description,omitempty"`
}

type UpdateBackupKubernetesRequest struct {
	Description string `json:"description,omitempty"`
}
