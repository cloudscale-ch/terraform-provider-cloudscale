---
page_title: "cloudscale.ch: cloudscale_network"
---

# cloudscale\_network

Provides access to cloudscale.ch private networks that are not managed by terraform.

## Example Usage

```hcl
data "cloudscale_network" "privnet" {
  name         = "privnet"
}

# Add a server with two interfaces:
#  - one attached to the public network
#  - one attached to the private network "privnet"
resource "cloudscale_server" "gw" {
  name                = "gateway"
  zone_slug           = "lpg1"
  flavor_slug         = "flex-4"
  image_slug          = "debian-9"
  volume_size_gb      = 20
  ssh_keys            = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIL2jzgla23DfRVLQr3KT20QQYovqCCN3clHrjm2ZuQFW user@example.com"]
  interfaces {
    type              = "public"
  }
  interfaces {
    type              = "private"
    network_uuid      = data.cloudscale_network.privnet.id
  }
}
```

## Argument Reference

The following arguments can be used to look up a network:

* `id` - (Optional) The UUID of a network.
* `name` - (Optional) The name of a network.
* `zone_slug` - (Optional) The zone slug of a network. Options include `lpg1` and `rma1`.


## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current network.
* `mtu` - The MTU size for the network.
* `subnets` -  A list of subnet objects that are used in this network. Each subnet object has the following attributes:
  * `cidr` - The CIDR notation of the subnet.
  * `href` - The cloudscale.ch API URL of this subnet.
  * `uuid` - The UUID of this subnet.
