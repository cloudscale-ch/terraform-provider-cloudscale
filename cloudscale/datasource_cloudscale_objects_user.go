package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleObjectsUser() *schema.Resource {
	recordSchema := getObjectsUserSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("Objects Users", recordSchema, getFetchFunc(
			listObjectsUsers,
			gatherObjectsUserResourceData,
		)),
		Schema: recordSchema,
	}
}

func listObjectsUsers(d *schema.ResourceData, meta any) ([]cloudscale.ObjectsUser, error) {
	client := meta.(*cloudscale.Client)
	return client.ObjectsUsers.List(context.Background())
}
