---
page_title: "cloudscale.ch: cloudscale_load_balancer"
---

# cloudscale\_load\_balancer

Provides access to cloudscale.ch load balancers that are not managed by terraform.

## Example Usage

```hcl
data "cloudscale_load_balancer" "lb1" {
  name = "web-lb1"
}
```

## Argument Reference

The following arguments can be used to look up a load balancer:

* `id` - (Optional) The UUID of the load balancer.
* `zone_slug` - (Optional) The slug of the zone in which the load balancer exists. Options include `lpg1` and `rma1`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current resource.
* `flavor_slug` - The slug (name) of the flavor to use for the new load balancer. Possible values can be found in our [API documentation](https://www.cloudscale.ch/en/api/v1#load-balancer-flavors).
* `name` - Name of the new load balancer.
* `status` - The current status of the load balancer.
* `vip_addresses` - A list of VIP address objects. This attributes needs to be specified if the load balancer should be assigned a VIP address in a subnet on a private network. If the  VIP address should be created on the public network, this attribute should be omitted. Each VIP address object has the following attributes:
    * `version` - The IP version, either `4` or `6`.
    * `subnet_href` - The cloudscale.ch API URL of the subnet the VIP address is part of.
    * `subnet_uuid` - The UUID of the subnet this VIP address should be part of.
    * `subnet_cidr` - The cidr of the subnet the VIP address is part of.
    * `address` - An VIP address that has been assigned to this load balancer.