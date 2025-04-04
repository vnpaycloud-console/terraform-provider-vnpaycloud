package secretv1

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vnpaycloud-console/gophercloud/v2"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/keymanager/v1/acls"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/keymanager/v1/secrets"
)

func keyManagerSecretV1WaitForSecretDeletion(ctx context.Context, kmClient *gophercloud.ServiceClient, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		err := secrets.Delete(ctx, kmClient, id).Err
		if err == nil {
			return "", "DELETED", nil
		}

		if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
			return "", "DELETED", nil
		}

		return nil, "ACTIVE", err
	}
}

func keyManagerSecretV1SecretType(v string) secrets.SecretType {
	var stype secrets.SecretType
	switch v {
	case "symmetric":
		stype = secrets.SymmetricSecret
	case "public":
		stype = secrets.PublicSecret
	case "private":
		stype = secrets.PrivateSecret
	case "passphrase":
		stype = secrets.PassphraseSecret
	case "certificate":
		stype = secrets.CertificateSecret
	case "opaque":
		stype = secrets.OpaqueSecret
	}

	return stype
}

func keyManagerSecretV1WaitForSecretCreation(ctx context.Context, kmClient *gophercloud.ServiceClient, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		secret, err := secrets.Get(ctx, kmClient, id).Extract()
		if err != nil {
			if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
				return "", "NOT_CREATED", nil
			}

			return "", "NOT_CREATED", err
		}

		if secret.Status == "ERROR" {
			return "", secret.Status, fmt.Errorf("Error creating secret")
		}

		return secret, secret.Status, nil
	}
}

func keyManagerSecretV1GetUUIDfromSecretRef(ref string) string {
	// secret ref has form https://{barbican_host}/v1/secrets/{secret_uuid}
	// so we are only interested in the last part
	refSplit := strings.Split(ref, "/")
	uuid := refSplit[len(refSplit)-1]
	return uuid
}

func flattenKeyManagerSecretV1Metadata(d *schema.ResourceData) map[string]string {
	m := make(map[string]string)
	for key, val := range d.Get("metadata").(map[string]interface{}) {
		m[key] = val.(string)
	}
	return m
}

func keyManagerSecretMetadataV1WaitForSecretMetadataCreation(ctx context.Context, kmClient *gophercloud.ServiceClient, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		metadata, err := secrets.GetMetadata(ctx, kmClient, id).Extract()
		if err != nil {
			if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
				return "", "NOT_CREATED", nil
			}

			return "", "NOT_CREATED", err
		}
		return metadata, "ACTIVE", nil
	}
}

func keyManagerSecretV1GetPayload(ctx context.Context, kmClient *gophercloud.ServiceClient, id, contentType string) string {
	opts := secrets.GetPayloadOpts{
		PayloadContentType: contentType,
	}
	payload, err := secrets.GetPayload(ctx, kmClient, id, opts).Extract()
	if err != nil {
		log.Printf("[DEBUG] Could not retrieve payload for secret with id %s: %s", id, err)
	}

	if !strings.HasPrefix(contentType, "text/") {
		return base64.StdEncoding.EncodeToString(payload)
	}

	return string(payload)
}

// So far only "read" is supported.
func getSupportedACLOperations() [1]string {
	return [1]string{"read"}
}

func getACLSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList, // the list, returned by Barbican, is always ordered
		Optional: true,
		Computed: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"project_access": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true, // defaults to true in OpenStack Barbican code
				},
				"users": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"created_at": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"updated_at": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
}

func expandKeyManagerV1ACL(v interface{}, aclType string) acls.SetOpt {
	users := []string{}
	iTrue := true // set default value to true
	res := acls.SetOpt{
		ProjectAccess: &iTrue,
		Users:         &users,
		Type:          aclType,
	}

	if v, ok := v.([]interface{}); ok {
		for _, v := range v {
			if v, ok := v.(map[string]interface{}); ok {
				if v, ok := v["project_access"]; ok {
					if v, ok := v.(bool); ok {
						res.ProjectAccess = &v
					}
				}
				if v, ok := v["users"]; ok {
					if v, ok := v.(*schema.Set); ok {
						for _, v := range v.List() {
							*res.Users = append(*res.Users, v.(string))
						}
					}
				}
			}
		}
	}
	return res
}

func expandKeyManagerV1ACLs(v interface{}) acls.SetOpts {
	var res []acls.SetOpt

	if v, ok := v.([]interface{}); ok {
		for _, v := range v {
			if v, ok := v.(map[string]interface{}); ok {
				for aclType, v := range v {
					acl := expandKeyManagerV1ACL(v, aclType)
					res = append(res, acl)
				}
			}
		}
	}

	return res
}

func flattenKeyManagerV1ACLs(acl *acls.ACL) []map[string][]map[string]interface{} {
	var m []map[string][]map[string]interface{}

	if acl != nil {
		allAcls := *acl
		for _, aclOp := range getSupportedACLOperations() {
			if v, ok := allAcls[aclOp]; ok {
				if m == nil {
					m = make([]map[string][]map[string]interface{}, 1)
					m[0] = make(map[string][]map[string]interface{})
				}
				if m[0][aclOp] == nil {
					m[0][aclOp] = make([]map[string]interface{}, 1)
				}
				m[0][aclOp][0] = map[string]interface{}{
					"project_access": v.ProjectAccess,
					"users":          v.Users,
					"created_at":     v.Created.UTC().Format(time.RFC3339),
					"updated_at":     v.Updated.UTC().Format(time.RFC3339),
				}
			}
		}
	}

	return m
}
