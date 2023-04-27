---
page_title: "cloudscale.ch: cloudscale_load_balancer_pool_member"
---

# cloudscale\_load\_balancer\_pool\_member

Provides access to cloudscale.ch load balancer pool members that are not managed by terraform.

## Example Usage

```hcl
data "cloudscale_load_balancer_pool_member" "member" {
  pool_uuid = "4ffda9cc-7ba5-4193-a104-0d377fb84c96" # required!
  name      = "web-lb1-pool-member-2"
}
```

## Argument Reference

The following arguments can be used to look up a load balancer pool member:

* `pool_uuid` - (Required) The UUID of the pool this member belongs to.
* `id` - (Optional) The UUID of the load balancer pool.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current load balancer.
* `enabled` - Pool member will not receive traffic if `false`.
* `pool_name` - The load balancer name of the pool.
* `pool_href` - The cloudscale.ch API URL of the pool's load balancer.
* `protocol_port` - The port to which actual traffic is sent.
* `monitor_port` - The port to which health monitor checks are sent. If not specified, `protocol_port` will be used.
* `address` - The IP address to which traffic is sent.
* `subnet_uuid` - The subnet UUID of the address must be specified here.
* `subnet_cidr` - The CIDR of the member's address subnet.
* `subnet_href` - The cloudscale.ch API URL of the member's address subnet.
* `monitor_status` - The status of the pool's health monitor check for this member. Can be `"up"`, `"down"`, `"changing"`, `"no_monitor"` and `"unknown"`.