package subnet

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// networkingSubnetStateRefreshFunc returns a standard retry.StateRefreshFunc to wait for subnet status.
func networkingSubnetStateRefreshFunc(ctx context.Context, subnetClient *client.Client, subnetID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		subnetResp := &dto.GetSubnetResponse{}
		_, err := subnetClient.Get(ctx, client.ApiPath.SubnetWithId(subnetID), subnetResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return subnetResp.Subnet, "DELETED", nil
			}

			return nil, "", err
		}

		return subnetResp.Subnet, "ACTIVE", nil
	}
}

// networkingSubnetStateRefreshFuncDelete returns a special case retry.StateRefreshFunc to try to delete a subnet.
func networkingSubnetStateRefreshFuncDelete(ctx context.Context, subnetClient *client.Client, subnetID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Attempting to delete vnpaycloud_networking_subnet %s", subnetID)

		subnetResp := &dto.GetSubnetResponse{}
		_, err := subnetClient.Get(ctx, client.ApiPath.SubnetWithId(subnetID), subnetResp, nil)
		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				log.Printf("[DEBUG] Successfully deleted vnpaycloud_networking_subnet %s", subnetID)
				return subnetResp.Subnet, "DELETED", nil
			}

			return subnetResp.Subnet, "ACTIVE", err
		}

		_, err = subnetClient.Delete(ctx, client.ApiPath.SubnetWithId(subnetID), nil)
		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				log.Printf("[DEBUG] Successfully deleted vnpaycloud_networking_subnet %s", subnetID)
				return subnetResp.Subnet, "DELETED", nil
			}
			// Subnet is still in use - we can retry.
			if util.ResponseCodeIs(err, http.StatusConflict) {
				return subnetResp.Subnet, "ACTIVE", nil
			}

			return subnetResp.Subnet, "ACTIVE", err
		}

		log.Printf("[DEBUG] vnpaycloud_networking_subnet %s is still active", subnetID)

		return subnetResp.Subnet, "ACTIVE", nil
	}
}

// expandNetworkingSubnetAllocationPools returns a slice of subnets.AllocationPool structs.
func expandNetworkingSubnetAllocationPools(allocationPools []interface{}) []dto.AllocationPool {
	result := make([]dto.AllocationPool, len(allocationPools))
	for i, raw := range allocationPools {
		rawMap := raw.(map[string]interface{})

		result[i] = dto.AllocationPool{
			Start: rawMap["start"].(string),
			End:   rawMap["end"].(string),
		}
	}

	return result
}

// flattenNetworkingSubnetAllocationPools allows to flatten slice of subnets.AllocationPool structs into
// a slice of maps.
func flattenNetworkingSubnetAllocationPools(allocationPools []dto.AllocationPool) []map[string]interface{} {
	result := make([]map[string]interface{}, len(allocationPools))
	for i, allocationPool := range allocationPools {
		pool := make(map[string]interface{})
		pool["start"] = allocationPool.Start
		pool["end"] = allocationPool.End

		result[i] = pool
	}

	return result
}

// flattenNetworkingSubnetHostRoutes allows to flatten slice of subnets.HostRoute structs into
// a slice of maps.
func flattenNetworkingSubnetHostRoutes(hostRoutes []dto.HostRoute) []map[string]interface{} {
	result := make([]map[string]interface{}, len(hostRoutes))
	for i, hostRoute := range hostRoutes {
		route := make(map[string]interface{})
		route["destination_cidr"] = hostRoute.DestinationCIDR
		route["next_hop"] = hostRoute.NextHop

		result[i] = route
	}

	return result
}

func networkingSubnetAllocationPoolsMatch(oldPools, newPools []interface{}) bool {
	if len(oldPools) != len(newPools) {
		return false
	}

	for _, newPool := range newPools {
		var found bool

		newPoolPool := newPool.(map[string]interface{})
		newStart := newPoolPool["start"].(string)
		newEnd := newPoolPool["end"].(string)

		for _, oldPool := range oldPools {
			oldPoolPool := oldPool.(map[string]interface{})
			oldStart := oldPoolPool["start"].(string)
			oldEnd := oldPoolPool["end"].(string)

			if oldStart == newStart && oldEnd == newEnd {
				found = true
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func networkingSubnetDNSNameserverAreUnique(raw []interface{}) error {
	set := make(map[string]struct{})
	for _, rawNS := range raw {
		nameserver, ok := rawNS.(string)
		if ok {
			if _, exists := set[nameserver]; exists {
				return fmt.Errorf("got duplicate nameserver %s", nameserver)
			}
			set[nameserver] = struct{}{}
		}
	}

	return nil
}
