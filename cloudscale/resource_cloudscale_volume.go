package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"size_gb": {
			Type:     schema.TypeInt,
			Required: true,
		},

		// Optional attributes

		"type": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"zone_slug": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"server_uuids": {
			Type:     schema.TypeList,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Optional: true,
		},

		// Computed attributes

		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func resourceVolumeCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.VolumeRequest{
		Name:   d.Get("name").(string),
		SizeGB: d.Get("size_gb").(int),
		Type:   d.Get("type").(string),
	}

	serverUUIDs := d.Get("server_uuids").([]interface{})
	s := make([]string, len(serverUUIDs))

	for i := range serverUUIDs {
		s[i] = serverUUIDs[i].(string)
	}

	if attr, ok := d.GetOk("zone_slug"); ok {
		opts.Zone = attr.(string)
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
	d.Set("zone_slug", volume.Zone.Slug)

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
		return CheckDeleted(d, err, "Error retrieving volume")
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
			opts := &cloudscale.VolumeRequest{}
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
		return CheckDeleted(d, err, "Error deleting volume")
	}
	return nil
}
