package maaspipeline

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceMaasPipeline() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMaasPipelineCreate,
		ReadContext:   resourceMaasPipelineRead,
		UpdateContext: resourceMaasPipelineUpdate,
		DeleteContext: resourceMaasPipelineDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: nameValidation,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"log", "metric"}, false),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: descriptionValidation,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMaasPipelineCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateMaasPipelineRequest{
		Name:        d.Get("name").(string),
		Type:        d.Get("type").(string),
		Description: d.Get("description").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_maas_pipeline create options", map[string]interface{}{
		"name": createOpts.Name,
		"type": createOpts.Type,
	})

	createResp := &dto.MaasPipelineResponse{}
	if _, err := cfg.Client.Post(ctx, client.ApiPath.MaasPipelines(cfg.ProjectID), createOpts, createResp, nil); err != nil {
		return diag.Errorf("Error creating vnpaycloud_maas_pipeline: %s", err)
	}

	d.SetId(createResp.Pipeline.ID)

	return resourceMaasPipelineRead(ctx, d, meta)
}

func resourceMaasPipelineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.MaasPipelineResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.MaasPipelineWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_maas_pipeline"))
	}

	p := resp.Pipeline
	d.Set("name", p.Name)
	d.Set("type", p.Type)
	d.Set("description", p.Description)
	d.Set("status", p.Status)
	d.Set("created_at", p.CreatedAt)

	return nil
}

func resourceMaasPipelineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChanges("name", "description") {
		updateOpts := dto.UpdateMaasPipelineRequest{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		}
		if _, err := cfg.Client.Put(ctx, client.ApiPath.MaasPipelineWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil); err != nil {
			return diag.Errorf("Error updating vnpaycloud_maas_pipeline %s: %s", d.Id(), err)
		}
	}

	return resourceMaasPipelineRead(ctx, d, meta)
}

func resourceMaasPipelineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.MaasPipelineWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_maas_pipeline"))
	}

	return nil
}
