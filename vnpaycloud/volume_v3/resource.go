package volumeV3

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/common"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/volumeactions"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &volumeV3Resource{}
	_ resource.ResourceWithConfigure = &volumeV3Resource{}
)

type volumeV3Resource struct {
	client        *common.Client
	serviceClient *gophercloud.ServiceClient
}

func NewVolumeV3Resource() resource.Resource {
	return &volumeV3Resource{}
}

func (r *volumeV3Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*common.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *common.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client

	var err error

	r.serviceClient, err = openstack.NewBlockStorageV3(r.client.GetProviderClient(), gophercloud.EndpointOpts{
		Region: common.EndpointOpts_Region,
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Initialize Block Storage Client",
			fmt.Sprintf("An error occurred while creating the Block Storage v3 client: %v. Please verify your provider configuration or credentials.", err),
		)

		return
	}
}

func (r *volumeV3Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (r *volumeV3Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: resourceDescriptions["name"],
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: resourceDescriptions["description"],
			},
			"size": schema.Int32Attribute{
				Required:    true,
				Description: resourceDescriptions["size"],
			},
			"volume_type": schema.StringAttribute{
				Required:      true,
				Description:   resourceDescriptions["volume_type"],
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, // NOTE: Same as Openstack but Vertix allows update
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: resourceDescriptions["status"],
			},
			"bootable": schema.BoolAttribute{
				Computed:    true,
				Description: resourceDescriptions["bootable"],
			},
			"source_volume_id": schema.StringAttribute{
				Computed:    true,
				Description: resourceDescriptions["source_volume_id"],
			},
		},
		Description: resourceDescriptions["resource"],
	}
}

func (r *volumeV3Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "volumeV3Resource.Create is running")
	var plan volumeV3ResourceConfigModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	createOpts := volumes.CreateOpts{
		AvailabilityZone: "nova",
		Description:      plan.Description.ValueString(),
		Metadata:         map[string]string{},
		Name:             plan.Name.ValueString(),
		Size:             int(plan.Size.ValueInt32()),
		VolumeType:       plan.VolumeType.ValueString(),
	}
	volume, err := volumes.Create(r.serviceClient, createOpts).Extract()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating volume",
			fmt.Sprintf("Could not create volume: %s", err.Error()),
		)
		return
	}

	state := volumeV3StateModel{
		Id:             types.StringValue(volume.ID),
		Name:           types.StringValue(volume.Name),
		Description:    types.StringValue(volume.Description),
		Size:           types.Int32Value(int32(volume.Size)),
		VolumeType:     types.StringValue(volume.VolumeType),
		Status:         types.StringValue(volume.Status),
		Bootable:       types.BoolValue(volume.Bootable == "true"),
		SourceVolumeId: types.StringValue(volume.SourceVolID),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *volumeV3Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "volumeV3Resource.Delete is running")
	var state volumeV3StateModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if state.Id.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing Volume ID",
			"Either a Volume ID must be provided.",
		)
		return
	}

	err := volumes.Delete(r.serviceClient, state.Id.ValueString(), volumes.DeleteOpts{Cascade: true}).ExtractErr()

	if err != nil {
		resp.Diagnostics.AddError(
			"Volume Deletion Failed",
			fmt.Sprintf(
				"The volume with ID '%s' could not be deleted. Error: %v. Please verify that the volume exists and is not in use.",
				state.Id.ValueString(),
				err,
			),
		)
		return
	}
}

func (r *volumeV3Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "volumeV3Resource.Read is running")
	var state volumeV3StateModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if state.Id.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing Volume ID",
			"Either a Volume ID must be provided.",
		)
		return
	}

	volume, err := volumes.Get(r.serviceClient, state.Id.ValueString()).Extract()

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Volume",
			fmt.Sprintf("An error occurred while reading the volume: %v", err),
		)
		return
	}

	newState := volumeV3StateModel{
		Id:             types.StringValue(volume.ID),
		Name:           types.StringValue(volume.Name),
		Size:           types.Int32Value(int32(volume.Size)),
		VolumeType:     types.StringValue(volume.VolumeType),
		Status:         types.StringValue(volume.Status),
		Bootable:       types.BoolValue(volume.Bootable == "true"),
		SourceVolumeId: types.StringValue(volume.SourceVolID),
	}

	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *volumeV3Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "volumeV3Resource.Update is running")
	var plan volumeV3ResourceConfigModel
	var state volumeV3StateModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	hasUpdate := false
	updateOpts := volumes.UpdateOpts{}

	if !state.Name.Equal(plan.Name) {
		_name := plan.Name.ValueString()
		updateOpts.Name = &_name
		state.Name = plan.Name
		hasUpdate = true
	}

	if !state.Description.Equal(plan.Description) {
		_desc := plan.Description.ValueString()
		updateOpts.Description = &_desc
		state.Description = plan.Description
		hasUpdate = true
	}

	if hasUpdate {
		if _, err := volumes.Update(r.serviceClient, state.Id.ValueString(), updateOpts).Extract(); err != nil {
			resp.Diagnostics.AddError(
				"Failed to Update Volume",
				fmt.Sprintf("An error occurred while updating the volume with ID '%s': %v", state.Id.ValueString(), err),
			)
			return
		}

		diags = resp.State.Set(ctx, &state)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !state.Size.Equal(plan.Size) {
		if err := volumeactions.ExtendSize(
			r.serviceClient,
			state.Id.ValueString(),
			volumeactions.ExtendSizeOpts{NewSize: int(plan.Size.ValueInt32())},
		).ExtractErr(); err != nil {
			resp.Diagnostics.AddError(
				"Failed to Extend Volume Size",
				fmt.Sprintf(
					"An error occurred while trying to extend the volume size for ID '%s'. Current size: %d GB, Requested size: %d GB. Error: %v",
					state.Id.ValueString(),
					state.Size.ValueInt32(),
					plan.Size.ValueInt32(),
					err,
				),
			)
			return
		}
	}

	if !state.VolumeType.Equal(plan.VolumeType) {
		if err := volumeactions.ChangeType(
			r.serviceClient,
			state.Id.ValueString(),
			volumeactions.ChangeTypeOpts{NewType: plan.VolumeType.ValueString()},
		).ExtractErr(); err != nil {
			resp.Diagnostics.AddError(
				"Failed to Change Volume Type",
				fmt.Sprintf(
					"An error occurred while trying to change the volume type for ID '%s'. Current type: '%s', Requested type: '%s'. Error: %v",
					state.Id.ValueString(),
					state.VolumeType.ValueString(),
					plan.VolumeType.ValueString(),
					err,
				),
			)
			return
		}
	}

}
