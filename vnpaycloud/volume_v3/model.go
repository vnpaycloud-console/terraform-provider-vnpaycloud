package volumeV3

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type volumeV3StateModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Size           types.Int32  `tfsdk:"size"`
	VolumeType     types.String `tfsdk:"volume_type"`
	Status         types.String `tfsdk:"status"`
	Bootable       types.Bool   `tfsdk:"bootable"`
	SourceVolumeId types.String `tfsdk:"source_volume_id"`
}

type volumeV3ResourceConfigModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Size           types.Int32  `tfsdk:"size"`
	VolumeType     types.String `tfsdk:"volume_type"`
	Status         types.String `tfsdk:"status"`
	Bootable       types.Bool   `tfsdk:"bootable"`
	SourceVolumeId types.String `tfsdk:"source_volume_id"`
}

type volumeV3DatasourceConfigModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Size           types.Int32  `tfsdk:"size"`
	VolumeType     types.String `tfsdk:"volume_type"`
	Status         types.String `tfsdk:"status"`
	Bootable       types.Bool   `tfsdk:"bootable"`
	SourceVolumeId types.String `tfsdk:"source_volume_id"`
}
