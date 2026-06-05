package client

import (
	"strings"
	"testing"
)

func TestApiPathBuilders(t *testing.T) {
	projectID := "proj-123"
	resourceID := "res-456"
	zone := "zone-a"

	tests := []struct {
		name     string
		got      string
		wantPfx  string
		wantSfx  string
		wantFull string
	}{
		// VPC
		{"VPCs", ApiPath.VPCs(projectID), "/v2/iac/projects/proj-123", "/vpcs", ""},
		{"VPCWithID", ApiPath.VPCWithID(projectID, resourceID), "", "", "/v2/iac/projects/proj-123/vpcs/res-456"},
		{"VPCSetSNAT", ApiPath.VPCSetSNAT(projectID, resourceID), "", "/snat", ""},

		// Subnet
		{"Subnets", ApiPath.Subnets(projectID), "/v2/iac/projects/proj-123", "/subnets", ""},
		{"SubnetWithID", ApiPath.SubnetWithID(projectID, resourceID), "", "", "/v2/iac/projects/proj-123/subnets/res-456"},
		{"SubnetEnableSNAT", ApiPath.SubnetEnableSNAT(projectID, resourceID), "", "/enable-snat", ""},
		{"SubnetDisableSNAT", ApiPath.SubnetDisableSNAT(projectID, resourceID), "", "/disable-snat", ""},

		// Security Group
		{"SecurityGroups", ApiPath.SecurityGroups(projectID), "/v2/iac/projects/proj-123", "/security-groups", ""},
		{"SecurityGroupWithID", ApiPath.SecurityGroupWithID(projectID, resourceID), "", "", "/v2/iac/projects/proj-123/security-groups/res-456"},

		// Security Group Rule
		{"SecurityGroupRules", ApiPath.SecurityGroupRules(projectID), "/v2/iac/projects/proj-123", "", ""},
		{"SecurityGroupRuleWithID", ApiPath.SecurityGroupRuleWithID(projectID, resourceID), "", resourceID, ""},

		// Floating IP
		{"FloatingIPs", ApiPath.FloatingIPs(projectID), "/v2/iac/projects/proj-123", "", ""},
		{"FloatingIPWithID", ApiPath.FloatingIPWithID(projectID, resourceID), "", resourceID, ""},
		{"FloatingIPAssociate", ApiPath.FloatingIPAssociate(projectID, resourceID), "", "/associate", ""},
		{"FloatingIPDisassociate", ApiPath.FloatingIPDisassociate(projectID, resourceID), "", "/disassociate", ""},

		// Network Interface
		{"NetworkInterfaces", ApiPath.NetworkInterfaces(projectID), "/v2/iac/projects/proj-123", "", ""},
		{"NetworkInterfaceWithID", ApiPath.NetworkInterfaceWithID(projectID, resourceID), "", resourceID, ""},
		{"NetworkInterfaceAttach", ApiPath.NetworkInterfaceAttach(projectID, resourceID), "", "/attach", ""},
		{"NetworkInterfaceDetach", ApiPath.NetworkInterfaceDetach(projectID, resourceID), "", "/detach", ""},

		// Volume
		{"Volumes", ApiPath.Volumes(projectID), "/v2/iac/projects/proj-123", "", ""},
		{"VolumeWithID", ApiPath.VolumeWithID(projectID, resourceID), "", resourceID, ""},
		{"VolumeResize", ApiPath.VolumeResize(projectID, resourceID), "", "/resize", ""},
		{"VolumeAttach", ApiPath.VolumeAttach(projectID, resourceID), "", "/attach", ""},
		{"VolumeDetach", ApiPath.VolumeDetach(projectID, resourceID), "", "/detach", ""},
		{"VolumeAttachments", ApiPath.VolumeAttachments(projectID), "", "", ""},
		{"VolumeAttachmentWithID", ApiPath.VolumeAttachmentWithID(projectID, resourceID), "", resourceID, ""},

		// Instance
		{"Instances", ApiPath.Instances(projectID), "/v2/iac/projects/proj-123", "", ""},
		{"InstanceWithID", ApiPath.InstanceWithID(projectID, resourceID), "", resourceID, ""},
		{"InstanceResize", ApiPath.InstanceResize(projectID, resourceID), "", "/resize", ""},

		// Server Group
		{"ServerGroups", ApiPath.ServerGroups(projectID), "/v2/iac/projects/proj-123", "", ""},
		{"ServerGroupWithID", ApiPath.ServerGroupWithID(projectID, resourceID), "", resourceID, ""},

		// KeyPair
		{"CreateKeyPair", ApiPath.CreateKeyPair(), "/v2/iac/", "", ""},
		{"KeyPairs", ApiPath.KeyPairs(projectID), "/v2/iac/projects/proj-123", "", ""},
		{"KeyPairWithName", ApiPath.KeyPairWithName(projectID, "mykey"), "", "mykey", ""},

		// Internet Gateway
		{"InternetGateways", ApiPath.InternetGateways(projectID), "/v2/iac/projects/proj-123", "", ""},
		{"InternetGatewayWithID", ApiPath.InternetGatewayWithID(projectID, resourceID), "", resourceID, ""},
		{"InternetGatewayAttachVPC", ApiPath.InternetGatewayAttachVPC(projectID, resourceID), "", "/attach-vpc", ""},
		{"InternetGatewayDetachVPC", ApiPath.InternetGatewayDetachVPC(projectID, resourceID), "", "/detach-vpc", ""},

		// Snapshot
		{"Snapshots", ApiPath.Snapshots(projectID), "/v2/iac/projects/proj-123", "", ""},
		{"SnapshotWithID", ApiPath.SnapshotWithID(projectID, resourceID), "", resourceID, ""},

		// Load Balancer
		{"LoadBalancers", ApiPath.LoadBalancers(projectID), "/v2/iac/projects/proj-123", "", ""},
		{"LoadBalancerWithID", ApiPath.LoadBalancerWithID(projectID, resourceID), "", resourceID, ""},
		{"LBFlavors", ApiPath.LBFlavors(projectID), "/v2/iac/projects/proj-123/lb-flavors", "", ""},

		// Listener
		{"Listeners", ApiPath.Listeners(projectID), "/v2/iac/projects/proj-123", "", ""},
		{"ListenerWithID", ApiPath.ListenerWithID(projectID, resourceID), "", resourceID, ""},

		// Pool
		{"Pools", ApiPath.Pools(projectID), "/v2/iac/projects/proj-123", "", ""},
		{"PoolWithID", ApiPath.PoolWithID(projectID, resourceID), "", resourceID, ""},

		// Health Monitor
		{"HealthMonitors", ApiPath.HealthMonitors(projectID), "/v2/iac/projects/proj-123", "", ""},
		{"HealthMonitorWithID", ApiPath.HealthMonitorWithID(projectID, resourceID), "", resourceID, ""},

		// Kubernetes
		{"Clusters", ApiPath.Clusters(projectID), "/v2/iac/projects/proj-123", "", ""},
		{"ClusterWithID", ApiPath.ClusterWithID(projectID, resourceID), "", resourceID, ""},
		{"ClusterKubeconfig", ApiPath.ClusterKubeconfig(projectID, resourceID), "", "/kubeconfig", ""},
		{"WorkerGroups", ApiPath.WorkerGroups(projectID, resourceID), "", "/worker-groups", ""},
		{"WorkerGroupWithID", ApiPath.WorkerGroupWithID(projectID, resourceID, "wg-1"), "", "wg-1", ""},

		// Registry
		{"RegistryProjects", ApiPath.RegistryProjects(projectID), "/v2/iac/projects/proj-123/registries", "", ""},
		{"RegistryProjectWithID", ApiPath.RegistryProjectWithID(projectID, resourceID), "/v2/iac/projects/proj-123/registries/", resourceID, ""},
		{"RobotAccounts", ApiPath.RobotAccounts(projectID), "/v2/iac/projects/proj-123/robot-accounts", "", ""},
		{"RobotAccountWithID", ApiPath.RobotAccountWithID(projectID, "robot-1"), "/v2/iac/projects/proj-123/robot-accounts/", "robot-1", ""},
		{"RegistryPermissions", ApiPath.RegistryPermissions(projectID), "/v2/iac/projects/proj-123/registry-permissions", "", ""},

		// Route Table
		{"RouteTables", ApiPath.RouteTables(projectID), "/v2/iac/projects/proj-123", "", ""},
		{"RouteTableWithID", ApiPath.RouteTableWithID(projectID, resourceID), "", resourceID, ""},

		// Private Gateway
		{"PrivateGateways", ApiPath.PrivateGateways(projectID), "/v2/iac/projects/proj-123", "", ""},
		{"PrivateGatewayWithID", ApiPath.PrivateGatewayWithID(projectID, resourceID), "", resourceID, ""},

		// VPC Peering (global, no project)
		{"PeeringConnections", ApiPath.PeeringConnections(), "/v2/iac/", "", ""},
		{"PeeringConnectionWithID", ApiPath.PeeringConnectionWithID(resourceID), "", resourceID, ""},

		// Flavor (global, zone-scoped)
		{"Flavors", ApiPath.Flavors(zone), "/v2/iac/", zone, ""},
		{"FlavorWithID", ApiPath.FlavorWithID(resourceID), "", resourceID, ""},

		// Image (global, zone-scoped)
		{"Images", ApiPath.Images(zone), "/v2/iac/", zone, ""},
		{"ImageWithID", ApiPath.ImageWithID(resourceID), "", resourceID, ""},

		// Volume Type (global, zone-scoped)
		{"VolumeTypes", ApiPath.VolumeTypes(zone), "/v2/iac/", zone, ""},
		{"VolumeTypeWithID", ApiPath.VolumeTypeWithID(resourceID), "", resourceID, ""},

		// S3 Bucket
		{"Buckets", ApiPath.Buckets(projectID), "/v2/iac/projects/proj-123", "", ""},
		{"BucketUsage", ApiPath.BucketUsage(projectID, "my-bucket"), "", "/usage", ""},
		{"BucketDelete", ApiPath.BucketDelete(projectID, "my-bucket", "us-east-1"), "", "us-east-1", ""},

		// Database Postgres Instance
		{"DatabasePostgresInstances", ApiPath.DatabasePostgresInstances(projectID), "/v2/iac/projects/proj-123", "/database/postgres-instances", ""},
		{"DatabasePostgresInstanceWithID", ApiPath.DatabasePostgresInstanceWithID(projectID, resourceID), "", "", "/v2/iac/projects/proj-123/database/postgres-instances/res-456"},
		{"DatabasePostgresInstanceScale", ApiPath.DatabasePostgresInstanceScale(projectID, resourceID), "", "/scale", ""},
		{"DatabasePostgresInstanceChangeFlavor", ApiPath.DatabasePostgresInstanceChangeFlavor(projectID, resourceID), "", "/change-flavor", ""},
		{"DatabasePostgresInstanceExpandVolume", ApiPath.DatabasePostgresInstanceExpandVolume(projectID, resourceID), "", "/expand-volume", ""},
		{"DatabasePostgresInstanceEnableAutoExpandVolume", ApiPath.DatabasePostgresInstanceEnableAutoExpandVolume(projectID, resourceID), "", "/enable-auto-expand-volume", ""},
		{"DatabasePostgresInstanceDisableAutoExpandVolume", ApiPath.DatabasePostgresInstanceDisableAutoExpandVolume(projectID, resourceID), "", "/disable-auto-expand-volume", ""},
		{"DatabasePostgresInstanceEnableTls", ApiPath.DatabasePostgresInstanceEnableTls(projectID, resourceID), "", "/enable-tls", ""},
		{"DatabasePostgresInstanceDisableTls", ApiPath.DatabasePostgresInstanceDisableTls(projectID, resourceID), "", "/disable-tls", ""},

		// Database Redis Instance
		{"DatabaseRedisInstances", ApiPath.DatabaseRedisInstances(projectID), "/v2/iac/projects/proj-123", "/database/redis-instances", ""},
		{"DatabaseRedisInstanceWithID", ApiPath.DatabaseRedisInstanceWithID(projectID, resourceID), "", "", "/v2/iac/projects/proj-123/database/redis-instances/res-456"},
		{"DatabaseRedisInstanceChangeFlavor", ApiPath.DatabaseRedisInstanceChangeFlavor(projectID, resourceID), "", "/change-flavor", ""},
		{"DatabaseRedisInstanceExpandVolume", ApiPath.DatabaseRedisInstanceExpandVolume(projectID, resourceID), "", "/expand-volume", ""},
		{"DatabaseRedisInstanceEnableAutoExpandVolume", ApiPath.DatabaseRedisInstanceEnableAutoExpandVolume(projectID, resourceID), "", "/enable-auto-expand-volume", ""},
		{"DatabaseRedisInstanceDisableAutoExpandVolume", ApiPath.DatabaseRedisInstanceDisableAutoExpandVolume(projectID, resourceID), "", "/disable-auto-expand-volume", ""},
		{"DatabaseRedisInstanceEnableTls", ApiPath.DatabaseRedisInstanceEnableTls(projectID, resourceID), "", "/enable-tls", ""},
		{"DatabaseRedisInstanceDisableTls", ApiPath.DatabaseRedisInstanceDisableTls(projectID, resourceID), "", "/disable-tls", ""},

		// Database Redis Sentinel Instance
		{"DatabaseRedisSentinelInstances", ApiPath.DatabaseRedisSentinelInstances(projectID), "/v2/iac/projects/proj-123", "/database/redis-sentinel-instances", ""},
		{"DatabaseRedisSentinelInstanceWithID", ApiPath.DatabaseRedisSentinelInstanceWithID(projectID, resourceID), "", "", "/v2/iac/projects/proj-123/database/redis-sentinel-instances/res-456"},
		{"DatabaseRedisSentinelInstanceScale", ApiPath.DatabaseRedisSentinelInstanceScale(projectID, resourceID), "", "/scale", ""},
		{"DatabaseRedisSentinelInstanceChangeFlavor", ApiPath.DatabaseRedisSentinelInstanceChangeFlavor(projectID, resourceID), "", "/change-flavor", ""},
		{"DatabaseRedisSentinelInstanceExpandVolume", ApiPath.DatabaseRedisSentinelInstanceExpandVolume(projectID, resourceID), "", "/expand-volume", ""},
		{"DatabaseRedisSentinelInstanceEnableAutoExpandVolume", ApiPath.DatabaseRedisSentinelInstanceEnableAutoExpandVolume(projectID, resourceID), "", "/enable-auto-expand-volume", ""},
		{"DatabaseRedisSentinelInstanceDisableAutoExpandVolume", ApiPath.DatabaseRedisSentinelInstanceDisableAutoExpandVolume(projectID, resourceID), "", "/disable-auto-expand-volume", ""},
		{"DatabaseRedisSentinelInstanceSentinelScale", ApiPath.DatabaseRedisSentinelInstanceSentinelScale(projectID, resourceID), "", "/sentinel-scale", ""},
		{"DatabaseRedisSentinelInstanceSentinelChangeFlavor", ApiPath.DatabaseRedisSentinelInstanceSentinelChangeFlavor(projectID, resourceID), "", "/sentinel-change-flavor", ""},
		{"DatabaseRedisSentinelInstanceEnableTls", ApiPath.DatabaseRedisSentinelInstanceEnableTls(projectID, resourceID), "", "/enable-tls", ""},
		{"DatabaseRedisSentinelInstanceDisableTls", ApiPath.DatabaseRedisSentinelInstanceDisableTls(projectID, resourceID), "", "/disable-tls", ""},

		// Database Flavor
		{"DatabaseFlavors", ApiPath.DatabaseFlavors(projectID), "/v2/iac/projects/proj-123", "/database/flavor-databases", ""},

		// Zone resolution
		{"ResolveProjectByZone", ApiPath.ResolveProjectByZone("zone-1"), "/v2/iac/", "/project", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got == "" {
				t.Fatal("path builder returned empty string")
			}
			if tt.wantFull != "" && tt.got != tt.wantFull {
				t.Errorf("expected %q, got %q", tt.wantFull, tt.got)
			}
			if tt.wantPfx != "" && !strings.HasPrefix(tt.got, tt.wantPfx) {
				t.Errorf("expected prefix %q, got %q", tt.wantPfx, tt.got)
			}
			if tt.wantSfx != "" && !strings.HasSuffix(tt.got, tt.wantSfx) {
				t.Errorf("expected suffix %q, got %q", tt.wantSfx, tt.got)
			}
		})
	}
}
