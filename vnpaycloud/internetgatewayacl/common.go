package internetgatewayacl

import (
	"fmt"
	"net"
)

// reservedCIDRs are the ranges the platform refuses to blacklist on an internet gateway.
var reservedCIDRs = []string{
	"0.0.0.0/8",
	"10.0.0.0/8",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"192.88.99.0/24",
	"192.168.0.0/16",
	"198.18.0.0/15",
	"198.51.100.0/24",
	"203.0.113.0/24",
	"224.0.0.0/4",
	"240.0.0.0/4",
	"255.255.255.255/32",
}

func validateBlacklistCIDR(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}

	ip, ipNet, err := net.ParseCIDR(v)
	if err != nil {
		if net.ParseIP(v) != nil {
			return nil, []error{fmt.Errorf("%q is not in CIDR format, write %s/32 instead", k, v)}
		}
		return nil, []error{fmt.Errorf("%q is not a valid CIDR block: %s", k, v)}
	}

	if ip.To4() == nil {
		return nil, []error{fmt.Errorf("%q must be an IPv4 CIDR block, got %s", k, v)}
	}

	for _, reserved := range reservedCIDRs {
		_, reservedNet, err := net.ParseCIDR(reserved)
		if err != nil {
			continue
		}
		if reservedNet.Contains(ipNet.IP) {
			return nil, []error{fmt.Errorf("%q cannot be blacklisted: %s falls inside the reserved range %s", k, v, reserved)}
		}
	}

	return nil, nil
}
