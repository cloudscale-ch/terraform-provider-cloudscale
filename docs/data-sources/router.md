---
page_title: "cloudscale.ch: cloudscale_router"
---

# cloudscale\_router

Provides access to cloudscale.ch private routers that are not managed by terraform.

## Example Usage

```hcl
data "cloudscale_router" "gw" {
  name         = "gw"
}
```

## Argument Reference

The following arguments can be used to look up a router:

* `id` - (Optional) The UUID of a router.
* `name` - (Optional) The name of a router.
* `zone_slug` - (Optional) The zone slug of a router. Options include `lpg1` and `rma1`.
* `tags` - (Optional) Filter by tags; the resource must have at least the specified key-value pairs (subset match).

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current router.
* `status` - The current status of the router.
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