---
page_title: "cloudscale.ch: cloudscale_network"
---

# cloudscale\_network

Provides a cloudscale.ch private network resource. This can be used to create, modify, import, and delete networks.

## Example Usage

```hcl
# Create a new private network
resource "cloudscale_network" "privnet" {
  name         = "privnet"
  zone_slug    = "lpg1"
  mtu          = "9000"
}

# Add a server with two interfaces:
#  - one attached to the public network
#  - one attached to the private network "privnet"
resource "cloudscale_server" "gw" {
  name                = "gateway"
  zone_slug           = "lpg1"
  flavor_slug         = "flex-8-4"
  image_slug          = "debian-11"
  volume_size_gb      = 20
  ssh_keys            = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIL2jzgla23DfRVLQr3KT20QQYovqCCN3clHrjm2ZuQFW user@example.com"]
  interfaces {
    type              = "public"
  }
  interfaces {
    type              = "private"
    network_uuid      = cloudscale_network.privnet.id
  }
}
```

## Argument Reference

The following arguments are supported when creating/changing networks:

* `name` - (Required) Name of the network.
* `zone_slug` - (Optional) The slug of the zone in which the new network will be created. Options include `lpg1` and `rma1`.
* `mtu` - (Optional) You can specify the MTU size for the network, defaults to 9000.
* `auto_create_ipv4_subnet` - (Optional) Automatically create an IPv4 Subnet on the network. Can be `true` (default) or `false`.
* `tags` - (Optional) Tags allow you to assign custom metadata to resources:
  ```
  tags = {
    foo = "bar"
  }
  ```
  Tags are always strings (both keys and values).


## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current network.
* `subnets` -  A list of subnet objects that are used in this network. Each subnet object has the following attributes:
  * `cidr` - The CIDR notation of the subnet.
  * `href` - The cloudscale.ch API URL of this subnet.
  * `uuid` - The UUID of this subnet.


## Import

Networks can be imported using the network's UUID:

```
terraform import cloudscale_network.network 48151623-42aa-aaaa-bbbb-caffeeeeeeee
```
