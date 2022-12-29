package cloudscale

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudscaleObjectsUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceObjectsUserCreate,
		Read:   resourceObjectsUserRead,
		Update: resourceObjectsUserUpdate,
		Delete: resourceObjectsUserDelete,

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

func resourceObjectsUserCreate(d *schema.ResourceData, meta any) error {
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

	fillObjectsUserResourceData(d, objectsUser)
	return nil
}

func fillObjectsUserResourceData(d *schema.ResourceData, objectsUser *cloudscale.ObjectsUser) {
	fillResourceData(d, gatherObjectsUserResourceData(objectsUser))
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

func resourceObjectsUserRead(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)

	objectsUser, err := client.ObjectsUsers.Get(context.Background(), d.Id())
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving objects user")
	}

	fillObjectsUserResourceData(d, objectsUser)
	return nil
}

func resourceObjectsUserUpdate(d *schema.ResourceData, meta any) error {
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
	return resourceObjectsUserRead(d, meta)
}

func resourceObjectsUserDelete(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	log.Printf("[INFO] Deleting objects user: %s", d.Id())
	// sending the next request immediately can cause errors, since the port cleanup process is still ongoing
	time.Sleep(5 * time.Second)
	err := client.ObjectsUsers.Delete(context.Background(), id)

	if err != nil {
		return CheckDeleted(d, err, "Error deleting objectsUser")
	}
	return nil
}
