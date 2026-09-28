package maaspipelineexecutor

import (
	"context"
	"regexp"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var descriptionRegex = regexp.MustCompile(`^[a-zA-Z0-9-_. ]*$`)

func ResourceMaasPipelineExecutor() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMaasPipelineExecutorCreate,
		ReadContext:   resourceMaasPipelineExecutorRead,
		UpdateContext: resourceMaasPipelineExecutorUpdate,
		DeleteContext: resourceMaasPipelineExecutorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customizeExecutorDiff,
		Schema: map[string]*schema.Schema{
			"pipeline_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"order": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 100),
			},
			"executor_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice(
					[]string{executorTypeStaticLabels, executorTypeRenameFields, executorTypeRenameLabels}, false),
			},
			"static_labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"rename_fields": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"rename_labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 255),
					validation.StringMatch(descriptionRegex, "must contain only letters, digits, spaces, hyphens, underscores and dots"),
				),
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
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

func pipelineLockKey(pipelineID string) string {
	return "vnpaycloud_maas_pipeline/" + pipelineID
}

func customizeExecutorDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	for _, key := range []string{"executor_type", "static_labels", "rename_fields", "rename_labels"} {
		if !d.NewValueKnown(key) {
			return nil
		}
	}

	return validateExecutorShape(
		d.Get("executor_type").(string),
		expandStringMap(d.Get("static_labels")),
		expandStringMap(d.Get("rename_fields")),
		expandStringMap(d.Get("rename_labels")),
	)
}

func resourceMaasPipelineExecutorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	lockKey := pipelineLockKey(d.Get("pipeline_id").(string))
	cfg.MutexKV.Lock(lockKey)
	defer cfg.MutexKV.Unlock(lockKey)

	createOpts := dto.CreateMaasPipelineExecutorRequest{
		PipelineID:   d.Get("pipeline_id").(string),
		Order:        d.Get("order").(int),
		Description:  d.Get("description").(string),
		ExecutorType: d.Get("executor_type").(string),
		StaticLabels: expandStringMap(d.Get("static_labels")),
		RenameFields: expandStringMap(d.Get("rename_fields")),
		RenameLabels: expandStringMap(d.Get("rename_labels")),
	}

	tflog.Debug(ctx, "vnpaycloud_maas_pipeline_executor create options", map[string]interface{}{
		"pipeline_id":   createOpts.PipelineID,
		"order":         createOpts.Order,
		"executor_type": createOpts.ExecutorType,
	})

	createResp := &dto.MaasPipelineExecutorResponse{}
	if _, err := cfg.Client.Post(ctx, client.ApiPath.MaasPipelineExecutors(cfg.ProjectID), createOpts, createResp, nil); err != nil {
		return diag.Errorf("Error creating vnpaycloud_maas_pipeline_executor: %s", err)
	}

	d.SetId(createResp.PipelineExecutor.ID)

	return resourceMaasPipelineExecutorRead(ctx, d, meta)
}

func resourceMaasPipelineExecutorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.MaasPipelineExecutorResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.MaasPipelineExecutorWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_maas_pipeline_executor"))
	}

	setExecutorData(d, &resp.PipelineExecutor)

	return nil
}

func resourceMaasPipelineExecutorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	lockKey := pipelineLockKey(d.Get("pipeline_id").(string))
	cfg.MutexKV.Lock(lockKey)
	defer cfg.MutexKV.Unlock(lockKey)

	if d.HasChanges("order", "description", "executor_type", "static_labels", "rename_fields", "rename_labels") {
		updateOpts := dto.UpdateMaasPipelineExecutorRequest{
			Order:        d.Get("order").(int),
			Description:  d.Get("description").(string),
			ExecutorType: d.Get("executor_type").(string),
			StaticLabels: expandStringMap(d.Get("static_labels")),
			RenameFields: expandStringMap(d.Get("rename_fields")),
			RenameLabels: expandStringMap(d.Get("rename_labels")),
		}
		if _, err := cfg.Client.Put(ctx, client.ApiPath.MaasPipelineExecutorWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil); err != nil {
			return diag.Errorf("Error updating vnpaycloud_maas_pipeline_executor %s: %s", d.Id(), err)
		}
	}

	return resourceMaasPipelineExecutorRead(ctx, d, meta)
}

func resourceMaasPipelineExecutorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.MaasPipelineExecutorWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_maas_pipeline_executor"))
	}

	return nil
}
