package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleSubnet() *schema.Resource {
	recordSchema := getSubnetSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("subnets", recordSchema, getFetchFunc(
			listSubnets,
			gatherSubnetResourceData,
		)),
		Schema: recordSchema,
	}
}

func listSubnets(d *schema.ResourceData, meta any) ([]cloudscale.Subnet, error) {
	client := meta.(*cloudscale.Client)
	return client.Subnets.List(context.Background())
}
