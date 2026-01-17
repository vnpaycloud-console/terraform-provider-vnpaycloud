package securityGroup

import (
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"context"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func NetworkingReadAttributesTags(d *schema.ResourceData, tags []string) {
	util.ExpandObjectReadTags(d, tags)
}

func NetworkingUpdateAttributesTags(d *schema.ResourceData) []string {
	return util.ExpandObjectUpdateTags(d)
}

func NetworkingAttributesTags(d *schema.ResourceData) []string {
	return util.ExpandObjectTags(d)
}

// networkingSecgroupStateRefreshFuncDelete returns a special case retry.StateRefreshFunc to try to delete a secgroup.
func networkingSecgroupStateRefreshFuncDelete(ctx context.Context, networkingClient *client.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Attempting to delete vnpaycloud_networking_secgroup %s", id)

		secGroupResp := &dto.GetSecurityGroupResponse{}
		_, err := networkingClient.Get(ctx, client.ApiPath.SecurityGroupWithId(id), secGroupResp, nil)
		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				log.Printf("[DEBUG] Successfully deleted vnpaycloud_networking_secgroup %s", id)
				return secGroupResp.SecurityGroup, "DELETED", nil
			}

			return secGroupResp.SecurityGroup, "ACTIVE", err
		}

		_, err = networkingClient.Delete(ctx, client.ApiPath.SecurityGroupWithId(id), nil)
		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				log.Printf("[DEBUG] Successfully deleted vnpaycloud_networking_secgroup %s", id)
				return secGroupResp.SecurityGroup, "DELETED", nil
			}
			if util.ResponseCodeIs(err, http.StatusConflict) {
				return secGroupResp.SecurityGroup, "ACTIVE", nil
			}

			return secGroupResp.SecurityGroup, "ACTIVE", err
		}

		log.Printf("[DEBUG] vnpaycloud_networking_secgroup %s is still active", id)

		return secGroupResp.SecurityGroup, "ACTIVE", nil
	}
}
