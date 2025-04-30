package applicationcredentials

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/vnpaycloud-console/gophercloud/v2"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/identity/v3/applicationcredentials"
)

type IdentityApplicationCredentialV3CreateOpts struct {
	applicationcredentials.CreateOpts
	ValueSpecs map[string]string `json:"value_specs,omitempty"`
}

func (opts IdentityApplicationCredentialV3CreateOpts) ToApplicationCredentialCreateMap() (map[string]interface{}, error) {
	return util.BuildRequest(opts, "application_credential")
}

func flattenIdentityApplicationCredentialRolesV3(roles []applicationcredentials.Role) []string {
	res := make([]string, 0, len(roles))
	for _, role := range roles {
		res = append(res, role.Name)
	}
	return res
}

func expandIdentityApplicationCredentialRolesV3(roles []interface{}) []applicationcredentials.Role {
	res := make([]applicationcredentials.Role, 0, len(roles))
	for _, role := range roles {
		res = append(res, applicationcredentials.Role{Name: role.(string)})
	}
	return res
}

func flattenIdentityApplicationCredentialAccessRulesV3(rules []applicationcredentials.AccessRule) []map[string]string {
	res := make([]map[string]string, 0, len(rules))
	for _, role := range rules {
		res = append(res, map[string]string{
			"id":      role.ID,
			"path":    role.Path,
			"method":  role.Method,
			"service": role.Service,
		})
	}
	return res
}

func expandIdentityApplicationCredentialAccessRulesV3(rules []interface{}) []applicationcredentials.AccessRule {
	res := make([]applicationcredentials.AccessRule, 0, len(rules))
	for _, v := range rules {
		rule := v.(map[string]interface{})
		res = append(res,
			applicationcredentials.AccessRule{
				ID:      rule["id"].(string),
				Path:    rule["path"].(string),
				Method:  rule["method"].(string),
				Service: rule["service"].(string),
			},
		)
	}
	return res
}

func applicationCredentialCleanupAccessRulesV3(ctx context.Context, client *gophercloud.ServiceClient, userID string, id string, rules []applicationcredentials.AccessRule) error {
	for _, rule := range rules {
		log.Printf("[DEBUG] Cleaning up %q access rule from the %q application credential", rule.ID, id)
		err := applicationcredentials.DeleteAccessRule(ctx, client, userID, rule.ID).ExtractErr()
		if err != nil {
			if gophercloud.ResponseCodeIs(err, http.StatusForbidden) {
				log.Printf("[DEBUG] Error delete %q access rule from the %q application credential: %s", rule.ID, id, err)
				continue
			}
			if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
				// access rule was already deleted
				continue
			}
			return fmt.Errorf("failed to delete %q access rule from the %q application credential: %s", rule.ID, id, err)
		}
	}
	return nil
}
