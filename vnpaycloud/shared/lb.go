package shared

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/vnpaycloud-console/gophercloud/v2"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/loadbalancer/v2/l7policies"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/loadbalancer/v2/listeners"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/loadbalancer/v2/loadbalancers"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/loadbalancer/v2/monitors"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/loadbalancer/v2/pools"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/ports"
)

const (
	lbPendingCreate = "PENDING_CREATE"
	lbPendingUpdate = "PENDING_UPDATE"
	lbPendingDelete = "PENDING_DELETE"
	lbActive        = "ACTIVE"
	lbError         = "ERROR"
)

// lbPendingStatuses are the valid statuses a LoadBalancer will be in while
// it's updating.
func GetLbPendingStatuses() []string {
	return []string{lbPendingCreate, lbPendingUpdate}
}

// lbPendingDeleteStatuses are the valid statuses a LoadBalancer will be before delete.
func GetLbPendingDeleteStatuses() []string {
	return []string{lbError, lbPendingUpdate, lbPendingDelete, lbActive}
}

func getLbSkipStatuses() []string {
	return []string{lbError, lbActive}
}

func FlattenLBMembersV2(members []pools.Member) []map[string]interface{} {
	m := make([]map[string]interface{}, len(members))

	for i, member := range members {
		m[i] = map[string]interface{}{
			"name":            member.Name,
			"weight":          member.Weight,
			"admin_state_up":  member.AdminStateUp,
			"subnet_id":       member.SubnetID,
			"address":         member.Address,
			"protocol_port":   member.ProtocolPort,
			"monitor_port":    member.MonitorPort,
			"monitor_address": member.MonitorAddress,
			"id":              member.ID,
			"backup":          member.Backup,
		}
	}

	return m
}

func ExpandLBMembersV2(members *schema.Set, lbClient *gophercloud.ServiceClient) []pools.BatchUpdateMemberOpts {
	var m []pools.BatchUpdateMemberOpts

	if members != nil {
		for _, raw := range members.List() {
			rawMap := raw.(map[string]interface{})
			name := rawMap["name"].(string)
			subnetID := rawMap["subnet_id"].(string)
			weight := rawMap["weight"].(int)
			adminStateUp := rawMap["admin_state_up"].(bool)

			member := pools.BatchUpdateMemberOpts{
				Address:      rawMap["address"].(string),
				ProtocolPort: rawMap["protocol_port"].(int),
				Name:         &name,
				SubnetID:     &subnetID,
				Weight:       &weight,
				AdminStateUp: &adminStateUp,
			}

			// backup requires octavia minor version 2.1. Only set when specified
			if val, ok := rawMap["backup"]; ok {
				backup := val.(bool)
				member.Backup = &backup
			}

			// Only set monitor_port and monitor_address when explicitly specified, as they are optional arguments
			if val, ok := rawMap["monitor_port"]; ok {
				monitorPort := val.(int)
				if monitorPort > 0 {
					member.MonitorPort = &monitorPort
				}
			}

			if val, ok := rawMap["monitor_address"]; ok {
				monitorAddress := val.(string)
				if monitorAddress != "" {
					member.MonitorAddress = &monitorAddress
				}
			}

			m = append(m, member)
		}
	}

	return m
}

func getListenerIDForL7Policy(ctx context.Context, lbClient *gophercloud.ServiceClient, id string) (string, error) {
	log.Printf("[DEBUG] Trying to get Listener ID associated with the %s L7 Policy ID", id)
	lbsPages, err := loadbalancers.List(lbClient, loadbalancers.ListOpts{}).AllPages(ctx)
	if err != nil {
		return "", fmt.Errorf("No Load Balancers were found: %s", err)
	}

	lbs, err := loadbalancers.ExtractLoadBalancers(lbsPages)
	if err != nil {
		return "", fmt.Errorf("Unable to extract Load Balancers list: %s", err)
	}

	for _, lb := range lbs {
		statuses, err := loadbalancers.GetStatuses(ctx, lbClient, lb.ID).Extract()
		if err != nil {
			return "", fmt.Errorf("Failed to get Load Balancer statuses: %s", err)
		}
		for _, listener := range statuses.Loadbalancer.Listeners {
			for _, l7policy := range listener.L7Policies {
				if l7policy.ID == id {
					return listener.ID, nil
				}
			}
		}
	}

	return "", fmt.Errorf("Unable to find Listener ID associated with the %s L7 Policy ID", id)
}

func waitForLBV2L7Rule(ctx context.Context, lbClient *gophercloud.ServiceClient, parentListener *listeners.Listener, parentL7policy *l7policies.L7Policy, l7rule *l7policies.Rule, target string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for l7rule %s to become %s.", l7rule.ID, target)

	if len(parentListener.Loadbalancers) == 0 {
		return fmt.Errorf("Unable to determine loadbalancer ID from listener %s", parentListener.ID)
	}

	lbID := parentListener.Loadbalancers[0].ID

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    resourceLBV2L7RuleRefreshFunc(ctx, lbClient, lbID, parentL7policy.ID, l7rule),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
			if target == "DELETED" {
				return nil
			}
		}

		return fmt.Errorf("Error waiting for l7rule %s to become %s: %s", l7rule.ID, target, err)
	}

	return nil
}

func resourceLBV2L7RuleRefreshFunc(ctx context.Context, lbClient *gophercloud.ServiceClient, lbID string, l7policyID string, l7rule *l7policies.Rule) retry.StateRefreshFunc {
	if l7rule.ProvisioningStatus == "" {
		return resourceLBV2LoadBalancerStatusRefreshFunc(ctx, lbClient, lbID, "l7rule", l7rule.ID, l7policyID)
	}

	return func() (interface{}, string, error) {
		lb, status, err := resourceLBV2LoadBalancerRefreshFunc(ctx, lbClient, lbID)()
		if err != nil {
			return lb, status, err
		}
		if !util.StrSliceContains(getLbSkipStatuses(), status) {
			return lb, status, nil
		}

		l7rule, err := l7policies.GetRule(ctx, lbClient, l7policyID, l7rule.ID).Extract()
		if err != nil {
			return nil, "", err
		}

		return l7rule, l7rule.ProvisioningStatus, nil
	}
}

func resourceLBV2ListenerRefreshFunc(ctx context.Context, lbClient *gophercloud.ServiceClient, lbID string, listener *listeners.Listener) retry.StateRefreshFunc {
	if listener.ProvisioningStatus == "" {
		return resourceLBV2LoadBalancerStatusRefreshFunc(ctx, lbClient, lbID, "listener", listener.ID, "")
	}

	return func() (interface{}, string, error) {
		lb, status, err := resourceLBV2LoadBalancerRefreshFunc(ctx, lbClient, lbID)()
		if err != nil {
			return lb, status, err
		}
		if !util.StrSliceContains(getLbSkipStatuses(), status) {
			return lb, status, nil
		}

		listener, err := listeners.Get(ctx, lbClient, listener.ID).Extract()
		if err != nil {
			return nil, "", err
		}

		return listener, listener.ProvisioningStatus, nil
	}
}

func WaitForLBV2Listener(ctx context.Context, lbClient *gophercloud.ServiceClient, listener *listeners.Listener, target string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for vnpaycloud_lb_listener %s to become %s.", listener.ID, target)

	if len(listener.Loadbalancers) == 0 {
		return fmt.Errorf("Failed to detect a vnpaycloud_lb_loadbalancer for the %s vnpaycloud_lb_listener", listener.ID)
	}

	lbID := listener.Loadbalancers[0].ID

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    resourceLBV2ListenerRefreshFunc(ctx, lbClient, lbID, listener),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
			if target == "DELETED" {
				return nil
			}
		}

		return fmt.Errorf("Error waiting for vnpaycloud_lb_listener %s to become %s: %s", listener.ID, target, err)
	}

	return nil
}

func lbV2FindLBIDviaPool(ctx context.Context, lbClient *gophercloud.ServiceClient, pool *pools.Pool) (string, error) {
	if len(pool.Loadbalancers) > 0 {
		return pool.Loadbalancers[0].ID, nil
	}

	if len(pool.Listeners) > 0 {
		listenerID := pool.Listeners[0].ID
		listener, err := listeners.Get(ctx, lbClient, listenerID).Extract()
		if err != nil {
			return "", err
		}

		if len(listener.Loadbalancers) > 0 {
			return listener.Loadbalancers[0].ID, nil
		}
	}

	return "", fmt.Errorf("Unable to determine loadbalancer ID from pool %s", pool.ID)
}

func resourceLBV2PoolRefreshFunc(ctx context.Context, lbClient *gophercloud.ServiceClient, lbID string, pool *pools.Pool) retry.StateRefreshFunc {
	if pool.ProvisioningStatus == "" {
		return resourceLBV2LoadBalancerStatusRefreshFunc(ctx, lbClient, lbID, "pool", pool.ID, "")
	}

	return func() (interface{}, string, error) {
		lb, status, err := resourceLBV2LoadBalancerRefreshFunc(ctx, lbClient, lbID)()
		if err != nil {
			return lb, status, err
		}
		if !util.StrSliceContains(getLbSkipStatuses(), status) {
			return lb, status, nil
		}

		pool, err := pools.Get(ctx, lbClient, pool.ID).Extract()
		if err != nil {
			return nil, "", err
		}

		return pool, pool.ProvisioningStatus, nil
	}
}

func WaitForLBV2Pool(ctx context.Context, lbClient *gophercloud.ServiceClient, pool *pools.Pool, target string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for pool %s to become %s.", pool.ID, target)

	lbID, err := lbV2FindLBIDviaPool(ctx, lbClient, pool)
	if err != nil {
		return err
	}

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    resourceLBV2PoolRefreshFunc(ctx, lbClient, lbID, pool),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
			if target == "DELETED" {
				return nil
			}
		}

		return fmt.Errorf("Error waiting for pool %s to become %s: %s", pool.ID, target, err)
	}

	return nil
}

func resourceLBV2LoadBalancerStatusRefreshFunc(ctx context.Context, lbClient *gophercloud.ServiceClient, lbID, resourceType, resourceID string, parentID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		statuses, err := loadbalancers.GetStatuses(ctx, lbClient, lbID).Extract()
		if err != nil {
			if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
				return nil, "", gophercloud.ErrUnexpectedResponseCode{
					Actual: http.StatusNotFound,
					BaseError: gophercloud.BaseError{
						DefaultErrString: fmt.Sprintf("Unable to get statuses from the Load Balancer %s statuses tree: %s", lbID, err),
					},
				}
			}
			return nil, "", fmt.Errorf("Unable to get statuses from the Load Balancer %s statuses tree: %s", lbID, err)
		}

		// Don't fail, when statuses returns "null"
		if statuses == nil || statuses.Loadbalancer == nil {
			statuses = new(loadbalancers.StatusTree)
			statuses.Loadbalancer = new(loadbalancers.LoadBalancer)
		} else if !util.StrSliceContains(getLbSkipStatuses(), statuses.Loadbalancer.ProvisioningStatus) {
			return statuses.Loadbalancer, statuses.Loadbalancer.ProvisioningStatus, nil
		}

		switch resourceType {
		case "listener":
			for _, listener := range statuses.Loadbalancer.Listeners {
				if listener.ID == resourceID {
					if listener.ProvisioningStatus != "" {
						return listener, listener.ProvisioningStatus, nil
					}
				}
			}
			listener, err := listeners.Get(ctx, lbClient, resourceID).Extract()
			return listener, "ACTIVE", err

		case "pool":
			for _, pool := range statuses.Loadbalancer.Pools {
				if pool.ID == resourceID {
					if pool.ProvisioningStatus != "" {
						return pool, pool.ProvisioningStatus, nil
					}
				}
			}
			pool, err := pools.Get(ctx, lbClient, resourceID).Extract()
			return pool, "ACTIVE", err

		case "monitor":
			for _, pool := range statuses.Loadbalancer.Pools {
				if pool.Monitor.ID == resourceID {
					if pool.Monitor.ProvisioningStatus != "" {
						return pool.Monitor, pool.Monitor.ProvisioningStatus, nil
					}
				}
			}
			monitor, err := monitors.Get(ctx, lbClient, resourceID).Extract()
			return monitor, "ACTIVE", err

		case "member":
			for _, pool := range statuses.Loadbalancer.Pools {
				for _, member := range pool.Members {
					if member.ID == resourceID {
						if member.ProvisioningStatus != "" {
							return member, member.ProvisioningStatus, nil
						}
					}
				}
			}
			member, err := pools.GetMember(ctx, lbClient, parentID, resourceID).Extract()
			return member, "ACTIVE", err

		case "l7policy":
			for _, listener := range statuses.Loadbalancer.Listeners {
				for _, l7policy := range listener.L7Policies {
					if l7policy.ID == resourceID {
						if l7policy.ProvisioningStatus != "" {
							return l7policy, l7policy.ProvisioningStatus, nil
						}
					}
				}
			}
			l7policy, err := l7policies.Get(ctx, lbClient, resourceID).Extract()
			return l7policy, "ACTIVE", err

		case "l7rule":
			for _, listener := range statuses.Loadbalancer.Listeners {
				for _, l7policy := range listener.L7Policies {
					for _, l7rule := range l7policy.Rules {
						if l7rule.ID == resourceID {
							if l7rule.ProvisioningStatus != "" {
								return l7rule, l7rule.ProvisioningStatus, nil
							}
						}
					}
				}
			}
			l7Rule, err := l7policies.GetRule(ctx, lbClient, parentID, resourceID).Extract()
			return l7Rule, "ACTIVE", err
		}

		return nil, "", fmt.Errorf("An unexpected error occurred querying the status of %s %s by loadbalancer %s", resourceType, resourceID, lbID)
	}
}

func waitForLBV2L7Policy(ctx context.Context, lbClient *gophercloud.ServiceClient, parentListener *listeners.Listener, l7policy *l7policies.L7Policy, target string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for l7policy %s to become %s.", l7policy.ID, target)

	if len(parentListener.Loadbalancers) == 0 {
		return fmt.Errorf("Unable to determine loadbalancer ID from listener %s", parentListener.ID)
	}

	lbID := parentListener.Loadbalancers[0].ID

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    resourceLBV2L7PolicyRefreshFunc(ctx, lbClient, lbID, l7policy),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
			if target == "DELETED" {
				return nil
			}
		}

		return fmt.Errorf("Error waiting for l7policy %s to become %s: %s", l7policy.ID, target, err)
	}

	return nil
}

func resourceLBV2L7PolicyRefreshFunc(ctx context.Context, lbClient *gophercloud.ServiceClient, lbID string, l7policy *l7policies.L7Policy) retry.StateRefreshFunc {
	if l7policy.ProvisioningStatus == "" {
		return resourceLBV2LoadBalancerStatusRefreshFunc(ctx, lbClient, lbID, "l7policy", l7policy.ID, "")
	}

	return func() (interface{}, string, error) {
		lb, status, err := resourceLBV2LoadBalancerRefreshFunc(ctx, lbClient, lbID)()
		if err != nil {
			return lb, status, err
		}
		if !util.StrSliceContains(getLbSkipStatuses(), status) {
			return lb, status, nil
		}

		l7policy, err := l7policies.Get(ctx, lbClient, l7policy.ID).Extract()
		if err != nil {
			return nil, "", err
		}

		return l7policy, l7policy.ProvisioningStatus, nil
	}
}

func WaitForLBV2Member(ctx context.Context, lbClient *gophercloud.ServiceClient, parentPool *pools.Pool, member *pools.Member, target string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for member %s to become %s.", member.ID, target)

	lbID, err := lbV2FindLBIDviaPool(ctx, lbClient, parentPool)
	if err != nil {
		return err
	}

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    resourceLBV2MemberRefreshFunc(ctx, lbClient, lbID, parentPool.ID, member),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
			if target == "DELETED" {
				return nil
			}
		}

		return fmt.Errorf("Error waiting for member %s to become %s: %s", member.ID, target, err)
	}

	return nil
}

func resourceLBV2MemberRefreshFunc(ctx context.Context, lbClient *gophercloud.ServiceClient, lbID string, poolID string, member *pools.Member) retry.StateRefreshFunc {
	if member.ProvisioningStatus == "" {
		return resourceLBV2LoadBalancerStatusRefreshFunc(ctx, lbClient, lbID, "member", member.ID, poolID)
	}

	return func() (interface{}, string, error) {
		lb, status, err := resourceLBV2LoadBalancerRefreshFunc(ctx, lbClient, lbID)()
		if err != nil {
			return lb, status, err
		}
		if !util.StrSliceContains(getLbSkipStatuses(), status) {
			return lb, status, nil
		}

		member, err := pools.GetMember(ctx, lbClient, poolID, member.ID).Extract()
		if err != nil {
			return nil, "", err
		}

		return member, member.ProvisioningStatus, nil
	}
}

func WaitForLBV2Monitor(ctx context.Context, lbClient *gophercloud.ServiceClient, parentPool *pools.Pool, monitor *monitors.Monitor, target string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for vnpaycloud_lb_monitor %s to become %s.", monitor.ID, target)

	lbID, err := lbV2FindLBIDviaPool(ctx, lbClient, parentPool)
	if err != nil {
		return err
	}

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    resourceLBV2MonitorRefreshFunc(ctx, lbClient, lbID, monitor),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
			if target == "DELETED" {
				return nil
			}
		}
		return fmt.Errorf("Error waiting for vnpaycloud_lb_monitor %s to become %s: %s", monitor.ID, target, err)
	}

	return nil
}

func resourceLBV2MonitorRefreshFunc(ctx context.Context, lbClient *gophercloud.ServiceClient, lbID string, monitor *monitors.Monitor) retry.StateRefreshFunc {
	if monitor.ProvisioningStatus == "" {
		return resourceLBV2LoadBalancerStatusRefreshFunc(ctx, lbClient, lbID, "monitor", monitor.ID, "")
	}

	return func() (interface{}, string, error) {
		lb, status, err := resourceLBV2LoadBalancerRefreshFunc(ctx, lbClient, lbID)()
		if err != nil {
			return lb, status, err
		}
		if !util.StrSliceContains(getLbSkipStatuses(), status) {
			return lb, status, nil
		}

		monitor, err := monitors.Get(ctx, lbClient, monitor.ID).Extract()
		if err != nil {
			return nil, "", err
		}

		return monitor, monitor.ProvisioningStatus, nil
	}
}

func ExpandLBPoolTLSVersionV2(v []interface{}) []pools.TLSVersion {
	versions := make([]pools.TLSVersion, len(v))
	for i, v := range v {
		versions[i] = pools.TLSVersion(v.(string))
	}
	return versions
}

func ExpandLBListenerTLSVersionV2(v []interface{}) []listeners.TLSVersion {
	versions := make([]listeners.TLSVersion, len(v))
	for i, v := range v {
		versions[i] = listeners.TLSVersion(v.(string))
	}
	return versions
}

func FlattenLBPoolPersistenceV2(p pools.SessionPersistence) []map[string]interface{} {
	if p == (pools.SessionPersistence{}) {
		return nil
	}
	return []map[string]interface{}{
		{
			"type":        p.Type,
			"cookie_name": p.CookieName,
		},
	}
}

func ExpandLBPoolPersistanceV2(p []interface{}) (*pools.SessionPersistence, error) {
	persistence := &pools.SessionPersistence{}

	for _, v := range p {
		v := v.(map[string]interface{})
		persistence.Type = v["type"].(string)

		if persistence.Type == "APP_COOKIE" {
			if v["cookie_name"].(string) == "" {
				return nil, fmt.Errorf("Persistence cookie_name needs to be set if using 'APP_COOKIE' persistence type")
			}
			persistence.CookieName = v["cookie_name"].(string)

			return persistence, nil
		}

		if v["cookie_name"].(string) != "" {
			return nil, fmt.Errorf("Persistence cookie_name can only be set if using 'APP_COOKIE' persistence type")
		}

		//nolint:staticcheck // we need the first element
		return persistence, nil
	}

	return persistence, nil
}

func WaitForLBV2LoadBalancer(ctx context.Context, lbClient *gophercloud.ServiceClient, lbID string, target string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for loadbalancer %s to become %s.", lbID, target)

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    resourceLBV2LoadBalancerRefreshFunc(ctx, lbClient, lbID),
		Timeout:    timeout,
		Delay:      0,
		MinTimeout: 1 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
			switch target {
			case "DELETED":
				return nil
			default:
				return fmt.Errorf("Error: loadbalancer %s not found: %s", lbID, err)
			}
		}
		return fmt.Errorf("Error waiting for loadbalancer %s to become %s: %s", lbID, target, err)
	}

	return nil
}

func resourceLBV2LoadBalancerRefreshFunc(ctx context.Context, lbClient *gophercloud.ServiceClient, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		lb, err := loadbalancers.Get(ctx, lbClient, id).Extract()
		if err != nil {
			return nil, "", err
		}

		return lb, lb.ProvisioningStatus, nil
	}
}

func ResourceLoadBalancerV2SetSecurityGroups(ctx context.Context, networkingClient *gophercloud.ServiceClient, vipPortID string, d *schema.ResourceData) error {
	if vipPortID != "" {
		if v, ok := d.GetOk("security_group_ids"); ok {
			securityGroups := util.ExpandToStringSlice(v.(*schema.Set).List())
			updateOpts := ports.UpdateOpts{
				SecurityGroups: &securityGroups,
			}

			log.Printf("[DEBUG] Adding security groups to vnpaycloud_lb_loadbalancer_v2 "+
				"VIP port %s: %#v", vipPortID, updateOpts)

			_, err := ports.Update(ctx, networkingClient, vipPortID, updateOpts).Extract()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func ResourceLoadBalancerV2GetSecurityGroups(ctx context.Context, networkingClient *gophercloud.ServiceClient, vipPortID string, d *schema.ResourceData) error {
	port, err := ports.Get(ctx, networkingClient, vipPortID).Extract()
	if err != nil {
		return err
	}

	d.Set("security_group_ids", port.SecurityGroups)

	return nil
}
