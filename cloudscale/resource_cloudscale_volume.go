package cloudscale

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCloudScaleVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceVolumeCreate,
		Read:   resourceVolumeRead,
		Update: resourceVolumeUpdate,
		Delete: resourceVolumeDelete,

		Schema: getVolumeSchema(),
	}
}

func getVolumeSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{

		// Required attributes

		"name": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},
		"size_gb": &schema.Schema{
			Type:     schema.TypeInt,
			Required: true,
		},

		// Optional attributes

		"type": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"server_uuids": {
			Type:     schema.TypeList,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Optional: true,
		},

		// Computed attributes

		"href": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func resourceVolumeCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.Volume{
		Name:   d.Get("name").(string),
		SizeGB: d.Get("size_gb").(int),
		Type:   d.Get("type").(string),
	}

	serverUUIDs := d.Get("server_uuids").([]interface{})
	s := make([]string, len(serverUUIDs))

	for i := range serverUUIDs {
		s[i] = serverUUIDs[i].(string)
	}

	opts.ServerUUIDs = &s

	log.Printf("[DEBUG] Volume create configuration: %#v", opts)

	volume, err := client.Volumes.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating volume: %s", err)
	}

	d.SetId(volume.UUID)

	log.Printf("[INFO] Volume ID %s", d.Id())

	err = fillVolumeResourceData(d, volume)
	if err != nil {
		return err
	}
	return nil
}

func fillVolumeResourceData(d *schema.ResourceData, volume *cloudscale.Volume) error {
	d.Set("href", volume.HREF)
	d.Set("name", volume.Name)
	d.Set("size_gb", volume.SizeGB)
	d.Set("type", volume.Type)

	err := d.Set("server_uuids", volume.ServerUUIDs)
	if err != nil {
		log.Printf("[DEBUG] Error setting server_uuids attribute: %#v, error: %#v", volume.ServerUUIDs, err)
		return fmt.Errorf("Error setting server_uuids attribute: %#v, error: %#v", volume.ServerUUIDs, err)
	}
	return nil
}

func resourceVolumeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	volume, err := client.Volumes.Get(context.Background(), d.Id())
	if err != nil {
		errorResponse, ok := err.(*cloudscale.ErrorResponse)
		if ok && errorResponse.StatusCode == http.StatusNotFound {
			log.Printf("[WARN] Cloudscale Volume (%s) not found", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving volume: %s", err)
	}

	err = fillVolumeResourceData(d, volume)
	if err != nil {
		return err
	}
	return nil
}

func resourceVolumeUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	for _, attribute := range []string{"name", "size_gb", "server_uuids"} {
		// cloudscale.ch volume attributes can only be changed one at a time.
		// This means that it's not possible to scale in the same call as
		// attaching the volume to a different server.
		if d.HasChange(attribute) {
			opts := &cloudscale.Volume{}
			if attribute == "server_uuids" {
				serverUUIDs := d.Get("server_uuids").([]interface{})
				s := make([]string, len(serverUUIDs))

				for i := range serverUUIDs {
					s[i] = serverUUIDs[i].(string)
				}
				opts.ServerUUIDs = &s
			} else if attribute == "name" {
				opts.Name = d.Get(attribute).(string)
			} else if attribute == "size_gb" {
				opts.SizeGB = d.Get(attribute).(int)
			}
			err := client.Volumes.Update(context.Background(), id, opts)
			if err != nil {
				return fmt.Errorf("Error updating the Volume (%s) status (%s) ", id, err)
			}
		}
	}
	return resourceVolumeRead(d, meta)
}

func resourceVolumeDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	log.Printf("[INFO] Deleting Volume: %s", d.Id())
	err := client.Volumes.Delete(context.Background(), id)

	if err != nil {
		errorResponse, ok := err.(*cloudscale.ErrorResponse)
		if ok && errorResponse.StatusCode == http.StatusNotFound {
			log.Printf("[WARN] Cloudscale Volume (%s) not found", d.Id())
		} else {
			return fmt.Errorf("Error deleting Volume: %s", err)
		}
	}

	d.SetId("")
	return nil
}
