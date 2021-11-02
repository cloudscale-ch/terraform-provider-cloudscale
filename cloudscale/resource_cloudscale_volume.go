package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudscaleVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceVolumeCreate,
		Read:   resourceVolumeRead,
		Update: resourceVolumeUpdate,
		Delete: resourceVolumeDelete,

		Schema: getVolumeSchema(false),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

	}
}

func getVolumeSchema(isDataSource bool) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: !isDataSource,
			Optional: isDataSource,
		},
		"size_gb": {
			Type:     schema.TypeInt,
			Required: !isDataSource,
			Computed: isDataSource,
		},
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
			Optional: !isDataSource,
			Computed: isDataSource,
		},
		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
	if isDataSource {
		m["id"] = &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		}
	}
	return m
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

	fillVolumeResourceData(d, volume)
	return nil
}

func fillVolumeResourceData(d *schema.ResourceData, volume *cloudscale.Volume) {
	fillResourceData(d, gatherVolumeResourceData(volume))
}

func gatherVolumeResourceData(volume *cloudscale.Volume) ResourceDataRaw {
	m := make(map[string]interface{})
	m["id"] = volume.UUID
	m["href"] = volume.HREF
	m["name"] = volume.Name
	m["size_gb"] = volume.SizeGB
	m["type"] = volume.Type
	m["zone_slug"] = volume.Zone.Slug
	m["server_uuids"] = volume.ServerUUIDs
	return m
}

func resourceVolumeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	volume, err := client.Volumes.Get(context.Background(), d.Id())
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving volume")
	}

	fillVolumeResourceData(d, volume)
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
