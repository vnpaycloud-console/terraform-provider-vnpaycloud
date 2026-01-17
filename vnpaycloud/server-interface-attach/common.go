package serverInterfaceAttach

import (
	"context"
	"log"
	"net/http"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func ComputeInterfaceAttachAttachFunc(ctx context.Context,
	computeClient *client.Client, instanceID, attachmentID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		interfaceResp := &dto.GetInterfaceResponse{}
		_, err := computeClient.Get(ctx, client.ApiPath.ServerInterfaceAttachWithId(instanceID, attachmentID), interfaceResp, nil)
		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return interfaceResp.Interface, "ATTACHING", nil
			}
			return interfaceResp.Interface, "", err
		}

		return interfaceResp.Interface, "ATTACHED", nil
	}
}

func ComputeInterfaceAttachDetachFunc(ctx context.Context,
	computeClient *client.Client, instanceID, attachmentID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Attempting to detach vnpaycloud_compute_interface_attach %s from instance %s",
			attachmentID, instanceID)

		interfaceResp := &dto.GetInterfaceResponse{}
		_, err := computeClient.Get(ctx, client.ApiPath.ServerInterfaceAttachWithId(instanceID, attachmentID), interfaceResp, nil)
		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return interfaceResp.Interface, "DETACHED", nil
			}
			return interfaceResp.Interface, "", err
		}

		_, err = computeClient.Delete(ctx, client.ApiPath.ServerInterfaceAttachWithId(instanceID, attachmentID), nil)
		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return interfaceResp.Interface, "DETACHED", nil
			}

			if util.ResponseCodeIs(err, http.StatusBadRequest) {
				return nil, "", nil
			}

			return nil, "", err
		}

		log.Printf("[DEBUG] vnpaycloud_compute_interface_attach %s is still active.", attachmentID)
		return nil, "", nil
	}
}
