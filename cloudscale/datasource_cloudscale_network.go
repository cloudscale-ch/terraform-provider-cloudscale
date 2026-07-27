package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v10"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleNetwork() *schema.Resource {
	recordSchema := getNetworkSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("networks", recordSchema, getFetchFunc(
			listNetworks,
			gatherNetworkResourceData,
		)),
		Schema: recordSchema,
	}
}

func listNetworks(ctx context.Context, d *schema.ResourceData, meta any) ([]cloudscale.Network, error) {
	client := meta.(*cloudscale.Client)
	return client.Networks.List(ctx)
}
