package cloudscale

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const volumeSnapshotHumanName = "volume snapshot"

func snapshotLockKey(volumeUUID string) string {
	return fmt.Sprintf("cloudscale/volume-snapshot/%s", volumeUUID)
}

// volumeSnapshotLockKey serializes snapshot operations on their source volume. The
// cloudscale API rejects concurrent snapshot operations on the same source volume, so
// create/update/delete are serialized per volume UUID. Because withLock wraps the
// whole operation, the lock is held through the full create/delete status-wait cycle
// (see resourceCloudscaleVolumeSnapshotCreate / deleteVolumeSnapshot), ensuring the
// volume is no longer busy before the next operation on it starts.
var volumeSnapshotLockKey = uuidLockKey("source_volume_uuid", snapshotLockKey)

var (
	resourceCloudscaleVolumeSnapshotRead   = getReadOperation(volumeSnapshotHumanName, getGenericResourceIdentifierFromSchema, readVolumeSnapshot, gatherVolumeSnapshotResourceData)
	resourceCloudscaleVolumeSnapshotUpdate = getUpdateOperation(volumeSnapshotHumanName, getGenericResourceIdentifierFromSchema, updateVolumeSnapshot, resourceCloudscaleVolumeSnapshotRead, gatherVolumeSnapshotUpdateRequest, volumeSnapshotLockKey)
	resourceCloudscaleVolumeSnapshotDelete = getDeleteOperation(volumeSnapshotHumanName, getGenericResourceIdentifierFromSchema, deleteVolumeSnapshot, volumeSnapshotLockKey)
)

func resourceCloudscaleVolumeSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateContext: withLock(volumeSnapshotLockKey, resourceCloudscaleVolumeSnapshotCreate),
		ReadContext:   resourceCloudscaleVolumeSnapshotRead,
		UpdateContext: resourceCloudscaleVolumeSnapshotUpdate,
		DeleteContext: resourceCloudscaleVolumeSnapshotDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: getVolumeSnapshotSchema(RESOURCE),
	}
}

func getVolumeSnapshotSchema(t SchemaType) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Optional: t.isDataSource(),
		},
		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"source_volume_uuid": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			ForceNew: t.isResource(),
			Optional: t.isDataSource(),
			Computed: t.isDataSource(),
		},
		"source_volume_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"source_volume_href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"size_gb": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"status": {
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

func resourceCloudscaleVolumeSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	timeout := d.Timeout(schema.TimeoutCreate)
	startTime := time.Now()

	client := meta.(*cloudscale.Client)

	sourceVolumeUUID := d.Get("source_volume_uuid").(string)

	opts := &cloudscale.VolumeSnapshotCreateRequest{
		Name:         d.Get("name").(string),
		SourceVolume: sourceVolumeUUID,
	}
	opts.Tags = CopyTags(d)

	log.Printf("[DEBUG] VolumeSnapshot create configuration: %#v", opts)

	snap, err := client.VolumeSnapshots.Create(ctx, opts)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error creating VolumeSnapshot: %s", err))
	}

	d.SetId(snap.UUID)

	log.Printf("[INFO] VolumeSnapshot ID: %s", d.Id())

	remainingTime := timeout - time.Since(startTime)
	_, err = waitForStatus(ctx, []string{}, "available", &remainingTime, newVolumeSnapshotRefreshFunc(ctx, d, "status", meta))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for volume snapshot (%s) to become available: %s", d.Id(), err))
	}

	return resourceCloudscaleVolumeSnapshotRead(ctx, d, meta)
}

func newVolumeSnapshotRefreshFunc(ctx context.Context, d *schema.ResourceData, attribute string, meta any) resource.StateRefreshFunc {
	client := meta.(*cloudscale.Client)
	return func() (any, string, error) {
		id := d.Id()

		// get the instance
		snap, err := client.VolumeSnapshots.Get(ctx, id)
		if err != nil {
			return nil, "", fmt.Errorf("error retrieving volume snapshot (%s) (refresh) %s", id, err)
		}

		data := gatherVolumeSnapshotResourceData(snap)
		attr, ok := data[attribute]
		if !ok {
			return nil, "", nil
		}

		// return attr
		return snap, attr.(string), nil
	}
}

func gatherVolumeSnapshotResourceData(snap *cloudscale.VolumeSnapshot) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = snap.UUID
	m["href"] = snap.HREF
	m["name"] = snap.Name
	m["source_volume_uuid"] = snap.SourceVolume.UUID
	m["source_volume_name"] = snap.SourceVolume.Name
	m["source_volume_href"] = snap.SourceVolume.HREF
	m["size_gb"] = snap.SizeGB
	m["status"] = snap.Status
	m["tags"] = snap.Tags
	return m
}

func readVolumeSnapshot(ctx context.Context, rId GenericResourceIdentifier, meta any) (*cloudscale.VolumeSnapshot, error) {
	client := meta.(*cloudscale.Client)
	return client.VolumeSnapshots.Get(ctx, rId.Id)
}

func updateVolumeSnapshot(ctx context.Context, rId GenericResourceIdentifier, meta any, updateRequest *cloudscale.VolumeSnapshotUpdateRequest) error {
	client := meta.(*cloudscale.Client)
	return client.VolumeSnapshots.Update(ctx, rId.Id, updateRequest)
}

func gatherVolumeSnapshotUpdateRequest(d *schema.ResourceData) []*cloudscale.VolumeSnapshotUpdateRequest {
	requests := make([]*cloudscale.VolumeSnapshotUpdateRequest, 0)

	for _, attribute := range []string{"name", "tags"} {
		if d.HasChange(attribute) {
			log.Printf("[INFO] Attribute %s changed", attribute)
			opts := &cloudscale.VolumeSnapshotUpdateRequest{}
			requests = append(requests, opts)

			if attribute == "name" {
				opts.Name = d.Get(attribute).(string)
			} else if attribute == "tags" {
				opts.Tags = CopyTags(d)
			}
		}
	}
	return requests
}

func deleteVolumeSnapshot(ctx context.Context, rId GenericResourceIdentifier, meta any) error {
	client := meta.(*cloudscale.Client)

	if err := client.VolumeSnapshots.Delete(ctx, rId.Id); err != nil {
		return err
	}
	// Unlike most cloudscale resources that disappear immediately after DELETE,
	// volume snapshots go through a background cleanup period. During this time
	// the snapshot still exists with status "deleting". We wait for it to be gone.
	return waitForVolumeSnapshotDeleted(ctx, rId.Id, meta)
}

func waitForVolumeSnapshotDeleted(ctx context.Context, id string, meta any) error {
	client := meta.(*cloudscale.Client)
	err := waitForDeleted(ctx, func() (exists bool, err error) {
		snapshot, err := client.VolumeSnapshots.Get(ctx, id)
		if err != nil {
			if errorResponse, ok := err.(*cloudscale.ErrorResponse); ok && errorResponse.StatusCode == http.StatusNotFound { // API returns 404 once fully deleted
				return false, nil // gone
			}
			return false, fmt.Errorf("error retrieving volume snapshot (%s) (delete refresh) %s", id, err)
		}
		log.Printf("[INFO] Status is %s", snapshot.Status)
		return true, nil // still exists
	})
	if err != nil {
		return fmt.Errorf("error waiting for volume snapshot (%s) to be deleted: %s", id, err)
	}
	return nil
}
