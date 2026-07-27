---
page_title: "cloudscale.ch: cloudscale_interface"
---

# cloudscale\_interface

Provides a cloudscale.ch interface resource. This currently attaches a router to a private network so
the router can route traffic for that network. It can be used to create, import, and delete
interfaces. Interfaces cannot be changed after creation; any change replaces the interface.

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
```

## Argument Reference

The following arguments are supported when creating interfaces:

* `router_uuid` - (Required) The router this interface is attached to.
* `network_uuid` - (Required) The network this interface connects to.
* `addresses` - (Required) A list of address blocks that assign IP addresses to the interface. If
  omitted, an address is assigned automatically. Each block supports:
  * `subnet_uuid` - (Required) The subnet the address is taken from.
  * `address` - (Required) A specific IP address to assign.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `network_name` - The name of the network this interface connects to.
* `network_href` - The cloudscale.ch API URL of this network.
* `type` - Whether this is a `public` or `private` interface.
* `mac_address` - The MAC address of this interface.
* `addresses` - In addition to the arguments above, each address block exports:
  * `address` - The IP address assigned to the interface (also exported when assigned
    automatically).
  * `subnet_href` - The cloudscale.ch API URL of the subnet the address belongs to.
  * `subnet_cidr` - The CIDR notation of that subnet.
  * `version` - The IP version of the address.
  * `reverse_ptr` - The reverse pointer of the address.

## Import

Interfaces can be imported using a combination of the router's UUID and the interface's
UUID, separated by a dot:

```
terraform import cloudscale_interface.gw 48151623-42aa-aaaa-bbbb-caffeeeeeeee.51518841-caff-eeee-bbbb-424242424242
```
