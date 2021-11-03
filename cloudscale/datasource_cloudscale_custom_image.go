package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleCustomImage() *schema.Resource {
	recordSchema := getCustomImageSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("custom images", recordSchema, customImagesRead),
		Schema:      recordSchema,
	}
}

func customImagesRead(meta interface{}) ([]ResourceDataRaw, error) {
	client := meta.(*cloudscale.Client)
	customImageList, err := client.CustomImages.List(context.Background())
	if err != nil {
		return nil, err
	}
	var rawItems []ResourceDataRaw
	for _, customImage := range customImageList {
		rawItems = append(rawItems, gatherCustomImageResourceData(&customImage))
	}
	return rawItems, nil
}
