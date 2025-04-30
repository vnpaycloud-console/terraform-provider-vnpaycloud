package network

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vnpaycloud-console/gophercloud/v2"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/extensions/dns"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/extensions/external"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/extensions/mtu"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/extensions/portsecurity"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/extensions/provider"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/extensions/qos/policies"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/extensions/vlantransparent"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/networks"
)

type networkExtended struct {
	networks.Network
	external.NetworkExternalExt
	vlantransparent.TransparentExt
	portsecurity.PortSecurityExt
	mtu.NetworkMTUExt
	dns.NetworkDNSExt
	policies.QoSPolicyExt
	provider.NetworkProviderExt
}

func expandNetworkingNetworkSegments(segments *schema.Set) []provider.Segment {
	rawSegments := segments.List()

	if len(rawSegments) == 1 {
		// unset segments
		rawMap := rawSegments[0].(map[string]interface{})
		if rawMap["physical_network"] == "" &&
			rawMap["network_type"] == "" &&
			rawMap["segmentation_id"] == 0 {
			return nil
		}
	}

	providerSegments := make([]provider.Segment, len(rawSegments))
	for i, raw := range rawSegments {
		rawMap := raw.(map[string]interface{})
		providerSegments[i] = provider.Segment{
			PhysicalNetwork: rawMap["physical_network"].(string),
			NetworkType:     rawMap["network_type"].(string),
			SegmentationID:  rawMap["segmentation_id"].(int),
		}
	}

	return providerSegments
}

func resourceNetworkingNetworkStateRefreshFunc(ctx context.Context, client *gophercloud.ServiceClient, networkID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		n, err := networks.Get(ctx, client, networkID).Extract()
		if err != nil {
			if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
				return n, "DELETED", nil
			}
			if gophercloud.ResponseCodeIs(err, http.StatusConflict) {
				return n, "ACTIVE", nil
			}

			return n, "", err
		}

		return n, n.Status, nil
	}
}

func flattenNetworkingNetworkSegments(network networkExtended) []map[string]interface{} {
	singleSegment := 0
	if network.NetworkType != "" ||
		network.PhysicalNetwork != "" ||
		network.SegmentationID != "" {
		singleSegment = 1
	}
	segmentsSet := make([]map[string]interface{}, len(network.Segments)+singleSegment)

	if singleSegment > 0 {
		segmentationID, err := strconv.Atoi(network.SegmentationID)
		if err != nil {
			log.Printf("[DEBUG] Unable to convert %q segmentation ID to an integer: %s", network.SegmentationID, err)
		}
		segmentsSet[0] = map[string]interface{}{
			"physical_network": network.PhysicalNetwork,
			"network_type":     network.NetworkType,
			"segmentation_id":  segmentationID,
		}
	}

	for i, segment := range network.Segments {
		segmentsSet[i+singleSegment] = map[string]interface{}{
			"physical_network": segment.PhysicalNetwork,
			"network_type":     segment.NetworkType,
			"segmentation_id":  segment.SegmentationID,
		}
	}

	return segmentsSet
}
