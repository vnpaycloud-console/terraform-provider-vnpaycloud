package securityGroupRule

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func ResourceNetworkingSecGroupRuleStateRefreshFunc(ctx context.Context, secGroupRuleclient *client.Client, sgRuleID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		securityGroupRuleResp := &dto.GetSecurityGroupRuleResponse{}
		_, err := secGroupRuleclient.Get(ctx, client.ApiPath.SecurityGroupRuleWithId(sgRuleID), securityGroupRuleResp, nil)
		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return securityGroupRuleResp.SecurityGroupRule, "DELETED", nil
			}

			return securityGroupRuleResp.SecurityGroupRule, "", err
		}

		return securityGroupRuleResp.SecurityGroupRule, "ACTIVE", nil
	}
}

func ResourceNetworkingSecGroupRuleDirection(v interface{}, k string) ([]string, []error) {
	switch dto.RuleDirection(v.(string)) {
	case dto.RuleDirIngress:
		return nil, nil
	case dto.RuleDirEgress:
		return nil, nil
	}

	return nil, []error{fmt.Errorf("unknown %q %s for vnpaycloud_networking_secgroup_rule", v, k)}
}

func ResourceNetworkingSecGroupRuleEtherType(v interface{}, k string) ([]string, []error) {
	switch dto.RuleEtherType(v.(string)) {
	case dto.RuleEtherType4:
		return nil, nil
	case dto.RuleEtherType6:
		return nil, nil
	}

	return nil, []error{fmt.Errorf("unknown %q %s for vnpaycloud_networking_secgroup_rule", v, k)}
}

func resourceNetworkingSecGroupRuleProtocol(v interface{}, k string) ([]string, []error) {
	//nolint:exhaustive // we need to override the rules.ProtocolAny case with an empty string
	switch dto.RuleProtocol(v.(string)) {
	case dto.RuleProtocolAH,
		dto.RuleProtocolDCCP,
		dto.RuleProtocolEGP,
		dto.RuleProtocolESP,
		dto.RuleProtocolGRE,
		dto.RuleProtocolICMP,
		dto.RuleProtocolIGMP,
		dto.RuleProtocolIPv6Encap,
		dto.RuleProtocolIPv6Frag,
		dto.RuleProtocolIPv6ICMP,
		dto.RuleProtocolIPv6NoNxt,
		dto.RuleProtocolIPv6Opts,
		dto.RuleProtocolIPv6Route,
		dto.RuleProtocolOSPF,
		dto.RuleProtocolPGM,
		dto.RuleProtocolRSVP,
		dto.RuleProtocolSCTP,
		dto.RuleProtocolTCP,
		dto.RuleProtocolUDP,
		dto.RuleProtocolUDPLite,
		dto.RuleProtocolVRRP,
		dto.RuleProtocolIPIP,
		"": // ProtocolAny
		return nil, nil
	}

	// If the protocol wasn't matched above, see if it's an integer.
	p, err := strconv.Atoi(v.(string))
	if err != nil {
		return nil, []error{fmt.Errorf("unknown %q %s for vnpaycloud_networking_secgroup_rule: %s", v, k, err)}
	}
	if p < 0 || p > 255 {
		return nil, []error{fmt.Errorf("unknown %q %s for vnpaycloud_networking_secgroup_rule", v, k)}
	}

	return nil, nil
}
