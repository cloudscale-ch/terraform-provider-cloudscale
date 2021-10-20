package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleSubnet() *schema.Resource {
	recordSchema := getSubnetSchema(true)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("subnets", recordSchema, subnetsRead),
		Schema:      recordSchema,
	}
}

func subnetsRead(meta interface{}) ([]ResourceDataRaw, error) {
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
