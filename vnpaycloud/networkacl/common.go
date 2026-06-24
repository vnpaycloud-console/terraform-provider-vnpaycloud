package networkacl

import (
	"context"
	"net/http"
	"sort"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func networkACLStateRefreshFunc(ctx context.Context, c *client.Client, projectID, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp := &dto.NetworkACLResponse{}
		_, err := c.Get(ctx, client.ApiPath.NetworkACLWithID(projectID, id), resp, nil)
		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return &dto.NetworkACL{}, "deleted", nil
			}
			return nil, "", err
		}

		if resp.NetworkACL.ID == "" {
			return &dto.NetworkACL{}, "deleted", nil
		}

		return &resp.NetworkACL, resp.NetworkACL.Status, nil
	}
}

func stringSetValues(set *schema.Set) []string {
	if set == nil {
		return nil
	}

	values := make([]string, 0, set.Len())
	for _, raw := range set.List() {
		if value, ok := raw.(string); ok && value != "" {
			values = append(values, value)
		}
	}
	sort.Strings(values)
	return values
}

func setNetworkACLAttributes(d *schema.ResourceData, acl dto.NetworkACL) {
	d.SetId(acl.ID)
	_ = d.Set("name", acl.Name)
	_ = d.Set("description", acl.Description)
	_ = d.Set("vpc_id", acl.VpcID)
	_ = d.Set("subnet_ids", acl.SubnetIDs)
	_ = d.Set("total_rules", acl.TotalRules)
	_ = d.Set("status", acl.Status)
	_ = d.Set("created_at", acl.CreatedAt)
}
