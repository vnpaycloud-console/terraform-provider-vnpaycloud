package kubernetescluster

import (
	"context"
	"encoding/base64"

	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DataSourceKubernetesKubeconfig exposes the admin kubeconfig of a
// Kubernetes cluster via GET /v2/iac/projects/{project_id}/clusters/{id}/kubeconfig.
func DataSourceKubernetesKubeconfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClusterKubeconfigRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the Kubernetes cluster to fetch the kubeconfig for.",
			},
			"is_private_access": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When true, return a kubeconfig whose API server points at the cluster's private IP (for access within the VPC).",
			},
			"kubeconfig_b64": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Base64-encoded kubeconfig as returned by the API.",
			},
			"content": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Decoded kubeconfig YAML, ready to write to a file.",
			},
		},
	}
}

func dataSourceClusterKubeconfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	clusterID := d.Get("cluster_id").(string)

	path := client.ApiPath.ClusterKubeconfig(cfg.ProjectID, clusterID)
	if d.Get("is_private_access").(bool) {
		path += "?is_private_access=true"
	}

	resp := &dto.KubeconfigResponse{}
	_, err := cfg.Client.Get(ctx, path, resp, nil)
	if err != nil {
		return diag.Errorf("Error fetching kubeconfig for vnpaycloud_kubernetes_cluster %s: %s", clusterID, err)
	}

	d.SetId(clusterID)
	d.Set("kubeconfig_b64", resp.Kubeconfig)

	if decoded, decErr := base64.StdEncoding.DecodeString(resp.Kubeconfig); decErr == nil {
		d.Set("content", string(decoded))
	} else {
		// API may already return plain YAML; fall back to the raw value.
		d.Set("content", resp.Kubeconfig)
	}

	return nil
}
