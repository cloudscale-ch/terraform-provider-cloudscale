package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleServerGroup() *schema.Resource {
	recordSchema := getServerGroupSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("server groups", recordSchema, serverGroupsRead),
		Schema:      recordSchema,
	}
}

func serverGroupsRead(meta interface{}) ([]ResourceDataRaw, error) {
	client := meta.(*cloudscale.Client)
	serverGroupList, err := client.ServerGroups.List(context.Background())
	if err != nil {
		return nil, err
	}
	var rawItems []ResourceDataRaw
	for _, serverGroup := range serverGroupList {
		rawItems = append(rawItems, gatherServerGroupResourceData(&serverGroup))
	}
	return rawItems, nil
}
