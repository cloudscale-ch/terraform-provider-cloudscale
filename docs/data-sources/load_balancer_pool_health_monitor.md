---
page_title: "cloudscale.ch: cloudscale_load_balancer_health_monitor"
---

# cloudscale\_load\_balancer\_health\_monitor

Provides access to cloudscale.ch load balancer health monitors that are not managed by terraform.

## Example Usage

```hcl
data "cloudscale_load_balancer_health_monitor" "lb1-health-monitor" {
  id = "d38ed4f8-6a8d-4b3d-a2ff-87e53e4434e1"
}
```

## Argument Reference

The following arguments can be used to look up a load balancer listener:

* `id` - (Optional) The UUID of the load balancer health monitor.
* `pool_uuid` - (Optional) The UUID of the pool this health monitor belongs to.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current resource.
* `pool_href` = The cloudscale.ch API URL of the listener's load balancer pool.
* `pool_name` = The load balancer pool name of the listener.
* `delay_s` - The delay between two successive checks in seconds.
* `timeout_s` - The maximum time allowed for an individual check in seconds.
* `up_threshold` - The number of checks that need to be successful before the `monitor_status` of a pool member changes to `"up"`.
* `down_threshold` - The number of checks that need to fail before the `monitor_status` of a pool member changes to `"down"`.
* `type` - The type of the health monitor. Options include: `"ping"`, `"tcp"`, `"http"`, `"https"` and `"tls-hello"`.
* `http_expected_codes` - The HTTP status codes allowed for a check to be considered successful. Can either be a list of status codes, for example `["200", "202"]`, or a list containing a single range, for example `["200-204"]`.
* `http_method` - The HTTP method used for the check. Options include `"CONNECT"`, `"DELETE"`, `"GET"`, `"HEAD"`, `"OPTIONS"`, `"PATCH"`, `"POST"`, `"PUT"` and `"TRACE"`.
* `http_url_path` - The URL used for the check.
* `http_version` - The HTTP version used for the check. Options include `"1.0"` and `"1.1"`.
* `http_host` - The server name in the HTTP Host: header used for the check.