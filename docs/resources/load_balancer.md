---
page_title: "cloudscale.ch: cloudscale_load_balancer"
---

# cloudscale\_load_balancer

Provides a cloudscale.ch load balancer resource. This can be used to create, modify, import, and delete load balancers. 

## Example Usage

```hcl
# Create a new load balancer
resource "cloudscale_load_balancer" "lb1" {
  name        = "web-lb1"
  flavor_slug = "lb-standard"
  zone_slug   = "lpg1"
}
```

## Argument Reference

The following arguments are supported when creating new load balancer:

* `name` - (Required) Name of the new load balancer.
* `flavor_slug` - (Required) The slug (name) of the flavor to use for the new load balancer. Possible values can be found in our [API documentation](https://www.cloudscale.ch/en/api/v1#flavors).
    **Note:** It's currently not possible to update the flavor after the load balancer has been created. It is therfore recommended to use load balancer in conjunction with a Floating IP.
* `zone_slug` - (Required) The slug of the zone in which the new load balancer will be created. Options include `lpg1` and `rma1`.
* `vip_addresses` - (Optional) A list of VIP address objects. This attributes needs to be specified if the load balancer should be assigned a VIP address in a subnet on a private network. If the  VIP address should be created on the public network, this attribute should be omitted. Each VIP address object has the following attributes:
    * `subnet_uuid` - (Optional) The UUID of the subnet this VIP address should be part of.
    * `address` - (Optional) An VIP address that has been assigned to this load balancer.
* `tags` - (Optional) Tags allow you to assign custom metadata to resources:
  ```
  tags = {
    foo = "bar"
  }
  ```
  Tags are always strings (both keys and values).

The following arguments are supported when updating load balancers:

* `name` - New name of the load balancer.
* `tags` - Change tags (see documentation above)

**Note on `vip_addresses`: It might be necessary to manually `terrafrom destroy` a load balancer in order
for Terraform to detect the change correctly. A replacement of the load balancer is required in all cases.**


## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The UUID of this load balancer.
* `href` - The cloudscale.ch API URL of the current resource.
* `status` - The current status of the load balancer.
* `vip_addresses` - A list of VIP address objects.  Each VIP address object has the following attributes:
    * `version` - The IP version, either `4` or `6`.
    * `subnet_cidr` - The cidr of the subnet the VIP address is part of.
    * `subnet_href` - The cloudscale.ch API URL of the subnet the VIP address is part of.

## Import

Load balancer can be imported using the load balancer's UUID:

```
terraform import cloudscale_load_balancer.lb 48151623-42aa-aaaa-bbbb-caffeeeeeeee
```
