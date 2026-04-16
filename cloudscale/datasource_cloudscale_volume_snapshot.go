package cloudscale

import (
	"context"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleVolumeSnapshot() *schema.Resource {
	recordSchema := getVolumeSnapshotSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("volume snapshots", recordSchema, getFetchFunc(
			listVolumeSnapshots,
			gatherVolumeSnapshotResourceData,
		)),
		Schema: recordSchema,
	}
}

func listVolumeSnapshots(d *schema.ResourceData, meta any) ([]cloudscale.VolumeSnapshot, error) {
	client := meta.(*cloudscale.Client)
	return client.VolumeSnapshots.List(context.Background())
}