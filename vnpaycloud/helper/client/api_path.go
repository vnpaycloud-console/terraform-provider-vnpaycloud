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
	Subnets           func(projectID string) string
	SubnetWithID      func(projectID, id string) string
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
	Volumes      func(projectID string) string
	VolumeWithID func(projectID, id string) string
	VolumeResize func(projectID, id string) string
	VolumeAttach func(projectID, id string) string
	VolumeDetach func(projectID, id string) string

	// Volume Attachment
	VolumeAttachments      func(projectID string) string
	VolumeAttachmentWithID func(projectID, id string) string

	// Instance
	Instances      func(projectID string) string
	InstanceWithID func(projectID, id string) string
	InstanceResize func(projectID, id string) string

	// Server Group
	ServerGroups      func(projectID string) string
	ServerGroupWithID func(projectID, id string) string

	// KeyPair (global resource — uses name, not ID)
	CreateKeyPair   func() string
	KeyPairs        func(projectID string) string
	KeyPairWithName func(projectID, name string) string

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

	// Robot Account (system-level, multi-project)
	RobotAccounts      func(projectID string) string
	RobotAccountWithID func(projectID, id string) string

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

	// Flavor (not project-scoped, filtered by zone)
	Flavors      func(zone string) string
	FlavorWithID func(id string) string

	// Image (not project-scoped, filtered by zone)
	Images      func(zone string) string
	ImageWithID func(id string) string

	// Volume Type (not project-scoped, filtered by zone)
	VolumeTypes      func(zone string) string
	VolumeTypeWithID func(id string) string

	// S3 Bucket
	Buckets      func(projectID string) string
	BucketUsage  func(projectID, bucketName string) string
	BucketDelete func(projectID, bucketName, region string) string

	// Database Postgres Instance
	DatabasePostgresInstances                      func(projectID string) string
	DatabasePostgresInstanceWithID                  func(projectID, id string) string
	DatabasePostgresInstanceScale                   func(projectID, id string) string
	DatabasePostgresInstanceChangeFlavor            func(projectID, id string) string
	DatabasePostgresInstanceExpandVolume            func(projectID, id string) string
	DatabasePostgresInstanceEnableAutoExpandVolume  func(projectID, id string) string
	DatabasePostgresInstanceDisableAutoExpandVolume func(projectID, id string) string
	DatabasePostgresInstanceEnableTls               func(projectID, id string) string
	DatabasePostgresInstanceDisableTls              func(projectID, id string) string

	// Database Redis Instance
	DatabaseRedisInstances                      func(projectID string) string
	DatabaseRedisInstanceWithID                  func(projectID, id string) string
	DatabaseRedisInstanceChangeFlavor            func(projectID, id string) string
	DatabaseRedisInstanceExpandVolume            func(projectID, id string) string
	DatabaseRedisInstanceEnableAutoExpandVolume  func(projectID, id string) string
	DatabaseRedisInstanceDisableAutoExpandVolume func(projectID, id string) string
	DatabaseRedisInstanceEnableTls               func(projectID, id string) string
	DatabaseRedisInstanceDisableTls              func(projectID, id string) string

	// Database Redis Sentinel Instance
	DatabaseRedisSentinelInstances                      func(projectID string) string
	DatabaseRedisSentinelInstanceWithID                  func(projectID, id string) string
	DatabaseRedisSentinelInstanceScale                   func(projectID, id string) string
	DatabaseRedisSentinelInstanceChangeFlavor            func(projectID, id string) string
	DatabaseRedisSentinelInstanceExpandVolume            func(projectID, id string) string
	DatabaseRedisSentinelInstanceEnableAutoExpandVolume  func(projectID, id string) string
	DatabaseRedisSentinelInstanceDisableAutoExpandVolume func(projectID, id string) string
	DatabaseRedisSentinelInstanceSentinelScale           func(projectID, id string) string
	DatabaseRedisSentinelInstanceSentinelChangeFlavor    func(projectID, id string) string
	DatabaseRedisSentinelInstanceEnableTls               func(projectID, id string) string
	DatabaseRedisSentinelInstanceDisableTls              func(projectID, id string) string

	// Database Flavor
	DatabaseFlavors func(projectID string) string

	// Zone → Project Resolution (not project-scoped)
	ResolveProjectByZone func(zoneID string) string
}{
	VPCs: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/vpcs", projectID)
	},
	VPCWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/vpcs/%s", projectID, id)
	},
	VPCSetSNAT: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/vpcs/%s/snat", projectID, id)
	},
	Subnets: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/subnets", projectID)
	},
	SubnetWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/subnets/%s", projectID, id)
	},
	SubnetEnableSNAT: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/subnets/%s/enable-snat", projectID, id)
	},
	SubnetDisableSNAT: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/subnets/%s/disable-snat", projectID, id)
	},
	SecurityGroups: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/security-groups", projectID)
	},
	SecurityGroupWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/security-groups/%s", projectID, id)
	},
	SecurityGroupRules: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/security-group-rules", projectID)
	},
	SecurityGroupRuleWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/security-group-rules/%s", projectID, id)
	},
	FloatingIPs: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/floating-ips", projectID)
	},
	FloatingIPWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/floating-ips/%s", projectID, id)
	},
	FloatingIPAssociate: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/floating-ips/%s/associate", projectID, id)
	},
	FloatingIPDisassociate: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/floating-ips/%s/disassociate", projectID, id)
	},
	NetworkInterfaces: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/network-interfaces", projectID)
	},
	NetworkInterfaceWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/network-interfaces/%s", projectID, id)
	},
	NetworkInterfaceAttach: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/network-interfaces/%s/attach", projectID, id)
	},
	NetworkInterfaceDetach: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/network-interfaces/%s/detach", projectID, id)
	},
	Volumes: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/volumes", projectID)
	},
	VolumeWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/volumes/%s", projectID, id)
	},
	VolumeResize: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/volumes/%s/resize", projectID, id)
	},
	VolumeAttach: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/volumes/%s/attach", projectID, id)
	},
	VolumeDetach: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/volumes/%s/detach", projectID, id)
	},
	VolumeAttachments: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/volume-attachments", projectID)
	},
	VolumeAttachmentWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/volume-attachments/%s", projectID, id)
	},
	Instances: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/instances", projectID)
	},
	InstanceWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/instances/%s", projectID, id)
	},
	InstanceResize: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/instances/%s/resize", projectID, id)
	},
	ServerGroups: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/server-groups", projectID)
	},
	ServerGroupWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/server-groups/%s", projectID, id)
	},
	CreateKeyPair: func() string {
		return "/v2/iac/key-pairs"
	},
	KeyPairs: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/key-pairs", projectID)
	},
	KeyPairWithName: func(projectID, name string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/key-pairs/%s", projectID, name)
	},
	InternetGateways: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/internet-gateways", projectID)
	},
	InternetGatewayWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/internet-gateways/%s", projectID, id)
	},
	InternetGatewayAttachVPC: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/internet-gateways/%s/attach-vpc", projectID, id)
	},
	InternetGatewayDetachVPC: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/internet-gateways/%s/detach-vpc", projectID, id)
	},
	Snapshots: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/snapshots", projectID)
	},
	SnapshotWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/snapshots/%s", projectID, id)
	},
	LoadBalancers: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/load-balancers", projectID)
	},
	LoadBalancerWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/load-balancers/%s", projectID, id)
	},
	Listeners: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/listeners", projectID)
	},
	ListenerWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/listeners/%s", projectID, id)
	},
	Pools: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/pools", projectID)
	},
	PoolWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/pools/%s", projectID, id)
	},
	HealthMonitors: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/health-monitors", projectID)
	},
	HealthMonitorWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/health-monitors/%s", projectID, id)
	},
	RegistryProjects: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/registries", projectID)
	},
	RegistryProjectWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/registries/%s", projectID, id)
	},
	RobotAccounts: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/robot-accounts", projectID)
	},
	RobotAccountWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/robot-accounts/%s", projectID, id)
	},
	Clusters: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/clusters", projectID)
	},
	ClusterWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/clusters/%s", projectID, id)
	},
	ClusterKubeconfig: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/clusters/%s/kubeconfig", projectID, id)
	},
	WorkerGroups: func(projectID, clusterID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/clusters/%s/worker-groups", projectID, clusterID)
	},
	WorkerGroupWithID: func(projectID, clusterID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/clusters/%s/worker-groups/%s", projectID, clusterID, id)
	},
	RouteTables: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/route-tables", projectID)
	},
	RouteTableWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/route-tables/%s", projectID, id)
	},
	PrivateGateways: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/private-gateways", projectID)
	},
	PrivateGatewayWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/private-gateways/%s", projectID, id)
	},
	PeeringConnections: func() string {
		return "/v2/iac/peering-connections"
	},
	PeeringConnectionWithID: func(id string) string {
		return fmt.Sprintf("/v2/iac/peering-connections/%s", id)
	},
	Flavors: func(zone string) string {
		return fmt.Sprintf("/v2/iac/flavors?zone=%s", zone)
	},
	FlavorWithID: func(id string) string {
		return fmt.Sprintf("/v2/iac/flavors/%s", id)
	},
	Images: func(zone string) string {
		return fmt.Sprintf("/v2/iac/images?zone=%s", zone)
	},
	ImageWithID: func(id string) string {
		return fmt.Sprintf("/v2/iac/images/%s", id)
	},
	VolumeTypes: func(zone string) string {
		return fmt.Sprintf("/v2/iac/volume-types?zone=%s", zone)
	},
	VolumeTypeWithID: func(id string) string {
		return fmt.Sprintf("/v2/iac/volume-types/%s", id)
	},
	Buckets: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/buckets", projectID)
	},
	BucketUsage: func(projectID, bucketName string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/buckets/%s/usage", projectID, bucketName)
	},
	BucketDelete: func(projectID, bucketName, region string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/buckets/%s?region=%s", projectID, bucketName, region)
	},
	// Database Postgres Instance
	DatabasePostgresInstances: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/postgres-instances", projectID)
	},
	DatabasePostgresInstanceWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/postgres-instances/%s", projectID, id)
	},
	DatabasePostgresInstanceScale: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/postgres-instances/%s/scale", projectID, id)
	},
	DatabasePostgresInstanceChangeFlavor: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/postgres-instances/%s/change-flavor", projectID, id)
	},
	DatabasePostgresInstanceExpandVolume: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/postgres-instances/%s/expand-volume", projectID, id)
	},
	DatabasePostgresInstanceEnableAutoExpandVolume: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/postgres-instances/%s/enable-auto-expand-volume", projectID, id)
	},
	DatabasePostgresInstanceDisableAutoExpandVolume: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/postgres-instances/%s/disable-auto-expand-volume", projectID, id)
	},
	DatabasePostgresInstanceEnableTls: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/postgres-instances/%s/enable-tls", projectID, id)
	},
	DatabasePostgresInstanceDisableTls: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/postgres-instances/%s/disable-tls", projectID, id)
	},

	// Database Redis Instance
	DatabaseRedisInstances: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-instances", projectID)
	},
	DatabaseRedisInstanceWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-instances/%s", projectID, id)
	},
	DatabaseRedisInstanceChangeFlavor: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-instances/%s/change-flavor", projectID, id)
	},
	DatabaseRedisInstanceExpandVolume: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-instances/%s/expand-volume", projectID, id)
	},
	DatabaseRedisInstanceEnableAutoExpandVolume: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-instances/%s/enable-auto-expand-volume", projectID, id)
	},
	DatabaseRedisInstanceDisableAutoExpandVolume: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-instances/%s/disable-auto-expand-volume", projectID, id)
	},
	DatabaseRedisInstanceEnableTls: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-instances/%s/enable-tls", projectID, id)
	},
	DatabaseRedisInstanceDisableTls: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-instances/%s/disable-tls", projectID, id)
	},

	// Database Redis Sentinel Instance
	DatabaseRedisSentinelInstances: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-sentinel-instances", projectID)
	},
	DatabaseRedisSentinelInstanceWithID: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-sentinel-instances/%s", projectID, id)
	},
	DatabaseRedisSentinelInstanceScale: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-sentinel-instances/%s/scale", projectID, id)
	},
	DatabaseRedisSentinelInstanceChangeFlavor: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-sentinel-instances/%s/change-flavor", projectID, id)
	},
	DatabaseRedisSentinelInstanceExpandVolume: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-sentinel-instances/%s/expand-volume", projectID, id)
	},
	DatabaseRedisSentinelInstanceEnableAutoExpandVolume: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-sentinel-instances/%s/enable-auto-expand-volume", projectID, id)
	},
	DatabaseRedisSentinelInstanceDisableAutoExpandVolume: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-sentinel-instances/%s/disable-auto-expand-volume", projectID, id)
	},
	DatabaseRedisSentinelInstanceSentinelScale: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-sentinel-instances/%s/sentinel-scale", projectID, id)
	},
	DatabaseRedisSentinelInstanceSentinelChangeFlavor: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-sentinel-instances/%s/sentinel-change-flavor", projectID, id)
	},
	DatabaseRedisSentinelInstanceEnableTls: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-sentinel-instances/%s/enable-tls", projectID, id)
	},
	DatabaseRedisSentinelInstanceDisableTls: func(projectID, id string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/redis-sentinel-instances/%s/disable-tls", projectID, id)
	},

	// Database Flavor
	DatabaseFlavors: func(projectID string) string {
		return fmt.Sprintf("/v2/iac/projects/%s/database/flavor-databases", projectID)
	},

	ResolveProjectByZone: func(zoneID string) string {
		return fmt.Sprintf("/v2/iac/zones/%s/project", zoneID)
	},
}
