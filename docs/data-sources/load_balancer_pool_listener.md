---
page_title: "cloudscale.ch: cloudscale_load_balancer_listener"
---

# cloudscale\_load\_balancer\_listener

Provides access to cloudscale.ch load balancer listeners that are not managed by terraform.

## Example Usage

```hcl
data "cloudscale_load_balancer_listener" "listener" {
  name = "web-lb1-listener"
}
```

## Argument Reference

The following arguments can be used to look up a load balancer listener:

* `id` - (Optional) The UUID of the load balancer listener.
* `name` = (Optional) Name of the load balancer listener.
* `pool_uuid` - (Optional) The UUID of the pool this listener belongs to.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current resource.
* `pool_href` = The cloudscale.ch API URL of the listener's load balancer pool.
* `pool_name` = The load balancer pool name of the listener.
* `protocol` = The protocol used for receiving traffic. Options include `"tcp"`.
* `protocol_port` = The port on which traffic is received.
* `timeout_client_data_ms` = Client inactivity timeout in milliseconds.
* `timeout_member_connect_ms` = Pool member connection timeout in milliseconds.
* `timeout_member_data_ms` = Pool member inactivity timeout in milliseconds.
* `allowed_cidrs` =  Restrict the allowed source IPs for this listener. `[]` means that any source IP is allowed. If the list is non-empty, traffic from source IPs not included is denied.
