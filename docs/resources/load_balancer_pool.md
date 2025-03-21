---
page_title: "cloudscale.ch: cloudscale_load_balancer_pool"
---

# cloudscale\_load\_balancer\_pool

Provides a cloudscale.ch load balancer pool resource. This can be used to create, modify, import, and delete load balancer pools. 

## Example Usage

```hcl
# Create a new load balancer
resource "cloudscale_load_balancer" "lb1" {
  name        = "web-lb1"
  flavor_slug = "lb-standard"
  zone_slug   = "lpg1"
}

# Create a new load balancer pool
resource "cloudscale_load_balancer_pool" "lb1-pool" {
  name               = "web-lb1-pool"
  algorithm          = "round_robin"
  protocol           = "tcp"
  load_balancer_uuid = cloudscale_load_balancer.lb1.id
}
```

## Argument Reference

The following arguments are supported when creating new load balancer pool:

* `name` - (Required) Name of the new load balancer pool.
* `algorithm` - (Required) The algorithm according to which the incoming traffic is distributed between the pool members. Options include `"round_robin"`, `"least_connections"` and `"source_ip"`.
* `protocol` - (Required) The protocol used for traffic between the load balancer and the pool members. Options include: `"tcp"`, `"proxy"` and `"proxyv2"`.
* `load_balancer_uuid` - (Required) The load balancer of the pool.
* `tags` - (Optional) Tags allow you to assign custom metadata to resources:
  ```hcl
  tags = {
    foo = "bar"
  }
  ```
  Tags are always strings (both keys and values).

The following arguments are supported when updating load balancer pools:

* `name` - New name of the load balancer pool.
* `tags` - Change tags (see documentation above)

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The UUID of this load balancer pool.
* `href` - The cloudscale.ch API URL of the current resource.
* `load_balancer_name` - The load balancer name of the pool.
* `load_balancer_href` - The cloudscale.ch API URL of the pool's load balancer.


## Import

Load balancer pools can be imported using the load balancer pool's UUID:

```
terraform import cloudscale_load_balancer_pool.lb1-pool 48151623-42aa-aaaa-bbbb-caffeeeeeeee
```
