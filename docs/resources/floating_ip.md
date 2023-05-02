---
page_title: "cloudscale.ch: cloudscale_floating_ip"
---

# cloudscale\_floating\_ip

Provides a cloudscale.ch Floating IP to represent a publicly-accessible static IP address or IP network that can be assigned to one of your cloudscale.ch servers. Floating IPs can be moved between servers. Possible use cases include: High-availability, non-disruptive maintenance, multiple IPs per server, or re-using the same IP after replacing a server.

## Example Usage

```hcl
# Create a new Server
resource "cloudscale_server" "web-worker01" {
  name        = "web-worker01"
  flavor_slug = "flex-8-4"
  image_slug  = "debian-9"
  ssh_keys    = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIL2jzgla23DfRVLQr3KT20QQYovqCCN3clHrjm2ZuQFW user@example.com"]
}

# Add a Floating IPv4 address to web-worker01
resource "cloudscale_floating_ip" "web-worker01-vip" {
  server      = cloudscale_server.web-worker01.id
  ip_version  = 4
  reverse_ptr = "vip.web-worker01.example.com"
}

# Add a Floating IPv6 network to web-worker01
resource "cloudscale_floating_ip" "web-worker01-net" {
  server        = cloudscale_server.web-worker01.id
  ip_version    = 6
  prefix_length = 56
}
```

## Argument Reference

The following arguments are supported when adding Floating IPs:

* `server` - (Optional) Assign the Floating IP to this server (UUID).
* `load_balancer` - (Optional) Assign the Floating IP to this load balancer (UUID).
* `ip_version` - (Required) `4` or `6`, for an IPv4 or IPv6 address or network respectively.
* `prefix_length` - (Optional) If you want to assign an entire network instead of a single IP address to your server, you must specify the prefix length. Currently, there is only support for `ip_version=6` and `prefix_length=56`.
* `type` - (Optional) You can specify the type. Options include `regional` (default) and `global`.
* `region_slug` - (Optional) The slug of the region in which the new Regional Floating IP will be created. Options include `lpg` and `rma`.
* `reverse_ptr` - (Optional) You can specify the PTR record (reverse DNS pointer) in case of a single Floating IP address.
* `tags` - (Optional) Tags allow you to assign custom metadata to resources:
  ```
  tags = {
    foo = "bar"
  }
  ```
  Tags are always strings (both keys and values).

The following arguments are supported when updating Floating IPs:

* `server` - (Optional) (Re-)Assign the Floating IP to this server (UUID).
* `load_balancer` - (Optional) (Re-)Assign the Floating IP to this load balancer (UUID).
* `reverse_ptr` - (Optional) You can specify the new PTR record (reverse DNS pointer) in case of a single Floating IP address.
* `tags` - (Optional) Change tags (see documentation above)

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current resource.
* `network` - The CIDR notation of the Floating IP address or network, e.g. `192.0.2.123/32`.
* `next_hop` - The IP address of the server or load balancer that your Floating IP is currently assigned to.


## Import

Floating IPs can be imported using the Floating IP's network IP:

```
terraform import cloudscale_floating_ip.floating_ip 192.0.2.24
```
