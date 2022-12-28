package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleSubnet() *schema.Resource {
	recordSchema := getSubnetSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("subnets", recordSchema, subnetsRead),
		Schema:      recordSchema,
	}
}

func subnetsRead(d *schema.ResourceData, meta any) ([]ResourceDataRaw, error) {
	client := meta.(*cloudscale.Client)
	subnetList, err := client.Subnets.List(context.Background())
	if err != nil {
		return nil, err
	}
	var rawItems []ResourceDataRaw
	for _, subnet := range subnetList {
		rawItems = append(rawItems, gatherSubnetResourceData(&subnet))
	}
	return rawItems, nil
}
