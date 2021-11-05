---
page_title: "cloudscale.ch: cloudscale_subnet"
---

# cloudscale\_subnet

Provides access to cloudscale.ch subnets that are not managed by terraform.

## Example Usage

```hcl
data "cloudscale_subnet" "privnet-subnet" {
  cidr         	  = "10.11.12.0/24"
}

# Create a server with fixed IP address
resource "cloudscale_server" "fixed" {
  name            = "fix"
  zone_slug       = "lpg1"
  flavor_slug     = "flex-2"
  image_slug      = "debian-9"
  interfaces      {
    type          = "public"
  }
  interfaces      {
    type          = "private"
    addresses {
      subnet_uuid = "${data.cloudscale_subnet.privnet-subnet.id}"     
      address     = "10.11.12.13"
    }
  }
  volume_size_gb  = 10
  ssh_keys        = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY=", "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}
```

## Argument Reference

The following arguments can be used to look up a subnet:

* `id` - (Optional) The UUID of the subnet.
* `cidr` - (Optional) The address range in CIDR notation.
* `network_uuid` - (Optional) The network UUID of the subnet.
* `network_name` - (Optional) The network name of the subnet.
* `gateway_address` - (Optional) The gateway address of the subnet.


## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current subnet.
* `dns_servers` - A list of DNS resolver IP addresses, that act as DNS servers.
* `network_href` - The cloudscale.ch API URL of the subnet's network.
