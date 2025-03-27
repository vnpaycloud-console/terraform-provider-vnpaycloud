package volumeV3

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/common"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &volumeV3DataSource{}
	_ datasource.DataSourceWithConfigure = &volumeV3DataSource{}
)

func NewVolumeV3DataSource() datasource.DataSource {
	return &volumeV3DataSource{}
}

type volumeV3DataSource struct {
	client        *common.Client
	serviceClient *gophercloud.ServiceClient
}

func (d *volumeV3DataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client

	var err error
	d.serviceClient, err = openstack.NewBlockStorageV3(d.client.GetProviderClient(), gophercloud.EndpointOpts{
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

func (d *volumeV3DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (d *volumeV3DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: datasourceDescriptions["name"],
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: datasourceDescriptions["description"],
			},
			"size": schema.Int32Attribute{
				Computed:    true,
				Description: datasourceDescriptions["size"],
			},
			"volume_type": schema.StringAttribute{
				Computed:    true,
				Description: datasourceDescriptions["volume_type"],
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: datasourceDescriptions["status"],
			},
			"bootable": schema.BoolAttribute{
				Computed:    true,
				Description: datasourceDescriptions["bootable"],
			},
			"source_volume_id": schema.StringAttribute{
				Computed:    true,
				Description: datasourceDescriptions["source_volume_id"],
			},
		},
		Description: datasourceDescriptions["datasource"],
	}
}

func (d *volumeV3DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "volumeV3DataSource.Read is running")
	var volumeConfig volumeV3DatasourceConfigModel

	diags := req.Config.Get(ctx, &volumeConfig)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if volumeConfig.Id.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing Volume ID",
			"Either a Volume ID must be provided.",
		)
		return
	}

	volume, err := volumes.Get(d.serviceClient, volumeConfig.Id.ValueString()).Extract()

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Volume",
			fmt.Sprintf("An error occurred while reading the volume: %v", err),
		)
		return
	}

	state := volumeV3StateModel{
		Id:             types.StringValue(volume.ID),
		Name:           types.StringValue(volume.Name),
		Size:           types.Int32Value(int32(volume.Size)),
		VolumeType:     types.StringValue(volume.VolumeType),
		Status:         types.StringValue(volume.Status),
		Bootable:       types.BoolValue(volume.Bootable == "true"),
		SourceVolumeId: types.StringValue(volume.SourceVolID),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}
