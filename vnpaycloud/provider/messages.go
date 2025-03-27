package provider

var attrErrorSummaryMsg = map[string]string{
	"unknown_auth_url": "Unknown VNPAY CLOUD Auth Url",

	"unknown_region": "Unknown VNPAY CLOUD Region",

	"unknown_application_credential_id": "Unknown VNPAY CLOUD API Application credential id",

	"unknown_application_credential_name": "Unknown VNPAY CLOUD API Application credential name",

	"unknown_application_credential_secret": "Unknown VNPAY CLOUD API Application credential secret",

	"missing_auth_url": "Missing VNPAY CLOUD Auth Url",

	"missing_region": "Missing VNPAY CLOUD Region",

	"missing_application_credential_id": "Missing VNPAY CLOUD API Application credential id",

	"missing_application_credential_name": "Missing VNPAY CLOUD API Application credential name",

	"missing_application_credential_secret": "Missing VNPAY CLOUD API Application credential secret",
}

var attrErrorDetailMsg = map[string]string{
	"unknown_auth_url": "The provider cannot create the VNPAY CLOUD API client as there is an unknown configuration value for the VNPAY CLOUD auth_url. " +
		"Either target apply the source of the value first, set the value statically in the configuration, or use the VNPAY_CLOUD_AUTH_URL environment variable.",

	"unknown_region": "The provider cannot create the VNPAY CLOUD API client as there is an unknown configuration value for the VNPAY CLOUD region. " +
		"Either target apply the source of the value first, set the value statically in the configuration, or use the VNPAY_CLOUD_REGION environment variable.",

	"unknown_application_credential_id": "The provider cannot create the VNPAY CLOUD API client as there is an unknown configuration value for the VNPAY CLOUD application_credential_id. " +
		"Either target apply the source of the value first, set the value statically in the configuration, or use the VNPAY_CLOUD_APPLICATION_CREDENTIAL_ID environment variable.",

	"unknown_application_credential_name": "The provider cannot create the VNPAY CLOUD API client as there is an unknown configuration value for the VNPAY CLOUD application_credential_name. " +
		"Either target apply the source of the value first, set the value statically in the configuration, or use the VNPAY_CLOUD_APPLICATION_CREDENTIAL_NAME environment variable.",

	"unknown_application_credential_secret": "The provider cannot create the VNPAY CLOUD API client as there is an unknown configuration value for the VNPAY CLOUD application_credential_secret. " +
		"Either target apply the source of the value first, set the value statically in the configuration, or use the VNPAY_CLOUD_APPLICATION_CREDENTIAL_SECRET environment variable.",

	"missing_auth_url": "The provider cannot create the VNPAY CLOUD API client as there is a missing or empty value for the VNPAY CLOUD auth_url. " +
		"Either target apply the source of the value first, set the value statically in the configuration, or use the VNPAY_CLOUD_AUTH_URL environment variable.",

	"missing_region": "The provider cannot create the VNPAY CLOUD API client as there is a missing or empty value for the VNPAY CLOUD region. " +
		"Either target apply the source of the value first, set the value statically in the configuration, or use the VNPAY_CLOUD_REGION environment variable.",

	"missing_application_credential_id": "The provider cannot create the VNPAY CLOUD API client as there is a missing or empty value for the VNPAY CLOUD application_credential_id. " +
		"Either target apply the source of the value first, set the value statically in the configuration, or use the VNPAY_CLOUD_APPLICATION_CREDENTIAL_ID environment variable.",

	"missing_application_credential_name": "The provider cannot create the VNPAY CLOUD API client as there is a missing or empty value for the VNPAY CLOUD application_credential_name. " +
		"Either target apply the source of the value first, set the value statically in the configuration, or use the VNPAY_CLOUD_APPLICATION_CREDENTIAL_NAME environment variable.",

	"missing_application_credential_secret": "The provider cannot create the VNPAY CLOUD API client as there is a missing or empty value for the VNPAY CLOUD application_credential_secret. " +
		"Either target apply the source of the value first, set the value statically in the configuration, or use the VNPAY_CLOUD_APPLICATION_CREDENTIAL_SECRET environment variable.",
}
