package serverInterfaceAttach

import (
	"context"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/vnpaycloud-console/gophercloud/v2"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/compute/v2/attachinterfaces"
)

func ComputeInterfaceAttachV2AttachFunc(ctx context.Context,
	computeClient *gophercloud.ServiceClient, instanceID, attachmentID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		va, err := attachinterfaces.Get(ctx, computeClient, instanceID, attachmentID).Extract()
		if err != nil {
			if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
				return va, "ATTACHING", nil
			}
			return va, "", err
		}

		return va, "ATTACHED", nil
	}
}

func ComputeInterfaceAttachV2DetachFunc(ctx context.Context,
	computeClient *gophercloud.ServiceClient, instanceID, attachmentID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Attempting to detach openstack_compute_interface_attach_v2 %s from instance %s",
			attachmentID, instanceID)

		va, err := attachinterfaces.Get(ctx, computeClient, instanceID, attachmentID).Extract()
		if err != nil {
			if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
				return va, "DETACHED", nil
			}
			return va, "", err
		}

		err = attachinterfaces.Delete(ctx, computeClient, instanceID, attachmentID).ExtractErr()
		if err != nil {
			if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
				return va, "DETACHED", nil
			}

			if gophercloud.ResponseCodeIs(err, http.StatusBadRequest) {
				return nil, "", nil
			}

			return nil, "", err
		}

		log.Printf("[DEBUG] openstack_compute_interface_attach_v2 %s is still active.", attachmentID)
		return nil, "", nil
	}
}
