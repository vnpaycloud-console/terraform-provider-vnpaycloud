package dto

// Instance matches the iac-proxy-v2 Instance proto message.
type Instance struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	ImageName           string   `json:"imageName"`
	ImageID             string   `json:"imageId"`
	FlavorName          string   `json:"flavorName"`
	VolumeIDs           []string `json:"volumeIds"`
	Status              string   `json:"status"`
	PowerState          string   `json:"powerState"`
	NetworkInterfaceIDs []string `json:"networkInterfaceIds"`
	KeyPairID           string   `json:"keyPairId"`
	SecurityGroupIDs    []string `json:"securityGroupIds"`
	ServerGroupID       string   `json:"serverGroupId"`
	CreatedAt           string   `json:"createdAt"`
	ProjectID           string   `json:"projectId"`
	ZoneID              string   `json:"zoneId"`
}

// CreateInstanceRequest matches the iac-proxy-v2 CreateInstanceRequest proto message.
// project_id is passed via URL path, not in the body.
type CreateInstanceRequest struct {
	Name                string   `json:"name"`
	Image               string   `json:"image,omitempty"`
	SnapshotID          string   `json:"snapshotId,omitempty"`
	Flavor              string   `json:"flavor,omitempty"`
	RootDiskGB          int32    `json:"rootDiskGb"`
	RootDiskVolumeType  string   `json:"rootDiskVolumeType"`
	KeyPair             string   `json:"keyPair,omitempty"`
	NetworkInterfaceIDs []string `json:"networkInterfaceIds,omitempty"`
	ServerGroupID       string   `json:"serverGroupId,omitempty"`
	UserData            string   `json:"userData,omitempty"`
	IsUserDataBase64    bool     `json:"isUserDataBase64,omitempty"`
}

// UpdateInstanceRequest matches the iac-proxy-v2 UpdateInstanceRequest proto message.
// project_id and id are passed via URL path.
type UpdateInstanceRequest struct {
	Name           string   `json:"name,omitempty"`
	SecurityGroups []string `json:"securityGroups,omitempty"`
}

// ResizeInstanceRequest matches the iac-proxy-v2 ResizeInstanceRequest proto message.
// project_id and id are passed via URL path.
type ResizeInstanceRequest struct {
	Flavor         string `json:"flavor,omitempty"`
	IsCustomFlavor bool   `json:"isCustomFlavor,omitempty"`
	CustomVCPUs    int32  `json:"customVcpus,omitempty"`
	CustomRAMMB    int32  `json:"customRamMb,omitempty"`
}

// InstanceResponse matches the iac-proxy-v2 InstanceResponse proto message.
type InstanceResponse struct {
	Instance Instance `json:"instance"`
}

// ListInstancesResponse matches the iac-proxy-v2 ListInstancesResponse proto message.
type ListInstancesResponse struct {
	Instances []Instance `json:"instances"`
}
