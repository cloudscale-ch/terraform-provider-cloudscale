package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const volumeHumanName = "volume"

var resourceCloudscaleVolumeRead = getReadOperation(volumeHumanName, getGenericResourceIdentifierFromSchema, readVolume, gatherVolumeResourceData)
var resourceCloudscaleVolumeDelete = getDeleteOperation(volumeHumanName, deleteVolume)

func resourceCloudscaleVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudscaleVolumeCreate,
		Read:   resourceCloudscaleVolumeRead,
		Update: resourceCloudscaleVolumeUpdate,
		Delete: resourceCloudscaleVolumeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: getVolumeSchema(RESOURCE),
	}
}

func getVolumeSchema(t SchemaType) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Optional: t.isDataSource(),
		},
		"size_gb": {
			Type:     schema.TypeInt,
			Required: t.isResource(),
			Computed: t.isDataSource(),
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
			Optional: t.isResource(),
			Computed: t.isDataSource(),
		},
		"href": {
			Type:     schema.TypeString,
			Computed: true,
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

func resourceCloudscaleVolumeCreate(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.VolumeRequest{
		Name:   d.Get("name").(string),
		SizeGB: d.Get("size_gb").(int),
		Type:   d.Get("type").(string),
	}

	serverUUIDs := d.Get("server_uuids").([]any)
	s := make([]string, len(serverUUIDs))

	for i := range serverUUIDs {
		s[i] = serverUUIDs[i].(string)
	}

	if attr, ok := d.GetOk("zone_slug"); ok {
		opts.Zone = attr.(string)
	}

	opts.ServerUUIDs = &s
	opts.Tags = CopyTags(d)

	log.Printf("[DEBUG] Volume create configuration: %#v", opts)

	volume, err := client.Volumes.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating volume: %s", err)
	}

	d.SetId(volume.UUID)

	log.Printf("[INFO] Volume ID %s", d.Id())

	err = resourceCloudscaleVolumeRead(d, meta)
	if err != nil {
		return fmt.Errorf("Error reading the volume (%s): %s", d.Id(), err)
	}
	return nil
}

func gatherVolumeResourceData(volume *cloudscale.Volume) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = volume.UUID
	m["href"] = volume.HREF
	m["name"] = volume.Name
	m["size_gb"] = volume.SizeGB
	m["type"] = volume.Type
	m["zone_slug"] = volume.Zone.Slug
	m["server_uuids"] = volume.ServerUUIDs
	m["tags"] = volume.Tags
	return m
}

func readVolume(rId GenericResourceIdentifier, meta any) (*cloudscale.Volume, error) {
	client := meta.(*cloudscale.Client)
	return client.Volumes.Get(context.Background(), rId.Id)
}

func resourceCloudscaleVolumeUpdate(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	for _, attribute := range []string{"name", "size_gb", "server_uuids", "tags"} {
		// cloudscale.ch volume attributes can only be changed one at a time.
		// This means that it's not possible to scale in the same call as
		// attaching the volume to a different server.
		if d.HasChange(attribute) {
			log.Printf("[INFO] Attribute %s changed", attribute)
			opts := &cloudscale.VolumeRequest{}
			if attribute == "server_uuids" {
				serverUUIDs := d.Get("server_uuids").([]any)
				s := make([]string, len(serverUUIDs))

				for i := range serverUUIDs {
					s[i] = serverUUIDs[i].(string)
				}
				opts.ServerUUIDs = &s
			} else if attribute == "name" {
				opts.Name = d.Get(attribute).(string)
			} else if attribute == "size_gb" {
				opts.SizeGB = d.Get(attribute).(int)
			} else if attribute == "tags" {
				opts.Tags = CopyTags(d)
			}
			err := client.Volumes.Update(context.Background(), id, opts)
			if err != nil {
				return fmt.Errorf("Error updating the Volume (%s) status (%s) ", id, err)
			}
		}
	}
	return resourceCloudscaleVolumeRead(d, meta)
}

func deleteVolume(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	return client.Volumes.Delete(context.Background(), id)
}
