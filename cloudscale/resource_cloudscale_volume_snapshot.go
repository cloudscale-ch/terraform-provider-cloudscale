package cloudscale

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const volumeSnapshotHumanName = "volume snapshot"

var (
	resourceCloudscaleVolumeSnapshotRead   = getReadOperation(volumeSnapshotHumanName, getGenericResourceIdentifierFromSchema, readVolumeSnapshot, gatherVolumeSnapshotResourceData)
	resourceCloudscaleVolumeSnapshotUpdate = getUpdateOperation(volumeSnapshotHumanName, getGenericResourceIdentifierFromSchema, updateVolumeSnapshot, resourceCloudscaleVolumeSnapshotRead, gatherVolumeSnapshotUpdateRequest)
	resourceCloudscaleVolumeSnapshotDelete = getDeleteOperation(volumeSnapshotHumanName, deleteVolumeSnapshot)
)

func resourceCloudscaleVolumeSnapshot() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudscaleVolumeSnapshotCreate,
		Read:   resourceCloudscaleVolumeSnapshotRead,
		Update: resourceCloudscaleVolumeSnapshotUpdate,
		Delete: resourceCloudscaleVolumeSnapshotDelete,

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
		"source_volume_uuid": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			ForceNew: t.isResource(),
			Optional: t.isDataSource(),
			Computed: t.isDataSource(),
		},
		"href": {
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

func snapshotLockKey(volumeUUID string) string {
	return fmt.Sprintf("cloudscale/volume-snapshot/%s", volumeUUID)
}

func resourceCloudscaleVolumeSnapshotCreate(d *schema.ResourceData, meta any) error {
	timeout := d.Timeout(schema.TimeoutCreate)
	startTime := time.Now()

	client := meta.(*cloudscale.Client)

	sourceVolumeUUID := d.Get("source_volume_uuid").(string)
	// The cloudscale API rejects concurrent snapshot operations on the same source
	// volume. Lock per volume UUID so that if two snapshots of the same volume are
	// created in the same apply, they are serialized. The lock is held through the
	// full create + status-wait cycle to ensure the volume is no longer busy before
	// the next operation starts.
	globalMu.Lock(snapshotLockKey(sourceVolumeUUID))
	defer globalMu.Unlock(snapshotLockKey(sourceVolumeUUID))

	opts := &cloudscale.VolumeSnapshotCreateRequest{
		Name:         d.Get("name").(string),
		SourceVolume: sourceVolumeUUID,
	}
	opts.Tags = CopyTags(d)

	log.Printf("[DEBUG] VolumeSnapshot create configuration: %#v", opts)

	snap, err := client.VolumeSnapshots.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating VolumeSnapshot: %s", err)
	}

	d.SetId(snap.UUID)

	log.Printf("[INFO] VolumeSnapshot ID: %s", d.Id())

	remainingTime := timeout - time.Since(startTime)
	_, err = waitForStatus([]string{}, "available", &remainingTime, newVolumeSnapshotRefreshFunc(d, "status", meta))
	if err != nil {
		return fmt.Errorf("error waiting for volume snapshot (%s) to become available: %s", d.Id(), err)
	}

	err = resourceCloudscaleVolumeSnapshotRead(d, meta)
	if err != nil {
		return fmt.Errorf("error reading the volume snapshot (%s): %s", d.Id(), err)
	}
	return nil
}

func newVolumeSnapshotRefreshFunc(d *schema.ResourceData, attribute string, meta any) resource.StateRefreshFunc {
	client := meta.(*cloudscale.Client)
	return func() (any, string, error) {
		id := d.Id()

		// read the latest data into d
		err := resourceCloudscaleVolumeSnapshotRead(d, meta)
		if err != nil {
			return nil, "", err
		}
		// get the instance
		snap, err := client.VolumeSnapshots.Get(context.Background(), id)
		if err != nil {
			return nil, "", fmt.Errorf("error retrieving volume snapshot (%s) (refresh) %s", id, err)
		}

		attr, ok := d.GetOk(attribute)
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
	m["size_gb"] = snap.SizeGB
	m["status"] = snap.Status
	m["tags"] = snap.Tags
	return m
}

func readVolumeSnapshot(rId GenericResourceIdentifier, meta any) (*cloudscale.VolumeSnapshot, error) {
	client := meta.(*cloudscale.Client)
	return client.VolumeSnapshots.Get(context.Background(), rId.Id)
}

func updateVolumeSnapshot(rId GenericResourceIdentifier, meta any, updateRequest *cloudscale.VolumeSnapshotUpdateRequest) error {
	client := meta.(*cloudscale.Client)
	return client.VolumeSnapshots.Update(context.Background(), rId.Id, updateRequest)
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

func deleteVolumeSnapshot(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	sourceVolumeUUID := d.Get("source_volume_uuid").(string)
	// Same API constraint as create: concurrent delete + create (or delete + delete)
	// on the same source volume are rejected. Lock so that the delete and its
	// background cleanup wait complete before any subsequent operation on this volume.
	globalMu.Lock(snapshotLockKey(sourceVolumeUUID))
	defer globalMu.Unlock(snapshotLockKey(sourceVolumeUUID))

	if err := client.VolumeSnapshots.Delete(context.Background(), id); err != nil {
		return err
	}
	// Unlike most cloudscale resources that disappear immediately after DELETE,
	// volume snapshots go through a background cleanup period. During this time
	// the snapshot still exists with status "deleting". We wait for it to be gone.
	return waitForVolumeSnapshotDeleted(id, meta, d.Timeout(schema.TimeoutDelete))
}

func waitForVolumeSnapshotDeleted(id string, meta any, remaining time.Duration) error {
	client := meta.(*cloudscale.Client)
	err := waitForDeleted(remaining, func() (exists bool, err error) {
		snapshot, err := client.VolumeSnapshots.Get(context.Background(), id)
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
