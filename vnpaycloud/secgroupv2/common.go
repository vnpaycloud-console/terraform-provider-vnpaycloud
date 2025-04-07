package secgroupv2

import (
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"context"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/vnpaycloud-console/gophercloud/v2"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/extensions/security/groups"
)

func networkingV2ReadAttributesTags(d *schema.ResourceData, tags []string) {
	util.ExpandObjectReadTags(d, tags)
}

func networkingV2UpdateAttributesTags(d *schema.ResourceData) []string {
	return util.ExpandObjectUpdateTags(d)
}

func networkingV2AttributesTags(d *schema.ResourceData) []string {
	return util.ExpandObjectTags(d)
}

// networkingSecgroupV2StateRefreshFuncDelete returns a special case retry.StateRefreshFunc to try to delete a secgroup.
func networkingSecgroupV2StateRefreshFuncDelete(ctx context.Context, networkingClient *gophercloud.ServiceClient, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Attempting to delete openstack_networking_secgroup_v2 %s", id)

		r, err := groups.Get(ctx, networkingClient, id).Extract()
		if err != nil {
			if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
				log.Printf("[DEBUG] Successfully deleted openstack_networking_secgroup_v2 %s", id)
				return r, "DELETED", nil
			}

			return r, "ACTIVE", err
		}

		err = groups.Delete(ctx, networkingClient, id).ExtractErr()
		if err != nil {
			if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
				log.Printf("[DEBUG] Successfully deleted openstack_networking_secgroup_v2 %s", id)
				return r, "DELETED", nil
			}
			if gophercloud.ResponseCodeIs(err, http.StatusConflict) {
				return r, "ACTIVE", nil
			}

			return r, "ACTIVE", err
		}

		log.Printf("[DEBUG] openstack_networking_secgroup_v2 %s is still active", id)

		return r, "ACTIVE", nil
	}
}
