package cloudscale

import (
	"context"
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

const objectsUserHumanName = "Objects User"

var resourceCloudscaleObjectsUserRead = getReadOperation(objectsUserHumanName, readObjectsUser, gatherObjectsUserResourceData)
var resourceCloudscaleObjectsUserDelete = getDeleteOperation(objectsUserHumanName, deleteObjectsUser)

func resourceCloudscaleObjectsUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudscaleObjectsUserCreate,
		Read:   resourceCloudscaleObjectsUserRead,
		Update: resourceCloudscaleObjectsUserUpdate,
		Delete: resourceCloudscaleObjectsUserDelete,

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

func resourceCloudscaleObjectsUserCreate(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.ObjectsUserRequest{
		DisplayName: d.Get("display_name").(string),
	}
	opts.Tags = CopyTags(d)

	objectsUser, err := client.ObjectsUsers.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating objects user: %s", err)
	}

	d.SetId(objectsUser.ID)

	log.Printf("[INFO] Objects user ID %s", d.Id())

	err = resourceCloudscaleObjectsUserRead(d, meta)
	if err != nil {
		return fmt.Errorf("Error reading the objects user (%s): %s", d.Id(), err)
	}
	return nil
}

func gatherObjectsUserResourceData(objectsUser *cloudscale.ObjectsUser) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = objectsUser.ID
	m["href"] = objectsUser.HREF
	m["user_id"] = objectsUser.ID
	m["display_name"] = objectsUser.DisplayName
	m["tags"] = objectsUser.Tags

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

func readObjectsUser(d *schema.ResourceData, meta any) (*cloudscale.ObjectsUser, error) {
	client := meta.(*cloudscale.Client)
	return client.ObjectsUsers.Get(context.Background(), d.Id())
}

func resourceCloudscaleObjectsUserUpdate(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	for _, attribute := range []string{"display_name", "tags"} {
		// cloudscale.ch objectsUser attributes can only be changed one at a time.
		if d.HasChange(attribute) {
			opts := &cloudscale.ObjectsUserRequest{}
			if attribute == "display_name" {
				opts.DisplayName = d.Get(attribute).(string)
			} else if attribute == "tags" {
				opts.Tags = CopyTags(d)
			}
			err := client.ObjectsUsers.Update(context.Background(), id, opts)
			if err != nil {
				return fmt.Errorf("Error updating the objects user (%s) status (%s) ", id, err)
			}
		}
	}
	return resourceCloudscaleObjectsUserRead(d, meta)
}

func deleteObjectsUser(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	return client.ObjectsUsers.Delete(context.Background(), id)
}
