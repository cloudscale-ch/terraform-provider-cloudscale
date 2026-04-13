---
page_title: "cloudscale.ch: cloudscale_volume_snapshot"
---

# cloudscale\_volume\_snapshot

Provides a cloudscale.ch volume snapshot resource. This can be used to create, modify, import, and delete volume snapshots.

## Example Usage

```hcl
# Create a volume to snapshot
resource "cloudscale_volume" "data" {
  name    = "data-vol"
  size_gb = 100
  type    = "ssd"
}

# Create a snapshot of the volume
resource "cloudscale_volume_snapshot" "data-snap" {
  name               = "data-snap-1"
  source_volume_uuid = cloudscale_volume.data.id
}
```

## Argument Reference

The following arguments are supported when creating a new volume snapshot:

* `name` - (Required) Name of the new volume snapshot.
* `source_volume_uuid` - (Required, Forces new resource) The UUID of the volume to snapshot. This field cannot be changed after creation.
* `tags` - (Optional) Tags allow you to assign custom metadata to resources:
  ```hcl
  tags = {
    foo = "bar"
  }
  ```
  Tags are always strings (both keys and values).

The following arguments are supported when updating volume snapshots:

* `name` - New name of the volume snapshot.
* `tags` - Change tags (see documentation above).

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The UUID of this volume snapshot.
* `href` - The cloudscale.ch API URL of the current resource.
* `size_gb` - The size of the snapshot in GB.
* `status` - The current status of the volume snapshot (e.g. `available`).

## Import

Volume snapshots can be imported using the snapshot's UUID:

```
terraform import cloudscale_volume_snapshot.snap 48151623-42aa-aaaa-bbbb-caffeeeeeeee
```