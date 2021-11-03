---
page_title: "cloudscale.ch: cloudscale_floating_ip"
---

# cloudscale\_floating\_ip

Provides access to cloudscale.ch Floating IP that are not managed by terraform.

## Example Usage

```hcl
data "cloudscale_floating_ip" "web-worker01-vip" {
  network = "192.0.2.42/32"
}

data "cloudscale_floating_ip" "web-worker01-net" {
  reverse_ptr = "vip.web-worker01.example.com"
  ip_version  = 6
}
```

## Argument Reference

The following arguments can be used to look up a Floating IP:

* `id` - (Optional) The network IP of the floating IP, e.g. `192.0.2.0` of the network `192.0.2.0/24`.
* `network` - The CIDR notation of the Floating IP address or network, e.g. `192.0.2.123/32`.
* `reverse_ptr` - The PTR record (reverse DNS pointer) in case of a single Floating IP address.
* `ip_version` - `4` or `6`, for an IPv4 or IPv6 address or network respectively.
* `region_slug` - The slug of the region in which a Regional Floating IP is assigned.
* `type` - Options include `regional` and `global`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current resource.
* `next_hop` - The IP address of the server that your Floating IP is currently assigned to.
* `server` - The UUID of the server that your Floating IP is currently assigned to.
* `prefix_length` - The prefix length of a Floating IP (e.g. /128 or /56, as an integer).
