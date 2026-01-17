package floatingip

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// networkingFloatingIPID retrieves floating IP ID by the provided IP address.
func networkingFloatingIPID(ctx context.Context, c *client.Client, floatingIP string) (string, error) {
	listOpts := dto.ListFloatingIPOpts{
		FloatingIP: floatingIP,
	}

	listResp := dto.ListFloatingIPResponse{}

	_, err := c.All(ctx, client.ApiPath.FloatingIPWithParams(listOpts), &listResp, nil)
	if err != nil {
		return "", err
	}

	allFloatingIPs := listResp.FloatingIPs

	if len(allFloatingIPs) == 0 {
		return "", fmt.Errorf("there are no vnpaycloud_networking_floatingip with %s IP", floatingIP)
	}
	if len(allFloatingIPs) > 1 {
		return "", fmt.Errorf("there are more than one vnpaycloud_networking_floatingip with %s IP", floatingIP)
	}

	return allFloatingIPs[0].ID, nil
}

func networkingFloatingIPStateRefreshFunc(ctx context.Context, networkingClient *client.Client, fipID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		fipResp := &dto.GetFloatingIPResponse{}
		_, err := networkingClient.Get(ctx, client.ApiPath.FloatingIPWithId(fipID), fipResp, nil)
		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return fipResp.FloatingIP, "DELETED", nil
			}

			return nil, "", err
		}

		return fipResp.FloatingIP, fipResp.FloatingIP.Status, nil
	}
}
