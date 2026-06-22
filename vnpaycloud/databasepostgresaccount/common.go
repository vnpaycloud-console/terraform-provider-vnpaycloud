package databasepostgresaccount

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func postgresAccountStateRefreshFunc(ctx context.Context, c *client.Client, projectID, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp := &dto.PostgresAccountResponse{}
		httpResp, err := c.Get(ctx, client.ApiPath.DatabasePostgresAccountWithID(projectID, id), resp, nil)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				return resp, "deleted", nil
			}
			return nil, "", err
		}

		status := resp.PostgresAccount.Status
		if status == "" {
			status = "unknown"
		}
		if status == "error" {
			return resp, status, fmt.Errorf("vnpaycloud_database_postgres_account %s is in error state", id)
		}
		return resp, status, nil
	}
}

// grantPostgresAccountPrivilege grants a single privilege block to the account.
func grantPostgresAccountPrivilege(ctx context.Context, cfg *config.Config, id string, m map[string]interface{}) error {
	grantOpts := dto.GrantPrivilegesPostgresAccountRequest{
		PrivilegeTemplate: m["privilege"].(string),
		DbName:            m["db_name"].(string),
		DbSchema:          m["db_schema"].(string),
	}
	if _, err := cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresAccountGrantPrivileges(cfg.ProjectID, id), grantOpts, nil, nil); err != nil {
		return fmt.Errorf("error granting privileges (db_name=%s, db_schema=%s) on vnpaycloud_database_postgres_account %s: %w", grantOpts.DbName, grantOpts.DbSchema, id, err)
	}
	return nil
}

// revokePostgresAccountPrivilege revokes a single privilege block from the account.
func revokePostgresAccountPrivilege(ctx context.Context, cfg *config.Config, id string, m map[string]interface{}) error {
	revokeOpts := dto.RevokePrivilegesPostgresAccountRequest{
		DbName:   m["db_name"].(string),
		DbSchema: m["db_schema"].(string),
	}
	if _, err := cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresAccountRevokePrivileges(cfg.ProjectID, id), revokeOpts, nil, nil); err != nil {
		return fmt.Errorf("error revoking privileges (db_name=%s, db_schema=%s) on vnpaycloud_database_postgres_account %s: %w", revokeOpts.DbName, revokeOpts.DbSchema, id, err)
	}
	return nil
}

// flattenPostgresAccountGrants converts DTO grants to Terraform state.
func flattenPostgresAccountGrants(grants []dto.PostgresAccountGrant) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(grants))
	for _, g := range grants {
		result = append(result, map[string]interface{}{
			"db_name":   g.DbName,
			"db_schema": g.DbSchema,
			"privilege": g.PrivilegeTemplate,
		})
	}
	return result
}
