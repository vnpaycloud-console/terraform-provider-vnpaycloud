package serverGroup

import (
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
)

const (
	antiAffinityPolicy = "anti-affinity"
	affinityPolicy     = "affinity"
)

func expandComputeServerGroupPolicies(client *client.Client, raw []interface{}) []string {
	policies := make([]string, len(raw))
	for i, v := range raw {

		policy := v.(string)
		policies[i] = policy

		// Set microversion for legacy policies to empty to not change behaviour for those policies
		if policy == antiAffinityPolicy || policy == affinityPolicy {
			// client.Microversion = ""
		}
	}

	return policies
}

func expandComputeServerGroupRulesMaxServerPerHost(raw []interface{}) int {
	for _, raw := range raw {
		raw, ok := raw.(map[string]interface{})
		if !ok {
			return 0
		}
		v, ok := raw["max_server_per_host"].(int)
		if !ok {
			return 0
		}
		//nolint:staticcheck // we need the first element
		return v
	}
	return 0
}
