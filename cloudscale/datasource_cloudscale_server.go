package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleServer() *schema.Resource {
	recordSchema := getServerSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("servers", recordSchema, serversRead),
		Schema:      recordSchema,
	}
}

func serversRead(meta interface{}) ([]ResourceDataRaw, error) {
	client := meta.(*cloudscale.Client)
	serverList, err := client.Servers.List(context.Background())
	if err != nil {
		return nil, err
	}
	var rawItems []ResourceDataRaw
	for _, server := range serverList {
		rawItems = append(rawItems, gatherServerResourceData(&server))
	}
	return rawItems, nil
}
