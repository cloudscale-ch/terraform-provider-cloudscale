---
layout: "cloudscale"
page_title: "Provider: cloudscale.ch"
sidebar_current: "docs-cloudscale-resource-floating-ip"
description: |-
  Provides a cloudscale.ch floating IP resource. It can be used to create, modify, and delete floating IPs.
---

# clouscale\_floating\_ip

Provides a cloudscale.ch floating IP to represent a publicly-accessible static
IP addresses that can be mapped to one of your cloudscale.ch servers.

## Example Usage

```hcl
# Creates a server
resource "cloudscale_server" "web_gateway" {
  name      			= "web-gateway"
  flavor    			= "flex-4"
  image     			= "debian-8"
  ssh_keys = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}
# Adds a floating IP to the server
resource "cloudscale_floating_ip" "gateway_floating_ip" {
  server 	= "${cloudscale_server.web_gateway.id}"
  ip_version   	= 4
}
```

## Argument Reference

The following arguments are supported for creating floating IPs:

* `ip_version` - (Required) `4` or `6 `for an IPv4 or IPv6 address.
* `server` - (Required) The floating IP is pointed to this server (UUID).
* `prefix_length` - (Optional) If you want a whole network instead of a single 
   IP routed to your server, specify the prefix length here. This is only 
   supported for IPv6 and prefix length `56`.
* `reverse_ptr` - (Optional) You can specify a reverse pointer.

The following arguments are supported for updating floating IPs:

* `server` - (Required) The floating IP is changed to this server (UUID).

## Attributes Reference

The following attributes are exported:

* `href` - The cloudscale.ch API URL for the current field.
* `server` - The floating IP is routed to this server (UUID).
* `network` - The CIDR notation of the network that is routed to your server,
   e.g. `192.0.2.123/32`.
* `next_hop` - Your floating IP is routed to this IP address.
* `reverse_ptr` - The reverse pointer for this floating IP address.
