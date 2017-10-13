---
layout: "cloudscale"
page_title: "cloudscale.ch: cloudscale_server"
sidebar_current: "docs-cloudscale-resource-server"
description: |-
  Provides a cloudscale.ch Server resource. This can be used to create, modify, and delete servers.
---

# cloudscale\_server

Provides a cloudscale.ch Server resource. This can be used to create, modify, and delete servers. 

## Example Usage

```hcl
# Create a new Server
resource "cloudscale_server" "web-worker01" {
  name                = "web-worker01"
  flavor_slug         = "flex-4"
  image_slug          = "debian-9"
  volume_size_gb      = 10
  bulk_volume_size_gb = 200
  ssh_keys            = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIL2jzgla23DfRVLQr3KT20QQYovqCCN3clHrjm2ZuQFW user@example.com"]
}
```

## Argument Reference

The following arguments are supported when creating new servers:

* `name` - (Required) Name of the new server. The name has to be a valid host name or a fully qualified domain name (FQDN).
* `flavor_slug` - (Required) The slug (name) of the flavor to use for the new server. Possible values can be found in our [API documentation](https://www.cloudscale.ch/en/api/v1#flavors).
* `image_slug` - (Required) The slug (name) of the image to use for the new server. Possible values can be found in our [API documentation](https://www.cloudscale.ch/en/api/v1#images).
* `ssh_keys` - (Required) A list of SSH public keys. Use the full content of your \*.pub file here.
* `volume_size_gb` - (Optional) The size in GB of the SSD root volume of the new server. If this parameter is not specified, the value will be set to 10. Valid values are either 10 or multiples of 50.
* `bulk_volume_size_gb` - (Optional) The size in GB of the bulk storage volume of the new server. If this parameter is not specified, no bulk storage volume will be attached to the server. Valid values are multiples of 100.
* `use_public_network` - (Optional) Attach/detach the public network interface to/from the new server. Can be `true` (default) or `false`.
* `use_private_network` - (Optional) Attach/detach the private network interface to/from the new server. Can be `true` or `false` (default).
* `use_ipv6` - (Optional) Enable/disable IPv6 on the public network interface of the new server. Can be `true` (default) or `false`.
* `anti_affinity_uuid` - (Optional) Pass the UUID of another server to either create a new anti-affinity group with that server or add the new server to the same (existing) group as the other server.
* `user_data` - (Optional) User data (custom cloud-config settings) to use for the new server. Needs to be valid YAML. A default configuration will be used if this parameter is not specified or set to null. Use only if you are an advanced user with knowledge of cloud-config and cloud-init.
* `status` - (Optional) The desired state of a server. Can be `running` (default) or `stopped`.

The following arguments are supported when updating servers:

* `status` - (Optional) The desired state of a server. Can be `running` (default) or `stopped`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The UUID of this server.
* `href` - The cloudscale.ch API URL of the current resource.
* `ssh_fingerprints` - A list of SSH host key fingerprints (strings) of this server.
* `ssh_host_keys` - A list of SSH host keys (strings) of this server.
* `anti_affinity_with` - A list of server UUIDs that belong to the same anti-affinity group as this server.
* `volumes` - A list of volume objects attached to this server. Each volume object has three attributes:
    * `device_path` - The path (string) to the volume on your server (e.g. `/dev/vda`)
    * `size_gb` - The size (int) of the volume in GB. Typically matches `volume_size_gb` or `bulk_volume_size_gb`.
    * `type` - A string. Either `ssd` or `bulk`.
* `interfaces` - A list of interface objects attached to this server. Each interface object has two attributes:
    * `type` - Either `public` or `private`. Public interfaces are connected to the Internet, while private interfaces are not.
    * `addresses` - A list of address objects:
        * `address` - An IPv4 or IPv6 address that has been assigned to this server.
        * `gateway` - An IPv4 or IPv6 address that represents the default gateway for this interface.
        * `prefix_length` - The prefix length for this IP address, typically 24 for IPv4 and 128 for IPv6.
        * `reverse_ptr` - The PTR record (reverse DNS pointer) for this IP address. If you use an FQDN as your server name it will automatically be used here.
        * `version` - The IP version, either `4` or `6`.
