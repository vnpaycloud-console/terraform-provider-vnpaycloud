package flavor

import (
	"context"
	"log"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceComputeFlavor() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceComputeFlavorRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"flavor_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name", "min_ram", "min_disk"},
			},

			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"flavor_id"},
			},

			"min_ram": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"flavor_id"},
			},

			"ram": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},

			"vcpus": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},

			"min_disk": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"flavor_id"},
			},

			"disk": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},

			"swap": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},

			"rx_tx_factor": {
				Type:     schema.TypeFloat,
				Optional: true,
				ForceNew: true,
			},

			"is_public": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			// Computed values
			"extra_specs": {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

// dataSourceComputeFlavorRead performs the flavor lookup.
func dataSourceComputeFlavorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	computeClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCloud compute client: %s", err)
	}

	var allFlavors []dto.Flavor
	if v := d.Get("flavor_id").(string); v != "" {
		var flavor dto.GetFlavorResponse
		_, err = computeClient.Get(ctx, client.ApiPath.FlavorWithId(v), &flavor, nil)
		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return diag.Errorf("No Flavor found")
			}
			return diag.Errorf("Unable to retrieve VNPAYCloud %s flavor: %s", v, err)
		}

		allFlavors = append(allFlavors, flavor.Flavor)
	} else {
		accessType := dto.FlavorAllAccess
		if v, ok := util.GetOkExists(d, "is_public"); ok {
			if v, ok := v.(bool); ok {
				if v {
					accessType = dto.FlavorPublicAccess
				} else {
					accessType = dto.FlavorPrivateAccess
				}
			}
		}
		listOpts := dto.ListFlavorParams{
			MinDisk:    d.Get("min_disk").(int),
			MinRAM:     d.Get("min_ram").(int),
			AccessType: accessType,
		}

		log.Printf("[DEBUG] vnpaycloud_compute_flavor ListOpts: %#v", listOpts)
		listResp := dto.ListFlavorsResponse{}
		_, err = computeClient.Get(ctx, client.ApiPath.FlavorDetailWithParams(listOpts), &listResp, nil)
		if err != nil {
			return diag.Errorf("Unable to query VNPAYCloud flavors: %s", err)
		}

		allFlavors = listResp.Flavors
	}

	// Loop through all flavors to find a more specific one.
	if len(allFlavors) > 0 {
		var filteredFlavors []dto.Flavor
		for _, flavor := range allFlavors {
			if v := d.Get("name").(string); v != "" {
				if flavor.Name != v {
					continue
				}
			}

			if v := d.Get("description").(string); v != "" {
				if flavor.Description != v {
					continue
				}
			}

			// d.GetOk is used because 0 might be a valid choice.
			if v, ok := d.GetOk("ram"); ok {
				if flavor.RAM != v.(int) {
					continue
				}
			}

			if v, ok := d.GetOk("vcpus"); ok {
				if flavor.VCPUs != v.(int) {
					continue
				}
			}

			if v, ok := d.GetOk("disk"); ok {
				if flavor.Disk != v.(int) {
					continue
				}
			}

			if v, ok := d.GetOk("swap"); ok {
				if flavor.Swap != v.(int) {
					continue
				}
			}

			if v, ok := d.GetOk("rx_tx_factor"); ok {
				if flavor.RxTxFactor != v.(float64) {
					continue
				}
			}

			filteredFlavors = append(filteredFlavors, flavor)
		}

		allFlavors = filteredFlavors
	}

	if len(allFlavors) < 1 {
		return diag.Errorf("Your query returned no results. " +
			"Please change your search criteria and try again.")
	}

	if len(allFlavors) > 1 {
		log.Printf("[DEBUG] Multiple results found: %#v", allFlavors)
		return diag.Errorf("Your query returned more than one result. " +
			"Please try a more specific search criteria")
	}

	return diag.FromErr(dataSourceComputeFlavorAttributes(ctx, d, computeClient, &allFlavors[0]))
}

// dataSourceComputeFlavorAttributes populates the fields of a Flavor resource.
func dataSourceComputeFlavorAttributes(ctx context.Context, d *schema.ResourceData, computeClient *client.Client, flavor *dto.Flavor) error {
	log.Printf("[DEBUG] Retrieved vnpaycloud_compute_flavor %s: %#v", flavor.ID, flavor)

	d.SetId(flavor.ID)
	d.Set("name", flavor.Name)
	d.Set("description", flavor.Description)
	d.Set("flavor_id", flavor.ID)
	d.Set("disk", flavor.Disk)
	d.Set("ram", flavor.RAM)
	d.Set("rx_tx_factor", flavor.RxTxFactor)
	d.Set("swap", flavor.Swap)
	d.Set("vcpus", flavor.VCPUs)
	d.Set("is_public", flavor.IsPublic)

	es := map[string]map[string]string{}

	_, err := computeClient.Get(ctx, client.ApiPath.FlavorExtraSpecs(d.Id()), &es, nil)
	if err != nil {
		return err
	}

	if err := d.Set("extra_specs", es["extra_specs"]); err != nil {
		log.Printf("[WARN] Unable to set extra_specs for vnpaycloud_compute_flavor %s: %s", d.Id(), err)
	}

	return nil
}
