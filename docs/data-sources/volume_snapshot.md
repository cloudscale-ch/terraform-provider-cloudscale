---
page_title: "cloudscale.ch: cloudscale_volume_snapshot"
---

# cloudscale\_volume\_snapshot

Provides access to cloudscale.ch volume snapshots that are not managed by terraform.

## Example Usage

```hcl
data "cloudscale_volume_snapshot" "snap" {
  name = "data-snap-1"
}
```

## Argument Reference

The following arguments can be used to look up a volume snapshot:

* `id` - (Optional) The UUID of the volume snapshot.
* `name` - (Optional) The name of the volume snapshot.
* `source_volume_uuid` - (Optional) The UUID of the source volume.
* `tags` - (Optional) Tags allow you to assign custom metadata to resources. Tags are always strings (both keys and values).

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current resource.
* `source_volume_uuid` - The UUID of the source volume.
* `size_gb` - The size of the snapshot in GB.
* `status` - The current status of the volume snapshot.