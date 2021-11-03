package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleVolume() *schema.Resource {
	recordSchema := getVolumeSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("volumes", recordSchema, volumesRead),
		Schema:      recordSchema,
	}
}

func volumesRead(meta interface{}) ([]ResourceDataRaw, error) {
	client := meta.(*cloudscale.Client)
	volumeList, err := client.Volumes.List(context.Background())
	if err != nil {
		return nil, err
	}
	var rawItems []ResourceDataRaw
	for _, volume := range volumeList {
		rawItems = append(rawItems, gatherVolumeResourceData(&volume))
	}
	return rawItems, nil
}
