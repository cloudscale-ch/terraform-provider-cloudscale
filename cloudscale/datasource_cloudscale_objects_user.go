package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleObjectsUser() *schema.Resource {
	recordSchema := getObjectsUserSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("Objects Users", recordSchema, objectsUsersRead),
		Schema:      recordSchema,
	}
}

func objectsUsersRead(meta interface{}) ([]ResourceDataRaw, error) {
	client := meta.(*cloudscale.Client)
	objectsUserList, err := client.ObjectsUsers.List(context.Background())
	if err != nil {
		return nil, err
	}
	var rawItems []ResourceDataRaw
	for _, objectsUser := range objectsUserList {
		rawItems = append(rawItems, gatherObjectsUserResourceData(&objectsUser))
	}
	return rawItems, nil
}
