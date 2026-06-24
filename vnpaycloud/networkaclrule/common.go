package networkaclrule

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func networkACLRuleStateRefreshFunc(ctx context.Context, c *client.Client, projectID, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp := &dto.NetworkACLRuleResponse{}
		_, err := c.Get(ctx, client.ApiPath.NetworkACLRuleWithID(projectID, id), resp, nil)
		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return &dto.NetworkACLRule{}, "deleted", nil
			}
			return nil, "", err
		}

		if resp.Rule.ID == "" {
			return &dto.NetworkACLRule{}, "deleted", nil
		}

		return &resp.Rule, resp.Rule.Status, nil
	}
}

var networkACLRuleTypes = []string{
	"ALL_TRAFFIC",
	"CUSTOM_TCP",
	"CUSTOM_UDP",
	"ICMP",
	"SSH",
	"TELNET",
	"SMTP",
	"DNS_TCP",
	"DNS_UDP",
	"HTTP",
	"HTTPS",
}

func setNetworkACLRuleAttributes(d *schema.ResourceData, rule dto.NetworkACLRule) {
	d.SetId(rule.ID)
	_ = d.Set("nacl_id", rule.NaclID)
	_ = d.Set("name", rule.Name)
	_ = d.Set("priority", rule.Priority)
	_ = d.Set("type", rule.Type)
	_ = d.Set("action", rule.Action)
	_ = d.Set("port_start", rule.PortStart)
	_ = d.Set("port_end", rule.PortEnd)
	_ = d.Set("source", rule.Source)
	_ = d.Set("destination", rule.Destination)
	_ = d.Set("icmp_type", rule.IcmpType)
	_ = d.Set("description", rule.Description)
	_ = d.Set("status", rule.Status)
}

func validateNetworkACLRuleConfig(d *schema.ResourceData) error {
	ruleType := d.Get("type").(string)
	portStart, hasPortStart := d.GetOk("port_start")
	portEnd, hasPortEnd := d.GetOk("port_end")

	if ruleType == "CUSTOM_TCP" || ruleType == "CUSTOM_UDP" {
		if !hasPortStart || !hasPortEnd {
			return fmt.Errorf("port_start and port_end are required for %s", ruleType)
		}
		if portStart.(int) < 1 || portEnd.(int) < 1 || portStart.(int) > 65535 || portEnd.(int) > 65535 {
			return fmt.Errorf("port_start and port_end must be in range 1-65535 for %s", ruleType)
		}
		if portEnd.(int) < portStart.(int) {
			return fmt.Errorf("port_end must be greater than or equal to port_start")
		}
	}

	if ruleType != "ICMP" {
		if icmpType, ok := d.GetOk("icmp_type"); ok && icmpType.(string) != "" {
			return fmt.Errorf("icmp_type can only be set when type is ICMP")
		}
	}

	return nil
}
