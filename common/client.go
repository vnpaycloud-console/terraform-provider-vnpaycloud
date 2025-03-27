package common

import (
	"context"
	"fmt"
	"net/url"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type AuthInfo struct {
	ApplicationCredentialId     string
	ApplicationCredentialName   string
	ApplicationCredentialSecret string
}

type Client struct {
	providerClient *gophercloud.ProviderClient
}

func (c *Client) GetProviderClient() *gophercloud.ProviderClient {
	return c.providerClient
}

func NewClient(ctx context.Context, baseUrl string, authInfo *AuthInfo) (*Client, error) {
	parsedURL, err := url.Parse(baseUrl)

	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %v", err)
	}

	authOpts := gophercloud.AuthOptions{
		IdentityEndpoint:            parsedURL.String(),
		ApplicationCredentialID:     authInfo.ApplicationCredentialId,
		ApplicationCredentialName:   authInfo.ApplicationCredentialName,
		ApplicationCredentialSecret: authInfo.ApplicationCredentialSecret,
		AllowReauth:                 true,
	}

	providerClient, err := openstack.AuthenticatedClient(authOpts)

	if err != nil {
		return nil, fmt.Errorf("Failed to authenticate with OpenStack: %v", err)
	}

	tflog.Debug(ctx, "Successfully authenticated with OpenStack.", map[string]interface{}{
		"token": providerClient.Token(),
	})

	return &Client{providerClient: providerClient}, nil
}
