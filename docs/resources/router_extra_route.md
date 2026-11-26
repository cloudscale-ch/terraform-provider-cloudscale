---
page_title: "cloudscale.ch: cloudscale_router_extra_route"
---

# cloudscale\_router\_extra\_route

Provides a cloudscale.ch router extra route resource. This can be used to create, import, and delete
extra routes on a router. Extra routes cannot be changed after creation; any change replaces the route.

## Example Usage

```hcl
resource "cloudscale_network" "test" {
  name                    = "test"
  zone_slug               = "rma1"
  auto_create_ipv4_subnet = false
}

resource "cloudscale_subnet" "test" {
  cidr         = "10.11.12.0/24"
  network_uuid = cloudscale_network.test.id
}

resource "cloudscale_router" "gw" {
  name             = "gw"
  zone_slug        = "rma1"
  internet_gateway = true
}

resource "cloudscale_interface" "gw" {
  router_uuid  = cloudscale_router.gw.id
  network_uuid = cloudscale_network.test.id

  addresses {
    subnet_uuid = cloudscale_subnet.test.id
    address     = "10.11.12.10"
  }
}

resource "cloudscale_router_extra_route" "peering_route" {
  router_uuid = cloudscale_router.gw.id
  destination = "198.51.100.0/24"
  next_hop    = "10.11.12.1"
}
```

## Argument Reference

The following arguments are supported when creating extra routes:

* `router_uuid` - (Required) The router this extra route is attached to.
* `destination` - (Required) The destination prefix in CIDR notation.
* `next_hop` - (Required) The IP address of the next hop.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of this extra route.

## Import

Extra routes can be imported using a combination of the router's UUID and the extra route's
UUID, separated by a dot:

```
terraform import cloudscale_router_extra_route.peering_route 48151623-42aa-aaaa-bbbb-caffeeeeeeee.51518841-caff-eeee-bbbb-424242424242
```
