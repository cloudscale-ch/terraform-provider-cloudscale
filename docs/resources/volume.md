---
page_title: "cloudscale.ch: cloudscale_volume"
---

# cloudscale\_volume

Provides a cloudscale.ch volume (block storage) resource. This can be used to create, modify, import, and delete volumes.

## Example Usage

```hcl
# Create a new Server
resource "cloudscale_server" "web-worker01" {
  name        = "web-worker01"
  flavor_slug = "flex-8-4"
  image_slug  = "debian-13"
  ssh_keys    = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIL2jzgla23DfRVLQr3KT20QQYovqCCN3clHrjm2ZuQFW user@example.com"]
}

# Add a volume to web-worker01
resource "cloudscale_volume" "web-worker01-volume" {
  name         = "web-worker-data"
  size_gb      = 100
  type         = "ssd"
  server_uuids = [cloudscale_server.web-worker01.id]
}
```

### Create a Volume from a Snapshot

```hcl
# Source volume
resource "cloudscale_volume" "data" {
  name    = "data-volume"
  size_gb = 50
  type    = "ssd"
}

# Snapshot of the source volume
resource "cloudscale_volume_snapshot" "data-snap" {
  name               = "data-snap"
  source_volume_uuid = cloudscale_volume.data.id
}

# Create a new volume from the snapshot, resized to 200 GB
resource "cloudscale_volume" "restored" {
  name                 = "restored-data"
  volume_snapshot_uuid = cloudscale_volume_snapshot.data-snap.id
  size_gb              = 200
}
```

## Argument Reference

The following arguments are supported when creating/changing volumes:

* `name` - (Required) Name of the new volume.
* `size_gb` - (Required, if `volume_snapshot_uuid` not set) The volume size in GB. Valid values are multiples of 1 for type "ssd" and multiples of 100 for type "bulk". When creating from a snapshot, this is optional and can be used to resize the volume after creation.
* `volume_snapshot_uuid` - (Optional, conflicts with `type`, `zone_slug`) The UUID of a volume snapshot to create the volume from. The new volume will contain the data stored in the snapshot. When set, `type` and `zone_slug` are inherited from the snapshot and cannot be specified.
* `zone_slug` - (Optional, conflicts with `volume_snapshot_uuid`) The slug of the zone in which the new volume will be created. Options include `lpg1` and `rma1`.
* `type` - (Optional, conflicts with `volume_snapshot_uuid`) For SSD/NVMe volumes specify "ssd" (default) or use "bulk" for our HDD cluster with NVMe caching. This is the only attribute that cannot be altered.
* `server_uuids` - (Optional) A list of server UUIDs. Default to an empty list. Currently a volume can only be attached to one server UUID.
* `tags` - (Optional) Tags allow you to assign custom metadata to resources:
  ```hcl
  tags = {
    foo = "bar"
  }
  ```
  Tags are always strings (both keys and values).

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current resource.


## Import

Volumes can be imported using the volume's UUID:

```
terraform import cloudscale_volume.volume 48151623-42aa-aaaa-bbbb-caffeeeeeeee
```
