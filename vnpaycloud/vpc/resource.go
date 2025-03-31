package vpc

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/vnpaycloud-console/gophercloud/v2"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/vpcs"
)

var (
	_ resource.Resource              = &VPCResource{}
	_ resource.ResourceWithConfigure = &VPCResource{}
)

type VPCResource struct {
	client        *common.Client
	serviceClient *gophercloud.ServiceClient
}

func NewVPCResource() resource.Resource {
	return &VPCResource{}
}

func (r *VPCResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	return
}

func (r *VPCResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

func (r *VPCResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	return
}

func (r *VPCResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Implementation for creating a VPC
}

func (r *VPCResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Implementation for reading a VPC
}

func (r *VPCResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Implementation for updating a VPC
}

func (r *VPCResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Implementation for deleting a VPC
}

func (r *VPCResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Implementation for importing a VPC
}
