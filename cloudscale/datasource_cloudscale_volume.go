package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v4"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleVolume() *schema.Resource {
	recordSchema := getVolumeSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("volumes", recordSchema, getFetchFunc(
			listVolumes,
			gatherVolumeResourceData,
		)),
		Schema: recordSchema,
	}
}

func listVolumes(d *schema.ResourceData, meta any) ([]cloudscale.Volume, error) {
	client := meta.(*cloudscale.Client)
	return client.Volumes.List(context.Background())
}
