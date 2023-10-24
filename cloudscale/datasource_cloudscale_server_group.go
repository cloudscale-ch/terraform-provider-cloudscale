package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v4"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleServerGroup() *schema.Resource {
	recordSchema := getServerGroupSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("server groups", recordSchema, getFetchFunc(
			listServerGroups,
			gatherServerGroupResourceData,
		)),
		Schema: recordSchema,
	}
}

func listServerGroups(d *schema.ResourceData, meta any) ([]cloudscale.ServerGroup, error) {
	client := meta.(*cloudscale.Client)
	return client.ServerGroups.List(context.Background())
}
