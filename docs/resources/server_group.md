---
page_title: "cloudscale.ch: cloudscale_server_group"
---

# cloudscale\_server\_group

Provides a cloudscale.ch server group resource. This can be used to create, modify, import, and delete server groups.

## Example Usage

```hcl
# Add a server group with anti affinity
resource "cloudscale_server_group" "web-worker-group" {
  name = "web-worker-group"
  type = "anti-affinity"
}

# Create three new servers in that group
resource "cloudscale_server" "web-worker01" {
  count            = 3
  name             = "web-worker${count.index}"
  flavor_slug      = "flex-8-4"
  image_slug       = "debian-11"
  server_group_ids = [cloudscale_server_group.web-worker-group.id]
  ssh_keys         = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIL2jzgla23DfRVLQr3KT20QQYovqCCN3clHrjm2ZuQFW user@example.com"]
}
```

## Argument Reference

The following arguments are supported when creating server groups:

* `name` - (Required) Name of the new server group.
* `type` - (Required) The type of the server group can currently only be `"anti-affinity"`.
* `zone_slug` - (Optional) The slug of the zone in which the new server group will be created. Options include `lpg1` and `rma1`.
* `tags` - (Optional) Tags allow you to assign custom metadata to resources:
  ```
  tags = {
    foo = "bar"
  }
  ```
  Tags are always strings (both keys and values).

The following arguments are supported when updating server groups:

* `name` -  The new name of the server group.
* `tags` - (Optional) Change tags (see documentation above)

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current resource.


## Import

Server groups can be imported using the server group's UUID:

```
terraform import cloudscale_server_group.server_group 48151623-42aa-aaaa-bbbb-caffeeeeeeee
```
