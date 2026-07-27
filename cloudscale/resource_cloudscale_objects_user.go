package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v10"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const objectsUserHumanName = "Objects User"

var (
	resourceCloudscaleObjectsUserCreate = getCreateOperation(createObjectsUser, nil)
	resourceCloudscaleObjectsUserRead   = getReadOperation(objectsUserHumanName, getGenericResourceIdentifierFromSchema, readObjectsUser, gatherObjectsUserResourceData)
	resourceCloudscaleObjectsUserUpdate = getUpdateOperation(objectsUserHumanName, getGenericResourceIdentifierFromSchema, updateObjectsUser, resourceCloudscaleObjectsUserRead, gatherObjectsUserUpdateRequest, nil)
	resourceCloudscaleObjectsUserDelete = getDeleteOperation(objectsUserHumanName, getGenericResourceIdentifierFromSchema, deleteObjectsUser, nil)
)

func resourceCloudscaleObjectsUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudscaleObjectsUserCreate,
		ReadContext:   resourceCloudscaleObjectsUserRead,
		UpdateContext: resourceCloudscaleObjectsUserUpdate,
		DeleteContext: resourceCloudscaleObjectsUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: getObjectsUserSchema(RESOURCE),
	}
}

func getObjectsUserSchema(t SchemaType) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
		"display_name": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Optional: t.isDataSource(),
		},
		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"user_id": {
			Type:     schema.TypeString,
			Optional: t.isDataSource(),
			Computed: true,
		},
		"keys": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"access_key": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"secret_key": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
			Computed:  true,
			Sensitive: true,
		},
		"tags": &TagsSchema,
	}
	if t.isDataSource() {
		m["id"] = &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		}
	}
	return m
}

func createObjectsUser(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.ObjectsUserRequest{
		DisplayName: d.Get("display_name").(string),
	}
	opts.Tags = TagsFromState(d)

	objectsUser, err := client.ObjectsUsers.Create(ctx, opts)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error creating objects user: %s", err))
	}

	d.SetId(objectsUser.ID)

	log.Printf("[INFO] Objects user ID %s", d.Id())

	return resourceCloudscaleObjectsUserRead(ctx, d, meta)
}

func gatherObjectsUserResourceData(objectsUser *cloudscale.ObjectsUser) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = objectsUser.ID
	m["href"] = objectsUser.HREF
	m["user_id"] = objectsUser.ID
	m["display_name"] = objectsUser.DisplayName
	m["tags"] = TagsToState(objectsUser.Tags)

	keys := make([]map[string]string, 0, len(objectsUser.Keys))
	for _, keyEntry := range objectsUser.Keys {
		g := map[string]string{}
		g["secret_key"] = keyEntry["secret_key"]
		g["access_key"] = keyEntry["access_key"]
		keys = append(keys, g)
	}
	m["keys"] = keys

	return m
}

func readObjectsUser(ctx context.Context, rId GenericResourceIdentifier, meta any) (*cloudscale.ObjectsUser, error) {
	client := meta.(*cloudscale.Client)
	return client.ObjectsUsers.Get(ctx, rId.Id)
}

func updateObjectsUser(ctx context.Context, rId GenericResourceIdentifier, meta any, updateRequest *cloudscale.ObjectsUserRequest) error {
	client := meta.(*cloudscale.Client)
	return client.ObjectsUsers.Update(ctx, rId.Id, updateRequest)
}

func gatherObjectsUserUpdateRequest(d *schema.ResourceData) []*cloudscale.ObjectsUserRequest {
	requests := make([]*cloudscale.ObjectsUserRequest, 0)

	for _, attribute := range []string{"display_name", "tags"} {
		if d.HasChange(attribute) {
			log.Printf("[INFO] Attribute %s changed", attribute)
			opts := &cloudscale.ObjectsUserRequest{}
			requests = append(requests, opts)
			if attribute == "display_name" {
				opts.DisplayName = d.Get(attribute).(string)
			} else if attribute == "tags" {
				opts.Tags = TagsFromState(d)
			}
		}
	}
	return requests
}

func deleteObjectsUser(ctx context.Context, rId GenericResourceIdentifier, meta any) error {
	client := meta.(*cloudscale.Client)
	return client.ObjectsUsers.Delete(ctx, rId.Id)
}
