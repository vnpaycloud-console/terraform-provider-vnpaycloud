package image

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
)

func IDFromName(ctx context.Context, imageClient *client.Client, name string) (string, error) {
	IDs, err := IDsFromName(ctx, imageClient, name)
	if err != nil {
		return "", err
	}

	switch count := len(IDs); count {
	case 0:
		return "", client.ErrResourceNotFound{Name: name, ResourceType: "image"}
	case 1:
		return IDs[0], nil
	default:
		return "", client.ErrMultipleResourcesFound{Name: name, Count: count, ResourceType: "image"}
	}
}

func IDsFromName(ctx context.Context, imageClient *client.Client, name string) ([]string, error) {
	params := dto.ListImagesParams{
		Name: name,
	}
	imagesResp := dto.ListImagesResponse{}
	_, err := imageClient.All(ctx, client.ApiPath.ImageWithParams(params), &imagesResp, nil)
	if err != nil {
		return nil, err
	}

	IDs := make([]string, len(imagesResp.Images))
	for i := range imagesResp.Images {
		IDs[i] = imagesResp.Images[i].ID
	}

	return IDs, nil
}
