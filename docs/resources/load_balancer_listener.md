---
page_title: "cloudscale.ch: cloudscale_load_balancer_listener"
---

# cloudscale\_load\_balancer\_listener

Provides a cloudscale.ch load balancer listener resource. This can be used to create, modify, import, and delete load balancer listeners. 

## Example Usage

```hcl
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

# Create a new load balancer listener
resource "cloudscale_load_balancer_listener" "lb1-listener" {
  name          = "web-lb1-listener"
  pool_uuid     = cloudscale_load_balancer_pool.lb1-pool.id
  protocol      = "tcp"
  protocol_port = 80
}
```

## Argument Reference

The following arguments are supported when creating new load balancer listener:

* `name` - (Required) Name of the new load balancer listener.
* `pool_uuid` - (Required) The pool of the listener.
* `protocol` - (Required) The protocol used for receiving traffic. Options include `"tcp"`.
* `protocol_port` - (Required) The port on which traffic is received.
* `timeout_client_data_ms` - (Optional) Client inactivity timeout in milliseconds.
* `timeout_member_connect_ms` - (Optional) Pool member connection timeout in milliseconds.
* `timeout_member_data_ms` - (Optional) Pool member inactivity timeout in milliseconds.
* `allowed_cidrs` - (Optional) Restrict the allowed source IPs for this listener. `[]` means that any source IP is allowed. If the list is non-empty, traffic from source IPs not included is denied.
* `tags` - (Optional) Tags allow you to assign custom metadata to resources:
  ```
  tags = {
    foo = "bar"
  }
  ```
  Tags are always strings (both keys and values).

The following arguments are supported when updating load balancer listener:

* `name` - New name of the load balancer listener.
* `protocol` - The protocol used for receiving traffic. Options include `"tcp"`.
* `protocol_port` - The port on which traffic is received.
* `timeout_client_data_ms` - Client inactivity timeout in milliseconds.
* `timeout_member_connect_ms` - Pool member connection timeout in milliseconds.
* `timeout_member_data_ms` - Pool member inactivity timeout in milliseconds.
* `allowed_cidrs` - Restrict the allowed source IPs for this listener. `[]` means that any source IP is allowed. If the list is non-empty, traffic from source IPs not included is denied.
* `tags` - Change tags (see documentation above)

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The UUID of this load balancer listner.
* `href` - The cloudscale.ch API URL of the current resource.
* `pool_name` - The load balancer pool name of the listener.
* `pool_href` - The cloudscale.ch API URL of the listener's load balancer pool.


## Import

Load balancer listener can be imported using the load balancer listener's UUID:

```
terraform import cloudscale_load_balancer_listener.lb1-listener 48151623-42aa-aaaa-bbbb-caffeeeeeeee
```
