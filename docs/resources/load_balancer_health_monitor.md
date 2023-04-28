---
page_title: "cloudscale.ch: cloudscale_load_balancer_health_monitor"
---

# cloudscale\_load\_balancer\_health_monitor

Provides a cloudscale.ch load balancer health monitor resource. This can be used to create, modify, import, and delete load balancer health monitors. 

## Example Usage

```hcl
# all resources from the cloudscale_load_balancer_pool_member example..
# ..followed by:

# Create a new load balancer health monitor
resource "cloudscale_load_balancer_health_monitor" "lb1-health-monitor" {
  pool_uuid        = cloudscale_load_balancer_pool.lb1-pool.id
  type             = "http"
  http_url_path    = "/"
  http_version     = "1.1"
  http_host        = "www.cloudscale.ch"
}
```

## Argument Reference

The following arguments are supported when creating new load balancer health monitor:

* `name` - (Required) Name of the new load balancer health monitor.
* `pool_uuid` - (Required) The pool of the health monitor.

* `type` - (Required) The type of the health monitor. Options include: `"ping"`, `"tcp"`, `"http"`, `"https"` and `"tls-hello"`.
* `delay_s` - (Optional) The delay between two successive checks in seconds.  Default is `2`.
* `timeout_s` - (Optional) The maximum time allowed for an individual check in seconds.  Default is `1`.
* `up_threshold` - (Optional) The number of checks that need to be successful before the `monitor_status` of a pool member changes to `"up"`. Default is `2`.
* `down_threshold` - (Optional) The number of checks that need to fail before the `monitor_status` of a pool member changes to `"down"`. Default is `3`.
* `http_expected_codes` - (Optional) The HTTP status codes allowed for a check to be considered successful. Can either be a list of status codes, for example `["200", "202"]`, or a list containing a single range, for example `["200-204"]`. Default is `["200"]`.
* `http_method` - (Optional) The HTTP method used for the check. Options include `"CONNECT"`, `"DELETE"`, `"GET"`, `"HEAD"`, `"OPTIONS"`, `"PATCH"`, `"POST"`, `"PUT"` and `"TRACE"`. Default is `"GET"`.
* `http_url_path` - (Optional) The URL used for the check. Default is `"/"`.
* `http_version` - (Optional) The HTTP version used for the check. Options include `"1.0"` and `"1.1"`. Default is `"1.1"`.
* `http_host` - (Optional) The server name in the HTTP Host: header used for the check. Requires version to be set to `"1.1"`.

* `tags` - (Optional) Tags allow you to assign custom metadata to resources:
  ```
  tags = {
    foo = "bar"
  }
  ```
  Tags are always strings (both keys and values).

The following arguments are supported when updating load balancer health monitor:

* `delay_s` - The delay between two successive checks in seconds.  Default is `2`.
* `timeout_s` - The maximum time allowed for an individual check in seconds.  Default is `1`.
* `up_threshold` - The number of checks that need to be successful before the `monitor_status` of a pool member changes to `"up"`. Default is `2`.
* `down_threshold` - The number of checks that need to fail before the `monitor_status` of a pool member changes to `"down"`. Default is `3`.
* `http_expected_codes` - The HTTP status codes allowed for a check to be considered successful. Can either be a list of status codes, for example `["200", "202"]`, or a list containing a single range, for example `["200-204"]`. Default is `["200"]`.
* `http_method` - The HTTP method used for the check. Options include `"CONNECT"`, `"DELETE"`, `"GET"`, `"HEAD"`, `"OPTIONS"`, `"PATCH"`, `"POST"`, `"PUT"` and `"TRACE"`. Default is `"GET"`.
* `http_url_path` - The URL used for the check. Default is `"/"`.
* `http_host` - The server name in the HTTP Host: header used for the check. Requires version to be set to `"1.1"`.
* `tags` - Change tags (see documentation above)

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The UUID of this load balancer health monitor.
* `href` - The cloudscale.ch API URL of the current resource.
* `pool_name` - The load balancer pool name of the health monitor.
* `pool_href` - The cloudscale.ch API URL of the health monitor's load balancer pool.


## Import

Load balancer health monitor can be imported using the load balancer health monitor's UUID:

```
terraform import cloudscale_load_balancer_health_monitor.lb1-health-monitor 48151623-42aa-aaaa-bbbb-caffeeeeeeee
```
