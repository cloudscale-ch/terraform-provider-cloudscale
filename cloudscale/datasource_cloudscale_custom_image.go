package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v4"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleCustomImage() *schema.Resource {
	recordSchema := getCustomImageSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("custom images", recordSchema, getFetchFunc(
			listCustomImages,
			gatherCustomImageResourceData,
		)),
		Schema: recordSchema,
	}
}

func listCustomImages(d *schema.ResourceData, meta any) ([]cloudscale.CustomImage, error) {
	client := meta.(*cloudscale.Client)
	return client.CustomImages.List(context.Background())
}
