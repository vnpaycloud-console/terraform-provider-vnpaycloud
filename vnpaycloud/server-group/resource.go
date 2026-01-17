package serverGroup

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
)

func ResourceComputeServerGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceComputeServerGroupCreate,
		ReadContext:   resourceComputeServerGroupRead,
		Update:        nil,
		DeleteContext: resourceComputeServerGroupDelete,
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

func resourceComputeServerGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	computeClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD compute client: %s", err)
	}

	name := d.Get("name").(string)

	rawPolicies := d.Get("policies").([]interface{})
	policies := expandComputeServerGroupPolicies(computeClient, rawPolicies)
	var policy string
	if len(policies) == 1 {
		policy = policies[0]
	}

	createOpts := dto.CreateServerGroupOpts{
		Name:       name,
		Policy:     policy,
		ValueSpecs: util.MapValueSpecs(d),
	}

	rulesVal, rulesPresent := d.GetOk("rules")
	if policy == "anti-affinity" && rulesPresent {
		createOpts.Rules = &dto.Rules{
			MaxServerPerHost: expandComputeServerGroupRulesMaxServerPerHost(rulesVal.([]interface{})),
		}
	}

	log.Printf("[DEBUG] vnpaycloud_compute_servergroup create options: %#v", createOpts)
	createReq := dto.CreateServerGroupRequest{
		ServerGroup: createOpts,
	}
	createResp := &dto.CreateServerGroupResponse{}
	reqOpts := &client.RequestOpts{
		OkCodes: []int{200},
	}
	_, err = computeClient.Post(ctx, client.ApiPath.ServerGroup, createReq, createResp, reqOpts)
	if err != nil {
		// return an error right away
		if createOpts.Rules != nil {
			return diag.Errorf("Error creating vnpaycloud_compute_servergroup %s: %s", name, err)
		}

		log.Printf("[DEBUG] Falling back to legacy API call due to: %#v", err)
		// fallback to legacy microversion
		createOpts = dto.CreateServerGroupOpts{
			Name:       name,
			Policies:   expandComputeServerGroupPolicies(computeClient, rawPolicies),
			ValueSpecs: util.MapValueSpecs(d),
		}
		log.Printf("[DEBUG] vnpaycloud_compute_servergroup create options: %#v", createOpts)
		createReq = dto.CreateServerGroupRequest{
			ServerGroup: createOpts,
		}
		_, err = computeClient.Post(ctx, client.ApiPath.ServerGroup, createReq, createResp, reqOpts)
		if err != nil {
			return diag.Errorf("Error creating vnpaycloud_compute_servergroup %s: %s", name, err)
		}
	}

	d.SetId(createResp.ServerGroup.ID)

	return resourceComputeServerGroupRead(ctx, d, meta)
}

func resourceComputeServerGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	computeClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD compute client: %s", err)
	}

	sgResp := &dto.GetServerGroupResponse{}
	_, err = computeClient.Get(ctx, client.ApiPath.ServerGroupWithId(d.Id()), sgResp, nil)
	if err != nil {
		log.Printf("[DEBUG] Falling back to legacy API call due to: %#v", err)

		_, err = computeClient.Get(ctx, client.ApiPath.ServerGroupWithId(d.Id()), sgResp, nil)
		if err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_compute_servergroup"))
		}
	}

	log.Printf("[DEBUG] Retrieved vnpaycloud_compute_servergroup %s: %#v", d.Id(), sgResp.ServerGroup)

	d.Set("name", sgResp.ServerGroup.Name)
	d.Set("members", sgResp.ServerGroup.Members)
	d.Set("region", util.GetRegion(d, config))
	if sgResp.ServerGroup.Policy != nil && *sgResp.ServerGroup.Policy != "" {
		d.Set("policies", []string{*sgResp.ServerGroup.Policy})
	} else {
		d.Set("policies", sgResp.ServerGroup.Policies)
	}
	if sgResp.ServerGroup.Rules != nil {
		d.Set("rules", []map[string]interface{}{{"max_server_per_host": sgResp.ServerGroup.Rules.MaxServerPerHost}})
	} else {
		d.Set("rules", nil)
	}

	return nil
}

func resourceComputeServerGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	computeClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD compute client: %s", err)
	}

	if _, err := computeClient.Delete(ctx, client.ApiPath.ServerGroupWithId(d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_compute_servergroup"))
	}

	return nil
}
