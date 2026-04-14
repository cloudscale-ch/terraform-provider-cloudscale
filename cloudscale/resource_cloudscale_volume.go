package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const volumeHumanName = "volume"

var (
	resourceCloudscaleVolumeRead   = getReadOperation(volumeHumanName, getGenericResourceIdentifierFromSchema, readVolume, gatherVolumeResourceData)
	resourceCloudscaleVolumeUpdate = getUpdateOperation(volumeHumanName, getGenericResourceIdentifierFromSchema, updateVolume, resourceCloudscaleVolumeRead, gatherVolumeUpdateRequests)
	resourceCloudscaleVolumeDelete = getDeleteOperation(volumeHumanName, deleteVolume)
)

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
	// For resources, "type" and "zone_slug" conflict with "volume_snapshot_uuid".
	// For data sources, there are no such conflicts.
	snapshotConflictsWith := []string{}
	if t.isResource() {
		snapshotConflictsWith = append(snapshotConflictsWith, "volume_snapshot_uuid")
	}

	m := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Optional: t.isDataSource(),
		},
		"size_gb": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"type": {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ForceNew:      true,
			ConflictsWith: snapshotConflictsWith,
		},
		"zone_slug": {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ForceNew:      true,
			ConflictsWith: snapshotConflictsWith,
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
	if t.isResource() {
		m["volume_snapshot_uuid"] = &schema.Schema{
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
			ConflictsWith: []string{"type", "zone_slug"},
			DiffSuppressFunc: func(_, old, new string, d *schema.ResourceData) bool {
				// volume_snapshot_uuid is write-only: the API accepts it at
				// creation but never returns it, and gatherVolumeResourceData
				// does not set it in state. Without this suppress, removing
				// the attribute from config (e.g. after deleting the source
				// snapshot) would trigger ForceNew and destroy the volume
				// that was created from that snapshot.
				return old != "" && new == ""
			},
		}
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

	opts := &cloudscale.VolumeCreateRequest{
		Name: d.Get("name").(string),
	}
	opts.Tags = CopyTags(d)

	snapshotUUID, fromSnapshot := d.GetOk("volume_snapshot_uuid")
	if fromSnapshot {
		// Create from snapshot: only name, volume_snapshot_uuid, and tags
		// are accepted by the API. Size, type, zone, and servers are
		// applied via subsequent Update calls below.
		log.Printf("[INFO] Volume will be created from snapshot %s", snapshotUUID.(string))
		opts.VolumeSnapshotUUID = snapshotUUID.(string)
	} else {
		log.Printf("[INFO] Volume will be created empty (not from snapshot)")
		opts.SizeGB = d.Get("size_gb").(int)
		opts.Type = d.Get("type").(string)

		if attr, ok := d.GetOk("zone_slug"); ok {
			opts.Zone = attr.(string)
		}

		serverUUIDs := d.Get("server_uuids").([]any)
		s := make([]string, len(serverUUIDs))
		for i := range serverUUIDs {
			s[i] = serverUUIDs[i].(string)
		}
		opts.ServerUUIDs = &s
	}

	log.Printf("[DEBUG] Volume create configuration: %#v", opts)

	volume, err := client.Volumes.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating volume: %s", err)
	}

	d.SetId(volume.UUID)
	log.Printf("[INFO] Volume ID %s", d.Id())

	if fromSnapshot {
		if v, ok := d.GetOk("size_gb"); ok {
			log.Printf("[INFO] Resizing volume %s to %d GB after creation from snapshot", volume.UUID, v.(int))
			updateReq := &cloudscale.VolumeUpdateRequest{SizeGB: v.(int)}
			if err := client.Volumes.Update(context.Background(), volume.UUID, updateReq); err != nil {
				return fmt.Errorf("Error resizing volume (%s) after creation from snapshot: %s", volume.UUID, err)
			}
		}

		serverUUIDs := d.Get("server_uuids").([]any)
		if len(serverUUIDs) > 0 {
			log.Printf("[INFO] Attaching volume %s to %d server(s) after creation from snapshot", volume.UUID, len(serverUUIDs))
			s := make([]string, len(serverUUIDs))
			for i := range serverUUIDs {
				s[i] = serverUUIDs[i].(string)
			}
			updateReq := &cloudscale.VolumeUpdateRequest{ServerUUIDs: &s}
			if err := client.Volumes.Update(context.Background(), volume.UUID, updateReq); err != nil {
				return fmt.Errorf("Error attaching volume (%s) after creation from snapshot: %s", volume.UUID, err)
			}
		}
	}

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

func updateVolume(rId GenericResourceIdentifier, meta any, updateRequest *cloudscale.VolumeUpdateRequest) error {
	client := meta.(*cloudscale.Client)
	return client.Volumes.Update(context.Background(), rId.Id, updateRequest)
}

func gatherVolumeUpdateRequests(d *schema.ResourceData) []*cloudscale.VolumeUpdateRequest {
	requests := make([]*cloudscale.VolumeUpdateRequest, 0)

	for _, attribute := range []string{"name", "size_gb", "server_uuids", "tags"} {
		if d.HasChange(attribute) {
			log.Printf("[INFO] Attribute %s changed", attribute)
			opts := &cloudscale.VolumeUpdateRequest{}
			requests = append(requests, opts)

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
		}
	}
	return requests
}

func deleteVolume(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	return client.Volumes.Delete(context.Background(), id)
}
