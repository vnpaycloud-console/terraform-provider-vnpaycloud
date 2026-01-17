package flavor

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
)

func IDFromName(ctx context.Context, flavorClient *client.Client, name string) (string, error) {
	IDs, err := IDsFromName(ctx, flavorClient, name)
	if err != nil {
		return "", err
	}

	switch count := len(IDs); count {
	case 0:
		return "", &client.ErrResourceNotFound{Name: name, ResourceType: "flavor"}
	case 1:
		return IDs[0], nil
	default:
		return "", &client.ErrMultipleResourcesFound{Name: name, Count: count, ResourceType: "flavor"}
	}
}

func IDsFromName(ctx context.Context, flavorClient *client.Client, name string) ([]string, error) {
	flavorsResp := dto.ListFlavorsResponse{}
	_, err := flavorClient.All(ctx, client.ApiPath.FlavorDetail, &flavorsResp, nil)
	if err != nil {
		return nil, err
	}

	IDs := make([]string, 0, len(flavorsResp.Flavors))
	for _, flavor := range flavorsResp.Flavors {
		if flavor.Name == name {
			IDs = append(IDs, flavor.ID)
		}
	}

	return IDs, nil
}
