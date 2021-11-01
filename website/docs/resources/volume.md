---
page_title: "cloudscale.ch: cloudscale_volume"
---

# cloudscale\_volume

Provides a cloudscale.ch volume (block storage) resource. This can be used to create, modify, and delete volumes.

## Example Usage

```hcl
# Create a new Server
resource "cloudscale_server" "web-worker01" {
  name        = "web-worker01"
  flavor_slug = "flex-4"
  image_slug  = "debian-9"
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

## Argument Reference

The following arguments are supported when creating/changing volumes:

* `name` - (Required) Name of the new volume.
* `size_gb` - (Required) The volume size in GB. Valid values are multiples of 1 for type "ssd" and multiples of 100 for type "bulk".
* `zone_slug` - (Optional) The slug of the zone in which the new volume will be created. Options include `lpg1` and `rma1`.
* `type` - (Optional) For SSD/NVMe volumes specify "ssd" (default) or use "bulk" for our HDD cluster with NVMe caching. This is the only attribute that cannot be altered.
* `server_uuids` - (Optional) A list of server UUIDs. Default to an empty list. Currently a volume can only be attached to one server UUID.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current resource.
