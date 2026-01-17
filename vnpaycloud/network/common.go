package network

import (
	"context"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandNetworkingNetworkSegments(segments *schema.Set) []dto.Segments {
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

	providerSegments := make([]dto.Segments, len(rawSegments))
	for i, raw := range rawSegments {
		rawMap := raw.(map[string]interface{})
		providerSegments[i] = dto.Segments{
			PhysicalNetwork: rawMap["physical_network"].(string),
			NetworkType:     rawMap["network_type"].(string),
			SegmentationID:  rawMap["segmentation_id"].(int),
		}
	}

	return providerSegments
}

func resourceNetworkingNetworkStateRefreshFunc(ctx context.Context, networkingClient *client.Client, networkID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		networkResp := &dto.GetNetworkResponse{}
		_, err := networkingClient.Get(ctx, client.ApiPath.NetworkWithId(networkID), networkResp, nil)
		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return networkResp, "DELETED", nil
			}
			if util.ResponseCodeIs(err, http.StatusConflict) {
				return networkResp, "ACTIVE", nil
			}

			return networkResp, "", err
		}

		return networkResp, networkResp.Network.Status, nil
	}
}

func flattenNetworkingNetworkSegments(network dto.Network) []map[string]interface{} {
	singleSegment := 0
	//if network.NetworkType != "" ||
	//	network.PhysicalNetwork != "" ||
	//	network.SegmentationID != "" {
	//	singleSegment = 1
	//}
	segmentsSet := make([]map[string]interface{}, len(network.Segments)+singleSegment)

	//if singleSegment > 0 {
	//	segmentationID, err := strconv.Atoi(network.SegmentationID)
	//	if err != nil {
	//		log.Printf("[DEBUG] Unable to convert %q segmentation ID to an integer: %s", network.SegmentationID, err)
	//	}
	//	segmentsSet[0] = map[string]interface{}{
	//		"physical_network": network.PhysicalNetwork,
	//		"network_type":     network.NetworkType,
	//		"segmentation_id":  segmentationID,
	//	}
	//}

	for i, segment := range network.Segments {
		segmentsSet[i+singleSegment] = map[string]interface{}{
			"physical_network": segment.PhysicalNetwork,
			"network_type":     segment.NetworkType,
			"segmentation_id":  segment.SegmentationID,
		}
	}

	return segmentsSet
}
