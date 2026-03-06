package client

import (
	"fmt"
)

// ApiPath provides REST API path builders for iac-api-gateway.
// All paths are project-scoped: /v2/projects/{project_id}/...
var ApiPath = struct {
	// VPC
	VPCs       func(projectID string) string
	VPCWithID  func(projectID, id string) string
	VPCSetSNAT func(projectID, id string) string

	// Subnet
	Subnets          func(projectID string) string
	SubnetWithID     func(projectID, id string) string
	SubnetEnableSNAT  func(projectID, id string) string
	SubnetDisableSNAT func(projectID, id string) string

	// Security Group
	SecurityGroups      func(projectID string) string
	SecurityGroupWithID func(projectID, id string) string

	// Security Group Rule
	SecurityGroupRules      func(projectID string) string
	SecurityGroupRuleWithID func(projectID, id string) string

	// Floating IP
	FloatingIPs            func(projectID string) string
	FloatingIPWithID       func(projectID, id string) string
	FloatingIPAssociate    func(projectID, id string) string
	FloatingIPDisassociate func(projectID, id string) string

	// Network Interface
	NetworkInterfaces      func(projectID string) string
	NetworkInterfaceWithID func(projectID, id string) string
	NetworkInterfaceAttach func(projectID, id string) string
	NetworkInterfaceDetach func(projectID, id string) string

	// Volume
	Volumes        func(projectID string) string
	VolumeWithID   func(projectID, id string) string
	VolumeResize   func(projectID, id string) string
	VolumeAttach   func(projectID, id string) string
	VolumeDetach   func(projectID, id string) string

	// Volume Attachment
	VolumeAttachments      func(projectID string) string
	VolumeAttachmentWithID func(projectID, id string) string

	// Instance
	Instances        func(projectID string) string
	InstanceWithID   func(projectID, id string) string
	InstanceResize   func(projectID, id string) string

	// KeyPair (global resource — uses name, not ID)
	CreateKeyPair     func() string
	KeyPairs          func(projectID string) string
	KeyPairWithName   func(projectID, name string) string

	// Internet Gateway
	InternetGateways         func(projectID string) string
	InternetGatewayWithID    func(projectID, id string) string
	InternetGatewayAttachVPC func(projectID, id string) string
	InternetGatewayDetachVPC func(projectID, id string) string

	// Snapshot
	Snapshots      func(projectID string) string
	SnapshotWithID func(projectID, id string) string

	// Load Balancer
	LoadBalancers      func(projectID string) string
	LoadBalancerWithID func(projectID, id string) string

	// Listener
	Listeners      func(projectID string) string
	ListenerWithID func(projectID, id string) string

	// Pool
	Pools      func(projectID string) string
	PoolWithID func(projectID, id string) string

	// Health Monitor
	HealthMonitors      func(projectID string) string
	HealthMonitorWithID func(projectID, id string) string

	// Registry Project
	RegistryProjects      func(projectID string) string
	RegistryProjectWithID func(projectID, id string) string

	// Robot Account
	RobotAccounts      func(projectID, registryID string) string
	RobotAccountWithID func(projectID, registryID, id string) string

	// Kubernetes Cluster
	Clusters          func(projectID string) string
	ClusterWithID     func(projectID, id string) string
	ClusterKubeconfig func(projectID, id string) string

	// Worker Group
	WorkerGroups      func(projectID, clusterID string) string
	WorkerGroupWithID func(projectID, clusterID, id string) string

	// Route Table
	RouteTables      func(projectID string) string
	RouteTableWithID func(projectID, id string) string

	// Private Gateway
	PrivateGateways      func(projectID string) string
	PrivateGatewayWithID func(projectID, id string) string

	// VPC Peering (not project-scoped)
	PeeringConnections      func() string
	PeeringConnectionWithID func(id string) string

	// S3 Bucket
	Buckets      func(projectID string) string
	BucketUsage  func(projectID, bucketName string) string
	BucketDelete func(projectID, bucketName, region string) string

	// Zone → Project Resolution (not project-scoped)
	ResolveProjectByZone func(zoneID string) string
}{
	VPCs: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/vpcs", projectID)
	},
	VPCWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/vpcs/%s", projectID, id)
	},
	VPCSetSNAT: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/vpcs/%s/snat", projectID, id)
	},
	Subnets: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/subnets", projectID)
	},
	SubnetWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/subnets/%s", projectID, id)
	},
	SubnetEnableSNAT: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/subnets/%s/enable-snat", projectID, id)
	},
	SubnetDisableSNAT: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/subnets/%s/disable-snat", projectID, id)
	},
	SecurityGroups: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/security-groups", projectID)
	},
	SecurityGroupWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/security-groups/%s", projectID, id)
	},
	SecurityGroupRules: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/security-group-rules", projectID)
	},
	SecurityGroupRuleWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/security-group-rules/%s", projectID, id)
	},
	FloatingIPs: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/floating-ips", projectID)
	},
	FloatingIPWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/floating-ips/%s", projectID, id)
	},
	FloatingIPAssociate: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/floating-ips/%s/associate", projectID, id)
	},
	FloatingIPDisassociate: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/floating-ips/%s/disassociate", projectID, id)
	},
	NetworkInterfaces: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/network-interfaces", projectID)
	},
	NetworkInterfaceWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/network-interfaces/%s", projectID, id)
	},
	NetworkInterfaceAttach: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/network-interfaces/%s/attach", projectID, id)
	},
	NetworkInterfaceDetach: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/network-interfaces/%s/detach", projectID, id)
	},
	Volumes: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/volumes", projectID)
	},
	VolumeWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/volumes/%s", projectID, id)
	},
	VolumeResize: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/volumes/%s/resize", projectID, id)
	},
	VolumeAttach: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/volumes/%s/attach", projectID, id)
	},
	VolumeDetach: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/volumes/%s/detach", projectID, id)
	},
	VolumeAttachments: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/volume-attachments", projectID)
	},
	VolumeAttachmentWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/volume-attachments/%s", projectID, id)
	},
	Instances: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/instances", projectID)
	},
	InstanceWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/instances/%s", projectID, id)
	},
	InstanceResize: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/instances/%s/resize", projectID, id)
	},
	CreateKeyPair: func() string {
		return "/v2/key-pairs"
	},
	KeyPairs: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/key-pairs", projectID)
	},
	KeyPairWithName: func(projectID, name string) string {
		return fmt.Sprintf("/v2/projects/%s/key-pairs/%s", projectID, name)
	},
	InternetGateways: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/internet-gateways", projectID)
	},
	InternetGatewayWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/internet-gateways/%s", projectID, id)
	},
	InternetGatewayAttachVPC: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/internet-gateways/%s/attach-vpc", projectID, id)
	},
	InternetGatewayDetachVPC: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/internet-gateways/%s/detach-vpc", projectID, id)
	},
	Snapshots: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/snapshots", projectID)
	},
	SnapshotWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/snapshots/%s", projectID, id)
	},
	LoadBalancers: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/load-balancers", projectID)
	},
	LoadBalancerWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/load-balancers/%s", projectID, id)
	},
	Listeners: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/listeners", projectID)
	},
	ListenerWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/listeners/%s", projectID, id)
	},
	Pools: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/pools", projectID)
	},
	PoolWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/pools/%s", projectID, id)
	},
	HealthMonitors: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/health-monitors", projectID)
	},
	HealthMonitorWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/health-monitors/%s", projectID, id)
	},
	RegistryProjects: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/registries", projectID)
	},
	RegistryProjectWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/registries/%s", projectID, id)
	},
	RobotAccounts: func(projectID, registryID string) string {
		return fmt.Sprintf("/v2/projects/%s/registries/%s/robot-accounts", projectID, registryID)
	},
	RobotAccountWithID: func(projectID, registryID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/registries/%s/robot-accounts/%s", projectID, registryID, id)
	},
	Clusters: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/clusters", projectID)
	},
	ClusterWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/clusters/%s", projectID, id)
	},
	ClusterKubeconfig: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/clusters/%s/kubeconfig", projectID, id)
	},
	WorkerGroups: func(projectID, clusterID string) string {
		return fmt.Sprintf("/v2/projects/%s/clusters/%s/worker-groups", projectID, clusterID)
	},
	WorkerGroupWithID: func(projectID, clusterID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/clusters/%s/worker-groups/%s", projectID, clusterID, id)
	},
	RouteTables: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/route-tables", projectID)
	},
	RouteTableWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/route-tables/%s", projectID, id)
	},
	PrivateGateways: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/private-gateways", projectID)
	},
	PrivateGatewayWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/projects/%s/private-gateways/%s", projectID, id)
	},
	PeeringConnections: func() string {
		return "/v2/peering-connections"
	},
	PeeringConnectionWithID: func(id string) string {
		return fmt.Sprintf("/v2/peering-connections/%s", id)
	},
	Buckets: func(projectID string) string {
		return fmt.Sprintf("/v2/projects/%s/buckets", projectID)
	},
	BucketUsage: func(projectID, bucketName string) string {
		return fmt.Sprintf("/v2/projects/%s/buckets/%s/usage", projectID, bucketName)
	},
	BucketDelete: func(projectID, bucketName, region string) string {
		return fmt.Sprintf("/v2/projects/%s/buckets/%s?region=%s", projectID, bucketName, region)
	},
	ResolveProjectByZone: func(zoneID string) string {
		return fmt.Sprintf("/v2/zones/%s/project", zoneID)
	},
}
