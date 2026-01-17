package server

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"maps"
	"net/http"
	"os"
	"strings"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/hashcode"
	serverInterfaceAttach "terraform-provider-vnpaycloud/vnpaycloud/server-interface-attach"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	flavorsutils "terraform-provider-vnpaycloud/vnpaycloud/util/flavor"
	imagesutils "terraform-provider-vnpaycloud/vnpaycloud/util/image"
)

func ResourceComputeInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceComputeInstanceCreate,
		ReadContext:   resourceComputeInstanceRead,
		UpdateContext: resourceComputeInstanceUpdate,
		DeleteContext: resourceComputeInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceComputeInstanceImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"image_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"image_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"flavor_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Computed: true,
			},
			"flavor_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Computed: true,
			},
			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				// just stash the hash for state & diff comparisons
				StateFunc: func(v interface{}) string {
					switch v := v.(type) {
					case string:
						hash := sha1.Sum([]byte(v))
						return hex.EncodeToString(hash[:])
					default:
						return ""
					}
				},
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"availability_zone_hints": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"availability_zone"},
			},
			"availability_zone": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Computed:         true,
				ConflictsWith:    []string{"availability_zone_hints"},
				DiffSuppressFunc: suppressAvailabilityZoneDetailDiffs,
			},
			"network_mode": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Computed:      false,
				ConflictsWith: []string{"network"},
				ValidateFunc: validation.StringInSlice([]string{
					"auto", "none",
				}, true),
			},
			"network": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uuid": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
						"fixed_ip_v4": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
						"fixed_ip_v6": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
						"mac": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"access_network": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"hypervisor_hostname": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"personality"},
			},
			"metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: false,
			},
			"config_drive": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"admin_pass": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				ForceNew:  false,
			},
			"access_ip_v4": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: false,
			},
			"access_ip_v6": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: false,
			},
			"key_pair": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"block_device": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"uuid": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"volume_size": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"destination_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"boot_index": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						//"delete_on_termination": {
						//	Type:     schema.TypeBool,
						//	Optional: true,
						//	Default:  false,
						//	ForceNew: true,
						//},
						"guest_format": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"volume_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"device_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"disk_bus": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"multiattach": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							ForceNew: true,
						},
					},
				},
			},
			"scheduler_hints": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"different_host": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"same_host": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"query": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"target_cell": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"different_cell": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"build_near_host_ip": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"additional_properties": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
						},
					},
				},
				Set: resourceComputeSchedulerHintsHash,
			},
			"personality": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"file": {
							Type:     schema.TypeString,
							Required: true,
						},
						"content": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Set:           resourceComputeInstancePersonalityHash,
				ConflictsWith: []string{"hypervisor_hostname"},
			},
			"stop_before_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"force_delete": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"all_metadata": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"power_state": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  "active",
				ValidateFunc: validation.StringInSlice([]string{
					"active", "shutoff", "shelved_offloaded", "paused",
				}, true),
				DiffSuppressFunc: suppressPowerStateDiffs,
			},
			"vendor_options": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ignore_resize_confirmation": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
						"detach_ports_before_destroy": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
					},
				},
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CustomizeDiff: customdiff.All(
			// VNPAYCLOUD cannot resize an instance, if its original flavor is deleted, that is why
			// we need to force recreation, if old flavor name or ID is reported as an empty string
			customdiff.ForceNewIfChange("flavor_id", func(ctx context.Context, old, new, meta interface{}) bool {
				return old.(string) == ""
			}),
			customdiff.ForceNewIfChange("flavor_name", func(ctx context.Context, old, new, meta interface{}) bool {
				return old.(string) == ""
			}),
			func(ctx context.Context, d *schema.ResourceDiff, _ interface{}) error {
				currentState, _ := d.GetChange("power_state")
				if currentState == "build" {
					// In "build" state, network and security groups are not yet available
					if err := d.Clear("network"); err != nil {
						return err
					}
					if err := d.Clear("security_groups"); err != nil {
						return err
					}
				}
				return nil
			},
		),
	}
}

func resourceComputeInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud client: %s", err)
	}

	var availabilityZone string
	var networks interface{}

	// Determines the Image ID using the following rules:
	// If a bootable block_device was specified, ignore the image altogether.
	// If an image_id was specified, use it.
	// If an image_name was specified, look up the image ID, report if error.
	imageID, err := getImageIDFromConfig(ctx, tfClient, d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Determines the Flavor ID using the following rules:
	// If a flavor_id was specified, use it.
	// If a flavor_name was specified, lookup the flavor ID, report if error.
	flavorID, err := getFlavorID(ctx, tfClient, d)
	if err != nil {
		return diag.FromErr(err)
	}

	// determine if block_device configuration is correct
	// this includes valid combinations and required attributes
	if err := checkBlockDeviceConfig(d); err != nil {
		return diag.FromErr(err)
	}

	if networkMode := d.Get("network_mode").(string); networkMode == "auto" || networkMode == "none" {
		// Use special string for network option
		// computeClient.Microversion = computeInstanceCreateServerWithNetworkModeMicroversion
		networks = networkMode
		log.Printf("[DEBUG] Create with network options %s", networks)
	} else {
		log.Printf("[DEBUG] Create with specified network options")
		// Build a list of networks with the information given upon creation.
		// Error out if an invalid network configuration was used.
		allInstanceNetworks, err := getAllInstanceNetworks(ctx, d, meta)
		if err != nil {
			return diag.FromErr(err)
		}

		// Build a []servers.Network to pass into the create options.
		networks = expandInstanceNetworks(allInstanceNetworks)
	}

	configDrive := d.Get("config_drive").(bool)

	// Retrieve tags and set microversion if they're provided.
	// instanceTags := computeInstanceTags(d)
	// if len(instanceTags) > 0 {
	// 	computeClient.Microversion = computeInstanceCreateServerWithTagsMicroversion
	// }

	var hypervisorHostname string
	if v, ok := util.GetOkExists(d, "hypervisor_hostname"); ok {
		hypervisorHostname = v.(string)
		// computeClient.Microversion = computeInstanceCreateServerWithHypervisorHostnameMicroversion
	}

	if v, ok := util.GetOkExists(d, "availability_zone"); ok {
		availabilityZone = v.(string)
	} else {
		availabilityZone = d.Get("availability_zone_hints").(string)
	}

	createOpts := &dto.CreateServerOpts{
		Name:               d.Get("name").(string),
		ImageRef:           imageID,
		FlavorRef:          flavorID,
		SecurityGroups:     resourceInstanceSecGroups(d),
		AvailabilityZone:   availabilityZone,
		Networks:           networks,
		HypervisorHostname: hypervisorHostname,
		Metadata:           resourceInstanceMetadata(d),
		ConfigDrive:        &configDrive,
		AdminPass:          d.Get("admin_pass").(string),
		UserData:           []byte(d.Get("user_data").(string)),
		Personality:        resourceInstancePersonality(d),
	}

	if vL, ok := d.GetOk("block_device"); ok {
		blockDevices, err := resourceInstanceBlockDevices(d, vL.([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}

		// // Check if Multiattach was set in any of the Block Devices.
		// // If so, set the client's microversion appropriately.
		// for _, bd := range d.Get("block_device").([]interface{}) {
		// 	if bd.(map[string]interface{})["multiattach"].(bool) {
		// 		computeClient.Microversion = computeInstanceBlockDeviceMultiattachMicroversion
		// 	}
		// }

		// // Check if VolumeType was set in any of the Block Devices.
		// // If so, set the client's microversion appropriately.
		// for _, bd := range blockDevices {
		// 	if bd.VolumeType != "" {
		// 		computeClient.Microversion = computeInstanceBlockDeviceVolumeTypeMicroversion
		// 	}
		// }

		createOpts.BlockDevice = blockDevices
	}

	var createOptsBuilder dto.CreateServerOptsBuilder = createOpts
	if keyName, ok := d.Get("key_pair").(string); ok && keyName != "" {
		createOptsBuilder = &dto.CreateServerOptsExt{
			CreateServerOptsBuilder: createOptsBuilder,
			KeyName:                 keyName,
		}
	}

	var schedulerHints dto.SchedulerHintOpts
	schedulerHintsRaw := d.Get("scheduler_hints").(*schema.Set).List()
	if len(schedulerHintsRaw) > 0 {
		log.Printf("[DEBUG] schedulerhints: %+v", schedulerHintsRaw)
		schedulerHints = resourceInstanceSchedulerHints(d, schedulerHintsRaw[0].(map[string]interface{}))
	}

	log.Printf("[DEBUG] Create Options: %#v", createOpts)

	builder, err := createOptsBuilder.ToServerCreateMap()
	if err != nil {
		return diag.Errorf("Error converting create options to map: %s", err)
	}

	schedulerHintsMap, err := schedulerHints.ToSchedulerHintsMap()
	if err != nil {
		return diag.Errorf("Error converting scheduler hints to map: %s", err)
	}

	maps.Copy(builder, schedulerHintsMap)

	serverResp := &dto.CreateServerResponse{}
	_, err = tfClient.Post(ctx, client.ApiPath.Server, builder, serverResp, &client.RequestOpts{
		OkCodes: []int{200, 202},
	})

	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD server: %s", err)
	}
	log.Printf("[INFO] Instance ID: %s", serverResp.Server.ID)

	// Store the ID now
	d.SetId(serverResp.Server.ID)

	// Wait for the instance to become running so we can get some attributes
	// that aren't available until later.
	log.Printf(
		"[DEBUG] Waiting for instance (%s) to become running",
		serverResp.Server.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"BUILD"},
		Target:     []string{"ACTIVE"},
		Refresh:    ServerStateRefreshFunc(ctx, tfClient, serverResp.Server.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	err = retry.RetryContext(ctx, stateConf.Timeout, func() *retry.RetryError {
		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			log.Printf("[DEBUG] Retrying after error: %s", err)
			return util.CheckForRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return diag.Errorf(
			"Error waiting for instance (%s) to become ready: %s",
			serverResp.Server.ID, err)
	}

	vmState := d.Get("power_state").(string)
	if strings.ToLower(vmState) == "shutoff" {
		_, err = tfClient.Post(ctx, client.ApiPath.ServerActionWithId(d.Id()), map[string]any{"os-stop": nil}, nil, nil)
		if err != nil {
			return diag.Errorf("Error stopping VNPAYCLOUD instance: %s", err)
		}
		stopStateConf := &retry.StateChangeConf{
			//Pending:    []string{"ACTIVE"},
			Target:     []string{"SHUTOFF"},
			Refresh:    ServerStateRefreshFunc(ctx, tfClient, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutCreate),
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		log.Printf("[DEBUG] Waiting for instance (%s) to stop", d.Id())
		_, err = stopStateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.Errorf("Error waiting for instance (%s) to become inactive(shutoff): %s", d.Id(), err)
		}
	}

	return resourceComputeInstanceRead(ctx, d, meta)
}

func resourceComputeInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD client: %s", err)
	}

	serverResp := &dto.GetServerResponse{}
	_, err = tfClient.Get(ctx, client.ApiPath.ServerWithId(d.Id()), serverResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "server"))
	}

	log.Printf("[DEBUG] Retrieved Server %s: %+v", d.Id(), serverResp.Server)

	d.Set("name", serverResp.Server.Name)
	d.Set("created", serverResp.Server.Created.String())
	d.Set("updated", serverResp.Server.Updated.String())

	// Get the instance network and address information
	networks, err := flattenInstanceNetworks(ctx, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Determine the best IPv4 and IPv6 addresses to access the instance with
	hostv4, hostv6 := getInstanceAccessAddresses(d, networks)

	// AccessIPv4/v6 isn't standard in VNPAYCLOUD, but there have been reports
	// of them being used in some environments.
	if serverResp.Server.AccessIPv4 != "" && hostv4 == "" {
		hostv4 = serverResp.Server.AccessIPv4
	}

	if serverResp.Server.AccessIPv6 != "" && hostv6 == "" {
		hostv6 = serverResp.Server.AccessIPv6
	}

	d.Set("network", networks)
	d.Set("access_ip_v4", hostv4)
	d.Set("access_ip_v6", hostv6)

	// Determine the best IP address to use for SSH connectivity.
	// Prefer IPv4 over IPv6.
	var preferredSSHAddress string
	if hostv4 != "" {
		preferredSSHAddress = hostv4
	} else if hostv6 != "" {
		preferredSSHAddress = hostv6
	}

	if preferredSSHAddress != "" {
		// Initialize the connection info
		d.SetConnInfo(map[string]string{
			"type": "ssh",
			"host": preferredSSHAddress,
		})
	}

	d.Set("all_metadata", serverResp.Server.Metadata)

	secGrpNames := []string{}
	for _, sg := range serverResp.Server.SecurityGroups {
		secGrpNames = append(secGrpNames, sg["name"].(string))
	}
	d.Set("security_groups", secGrpNames)

	d.Set("key_pair", serverResp.Server.KeyName)

	flavorID, ok := serverResp.Server.Flavor["id"].(string)
	if !ok {
		return diag.Errorf("Error setting VNPAYCLOUD server's flavor: %v", serverResp.Server.Flavor)
	}
	d.Set("flavor_id", flavorID)

	flavorResp := &dto.GetFlavorResponse{}
	_, err = tfClient.Get(ctx, client.ApiPath.FlavorWithId(flavorID), flavorResp, nil)
	if err != nil {
		if util.ResponseCodeIs(err, http.StatusNotFound) {
			// Original flavor was deleted, but it is possible that instance started
			// with this flavor is still running
			log.Printf("[DEBUG] Original instance flavor id %s could not be found", d.Id())
			d.Set("flavor_id", "")
			d.Set("flavor_name", "")
		} else {
			return diag.FromErr(err)
		}
	} else {
		d.Set("flavor_name", flavorResp.Flavor.Name)
	}

	// Set the instance's image information appropriately
	if err := setImageInformation(ctx, tfClient, &serverResp.Server, d); err != nil {
		return diag.FromErr(err)
	}

	// Set the availability zone
	d.Set("availability_zone", serverResp.Server.AvailabilityZone)

	// Set the region
	d.Set("region", util.GetRegion(d, config))

	// Set the current power_state
	currentStatus := strings.ToLower(serverResp.Server.Status)
	switch currentStatus {
	case "active", "shutoff", "error", "migrating", "shelved_offloaded", "shelved", "build", "paused":
		d.Set("power_state", currentStatus)
	default:
		return diag.Errorf("Invalid power_state for instance %s: %s", d.Id(), serverResp.Server.Status)
	}

	// Populate tags.
	// computeClient.Microversion = computeTagsExtensionMicroversion

	// Set the hypervisor hostname
	d.Set("hypervisor_hostname", serverResp.Server.HypervisorHostname)

	return nil
}

func resourceComputeInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceComputeInstanceRead(ctx, d, meta)
}

func resourceComputeInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	computeClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD client: %s", err)
	}

	if d.Get("stop_before_destroy").(bool) {
		_, err = computeClient.Post(ctx, client.ApiPath.ServerActionWithId(d.Id()), map[string]any{"os-stop": nil}, nil, nil)
		if err != nil {
			log.Printf("[WARN] Error stopping vnpaycloud_compute_server: %s", err)
		} else {
			stopStateConf := &retry.StateChangeConf{
				Pending:    []string{"ACTIVE"},
				Target:     []string{"SHUTOFF"},
				Refresh:    ServerStateRefreshFunc(ctx, computeClient, d.Id()),
				Timeout:    d.Timeout(schema.TimeoutDelete),
				Delay:      10 * time.Second,
				MinTimeout: 3 * time.Second,
			}
			log.Printf("[DEBUG] Waiting for instance (%s) to stop", d.Id())
			_, err = stopStateConf.WaitForStateContext(ctx)
			if err != nil {
				log.Printf("[WARN] Error waiting for instance (%s) to stop: %s, proceeding to delete", d.Id(), err)
			}
		}
	}
	vendorOptionsRaw := d.Get("vendor_options").(*schema.Set)
	var detachPortBeforeDestroy bool
	if vendorOptionsRaw.Len() > 0 {
		vendorOptions := util.ExpandVendorOptions(vendorOptionsRaw.List())
		detachPortBeforeDestroy = vendorOptions["detach_ports_before_destroy"].(bool)
	}
	if detachPortBeforeDestroy {
		allInstanceNetworks, err := getAllInstanceNetworks(ctx, d, meta)
		if err != nil {
			log.Printf("[WARN] Unable to get vnpaycloud_compute_server ports: %s", err)
		} else {
			for _, network := range allInstanceNetworks {
				if network.Port != "" {
					stateConf := &retry.StateChangeConf{
						Pending:    []string{""},
						Target:     []string{"DETACHED"},
						Refresh:    serverInterfaceAttach.ComputeInterfaceAttachDetachFunc(ctx, computeClient, d.Id(), network.Port),
						Timeout:    d.Timeout(schema.TimeoutDelete),
						Delay:      5 * time.Second,
						MinTimeout: 5 * time.Second,
					}
					if _, err = stateConf.WaitForStateContext(ctx); err != nil {
						return diag.Errorf("Error detaching vnpaycloud_compute_server %s: %s", d.Id(), err)
					}
				}
			}
		}
	}
	if d.Get("force_delete").(bool) {
		log.Printf("[DEBUG] Force deleting VNPAYCLOUD Instance %s", d.Id())
		_, err = computeClient.Post(ctx, client.ApiPath.ServerActionWithId(d.Id()), map[string]any{"forceDelete": ""}, nil, nil)
		if err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error force deleting vnpaycloud_compute_server"))
		}
	} else {
		log.Printf("[DEBUG] Deleting VNPAYCLOUD Instance %s", d.Id())
		_, err = computeClient.Delete(ctx, client.ApiPath.ServerWithId(d.Id()), nil)
		if err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_compute_server"))
		}
	}

	// Wait for the instance to delete before moving on.
	log.Printf("[DEBUG] Waiting for instance (%s) to delete", d.Id())

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"ACTIVE", "SHUTOFF"},
		Target:     []string{"DELETED", "SOFT_DELETED"},
		Refresh:    ServerStateRefreshFunc(ctx, computeClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf(
			"Error waiting for instance (%s) to Delete:  %s",
			d.Id(), err)
	}

	return nil
}

func resourceComputeInstanceImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return nil, fmt.Errorf("Error creating VNPAYCLOUD compute client: %s", err)
	}

	results := make([]*schema.ResourceData, 1)
	diagErr := resourceComputeInstanceRead(ctx, d, meta)
	if diagErr != nil {
		return nil, fmt.Errorf("Error reading vnpaycloud_compute_server %s: %v", d.Id(), diagErr)
	}

	serverResp := &dto.GetServerResponse{}
	_, err = tfClient.Get(ctx, client.ApiPath.ServerWithId(d.Id()), serverResp, nil)
	if err != nil {
		return nil, util.CheckDeleted(d, err, "vnpaycloud_compute_server")
	}

	log.Printf("[DEBUG] Retrieved vnpaycloud_compute_server %s volume attachments: %#v",
		d.Id(), serverResp.Server.AttachedVolumes)

	bds := []map[string]interface{}{}
	if len(serverResp.Server.AttachedVolumes) > 0 {
		if err == nil {
			for i, b := range serverResp.Server.AttachedVolumes {
				volumeResp := &dto.GetVolumeResponse{}
				_, err = tfClient.Get(ctx, client.ApiPath.VolumeWithId(tfClient.GetProjectID(), b.ID), volumeResp, nil)

				log.Printf("[DEBUG] retrieved volume %+v", volumeResp.Volume)
				v := map[string]interface{}{
					//"delete_on_termination": true,
					"uuid":             volumeResp.Volume.VolumeImageMetadata["image_id"],
					"boot_index":       i,
					"destination_type": "volume",
					"source_type":      "image",
					"volume_size":      volumeResp.Volume.Size,
					"disk_bus":         "",
					"volume_type":      "",
					"device_type":      "",
				}

				if volumeResp.Volume.Bootable == "true" {
					bds = append(bds, v)
				}
			}
		} else {
			log.Print("[DEBUG] Could not create BlockStorageV3 client, trying BlockStorage")
			for i, b := range serverResp.Server.AttachedVolumes {
				volumeResp := &dto.GetVolumeResponse{}
				_, err = tfClient.Get(ctx, client.ApiPath.VolumeWithId(tfClient.GetProjectID(), b.ID), volumeResp, nil)

				log.Printf("[DEBUG] retrieved volume%+v", volumeResp.Volume)
				v := map[string]interface{}{
					//"delete_on_termination": true,
					"uuid":             volumeResp.Volume.VolumeImageMetadata["image_id"],
					"boot_index":       i,
					"destination_type": "volume",
					"source_type":      "image",
					"volume_size":      volumeResp.Volume.Size,
					"disk_bus":         "",
					"volume_type":      "",
					"device_type":      "",
				}

				if volumeResp.Volume.Bootable == "true" {
					bds = append(bds, v)
				}
			}
		}

		d.Set("block_device", bds)
	}

	metadata := make(map[string]string)
	_, err = tfClient.Get(ctx, client.ApiPath.ServerMetadataWithId(d.Id()), &metadata, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to read metadata for vnpaycloud_compute_server %s: %s", d.Id(), err)
	}

	d.Set("metadata", metadata)

	results[0] = d

	return results, nil
}

// ServerStateRefreshFunc returns a retry.StateRefreshFunc that is used to watch
// an VNPAYCLOUD instance.
func ServerStateRefreshFunc(ctx context.Context, computeClient *client.Client, instanceID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		serverResp := &dto.GetServerResponse{}
		_, err := computeClient.Get(ctx, client.ApiPath.ServerWithId(instanceID), serverResp, nil)
		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return serverResp.Server, "DELETED", nil
			}
			return nil, "", err
		}

		return serverResp.Server, serverResp.Server.Status, nil
	}
}

func resourceInstanceSecGroups(d *schema.ResourceData) []string {
	rawSecGroups := d.Get("security_groups").(*schema.Set).List()
	res := make([]string, len(rawSecGroups))
	for i, raw := range rawSecGroups {
		res[i] = raw.(string)
	}
	return res
}

func resourceInstanceMetadata(d *schema.ResourceData) map[string]string {
	m := make(map[string]string)
	for key, val := range d.Get("metadata").(map[string]interface{}) {
		m[key] = val.(string)
	}
	return m
}

func resourceInstanceBlockDevices(_ *schema.ResourceData, bds []interface{}) ([]dto.BlockDevice, error) {
	blockDeviceOpts := make([]dto.BlockDevice, len(bds))
	for i, bd := range bds {
		bdM := bd.(map[string]interface{})
		blockDeviceOpts[i] = dto.BlockDevice{
			UUID:       bdM["uuid"].(string),
			VolumeSize: bdM["volume_size"].(int),
			BootIndex:  bdM["boot_index"].(int),
			//DeleteOnTermination: bdM["delete_on_termination"].(bool),
			GuestFormat: bdM["guest_format"].(string),
			VolumeType:  bdM["volume_type"].(string),
			DeviceType:  bdM["device_type"].(string),
			DiskBus:     bdM["disk_bus"].(string),
		}

		sourceType := bdM["source_type"].(string)
		switch sourceType {
		case "blank":
			blockDeviceOpts[i].SourceType = dto.SourceBlank
		case "image":
			blockDeviceOpts[i].SourceType = dto.SourceImage
		case "snapshot":
			blockDeviceOpts[i].SourceType = dto.SourceSnapshot
		case "volume":
			blockDeviceOpts[i].SourceType = dto.SourceVolume
		default:
			return blockDeviceOpts, fmt.Errorf("unknown block device source type %s", sourceType)
		}

		destinationType := bdM["destination_type"].(string)
		switch destinationType {
		case "local":
			blockDeviceOpts[i].DestinationType = dto.DestinationLocal
		case "volume":
			blockDeviceOpts[i].DestinationType = dto.DestinationVolume
		default:
			return blockDeviceOpts, fmt.Errorf("unknown block device destination type %s", destinationType)
		}
	}

	log.Printf("[DEBUG] Block Device Options: %+v", blockDeviceOpts)
	return blockDeviceOpts, nil
}

func resourceInstanceSchedulerHints(ctx *schema.ResourceData, schedulerHintsRaw map[string]interface{}) dto.SchedulerHintOpts {
	differentHost := []string{}
	if v, ok := schedulerHintsRaw["different_host"].([]interface{}); ok {
		for _, dh := range v {
			differentHost = append(differentHost, dh.(string))
		}
	}

	sameHost := []string{}
	if v, ok := schedulerHintsRaw["same_host"].([]interface{}); ok {
		for _, sh := range v {
			sameHost = append(sameHost, sh.(string))
		}
	}

	query := []interface{}{}
	if v, ok := schedulerHintsRaw["query"].([]interface{}); ok {
		for _, q := range v {
			query = append(query, q.(string))
		}
	}

	differentCell := []string{}
	if v, ok := schedulerHintsRaw["different_cell"].([]interface{}); ok {
		for _, dh := range v {
			differentCell = append(differentCell, dh.(string))
		}
	}

	schedulerHints := dto.SchedulerHintOpts{
		Group:                schedulerHintsRaw["group"].(string),
		DifferentHost:        differentHost,
		SameHost:             sameHost,
		Query:                query,
		TargetCell:           schedulerHintsRaw["target_cell"].(string),
		DifferentCell:        differentCell,
		BuildNearHostIP:      schedulerHintsRaw["build_near_host_ip"].(string),
		AdditionalProperties: schedulerHintsRaw["additional_properties"].(map[string]interface{}),
	}

	return schedulerHints
}

func getImageIDFromConfig(ctx context.Context, imageClient *client.Client, d *schema.ResourceData) (string, error) {
	// If block_device was used, an Image does not need to be specified, unless an image/local
	// combination was used. This emulates normal boot behavior. Otherwise, ignore the image altogether.
	if vL, ok := d.GetOk("block_device"); ok {
		needImage := false
		for _, v := range vL.([]interface{}) {
			vM := v.(map[string]interface{})
			if vM["source_type"] == "image" && vM["destination_type"] == "local" {
				needImage = true
			}
		}
		if !needImage {
			return "", nil
		}
	}

	if imageID := d.Get("image_id").(string); imageID != "" {
		return imageID, nil
	}
	// try the OS_IMAGE_ID environment variable
	if v := os.Getenv("OS_IMAGE_ID"); v != "" {
		return v, nil
	}

	imageName := d.Get("image_name").(string)
	if imageName == "" {
		// try the OS_IMAGE_NAME environment variable
		if v := os.Getenv("OS_IMAGE_NAME"); v != "" {
			imageName = v
		}
	}

	if imageName != "" {
		imageID, err := imagesutils.IDFromName(ctx, imageClient, imageName)
		if err != nil {
			return "", err
		}
		return imageID, nil
	}

	return "", fmt.Errorf("Neither a boot device, image ID, or image name were able to be determined")
}

func setImageInformation(ctx context.Context, imageClient *client.Client, server *dto.Server, d *schema.ResourceData) error {
	// If block_device was used, an Image does not need to be specified, unless an image/local
	// combination was used. This emulates normal boot behavior. Otherwise, ignore the image altogether.
	if vL, ok := d.GetOk("block_device"); ok {
		needImage := false
		for _, v := range vL.([]interface{}) {
			vM := v.(map[string]interface{})
			if vM["source_type"] == "image" && vM["destination_type"] == "local" {
				needImage = true
			}
		}
		if !needImage {
			d.Set("image_id", "Attempt to boot from volume - no image supplied")
			return nil
		}
	}

	if server.Image["id"] != nil {
		imageID := server.Image["id"].(string)
		if imageID != "" {
			d.Set("image_id", imageID)
			imageResp := &dto.GetImageResponse{}
			_, err := imageClient.Get(ctx, client.ApiPath.ImageWithId(imageID), imageResp, nil)
			if err != nil {
				if util.ResponseCodeIs(err, http.StatusNotFound) {
					// If the image name can't be found, set the value to "Image not found".
					// The most likely scenario is that the image no longer exists in the Image Service
					// but the instance still has a record from when it existed.
					d.Set("image_name", "Image not found")
					return nil
				}
				return err
			}
			d.Set("image_name", imageResp.Image.Name)
		}
	}

	return nil
}

func getFlavorID(ctx context.Context, computeClient *client.Client, d *schema.ResourceData) (string, error) {
	if flavorID := d.Get("flavor_id").(string); flavorID != "" {
		return flavorID, nil
	}
	// Try the OS_FLAVOR_ID environment variable
	if v := os.Getenv("OS_FLAVOR_ID"); v != "" {
		return v, nil
	}

	flavorName := d.Get("flavor_name").(string)
	if flavorName == "" {
		// Try the OS_FLAVOR_NAME environment variable
		if v := os.Getenv("OS_FLAVOR_NAME"); v != "" {
			flavorName = v
		}
	}

	if flavorName != "" {
		flavorID, err := flavorsutils.IDFromName(ctx, computeClient, flavorName)
		if err != nil {
			return "", err
		}
		return flavorID, nil
	}

	return "", fmt.Errorf("Neither a flavor_id or flavor_name could be determined")
}

func resourceComputeSchedulerHintsHash(v interface{}) int {
	var buf bytes.Buffer

	m, ok := v.(map[string]interface{})
	if !ok {
		return hashcode.String(buf.String())
	}
	if m == nil {
		return hashcode.String(buf.String())
	}

	if m["group"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", m["group"].(string)))
	}

	if m["target_cell"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", m["target_cell"].(string)))
	}

	if m["build_host_near_ip"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", m["build_host_near_ip"].(string)))
	}

	if m["additional_properties"] != nil {
		for _, v := range m["additional_properties"].(map[string]interface{}) {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}

	buf.WriteString(fmt.Sprintf("%s-", m["different_host"].([]interface{})))
	buf.WriteString(fmt.Sprintf("%s-", m["same_host"].([]interface{})))
	buf.WriteString(fmt.Sprintf("%s-", m["query"].([]interface{})))
	buf.WriteString(fmt.Sprintf("%s-", m["different_cell"].([]interface{})))

	return hashcode.String(buf.String())
}

func checkBlockDeviceConfig(d *schema.ResourceData) error {
	if vL, ok := d.GetOk("block_device"); ok {
		for _, v := range vL.([]interface{}) {
			vM := v.(map[string]interface{})

			if vM["source_type"] != "blank" && vM["uuid"] == "" {
				return fmt.Errorf("You must specify a uuid for %s block device types", vM["source_type"])
			}

			if vM["source_type"] == "image" && vM["destination_type"] == "volume" {
				if vM["volume_size"] == 0 {
					return fmt.Errorf("You must specify a volume_size when creating a volume from an image")
				}
			}

			if vM["source_type"] == "blank" && vM["destination_type"] == "local" {
				if vM["volume_size"] == 0 {
					return fmt.Errorf("You must specify a volume_size when creating a blank block device")
				}
			}
		}
	}

	return nil
}

func resourceComputeInstancePersonalityHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["file"].(string)))

	return hashcode.String(buf.String())
}

func resourceInstancePersonality(d *schema.ResourceData) dto.Personality {
	var personalities dto.Personality

	if v := d.Get("personality"); v != nil {
		personalityList := v.(*schema.Set).List()
		if len(personalityList) > 0 {
			for _, p := range personalityList {
				rawPersonality := p.(map[string]interface{})
				file := dto.File{
					Path:     rawPersonality["file"].(string),
					Contents: []byte(rawPersonality["content"].(string)),
				}

				log.Printf("[DEBUG] VNPAYCLOUD Compute Instance Personality: %+v", file)

				personalities = append(personalities, &file)
			}
		}
	}

	return personalities
}

// suppressAvailabilityZoneDetailDiffs will suppress diffs when a user specifies an
// availability zone in the format of `az:host:node` and Nova/Compute responds with
// only `az`.
func suppressAvailabilityZoneDetailDiffs(_, old, new string, _ *schema.ResourceData) bool {
	if strings.Contains(new, ":") {
		parts := strings.Split(new, ":")
		az := parts[0]

		if az == old {
			return true
		}
	}

	return false
}

// suppressPowerStateDiffs will allow a state of "error" or "migrating" even though we don't
// allow them as a user input.
func suppressPowerStateDiffs(_, old, _ string, _ *schema.ResourceData) bool {
	if old == "error" || old == "migrating" {
		return true
	}

	return false
}
