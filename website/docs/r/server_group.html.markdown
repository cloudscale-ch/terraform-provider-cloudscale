---
layout: "cloudscale"
page_title: "cloudscale.ch: cloudscale_server_group"
sidebar_current: "docs-cloudscale-resource-server_group"
description: |-
  Provides a cloudscale.ch server group resource. This can be used to create, modify, and delete server groups.
---

# cloudscale\_server\_group

Provides a cloudscale.ch server group resource. This can be used to create, and delete server groups.

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
  name             = "web-worker01"
  flavor_slug      = "flex-4"
  image_slug       = "debian-9"
  server_group_ids = ["${cloudscale_server_group.web-worker-group.id}"]
  ssh_keys         = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIL2jzgla23DfRVLQr3KT20QQYovqCCN3clHrjm2ZuQFW user@example.com"]
}
```

## Argument Reference

The following arguments are supported when creating server groups:

* `name` - (Required) Name of the new server group.
* `type` - (Required) The type of the server group can currently only be `"anti-affinity"`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current resource.
