package serverGroup

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/vnpaycloud-console/gophercloud/v2/openstack/compute/v2/servergroups"
)

func ResourceComputeServerGroupV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceComputeServerGroupV2Create,
		ReadContext:   resourceComputeServerGroupV2Read,
		Update:        nil,
		DeleteContext: resourceComputeServerGroupV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
				ForceNew: true,
				Required: true,
			},

			"policies": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"rules": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Computed: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_server_per_host": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},

			"members": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"value_specs": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceComputeServerGroupV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	computeClient, err := config.ComputeV2Client(ctx, util.GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD compute client: %s", err)
	}

	name := d.Get("name").(string)

	rawPolicies := d.Get("policies").([]interface{})
	policies := expandComputeServerGroupV2Policies(computeClient, rawPolicies)
	var policy string
	if len(policies) == 1 {
		policy = policies[0]
	}

	createOpts := ComputeServerGroupV2CreateOpts{
		servergroups.CreateOpts{
			Name:   name,
			Policy: policy,
		},
		util.MapValueSpecs(d),
	}

	rulesVal, rulesPresent := d.GetOk("rules")
	if policy == "anti-affinity" && rulesPresent {
		computeClient.Microversion = "2.64"
		createOpts.CreateOpts.Rules = &servergroups.Rules{
			MaxServerPerHost: expandComputeServerGroupV2RulesMaxServerPerHost(rulesVal.([]interface{})),
		}
	}

	log.Printf("[DEBUG] vnpaycloud_compute_servergroup create options: %#v", createOpts)
	newSG, err := servergroups.Create(ctx, computeClient, createOpts).Extract()
	if err != nil {
		// return an error right away
		if createOpts.CreateOpts.Rules != nil {
			return diag.Errorf("Error creating vnpaycloud_compute_servergroup %s: %s", name, err)
		}

		log.Printf("[DEBUG] Falling back to legacy API call due to: %#v", err)
		// fallback to legacy microversion
		createOpts = ComputeServerGroupV2CreateOpts{
			servergroups.CreateOpts{
				Name:     name,
				Policies: expandComputeServerGroupV2Policies(computeClient, rawPolicies),
			},
			util.MapValueSpecs(d),
		}
		log.Printf("[DEBUG] vnpaycloud_compute_servergroup create options: %#v", createOpts)
		newSG, err = servergroups.Create(ctx, computeClient, createOpts).Extract()
		if err != nil {
			return diag.Errorf("Error creating vnpaycloud_compute_servergroup %s: %s", name, err)
		}
	}

	d.SetId(newSG.ID)

	return resourceComputeServerGroupV2Read(ctx, d, meta)
}

func resourceComputeServerGroupV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	computeClient, err := config.ComputeV2Client(ctx, util.GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD compute client: %s", err)
	}

	// Attempt to read with microversion 2.64
	computeClient.Microversion = "2.64"
	sg, err := servergroups.Get(ctx, computeClient, d.Id()).Extract()
	if err != nil {
		log.Printf("[DEBUG] Falling back to legacy API call due to: %#v", err)
		// fallback to legacy microversion
		computeClient.Microversion = ""

		sg, err = servergroups.Get(ctx, computeClient, d.Id()).Extract()
		if err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_compute_servergroup"))
		}
	}

	log.Printf("[DEBUG] Retrieved vnpaycloud_compute_servergroup %s: %#v", d.Id(), sg)

	d.Set("name", sg.Name)
	d.Set("members", sg.Members)
	d.Set("region", util.GetRegion(d, config))
	if sg.Policy != nil && *sg.Policy != "" {
		d.Set("policies", []string{*sg.Policy})
	} else {
		d.Set("policies", sg.Policies)
	}
	if sg.Rules != nil {
		d.Set("rules", []map[string]interface{}{{"max_server_per_host": sg.Rules.MaxServerPerHost}})
	} else {
		d.Set("rules", nil)
	}

	return nil
}

func resourceComputeServerGroupV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	computeClient, err := config.ComputeV2Client(ctx, util.GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD compute client: %s", err)
	}

	if err := servergroups.Delete(ctx, computeClient, d.Id()).ExtractErr(); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_compute_servergroup"))
	}

	return nil
}
