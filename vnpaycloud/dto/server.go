package dto

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"
)

type PowerState int

type DestinationType string

type SourceType string

type Fault struct {
	Code    int       `json:"code"`
	Created time.Time `json:"created"`
	Details string    `json:"details"`
	Message string    `json:"message"`
}

type AttachedVolume struct {
	ID string `json:"id"`
}

type DiskConfig string

const (
	// Auto builds a server with a single partition the size of the target flavor
	// disk and automatically adjusts the filesystem to fit the entire partition.
	// Auto may only be used with images and servers that use a single EXT3
	// partition.
	Auto DiskConfig = "AUTO"

	// Manual builds a server using whatever partition scheme and filesystem are
	// present in the source image. If the target flavor disk is larger, the
	// remaining space is left unpartitioned. This enables images to have non-EXT3
	// filesystems, multiple partitions, and so on, and enables you to manage the
	// disk configuration. It also results in slightly shorter boot times.
	Manual DiskConfig = "MANUAL"
)

const (
	// DestinationLocal DestinationType is for using an ephemeral disk as the
	// destination.
	DestinationLocal DestinationType = "local"

	// DestinationVolume DestinationType is for using a volume as the destination.
	DestinationVolume DestinationType = "volume"

	// SourceBlank SourceType is for a "blank" or empty source.
	SourceBlank SourceType = "blank"

	// SourceImage SourceType is for using images as the source of a block device.
	SourceImage SourceType = "image"

	// SourceSnapshot SourceType is for using a volume snapshot as the source of
	// a block device.
	SourceSnapshot SourceType = "snapshot"

	// SourceVolume SourceType is for using a volume as the source of block
	// device.
	SourceVolume SourceType = "volume"
)

type File struct {
	// Path of the file.
	Path string

	// Contents of the file. Maximum content size is 255 bytes.
	Contents []byte
}

type Personality []*File

type BlockDevice struct {
	// SourceType must be one of: "volume", "snapshot", "image", or "blank".
	SourceType SourceType `json:"source_type" required:"true"`

	// UUID is the unique identifier for the existing volume, snapshot, or
	// image (see above).
	UUID string `json:"uuid,omitempty"`

	// BootIndex is the boot index. It defaults to 0.
	BootIndex int `json:"boot_index"`

	// DeleteOnTermination specifies whether or not to delete the attached volume
	// when the server is deleted. Defaults to `false`.
	DeleteOnTermination bool `json:"delete_on_termination"`

	// DestinationType is the type that gets created. Possible values are "volume"
	// and "local".
	DestinationType DestinationType `json:"destination_type,omitempty"`

	// GuestFormat specifies the format of the block device.
	// Not specifying this will cause the device to be formatted to the default in Nova
	// which is currently vfat.
	GuestFormat string `json:"guest_format,omitempty"`

	// VolumeSize is the size of the volume to create (in gigabytes). This can be
	// omitted for existing volumes.
	VolumeSize int `json:"volume_size,omitempty"`

	// DeviceType specifies the device type of the block devices.
	// Examples of this are disk, cdrom, floppy, lun, etc.
	DeviceType string `json:"device_type,omitempty"`

	// DiskBus is the bus type of the block devices.
	// Examples of this are ide, usb, virtio, scsi, etc.
	DiskBus string `json:"disk_bus,omitempty"`

	// VolumeType is the volume type of the block device.
	// This requires Compute API microversion 2.67 or later.
	VolumeType string `json:"volume_type,omitempty"`

	// Tag is an arbitrary string that can be applied to a block device.
	// Information about the device tags can be obtained from the metadata API
	// and the config drive, allowing devices to be easily identified.
	// This requires Compute API microversion 2.42 or later.
	Tag string `json:"tag,omitempty"`
}

// SchedulerHintOpts represents a set of scheduling hints that are passed to the scheduler.
type SchedulerHintOpts struct {
	// Group specifies a Server Group to place the instance in.
	Group string

	// DifferentHost will place the instance on a compute node that does not
	// host the given instances.
	DifferentHost []string

	// SameHost will place the instance on a compute node that hosts the given
	// instances.
	SameHost []string

	// Query is a conditional statement that results in compute nodes able to
	// host the instance.
	Query []any

	// TargetCell specifies a cell name where the instance will be placed.
	TargetCell string `json:"target_cell,omitempty"`

	// DifferentCell specifies cells names where an instance should not be placed.
	DifferentCell []string `json:"different_cell,omitempty"`

	// BuildNearHostIP specifies a subnet of compute nodes to host the instance.
	BuildNearHostIP string

	// AdditionalProperies are arbitrary key/values that are not validated by nova.
	AdditionalProperties map[string]any
}

// SchedulerHintOptsBuilder builds the scheduler hints into a serializable format.
type SchedulerHintOptsBuilder interface {
	ToSchedulerHintsMap() (map[string]any, error)
}

// ToSchedulerHintsMap assembles a request body for scheduler hints.
func (opts SchedulerHintOpts) ToSchedulerHintsMap() (map[string]any, error) {
	sh := make(map[string]any)

	uuidRegex, _ := regexp.Compile("^[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}$")

	if opts.Group != "" {
		if !uuidRegex.MatchString(opts.Group) {
			err := client.ErrInvalidInput{}
			err.Argument = "servers.schedulerhints.SchedulerHintOpts.Group"
			err.Value = opts.Group
			err.Info = "Group must be a UUID"
			return nil, err
		}
		sh["group"] = opts.Group
	}

	if len(opts.DifferentHost) > 0 {
		for _, diffHost := range opts.DifferentHost {
			if !uuidRegex.MatchString(diffHost) {
				err := client.ErrInvalidInput{}
				err.Argument = "servers.schedulerhints.SchedulerHintOpts.DifferentHost"
				err.Value = opts.DifferentHost
				err.Info = "The hosts must be in UUID format."
				return nil, err
			}
		}
		sh["different_host"] = opts.DifferentHost
	}

	if len(opts.SameHost) > 0 {
		for _, sameHost := range opts.SameHost {
			if !uuidRegex.MatchString(sameHost) {
				err := client.ErrInvalidInput{}
				err.Argument = "servers.schedulerhints.SchedulerHintOpts.SameHost"
				err.Value = opts.SameHost
				err.Info = "The hosts must be in UUID format."
				return nil, err
			}
		}
		sh["same_host"] = opts.SameHost
	}

	/*
		Query can be something simple like:
			 [">=", "$free_ram_mb", 1024]

			Or more complex like:
				['and',
					['>=', '$free_ram_mb', 1024],
					['>=', '$free_disk_mb', 200 * 1024]
				]

		Because of the possible complexity, just make sure the length is a minimum of 3.
	*/
	if len(opts.Query) > 0 {
		if len(opts.Query) < 3 {
			err := client.ErrInvalidInput{}
			err.Argument = "servers.schedulerhints.SchedulerHintOpts.Query"
			err.Value = opts.Query
			err.Info = "Must be a conditional statement in the format of [op,variable,value]"
			return nil, err
		}

		// The query needs to be sent as a marshalled string.
		b, err := json.Marshal(opts.Query)
		if err != nil {
			err := client.ErrInvalidInput{}
			err.Argument = "servers.schedulerhints.SchedulerHintOpts.Query"
			err.Value = opts.Query
			err.Info = "Must be a conditional statement in the format of [op,variable,value]"
			return nil, err
		}

		sh["query"] = string(b)
	}

	if opts.TargetCell != "" {
		sh["target_cell"] = opts.TargetCell
	}

	if len(opts.DifferentCell) > 0 {
		sh["different_cell"] = opts.DifferentCell
	}

	if opts.BuildNearHostIP != "" {
		if _, _, err := net.ParseCIDR(opts.BuildNearHostIP); err != nil {
			err := client.ErrInvalidInput{}
			err.Argument = "servers.schedulerhints.SchedulerHintOpts.BuildNearHostIP"
			err.Value = opts.BuildNearHostIP
			err.Info = "Must be a valid subnet in the form 192.168.1.1/24"
			return nil, err
		}
		ipParts := strings.Split(opts.BuildNearHostIP, "/")
		sh["build_near_host_ip"] = ipParts[0]
		sh["cidr"] = "/" + ipParts[1]
	}

	if opts.AdditionalProperties != nil {
		for k, v := range opts.AdditionalProperties {
			sh[k] = v
		}
	}

	if len(sh) == 0 {
		return sh, nil
	}

	return map[string]any{"os:scheduler_hints": sh}, nil
}

// ServerNetwork is used within CreateOpts to control a new server's network
// attachments.
type ServerNetwork struct {
	// UUID of a network to attach to the newly provisioned server.
	// Required unless Port is provided.
	UUID string

	// Port of a neutron network to attach to the newly provisioned server.
	// Required unless UUID is provided.
	Port string

	// FixedIP specifies a fixed IPv4 address to be used on this network.
	FixedIP string

	// Tag may contain an optional device role tag for the server's virtual
	// network interface. This can be used to identify network interfaces when
	// multiple networks are connected to one server.
	//
	// Requires microversion 2.32 through 2.36 or 2.42 or later.
	Tag string
}

type CreateServerOpts struct {
	// Name is the name to assign to the newly launched server.
	Name string `json:"name" required:"true"`

	// ImageRef is the ID or full URL to the image that contains the
	// server's OS and initial state.
	// Also optional if using the boot-from-volume extension.
	ImageRef string `json:"imageRef"`

	// FlavorRef is the ID or full URL to the flavor that describes the server's specs.
	FlavorRef string `json:"flavorRef"`

	// SecurityGroups lists the names of the security groups to which this server
	// should belong.
	SecurityGroups []string `json:"-"`

	// UserData contains configuration information or scripts to use upon launch.
	// Create will base64-encode it for you, if it isn't already.
	UserData []byte `json:"-"`

	// AvailabilityZone in which to launch the server.
	AvailabilityZone string `json:"availability_zone,omitempty"`

	// Networks dictates how this server will be attached to available networks.
	// By default, the server will be attached to all isolated networks for the
	// tenant.
	// Starting with microversion 2.37 networks can also be an "auto" or "none"
	// string.
	Networks any `json:"-"`

	// Metadata contains key-value pairs (up to 255 bytes each) to attach to the
	// server.
	Metadata map[string]string `json:"metadata,omitempty"`

	// Personality includes files to inject into the server at launch.
	// Create will base64-encode file contents for you.
	Personality Personality `json:"personality,omitempty"`

	// ConfigDrive enables metadata injection through a configuration drive.
	ConfigDrive *bool `json:"config_drive,omitempty"`

	// AdminPass sets the root user password. If not set, a randomly-generated
	// password will be created and returned in the response.
	AdminPass string `json:"adminPass,omitempty"`

	// AccessIPv4 specifies an IPv4 address for the instance.
	AccessIPv4 string `json:"accessIPv4,omitempty"`

	// AccessIPv6 specifies an IPv6 address for the instance.
	AccessIPv6 string `json:"accessIPv6,omitempty"`

	// Min specifies Minimum number of servers to launch.
	Min int `json:"min_count,omitempty"`

	// Max specifies Maximum number of servers to launch.
	Max int `json:"max_count,omitempty"`

	// Tags allows a server to be tagged with single-word metadata.
	// Requires microversion 2.52 or later.
	Tags []string `json:"tags,omitempty"`

	// (Available from 2.90) Hostname specifies the hostname to configure for the
	// instance in the metadata service. Starting with microversion 2.94, this can
	// be a Fully Qualified Domain Name (FQDN) of up to 255 characters in length.
	// If not set, VNPAYCloud will derive the server's hostname from the Name field.
	Hostname string `json:"hostname,omitempty"`

	// BlockDevice describes the mapping of various block devices.
	BlockDevice []BlockDevice `json:"block_device_mapping_v2,omitempty"`

	// DiskConfig [optional] controls how the created server's disk is partitioned.
	DiskConfig DiskConfig `json:"OS-DCF:diskConfig,omitempty"`

	// HypervisorHostname is the name of the hypervisor to which the server is scheduled.
	HypervisorHostname string `json:"hypervisor_hostname,omitempty"`
}

// CreateOptsBuilder allows extensions to add additional parameters to the
// Create request.
type CreateServerOptsBuilder interface {
	ToServerCreateMap() (map[string]any, error)
}

// ToServerCreateMap assembles a request body based on the contents of a
// CreateOpts.
func (opts CreateServerOpts) ToServerCreateMap() (map[string]any, error) {
	// We intentionally don't envelope the body here since we want to strip
	// some fields out and modify others
	b, err := util.BuildRequestBody(opts, "")
	if err != nil {
		return nil, err
	}

	if opts.UserData != nil {
		var userData string
		if _, err := base64.StdEncoding.DecodeString(string(opts.UserData)); err != nil {
			userData = base64.StdEncoding.EncodeToString(opts.UserData)
		} else {
			userData = string(opts.UserData)
		}
		b["user_data"] = &userData
	}

	if len(opts.SecurityGroups) > 0 {
		securityGroups := make([]map[string]any, len(opts.SecurityGroups))
		for i, groupName := range opts.SecurityGroups {
			securityGroups[i] = map[string]any{"name": groupName}
		}
		b["security_groups"] = securityGroups
	}

	switch v := opts.Networks.(type) {
	case []ServerNetwork:
		if len(v) > 0 {
			networks := make([]map[string]any, len(v))
			for i, net := range v {
				networks[i] = make(map[string]any)
				if net.UUID != "" {
					networks[i]["uuid"] = net.UUID
				}
				if net.Port != "" {
					networks[i]["port"] = net.Port
				}
				if net.FixedIP != "" {
					networks[i]["fixed_ip"] = net.FixedIP
				}
				if net.Tag != "" {
					networks[i]["tag"] = net.Tag
				}
			}
			b["networks"] = networks
		}
	case string:
		if v == "auto" || v == "none" {
			b["networks"] = v
		} else {
			return nil, fmt.Errorf(`networks must be a slice of Network struct or a string with "auto" or "none" values, current value is %q`, v)
		}
	}

	if opts.Min != 0 {
		b["min_count"] = opts.Min
	}

	if opts.Max != 0 {
		b["max_count"] = opts.Max
	}

	// Now we do our enveloping
	b = map[string]any{"server": b}

	return b, nil
}

type CreateServerOptsExt struct {
	CreateServerOptsBuilder

	// KeyName is the name of the key pair.
	KeyName string `json:"key_name,omitempty"`
}

// ToServerCreateMap adds the key_name to the base server creation options.
func (opts CreateServerOptsExt) ToServerCreateMap() (map[string]any, error) {
	base, err := opts.CreateServerOptsBuilder.ToServerCreateMap()
	if err != nil {
		return nil, err
	}

	if opts.KeyName == "" {
		return base, nil
	}

	serverMap := base["server"].(map[string]any)
	serverMap["key_name"] = opts.KeyName

	return base, nil
}

type Server struct {
	// ID uniquely identifies this server amongst all other servers,
	// including those not accessible to the current tenant.
	ID string `json:"id"`

	// TenantID identifies the tenant owning this server resource.
	TenantID string `json:"tenant_id"`

	// UserID uniquely identifies the user account owning the tenant.
	UserID string `json:"user_id"`

	// Name contains the human-readable name for the server.
	Name string `json:"name"`

	// Updated and Created contain ISO-8601 timestamps of when the state of the
	// server last changed, and when it was created.
	Updated time.Time `json:"updated"`
	Created time.Time `json:"created"`

	// HostID is the host where the server is located in the cloud.
	HostID string `json:"hostid"`

	// Status contains the current operational status of the server,
	// such as IN_PROGRESS or ACTIVE.
	Status string `json:"status"`

	// Progress ranges from 0..100.
	// A request made against the server completes only once Progress reaches 100.
	Progress int `json:"progress"`

	// AccessIPv4 and AccessIPv6 contain the IP addresses of the server,
	// suitable for remote access for administration.
	AccessIPv4 string `json:"accessIPv4"`
	AccessIPv6 string `json:"accessIPv6"`

	// Image refers to a JSON object, which itself indicates the OS image used to
	// deploy the server.
	Image map[string]any `json:"-"`

	// Flavor refers to a JSON object, which itself indicates the hardware
	// configuration of the deployed server.
	Flavor map[string]any `json:"flavor"`

	// Addresses includes a list of all IP addresses assigned to the server,
	// keyed by pool.
	Addresses map[string]any `json:"addresses"`

	// Metadata includes a list of all user-specified key-value pairs attached
	// to the server.
	Metadata map[string]string `json:"metadata"`

	// Links includes HTTP references to the itself, useful for passing along to
	// other APIs that might want a server reference.
	Links []any `json:"links"`

	// KeyName indicates which public key was injected into the server on launch.
	KeyName string `json:"key_name"`

	// AdminPass will generally be empty ("").  However, it will contain the
	// administrative password chosen when provisioning a new server without a
	// set AdminPass setting in the first place.
	// Note that this is the ONLY time this field will be valid.
	AdminPass string `json:"adminPass"`

	// SecurityGroups includes the security groups that this instance has applied
	// to it.
	SecurityGroups []map[string]any `json:"security_groups"`

	// AttachedVolumes includes the volume attachments of this instance
	AttachedVolumes []AttachedVolume `json:"os-extended-volumes:volumes_attached"`

	// Fault contains failure information about a server.
	Fault Fault `json:"fault"`

	// Tags is a slice/list of string tags in a server.
	// The requires microversion 2.26 or later.
	Tags *[]string `json:"tags"`

	// ServerGroups is a slice of strings containing the UUIDs of the
	// server groups to which the server belongs. Currently this can
	// contain at most one entry.
	// New in microversion 2.71
	ServerGroups *[]string `json:"server_groups"`

	// Host is the host/hypervisor that the instance is hosted on.
	Host string `json:"OS-EXT-SRV-ATTR:host"`

	// InstanceName is the name of the instance.
	InstanceName string `json:"OS-EXT-SRV-ATTR:instance_name"`

	// HypervisorHostname is the hostname of the host/hypervisor that the
	// instance is hosted on.
	HypervisorHostname string `json:"OS-EXT-SRV-ATTR:hypervisor_hostname"`

	// ReservationID is the reservation ID of the instance.
	// This requires microversion 2.3 or later.
	ReservationID *string `json:"OS-EXT-SRV-ATTR:reservation_id"`

	// LaunchIndex is the launch index of the instance.
	// This requires microversion 2.3 or later.
	LaunchIndex *int `json:"OS-EXT-SRV-ATTR:launch_index"`

	// RAMDiskID is the ID of the RAM disk image of the instance.
	// This requires microversion 2.3 or later.
	RAMDiskID *string `json:"OS-EXT-SRV-ATTR:ramdisk_id"`

	// KernelID is the ID of the kernel image of the instance.
	// This requires microversion 2.3 or later.
	KernelID *string `json:"OS-EXT-SRV-ATTR:kernel_id"`

	// Hostname is the hostname of the instance.
	// This requires microversion 2.3 or later.
	Hostname *string `json:"OS-EXT-SRV-ATTR:hostname"`

	// RootDeviceName is the name of the root device of the instance.
	// This requires microversion 2.3 or later.
	RootDeviceName *string `json:"OS-EXT-SRV-ATTR:root_device_name"`

	// Userdata is the userdata of the instance.
	// This requires microversion 2.3 or later.
	Userdata *string `json:"OS-EXT-SRV-ATTR:user_data"`

	TaskState  string     `json:"OS-EXT-STS:task_state"`
	VmState    string     `json:"OS-EXT-STS:vm_state"`
	PowerState PowerState `json:"OS-EXT-STS:power_state"`

	LaunchedAt   time.Time `json:"-"`
	TerminatedAt time.Time `json:"-"`

	// DiskConfig is the disk configuration of the server.
	DiskConfig DiskConfig `json:"OS-DCF:diskConfig"`

	// AvailabilityZone is the availabilty zone the server is in.
	AvailabilityZone string `json:"OS-EXT-AZ:availability_zone"`

	// Locked indicates the lock status of the server
	// This requires microversion 2.9 or later
	Locked *bool `json:"locked"`
}

type CreateServerResponse struct {
	Server Server `json:"server"`
}

type GetServerResponse struct {
	Server Server `json:"server"`
}
