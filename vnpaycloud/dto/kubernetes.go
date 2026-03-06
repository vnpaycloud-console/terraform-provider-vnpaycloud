package dto

// K8sCluster matches the iac-proxy-v2 K8sCluster proto message.
type K8sCluster struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Zone        string `json:"zone"`
	K8sVersion  string `json:"k8sVersion"`
	Purpose     string `json:"purpose"`
	SubnetID    string `json:"subnetId"`
	CniPlugin   string `json:"cniPlugin"`
	PodCidr     string `json:"podCidr"`
	ServiceCidr string `json:"serviceCidr"`
	PrivateGwID string `json:"privateGwId"`
	ClusterSize string `json:"clusterSize"`
	ApiEndpoint string `json:"apiEndpoint"`
	PrivateIP   string `json:"privateIp"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
}

// CreateK8sClusterRequest matches the iac-proxy-v2 CreateK8sClusterRequest proto message.
// project_id is passed via URL path.
type CreateK8sClusterRequest struct {
	ClusterInformation     K8sClusterInformation     `json:"clusterInformation"`
	NetworkInformation     K8sNetworkInformation     `json:"networkInformation"`
	MasterInformation      K8sMasterInformation      `json:"masterInformation"`
	WorkerGroupInformation K8sWorkerGroupInformation `json:"workerGroupInformation"`
}

type K8sClusterInformation struct {
	Name        string `json:"name"`
	K8sVersion  string `json:"k8sVersion,omitempty"`
	Purpose     string `json:"purpose,omitempty"`
	PrivateGwID string `json:"privateGwId,omitempty"`
}

type K8sNetworkInformation struct {
	SubnetID    string `json:"subnetId"`
	CniPlugin   string `json:"cniPlugin,omitempty"`
	PodCidr     string `json:"podCidr,omitempty"`
	ServiceCidr string `json:"serviceCidr,omitempty"`
}

type K8sMasterInformation struct {
	ClusterSize string `json:"clusterSize,omitempty"`
}

type K8sWorkerGroupInformation struct {
	Name        string `json:"name,omitempty"`
	NumWorkers  int    `json:"numWorkers,omitempty"`
	AutoScaling bool   `json:"autoScaling,omitempty"`
	MinWorkers  int    `json:"minWorkers,omitempty"`
	MaxWorkers  int    `json:"maxWorkers,omitempty"`
	Flavor      string `json:"flavor"`
	VolumeType  string `json:"volumeType,omitempty"`
	VolumeSize  int    `json:"volumeSize,omitempty"`
	SshKeyID    string `json:"sshKeyId,omitempty"`
	Labels      string `json:"labels,omitempty"`
}

// K8sClusterResponse matches the iac-proxy-v2 K8sClusterResponse proto message.
type K8sClusterResponse struct {
	Cluster K8sCluster `json:"cluster"`
}

// ListK8sClustersResponse matches the iac-proxy-v2 ListK8sClustersResponse proto message.
type ListK8sClustersResponse struct {
	Clusters []K8sCluster `json:"clusters"`
}

// KubeconfigResponse matches the iac-proxy-v2 KubeconfigResponse proto message.
type KubeconfigResponse struct {
	Kubeconfig string `json:"kubeconfig"`
}

// WorkerGroup matches the iac-proxy-v2 WorkerGroup proto message.
type WorkerGroup struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ClusterID   string `json:"clusterId"`
	Flavor      string `json:"flavor"`
	NumWorkers  int    `json:"numWorkers"`
	MinWorkers  int    `json:"minWorkers"`
	MaxWorkers  int    `json:"maxWorkers"`
	AutoScaling bool   `json:"autoScaling"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
}

// CreateWorkerGroupRequest matches the iac-proxy-v2 CreateWorkerGroupRequest proto message.
// project_id and cluster_id are passed via URL path.
type CreateWorkerGroupRequest struct {
	Name        string            `json:"name"`
	Flavor      string            `json:"flavor"`
	NumWorkers  int               `json:"numWorkers"`
	AutoScaling bool              `json:"autoScaling,omitempty"`
	MinWorkers  int               `json:"minWorkers,omitempty"`
	MaxWorkers  int               `json:"maxWorkers,omitempty"`
	VolumeType  string            `json:"volumeType,omitempty"`
	VolumeSize  int               `json:"volumeSize,omitempty"`
	SshKeyID    string            `json:"sshKeyId,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// UpdateWorkerGroupRequest matches the iac-proxy-v2 UpdateWorkerGroupRequest proto message.
type UpdateWorkerGroupRequest struct {
	NumWorkers  int  `json:"numWorkers"`
	AutoScaling bool `json:"autoScaling,omitempty"`
	MinWorkers  int  `json:"minWorkers,omitempty"`
	MaxWorkers  int  `json:"maxWorkers,omitempty"`
}

// WorkerGroupResponse matches the iac-proxy-v2 WorkerGroupResponse proto message.
type WorkerGroupResponse struct {
	WorkerGroup WorkerGroup `json:"workerGroup"`
}

// ListWorkerGroupsResponse matches the iac-proxy-v2 ListWorkerGroupsResponse proto message.
type ListWorkerGroupsResponse struct {
	WorkerGroups []WorkerGroup `json:"workerGroups"`
}
