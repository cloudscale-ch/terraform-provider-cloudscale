---
layout: "cloudscale"
page_title: "Provider: cloudscale.ch"
sidebar_current: "docs-cloudscale-resource-server"
description: |-
  Provides a cloudscale.ch server resource. It can be used to create, modify, and delete servers.
---

# cloudscale\_server

Provides a cloudscale.ch server resource. This can be used to create, modify,
and delete servers. 

## Example Usage

```hcl
# Creates a new server
resource "cloudscale_server" "db_server" {
  name      			= "db-server"
  flavor    			= "flex-4"
  image     			= "debian-8"
  volume_size_gb	    = 10
  ssh_keys = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}
```

## Argument Reference

The following arguments are supported:

* `image_slug` - (Required) The slug (name) of the image to use for a new server.
   Possible values can be found [here](https://www.cloudscale.ch/en/api/v1#images).
* `name` - (Required) Name to use for the new server. The name has to be a
   valid host name or a fully qualified domain name (FQDN).
* `flavor_slug` - (Required) The slug (name) of the flavor to use for the new server.
   Possible values are can be found [here](https://www.cloudscale.ch/en/api/v1#flavors).
* `ssh_keys` - (Required) A list of SSH public keys. Use the full content of 
   your \*.pub files.
* `volume_size_gb` - (Optional) The size in GB of the SSD root volume to use
   for the new server. If this parameter is not specified, the size will be set
   to 10 GB. Valid values are multiples of 50.
* `bulk_volume_size_gb` - (Optional) The size in GB of the bulk storage volume 
   to use for the new server. If this parameter is not specified, no bulk
   storage volume will be attached to the server. Valid values are multiples of 100.
* `use_public_network` - (Optional) Attaches/detaches the public network 
   interface from the new server. Can be `true` (the default) or `false`.
* `use_private_network` - (Optional) Attaches/detaches the attach the private 
   network interface from the new server. Can be `true` or `false` (the default).
* `use_ipv6` - (Optional) Enables/disables IPv6 on the public network interface 
   of the new server. Can be `true` (the default) or `false`.
* `anti_affinity_with	` (Optional) - Pass the UUID of another server to
   create an anti-affinity group with that server or add it to the same group
   as that server.
* `user_data` (Optional) - Cloud-init configuration (cloud-config) data to use 
   for the new server. Needs to be valid YAML. A default configuration is used 
   if this parameter is not specified or is set to null. Use only if you are an 
   advanced users with knowledge of cloud-config.
* `status` (Optional) - The desired state of a server, can be `running` or `stopped`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes
are exported:

* `id` - The UUID of the server
* `href` - The cloudscale.ch API URL of the server
* `ssh_fingerprints` - A list of SSH host key fingerprints (strings) of your server.
* `ssh_host_keys` - A list of SSH host keys (strings) of your server.
* `anti_affinity_with` - 
* `volumes` - A list of volume objects attached to this server. A volume object
   has three attributes:
    * `device_path` - A string, describes the path on your server (e.g. `/dev/vda`)
    * `size_gb` - The size (int) of the volume, typically matches
       `volume_size_gb` or `bulk_volume_size_gb`.
    * `type` - A string. Either `ssd` or `bulk`.
* `interfaces` - A list of interface objects. An interface object has two attributes:
    * `type` - Either `public` or `private`. Public interfaces are connected
       to the internet, while private interfaces are not.
    * `addresses` - A list of address objects:
        * address - An IPv4 or IPv6 address that is available for that server.
        * gateway - Also an IPv4 or IPv6 address.
        * prefix_length - The prefix length of the IP address, typically 24 for
          IPv4 and 128 for IPv6.
        * reverse_ptr - The reverse pointer for this IP address. If you use an
          FQDN as your server name it will be used here.
        * version - The IP version, either `4` or `6`.
