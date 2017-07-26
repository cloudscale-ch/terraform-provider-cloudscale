---
layout: "cloudscale"
page_title: "CloudScale: cloudscale"
sidebar_current: "docs-cloudscale-resource-server"
description: |-
  Provides a CloudScale Seerver resource. This can be used to create, modify, and delete Servers.
---

# cloudscale\_server

Provides a CloudScale Server resource. This can be used to create,
modify, and delete Servers. Droplets also support

## Example Usage

```hcl
# Create a new Web Droplet in the nyc2 region
resource "cloudscale_server" "basic" {
  name      			= "db-master"
  flavor    			= "flex-4"
  image     			= "debian-8"
  volume_size_gb	= 10
  ssh_keys = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}
```

## Argument Reference

The following arguments are supported:

* `image` - (Required) Slug of the image to use for the new server.
* `name` - (Required) Name to use for the new server. The name has to be a valid 
   host name or fully qualified domain name (FQDN).
* `flavor` - (Required) Slug of the flavor to use for the new server.
* `ssh_keys` - (Required) A list of SSH public keys. Use the full content of 
  your .pub file here.
* `volume_size_gb` - (Optional) The instance size to start
* `backups` - (Optional) The size in GB of the root volume to use for the new 
   server. If this parameter is not specified, the size will be set to 10 GB.
* `bulk_volume_size_gb` - (Optional) The size in GB of the bulk storage volume 
   to use for the new server. If this parameter is not specified or is set to 
   `null`, no bulk storage volume will be attached to the server.
* `use_public_network` - (Optional) Attaches/detaches the public network 
   interface from the new server. Can be `true` (the default) or `false`.
* `use_private_network` - (Optional) Attaches/detaches the attach the public 
   network interface from the new server. Can be `true` or `false` (the default).
* `use_ipv6` - (Optional) Enables/disables IPv6 on the public network interface 
   of the new server. Can be `true` (the default) or `false`.
* `user_data` (Optional) - A string of the desired User Data for the Droplet.
* `anti_affinity_with	` (Optional) - Pass the UUID of another server to create 
   an anti-affinity group with that server or add it to the same group as that 
  server.
* `user_data	` (Optional) - Cloud-init configuration (cloud-config) data to use 
   for the new server. Needs to be valid YAML. A default configuration is used 
   if this parameter is not specified or is set to null. Use only if you are an 
   advanced users with knowledge of cloud-config..

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the S erver
* `name`- The name of the Server
* `href` - The url of the Server
* `image` - The image of the Server
* `flavor` - The flavor of the Server
* `use_ipv6` - Is IPv6 enabled
* `ipv6_address` - The IPv6 address
* `ipv6_private_address` - The private networking IPv6 address
* `ipv4_address` - The IPv4 address
* `ipv4_private_address` - The private networking IPv4 address
* `ssh_fingerprints` - A list of SSH host key fingerprints
* `ssh_host_keys` - A list of SSH host keys
* `anti_affinity_with` - Droplet hourly price
* `price_monthly` - Droplet monthly price
* `size` - The instance size
* `disk` - The size of the instance's disk in GB
* `vcpus` - The number of the instance's virtual CPUs
* `status` - The status of the droplet
* `tags` - The tags associated with the droplet
* `volumes` -  A list of volumes attached to this server.

