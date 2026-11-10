package kubernetesapiwhitelist

import (
	"fmt"
	"net"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const maxNoteLength = 64

// validateWhitelistCIDR mirrors the backend rules so the stored value equals
// the configured one: IPv4 network address in CIDR notation, not 0.0.0.0/0.
func validateWhitelistCIDR(v interface{}, k string) ([]string, []error) {
	s, ok := v.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}
	ip, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		return nil, []error{fmt.Errorf("%q must be an IPv4 CIDR such as 203.0.113.0/24 or 203.0.113.5/32, got %q", k, s)}
	}
	if ip.To4() == nil {
		return nil, []error{fmt.Errorf("%q only supports IPv4, got %q", k, s)}
	}
	if !ip.Equal(ipNet.IP) {
		return nil, []error{fmt.Errorf("%q %s is not a network address, use %s", k, s, ipNet.String())}
	}
	if ones, _ := ipNet.Mask.Size(); ones == 0 {
		return nil, []error{fmt.Errorf("%q 0.0.0.0/0 is not allowed; destroy the resource to allow every IP", k)}
	}
	return nil, nil
}

func entrySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cidr": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateWhitelistCIDR,
			},
			"note": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, maxNoteLength),
			},
		},
	}
}

func expandEntries(v interface{}) []dto.ApiWhitelistEntry {
	entries := []dto.ApiWhitelistEntry{}
	set, ok := v.(*schema.Set)
	if !ok {
		return entries
	}
	for _, raw := range set.List() {
		m := raw.(map[string]interface{})
		entries = append(entries, dto.ApiWhitelistEntry{
			Cidr: m["cidr"].(string),
			Note: m["note"].(string),
		})
	}
	return entries
}

func flattenEntries(entries []dto.ApiWhitelistEntry) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(entries))
	for _, e := range entries {
		out = append(out, map[string]interface{}{
			"cidr": e.Cidr,
			"note": e.Note,
		})
	}
	return out
}

func setWhitelistData(d *schema.ResourceData, resp *dto.ClusterApiWhitelistResponse) {
	d.Set("entry", flattenEntries(resp.Entries))
	d.Set("endpoint", resp.Endpoint)
	d.Set("private_access", resp.PrivateAccess)
	d.Set("max_entries", resp.MaxEntries)
	d.Set("supported", resp.Supported)
	d.Set("unsupported_reason", resp.UnsupportedReason)
}

func computedSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"endpoint":           {Type: schema.TypeString, Computed: true},
		"private_access":     {Type: schema.TypeBool, Computed: true},
		"max_entries":        {Type: schema.TypeInt, Computed: true},
		"supported":          {Type: schema.TypeBool, Computed: true},
		"unsupported_reason": {Type: schema.TypeString, Computed: true},
	}
}
