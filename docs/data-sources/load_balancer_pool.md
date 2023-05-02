---
page_title: "cloudscale.ch: cloudscale_load_balancer_pool"
---

# cloudscale\_load\_balancer\_pool

Provides access to cloudscale.ch load balancer pools that are not managed by terraform.

## Example Usage

```hcl
data "cloudscale_load_balancer_pool" "pool" {
  name = "web-lb1-pool"
}
```

## Argument Reference

The following arguments can be used to look up a load balancer pool:

* `id` - (Optional) The UUID of the load balancer pool.
* `name` - (Optional) Name of the load balancer pool.
* `load_balancer_uuid` - (Optional) The load balancer of the pool.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current resource.
* `load_balancer_name` - The load balancer name of the pool.
* `load_balancer_href` - The cloudscale.ch API URL of the pool's load balancer.
* `algorithm` - The algorithm according to which the incoming traffic is distributed between the pool members. Options include `"round_robin"`, `"least_connections"` and `"source_ip"`.
* `protocol` - The protocol used for traffic between the load balancer and the pool members. Options include: `"tcp"`, `"proxy"` and `"proxyv2"`.
