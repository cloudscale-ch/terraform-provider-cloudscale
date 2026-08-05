package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleRouter() *schema.Resource {
	recordSchema := getRouterSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("routers", recordSchema, getFetchFunc(
			listRouters,
			gatherRouterResourceData,
		)),
		Schema: recordSchema,
	}
}

func listRouters(ctx context.Context, d *schema.ResourceData, meta any) ([]cloudscale.Router, error) {
	client := meta.(*cloudscale.Client)
	return client.Routers.List(ctx)
}
