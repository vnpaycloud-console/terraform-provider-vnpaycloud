package shared

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func FlattenLBMembers(members []dto.Member) []map[string]interface{} {
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

func ExpandLBMembers(members *schema.Set, lbClient *client.Client) []dto.BatchUpdateMemberOpts {
	var m []dto.BatchUpdateMemberOpts

	if members != nil {
		for _, raw := range members.List() {
			rawMap := raw.(map[string]interface{})
			name := rawMap["name"].(string)
			subnetID := rawMap["subnet_id"].(string)
			weight := rawMap["weight"].(int)

			member := dto.BatchUpdateMemberOpts{
				Address:      rawMap["address"].(string),
				ProtocolPort: rawMap["protocol_port"].(int),
				Name:         &name,
				SubnetID:     &subnetID,
				Weight:       &weight,
			}

			// backup requires octavia minor version 2.1. Only set when specified
			//if val, ok := rawMap["backup"]; ok {
			//	backup := val.(bool)
			//	member.Backup = &backup
			//}

			// Only set monitor_port and monitor_address when explicitly specified, as they are optional arguments
			//if val, ok := rawMap["monitor_port"]; ok {
			//	monitorPort := val.(int)
			//	if monitorPort > 0 {
			//		member.MonitorPort = &monitorPort
			//	}
			//}

			//if val, ok := rawMap["monitor_address"]; ok {
			//	monitorAddress := val.(string)
			//	if monitorAddress != "" {
			//		member.MonitorAddress = &monitorAddress
			//	}
			//}

			m = append(m, member)
		}
	}

	return m
}

func getListenerIDForL7Policy(ctx context.Context, lbClient *client.Client, id string) (string, error) {
	log.Printf("[DEBUG] Trying to get Listener ID associated with the %s L7 Policy ID", id)
	listResp := dto.ListLoadBalancerResponse{}
	_, err := lbClient.Get(ctx, client.ApiPath.LbaasLoadBalancerWithParams(dto.ListLoadBalancerParams{}), &listResp, nil)
	if err != nil {
		return "", fmt.Errorf("No Load Balancers were found: %s", err)
	}

	lbs := listResp.LoadBalancers

	for _, lb := range lbs {
		getStatusResp := dto.GetLoadBalancerStatusResponse{}
		_, err := lbClient.Get(ctx, client.ApiPath.LbaasLoadBalancerWithId(lb.ID), &getStatusResp, nil)
		if err != nil {
			return "", fmt.Errorf("Failed to get Load Balancer statuses: %s", err)
		}
		for _, listener := range getStatusResp.Statuses.Loadbalancer.Listeners {
			for _, l7policy := range listener.L7Policies {
				if l7policy.ID == id {
					return listener.ID, nil
				}
			}
		}
	}

	return "", fmt.Errorf("Unable to find Listener ID associated with the %s L7 Policy ID", id)
}

func waitForLBL7Rule(ctx context.Context, lbClient *client.Client, parentListener *dto.Listener, parentL7policy *dto.L7Policy, l7rule *dto.Rule, target string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for l7rule %s to become %s.", l7rule.ID, target)

	if len(parentListener.Loadbalancers) == 0 {
		return fmt.Errorf("Unable to determine loadbalancer ID from listener %s", parentListener.ID)
	}

	lbID := parentListener.Loadbalancers[0].ID

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    resourceLBL7RuleRefreshFunc(ctx, lbClient, lbID, parentL7policy.ID, l7rule),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if client.ResponseCodeIs(err, http.StatusNotFound) {
			if target == "DELETED" {
				return nil
			}
		}

		return fmt.Errorf("Error waiting for l7rule %s to become %s: %s", l7rule.ID, target, err)
	}

	return nil
}

func resourceLBL7RuleRefreshFunc(ctx context.Context, lbClient *client.Client, lbID string, l7policyID string, l7rule *dto.Rule) retry.StateRefreshFunc {
	if l7rule.ProvisioningStatus == "" {
		return resourceLBLoadBalancerStatusRefreshFunc(ctx, lbClient, lbID, "l7rule", l7rule.ID, l7policyID)
	}

	return func() (interface{}, string, error) {
		lb, status, err := resourceLBLoadBalancerRefreshFunc(ctx, lbClient, lbID)()
		if err != nil {
			return lb, status, err
		}
		if !util.StrSliceContains(getLbSkipStatuses(), status) {
			return lb, status, nil
		}

		l7ruleResp := dto.GetRuleResponse{}
		_, err = lbClient.Get(ctx, client.ApiPath.LbaasL7RuleWithId(l7policyID, l7rule.ID), &l7ruleResp, nil)
		if err != nil {
			return nil, "", err
		}

		return l7ruleResp.Rule, l7ruleResp.Rule.ProvisioningStatus, nil
	}
}

func resourceLBListenerRefreshFunc(ctx context.Context, lbClient *client.Client, lbID string, listener *dto.Listener) retry.StateRefreshFunc {
	if listener.ProvisioningStatus == "" {
		return resourceLBLoadBalancerStatusRefreshFunc(ctx, lbClient, lbID, "listener", listener.ID, "")
	}

	return func() (interface{}, string, error) {
		lb, status, err := resourceLBLoadBalancerRefreshFunc(ctx, lbClient, lbID)()
		if err != nil {
			return lb, status, err
		}
		if !util.StrSliceContains(getLbSkipStatuses(), status) {
			return lb, status, nil
		}

		listenerResp := dto.GetListenerResponse{}
		_, err = lbClient.Get(ctx, client.ApiPath.LbaasListenerWithId(listener.ID), &listenerResp, nil)
		if err != nil {
			return nil, "", err
		}

		return listenerResp.Listener, listenerResp.Listener.ProvisioningStatus, nil
	}
}

func WaitForLBListener(ctx context.Context, lbClient *client.Client, listener *dto.Listener, target string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for vnpaycloud_lb_listener %s to become %s.", listener.ID, target)

	if len(listener.Loadbalancers) == 0 {
		return fmt.Errorf("Failed to detect a vnpaycloud_lb_loadbalancer for the %s vnpaycloud_lb_listener", listener.ID)
	}

	lbID := listener.Loadbalancers[0].ID

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    resourceLBListenerRefreshFunc(ctx, lbClient, lbID, listener),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if client.ResponseCodeIs(err, http.StatusNotFound) {
			if target == "DELETED" {
				return nil
			}
		}

		return fmt.Errorf("Error waiting for vnpaycloud_lb_listener %s to become %s: %s", listener.ID, target, err)
	}

	return nil
}

func lbFindLBIDviaPool(ctx context.Context, lbClient *client.Client, pool *dto.Pool) (string, error) {
	if len(pool.Loadbalancers) > 0 {
		return pool.Loadbalancers[0].ID, nil
	}

	if len(pool.Listeners) > 0 {
		listenerID := pool.Listeners[0].ID
		listenerResp := dto.GetListenerResponse{}
		_, err := lbClient.Get(ctx, client.ApiPath.LbaasListenerWithId(listenerID), &listenerResp, nil)
		if err != nil {
			return "", err
		}

		listener := listenerResp.Listener

		if len(listener.Loadbalancers) > 0 {
			return listener.Loadbalancers[0].ID, nil
		}
	}

	return "", fmt.Errorf("Unable to determine loadbalancer ID from pool %s", pool.ID)
}

func resourceLBPoolRefreshFunc(ctx context.Context, lbClient *client.Client, lbID string, pool *dto.Pool) retry.StateRefreshFunc {
	if pool.ProvisioningStatus == "" {
		return resourceLBLoadBalancerStatusRefreshFunc(ctx, lbClient, lbID, "pool", pool.ID, "")
	}

	return func() (interface{}, string, error) {
		lb, status, err := resourceLBLoadBalancerRefreshFunc(ctx, lbClient, lbID)()
		if err != nil {
			return lb, status, err
		}
		if !util.StrSliceContains(getLbSkipStatuses(), status) {
			return lb, status, nil
		}

		poolResp := dto.GetPoolResponse{}
		_, err = lbClient.Get(ctx, client.ApiPath.LbaasPoolWithId(pool.ID), &poolResp, nil)
		if err != nil {
			return nil, "", err
		}

		pool := poolResp.Pool

		return pool, pool.ProvisioningStatus, nil
	}
}

func WaitForLBPool(ctx context.Context, lbClient *client.Client, pool *dto.Pool, target string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for pool %s to become %s.", pool.ID, target)

	lbID, err := lbFindLBIDviaPool(ctx, lbClient, pool)
	if err != nil {
		return err
	}

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    resourceLBPoolRefreshFunc(ctx, lbClient, lbID, pool),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		if client.ResponseCodeIs(err, http.StatusNotFound) {
			if target == "DELETED" {
				return nil
			}
		}

		return fmt.Errorf("Error waiting for pool %s to become %s: %s", pool.ID, target, err)
	}

	return nil
}

func resourceLBLoadBalancerStatusRefreshFunc(ctx context.Context, lbClient *client.Client, lbID, resourceType, resourceID string, parentID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		statusesResp := dto.GetLoadBalancerStatusResponse{}
		_, err := lbClient.Get(ctx, client.ApiPath.LbaasLoadBalancerWithId(lbID), &statusesResp, nil)
		if err != nil {
			if client.ResponseCodeIs(err, http.StatusNotFound) {
				return nil, "", client.ErrUnexpectedResponseCode{
					Actual: http.StatusNotFound,
					BaseError: client.BaseError{
						DefaultErrString: fmt.Sprintf("Unable to get statuses from the Load Balancer %s statuses tree: %s", lbID, err),
					},
				}
			}
			return nil, "", fmt.Errorf("Unable to get statuses from the Load Balancer %s statuses tree: %s", lbID, err)
		}

		statuses := statusesResp.Statuses

		// Don't fail, when statuses returns "null"
		if statuses == nil || statuses.Loadbalancer == nil {
			statuses = new(dto.LoadBalancerStatusTree)
			statuses.Loadbalancer = new(dto.LoadBalancer)
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
			listenerResp := dto.GetListenerResponse{}
			_, err := lbClient.Get(ctx, client.ApiPath.LbaasListenerWithId(resourceID), &listenerResp, nil)
			if err != nil {
				return nil, "", err
			}
			listener := listenerResp.Listener

			return listener, "ACTIVE", err

		case "pool":
			for _, pool := range statuses.Loadbalancer.Pools {
				if pool.ID == resourceID {
					if pool.ProvisioningStatus != "" {
						return pool, pool.ProvisioningStatus, nil
					}
				}
			}
			poolResp := dto.GetPoolResponse{}
			_, err := lbClient.Get(ctx, client.ApiPath.LbaasPoolWithId(resourceID), &poolResp, nil)
			if err != nil {
				return nil, "", err
			}
			pool := poolResp.Pool

			return pool, "ACTIVE", err

		case "monitor":
			for _, pool := range statuses.Loadbalancer.Pools {
				if pool.Monitor.ID == resourceID {
					if pool.Monitor.ProvisioningStatus != "" {
						return pool.Monitor, pool.Monitor.ProvisioningStatus, nil
					}
				}
			}
			monitorResp := dto.GetMonitorResponse{}
			_, err := lbClient.Get(ctx, client.ApiPath.LbaasHealthMonitorWithId(resourceID), &monitorResp, nil)
			if err != nil {
				return nil, "", err
			}
			monitor := monitorResp.HealthMonitor

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
			memberResp := dto.GetMemberResponse{}
			_, err := lbClient.Get(ctx, client.ApiPath.LbaasPoolMemberWithId(parentID, resourceID), &memberResp, nil)
			if err != nil {
				return nil, "", err
			}
			member := memberResp.Member

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

			l7policyResp := dto.GetL7PolicyResponse{}
			_, err := lbClient.Get(ctx, client.ApiPath.LbaasL7PolicyWithId(resourceID), &l7policyResp, nil)

			if err != nil {
				return nil, "", err
			}

			l7policy := l7policyResp.L7Policy

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
			l7ruleResp := dto.GetRuleResponse{}
			_, err := lbClient.Get(ctx, client.ApiPath.LbaasL7RuleWithId(parentID, resourceID), &l7ruleResp, nil)
			if err != nil {
				return nil, "", err
			}
			l7Rule := l7ruleResp.Rule

			return l7Rule, "ACTIVE", err
		}

		return nil, "", fmt.Errorf("An unexpected error occurred querying the status of %s %s by loadbalancer %s", resourceType, resourceID, lbID)
	}
}

func waitForLBL7Policy(ctx context.Context, lbClient *client.Client, parentListener *dto.Listener, l7policy *dto.L7Policy, target string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for l7policy %s to become %s.", l7policy.ID, target)

	if len(parentListener.Loadbalancers) == 0 {
		return fmt.Errorf("Unable to determine loadbalancer ID from listener %s", parentListener.ID)
	}

	lbID := parentListener.Loadbalancers[0].ID

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    resourceLBL7PolicyRefreshFunc(ctx, lbClient, lbID, l7policy),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if client.ResponseCodeIs(err, http.StatusNotFound) {
			if target == "DELETED" {
				return nil
			}
		}

		return fmt.Errorf("Error waiting for l7policy %s to become %s: %s", l7policy.ID, target, err)
	}

	return nil
}

func resourceLBL7PolicyRefreshFunc(ctx context.Context, lbClient *client.Client, lbID string, l7policy *dto.L7Policy) retry.StateRefreshFunc {
	if l7policy.ProvisioningStatus == "" {
		return resourceLBLoadBalancerStatusRefreshFunc(ctx, lbClient, lbID, "l7policy", l7policy.ID, "")
	}

	return func() (interface{}, string, error) {
		lb, status, err := resourceLBLoadBalancerRefreshFunc(ctx, lbClient, lbID)()
		if err != nil {
			return lb, status, err
		}
		if !util.StrSliceContains(getLbSkipStatuses(), status) {
			return lb, status, nil
		}

		l7policyResp := dto.GetL7PolicyResponse{}
		_, err = lbClient.Get(ctx, client.ApiPath.LbaasL7PolicyWithId(l7policy.ID), &l7policyResp, nil)
		if err != nil {
			return nil, "", err
		}

		l7policy := l7policyResp.L7Policy

		return l7policy, l7policy.ProvisioningStatus, nil
	}
}

func WaitForLBMember(ctx context.Context, lbClient *client.Client, parentPool *dto.Pool, member *dto.Member, target string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for member %s to become %s.", member.ID, target)

	lbID, err := lbFindLBIDviaPool(ctx, lbClient, parentPool)
	if err != nil {
		return err
	}

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    resourceLBMemberRefreshFunc(ctx, lbClient, lbID, parentPool.ID, member),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		if client.ResponseCodeIs(err, http.StatusNotFound) {
			if target == "DELETED" {
				return nil
			}
		}

		return fmt.Errorf("Error waiting for member %s to become %s: %s", member.ID, target, err)
	}

	return nil
}

func resourceLBMemberRefreshFunc(ctx context.Context, lbClient *client.Client, lbID string, poolID string, member *dto.Member) retry.StateRefreshFunc {
	if member.ProvisioningStatus == "" {
		return resourceLBLoadBalancerStatusRefreshFunc(ctx, lbClient, lbID, "member", member.ID, poolID)
	}

	return func() (interface{}, string, error) {
		lb, status, err := resourceLBLoadBalancerRefreshFunc(ctx, lbClient, lbID)()
		if err != nil {
			return lb, status, err
		}
		if !util.StrSliceContains(getLbSkipStatuses(), status) {
			return lb, status, nil
		}

		memberResp := dto.GetMemberResponse{}
		_, err = lbClient.Get(ctx, client.ApiPath.LbaasPoolMemberWithId(poolID, member.ID), &memberResp, nil)
		if err != nil {
			return nil, "", err
		}
		member := memberResp.Member

		return member, member.ProvisioningStatus, nil
	}
}

func WaitForLBMonitor(ctx context.Context, lbClient *client.Client, parentPool *dto.Pool, monitor *dto.Monitor, target string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for vnpaycloud_lb_monitor %s to become %s.", monitor.ID, target)

	lbID, err := lbFindLBIDviaPool(ctx, lbClient, parentPool)
	if err != nil {
		return err
	}

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    resourceLBMonitorRefreshFunc(ctx, lbClient, lbID, monitor),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		if client.ResponseCodeIs(err, http.StatusNotFound) {
			if target == "DELETED" {
				return nil
			}
		}
		return fmt.Errorf("Error waiting for vnpaycloud_lb_monitor %s to become %s: %s", monitor.ID, target, err)
	}

	return nil
}

func resourceLBMonitorRefreshFunc(ctx context.Context, lbClient *client.Client, lbID string, monitor *dto.Monitor) retry.StateRefreshFunc {
	if monitor.ProvisioningStatus == "" {
		return resourceLBLoadBalancerStatusRefreshFunc(ctx, lbClient, lbID, "monitor", monitor.ID, "")
	}

	return func() (interface{}, string, error) {
		lb, status, err := resourceLBLoadBalancerRefreshFunc(ctx, lbClient, lbID)()
		if err != nil {
			return lb, status, err
		}
		if !util.StrSliceContains(getLbSkipStatuses(), status) {
			return lb, status, nil
		}

		monitorResp := dto.GetMonitorResponse{}
		_, err = lbClient.Get(ctx, client.ApiPath.LbaasHealthMonitorWithId(monitor.ID), &monitorResp, nil)
		if err != nil {
			return nil, "", err
		}
		monitor := monitorResp.HealthMonitor
		return monitor, monitor.ProvisioningStatus, nil
	}
}

func ExpandLBPoolTLSVersion(v []interface{}) []dto.TLSVersion {
	versions := make([]dto.TLSVersion, len(v))
	for i, v := range v {
		versions[i] = dto.TLSVersion(v.(string))
	}
	return versions
}

func ExpandLBListenerTLSVersion(v []interface{}) []dto.TLSVersion {
	versions := make([]dto.TLSVersion, len(v))
	for i, v := range v {
		versions[i] = dto.TLSVersion(v.(string))
	}
	return versions
}

func FlattenLBPoolPersistence(p dto.SessionPersistence) []map[string]interface{} {
	if p == (dto.SessionPersistence{}) {
		return nil
	}
	return []map[string]interface{}{
		{
			"type":        p.Type,
			"cookie_name": p.CookieName,
		},
	}
}

func ExpandLBPoolPersistance(p []interface{}) (*dto.SessionPersistence, error) {
	persistence := &dto.SessionPersistence{}

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

func WaitForLBLoadBalancer(ctx context.Context, lbClient *client.Client, lbID string, target string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for loadbalancer %s to become %s.", lbID, target)

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    resourceLBLoadBalancerRefreshFunc(ctx, lbClient, lbID),
		Timeout:    timeout,
		Delay:      0,
		MinTimeout: 1 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if client.ResponseCodeIs(err, http.StatusNotFound) {
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

func resourceLBLoadBalancerRefreshFunc(ctx context.Context, lbClient *client.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		lbResp := dto.GetLoadBalancerResponse{}
		_, err := lbClient.Get(ctx, client.ApiPath.LbaasLoadBalancerWithId(id), &lbResp, nil)
		if err != nil {
			return nil, "", err
		}
		lb := lbResp.LoadBalancer

		return lb, lb.ProvisioningStatus, nil
	}
}

func ResourceLoadBalancerSetSecurityGroups(ctx context.Context, networkingClient *client.Client, vipPortID string, d *schema.ResourceData) error {
	if vipPortID != "" {
		if v, ok := d.GetOk("security_group_ids"); ok {
			securityGroups := util.ExpandToStringSlice(v.(*schema.Set).List())
			updateOpts := dto.UpdatePortOpts{
				SecurityGroups: &securityGroups,
			}

			log.Printf("[DEBUG] Adding security groups to vnpaycloud_lb_loadbalancer_v2 "+
				"VIP port %s: %#v", vipPortID, updateOpts)

			_, err := networkingClient.Put(ctx, client.ApiPath.PortWithId(vipPortID), updateOpts, nil, nil)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func ResourceLoadBalancerGetSecurityGroups(ctx context.Context, networkingClient *client.Client, vipPortID string, d *schema.ResourceData) error {
	portResp := dto.GetPortResponse{}
	_, err := networkingClient.Get(ctx, client.ApiPath.PortWithId(vipPortID), &portResp, nil)
	if err != nil {
		return err
	}

	d.Set("security_group_ids", portResp.Port.SecurityGroups)

	return nil
}
