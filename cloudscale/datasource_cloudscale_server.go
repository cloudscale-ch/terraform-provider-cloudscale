package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleServer() *schema.Resource {
	recordSchema := getServerSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("servers", recordSchema, getFetchFunc(
			listServers,
			gatherServerResourceData,
		)),
		Schema: recordSchema,
	}
}

func listServers(d *schema.ResourceData, meta any) ([]cloudscale.Server, error) {
	client := meta.(*cloudscale.Client)
	return client.Servers.List(context.Background())
}
