package provider

import (
	"context"
	"os"
	"terraform-provider-vnpaycloud/common"
	volumeV3 "terraform-provider-vnpaycloud/vnpaycloud/volume_v3"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ provider.Provider = &vnpayCloudProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &vnpayCloudProvider{
			version: version,
		}
	}
}

type vnpayCloudProviderConfig struct {
	BaseUrl                     types.String `tfsdk:"auth_url"`
	ApplicationCredentialId     types.String `tfsdk:"application_credential_id"`
	ApplicationCredentialName   types.String `tfsdk:"application_credential_name"`
	ApplicationCredentialSecret types.String `tfsdk:"application_credential_secret"`
}

type vnpayCloudProvider struct {
	version string
}

func (v *vnpayCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config vnpayCloudProviderConfig

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.BaseUrl.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("auth_url"),
			attrErrorSummaryMsg["unknown_auth_url"],
			attrErrorDetailMsg["unknown_auth_url"],
		)
	}

	if config.ApplicationCredentialId.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("application_credential_id"),
			attrErrorSummaryMsg["unknown_application_credential_id"],
			attrErrorDetailMsg["unknown_application_credential_id"],
		)
	}

	if config.ApplicationCredentialName.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("application_credential_name"),
			attrErrorSummaryMsg["unknown_application_credential_name"],
			attrErrorDetailMsg["unknown_application_credential_name"],
		)
	}

	if config.ApplicationCredentialSecret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("application_credential_secret"),
			attrErrorSummaryMsg["unknown_application_credential_secret"],
			attrErrorDetailMsg["unknown_application_credential_secret"],
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	baseUrl := os.Getenv("VNPAY_CLOUD_AUTH_URL")
	applicationCredentialId := os.Getenv("VNPAY_CLOUD_APPLICATION_CREDENTIAL_ID")
	applicationCredentialName := os.Getenv("VNPAY_CLOUD_APPLICATION_CREDENTIAL_NAME")
	applicationCredentialSecret := os.Getenv("VNPAY_CLOUD_APPLICATION_CREDENTIAL_SECRET")

	if !config.BaseUrl.IsNull() {
		baseUrl = config.BaseUrl.ValueString()
	}

	if !config.ApplicationCredentialId.IsNull() {
		applicationCredentialId = config.ApplicationCredentialId.ValueString()
	}

	if !config.ApplicationCredentialName.IsNull() {
		applicationCredentialName = config.ApplicationCredentialName.ValueString()
	}

	if !config.ApplicationCredentialSecret.IsNull() {
		applicationCredentialSecret = config.ApplicationCredentialSecret.ValueString()
	}

	if baseUrl == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("auth_url"),
			attrErrorSummaryMsg["missing_auth_url"],
			attrErrorDetailMsg["missing_auth_url"],
		)
	}

	if applicationCredentialId == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("application_credential_id"),
			attrErrorSummaryMsg["missing_application_credential_id"],
			attrErrorDetailMsg["missing_application_credential_id"],
		)
	}

	if applicationCredentialName == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("application_credential_name"),
			attrErrorSummaryMsg["missing_application_credential_name"],
			attrErrorDetailMsg["missing_application_credential_name"],
		)
	}

	if applicationCredentialSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("application_credential_secret"),
			attrErrorSummaryMsg["missing_application_credential_secret"],
			attrErrorDetailMsg["missing_application_credential_secret"],
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := common.NewClient(
		ctx,
		baseUrl,
		&common.AuthInfo{
			ApplicationCredentialId:     applicationCredentialId,
			ApplicationCredentialName:   applicationCredentialName,
			ApplicationCredentialSecret: applicationCredentialSecret,
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create VNPAY CLOUD API Client",
			"An unexpected error occurred when creating the VNPAY CLOUD API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"VNPAY CLOUD Client Error: "+err.Error(),
		)
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (v *vnpayCloudProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "vnpaycloud"
	resp.Version = v.version
}

func (v *vnpayCloudProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		volumeV3.NewVolumeV3DataSource,
	}
}

func (v *vnpayCloudProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		volumeV3.NewVolumeV3Resource,
	}
}

func (v *vnpayCloudProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"auth_url": schema.StringAttribute{
				Required:    true,
				Description: descriptions["auth_url"],
			},
			"application_credential_id": schema.StringAttribute{
				Optional:    true,
				Description: descriptions["application_credential_id"],
			},
			"application_credential_name": schema.StringAttribute{
				Optional:    true,
				Description: descriptions["application_credential_name"],
			},
			"application_credential_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: descriptions["application_credential_secret"],
			},
		},
	}
}
