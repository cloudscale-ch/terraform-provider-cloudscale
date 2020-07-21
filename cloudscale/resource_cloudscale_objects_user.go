package cloudscale

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceCloudScaleObjectsUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceObjectsUserCreate,
		Read:   resourceObjectsUserRead,
		Update: resourceObjectsUserUpdate,
		Delete: resourceObjectsUserDelete,

		Schema: getObjectsUserSchema(),
	}
}

func getObjectsUserSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// Required attributes

		"display_name": {
			Type:     schema.TypeString,
			Required: true,
		},

		// Computed attributes

		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"user_id": {
			Type:     schema.TypeString,
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
			Computed: true,
		},
	}
}

func resourceObjectsUserCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.ObjectsUserRequest{
		DisplayName: d.Get("display_name").(string),
	}

	objectsUser, err := client.ObjectsUsers.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating objects user: %s", err)
	}

	d.SetId(objectsUser.ID)

	log.Printf("[INFO] Objects user ID %s", d.Id())

	err = fillObjectsUserResourceData(d, objectsUser)
	if err != nil {
		return err
	}
	return nil
}

func fillObjectsUserResourceData(d *schema.ResourceData, objectsUser *cloudscale.ObjectsUser) error {
	d.Set("href", objectsUser.HREF)
	d.Set("user_id", objectsUser.ID)
	d.Set("display_name", objectsUser.DisplayName)

	keys := make([]map[string]string, 0, len(objectsUser.Keys))
	for _, keyEntry := range objectsUser.Keys {
		g := map[string]string{}
		g["secret_key"] = keyEntry["secret_key"]
		g["access_key"] = keyEntry["access_key"]
		keys = append(keys, g)
	}
	err := d.Set("keys", keys)

	return err
}

func resourceObjectsUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	objectsUser, err := client.ObjectsUsers.Get(context.Background(), d.Id())
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving objects user")
	}

	err = fillObjectsUserResourceData(d, objectsUser)
	if err != nil {
		return err
	}
	return nil
}

func resourceObjectsUserUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	for _, attribute := range []string{"display_name"} {
		// cloudscale.ch objectsUser attributes can only be changed one at a time.
		if d.HasChange(attribute) {
			opts := &cloudscale.ObjectsUserRequest{}
			if attribute == "display_name" {
				opts.DisplayName = d.Get(attribute).(string)
			}
			err := client.ObjectsUsers.Update(context.Background(), id, opts)
			if err != nil {
				return fmt.Errorf("Error updating the objects user (%s) status (%s) ", id, err)
			}
		}
	}
	return resourceObjectsUserRead(d, meta)
}

func resourceObjectsUserDelete(d *schema.ResourceData, meta interface{}) error {
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
