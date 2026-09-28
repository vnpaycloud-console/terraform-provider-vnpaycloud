package dto

type CreateBackupKubernetesRestoreRequest struct {
	DestClusterID                string   `json:"destClusterId"`
	RestorePointID               string   `json:"restorePointId"`
	RestoreType                  string   `json:"restoreType,omitempty"`
	TransformLbToClusterIP       bool     `json:"transformLbToClusterIp"`
	KeepNodePortNumbers          bool     `json:"keepNodePortNumbers"`
	RestorePersistentVolumes     bool     `json:"restorePersistentVolumes"`
	IncludeClusterScopedResource bool     `json:"includeClusterScopedResource"`
	IncludedNamespaces           []string `json:"includedNamespaces,omitempty"`
	ExcludedNamespaces           []string `json:"excludedNamespaces,omitempty"`
	IncludedResources            []string `json:"includedResources,omitempty"`
	ExcludedResources            []string `json:"excludedResources,omitempty"`
}

type BackupKubernetesRestoreResponse struct {
	DestClusterID  string `json:"destClusterId"`
	RestorePointID string `json:"restorePointId"`
}

type BackupKubernetesRestorePoint struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	ClusterID         string   `json:"clusterId"`
	ClusterName       string   `json:"clusterName"`
	ClusterBackupID   string   `json:"clusterBackupId"`
	PVCBackupPolicyID string   `json:"pvcBackupPolicyId"`
	ProjectID         string   `json:"projectId"`
	ZoneID            string   `json:"zoneId"`
	BackupVaultIDs    []string `json:"backupVaultIds"`
	VeleroBackupName  string   `json:"veleroBackupName"`
	BackupPoint       string   `json:"backupPoint"`
	KubernetesVersion string   `json:"kubernetesVersion"`
	Status            string   `json:"status"`
}

type ListBackupKubernetesRestorePointsResponse struct {
	RestorePoints []BackupKubernetesRestorePoint `json:"restorePoints"`
}

type BackupKubernetesRestorePointResponse struct {
	RestorePoint BackupKubernetesRestorePoint `json:"restorePoint"`
}
