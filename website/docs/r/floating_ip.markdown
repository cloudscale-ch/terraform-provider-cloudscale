---
layout: "cloudscale"
page_title: "CloudScale: cloudscale_floating_ip"
sidebar_current: "docs-cloudscale-resource-floating-ip"
description: |-
  Provides a CloudScale Floating IP resource.
---

# clouscale\_floating_ip

Provides a CloudScale Floating IP to represent a publicly-accessible static IP addresses that can be mapped to one of your Servres.

## Example Usage

```hcl
resource "cloudscale_server" "basic" {
  name      			= "db-master"
  flavor    			= "flex-2"
  image     			= "debian-8"
  volume_size_gb	= 10
  ssh_keys = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}
resource "cloudscale_floating_ip" "gateway" {
  server 					= "${cloudscale_server.basic.id}"
  ip_version     	= 4
}
```

## Argument Reference

The following arguments are supported:

* `ip_version` - (Required) `4` or `6 `for an IPv4 or IPv6 address.
* `server` - (Optional) The server UUID the floating IP is pointed to.
* `server` - (Required) The server UUID the floating IP is pointed to.
* `prefix_length` - (Optional) If you want a whole network instead of a single 
   IP routed to your server, specify the prefix length here. This is only 
   supported for IPv6 and prefix length `56`.
* `reverse_ptr` - (Optional) You can optionally specify a reverse pointer.



## Attributes Reference

The following attributes are exported:

* `href` - The URL for the current field
* `network` - The IP/network that is routed to your server
* `next_hop` - The IP of the server that your Floating IP is assigned to
* `href` - The reverse pointer for this IP address
* `href` - The server to which the floating IP is routed