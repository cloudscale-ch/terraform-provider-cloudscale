---
page_title: "cloudscale.ch: cloudscale_volume"
---

# cloudscale\_volume

Provides access to cloudscale.ch volumes (block storage) that are not managed by terraform.

## Example Usage

```hcl
data "cloudscale_volume" "web-worker01-volume" {
  name         = "web-worker-data"
}
```

## Argument Reference

The following arguments can be used to look up a network:

* `id` - (Optional) The UUID of a volume.
* `name` - (Optional) The Name of the volume.
* `zone_slug` - (Optional) The slug of the zone in which the new volume will be created. Options include `lpg1` and `rma1`.
* `type` - (Optional) For SSD/NVMe volumes "ssd" (default); or "bulk" for our HDD cluster with NVMe caching.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current resource.
* `size_gb` - The volume size in GB. Valid values are multiples of 1 for type "ssd" and multiples of 100 for type "bulk".
* `server_uuids` - (Optional) A list of server UUIDs.
