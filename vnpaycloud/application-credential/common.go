package applicationcredentials

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
)

func flattenIdentityApplicationCredentialRolesV3(roles []dto.Role) []string {
	res := make([]string, 0, len(roles))
	for _, role := range roles {
		res = append(res, role.Name)
	}
	return res
}

func expandIdentityApplicationCredentialRolesV3(roles []interface{}) []dto.Role {
	res := make([]dto.Role, 0, len(roles))
	for _, role := range roles {
		res = append(res, dto.Role{Name: role.(string)})
	}
	return res
}

func flattenIdentityApplicationCredentialAccessRulesV3(rules []dto.AccessRule) []map[string]string {
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

func expandIdentityApplicationCredentialAccessRulesV3(rules []interface{}) []dto.AccessRule {
	res := make([]dto.AccessRule, 0, len(rules))
	for _, v := range rules {
		rule := v.(map[string]interface{})
		res = append(res,
			dto.AccessRule{
				ID:      rule["id"].(string),
				Path:    rule["path"].(string),
				Method:  rule["method"].(string),
				Service: rule["service"].(string),
			},
		)
	}
	return res
}

func applicationCredentialCleanupAccessRulesV3(ctx context.Context, appClient *client.Client, userID string, id string, rules []dto.AccessRule) error {
	for _, rule := range rules {
		log.Printf("[DEBUG] Cleaning up %q access rule from the %q application credential", rule.ID, id)
		_, err := appClient.Delete(ctx, client.ApiPath.ApplicationCredentialAccessRuleWithId(userID, rule.ID), nil)
		if err != nil {
			if util.ResponseCodeIs(err, http.StatusForbidden) {
				log.Printf("[DEBUG] Error delete %q access rule from the %q application credential: %s", rule.ID, id, err)
				continue
			}
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				// access rule was already deleted
				continue
			}
			return fmt.Errorf("failed to delete %q access rule from the %q application credential: %s", rule.ID, id, err)
		}
	}
	return nil
}
