package cloudscale

import (
       "context"

       "github.com/cloudscale-ch/cloudscale-go-sdk"
       "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudScaleNetwork() *schema.Resource {
	recordSchema := customImageSchema()

	for _, f := range recordSchema {
		f.Computed = true
	}

	return &schema.Resource{
		ReadContext: dataSourceResourceRead(networksRead),
		Schema:      recordSchema,
	}
}


func networksRead(meta interface{}) ([]ResourceDataRaw, error) {
	client := meta.(*cloudscale.Client)
	networkList, err := client.Networks.List(context.Background())
	if err != nil {
		return nil, err
	}
	var rawItems []ResourceDataRaw
	for _, network := range networkList {
		rawItems = append(rawItems, gatherNetworkResourceData(&network))
	}
	return rawItems, nil
}
