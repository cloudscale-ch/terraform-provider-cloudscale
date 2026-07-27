---
page_title: "cloudscale.ch: cloudscale_router"
---

# cloudscale\_router

Provides a cloudscale.ch router resource. This can be used to create, modify, import, and delete routers.

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

The following arguments are supported when creating/changing routers:

* `name` - (Required) Name of the router.
* `zone_slug` - (Required) The slug of the zone in which the new router will be created. Options include `lpg1` and `rma1`.
* `internet_gateway` - (Optional) If set to true the router acts as an internet gateway.
* `tags` - (Optional) Tags allow you to assign custom metadata to resources:
  ```hcl
  tags = {
    foo = "bar"
  }
  ```
  Tags are always strings (both keys and values).


## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current router.
* `tags` - The tags assigned to this router.
* `status` - The current status of the router.
* `internet_gateway` - If set to true the router acts as an internet gateway.
* `internet_gateway_addresses` - A list of [address objects](#address-object) describing the addresses assigned to the router on the public network.
* `interfaces` - A list of interface objects describing the router's interfaces, including their assigned IP addresses:
  * `uuid` - The UUID of this interface.
  * `network_uuid` - The UUID of this interface's network.
  * `network_name` - The name of that network.
  * `network_href` - The cloudscale.ch API URL of that network.
  * `addresses` - A list of [address objects](#address-object) describing the addresses assigned to the router on this network.
  * `type` - Whether this is a `public` or `private` interface.
  * `mac_address` - The MAC address of this interface.

### Address object

Both `internet_gateway_addresses` and the `addresses` of each entry in `interfaces` are
address objects with the following attributes:

* `address` - The IP address assigned to this router.
* `subnet_uuid` - The UUID of the subnet this address belongs to.
* `subnet_cidr` - The CIDR notation of that subnet.
* `subnet_href` - The cloudscale.ch API URL of that subnet.
* `version` - The IP version of this address.
* `reverse_ptr` - The reverse pointer of this address.

## Import

Routers can be imported using the router's UUID:

```
terraform import cloudscale_router.router 48151623-42aa-aaaa-bbbb-caffeeeeeeee
```
