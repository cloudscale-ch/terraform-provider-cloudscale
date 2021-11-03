---
page_title: "cloudscale.ch: cloudscale_server"
---

# cloudscale\_server

Provides access to cloudscale.ch servers that are not managed by terraform.

## Example Usage

```hcl
data "cloudscale_server" "web-worker01" {
  name = "web-worker01"
}
```

## Argument Reference

The following arguments can be used to look up a server:

* `id` - (Optional) The UUID of a server.
* `name` - (Optional) Name of the server.
* `zone_slug` - (Optional) The slug of the zone in which the server exists. Options include `lpg1` and `rma1`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The UUID of this server.
* `href` - The cloudscale.ch API URL of the current resource.
* `ssh_fingerprints` - A list of SSH host key fingerprints (strings) of this server.
* `ssh_host_keys` - A list of SSH host keys (strings) of this server.
* `flavor_slug` - The slug (name) of the flavor used by server.
* `image_slug` - The slug (name) of the image (or custom image) used by the server.
* `volumes` - A list of volume objects attached to this server. Each volume object has three attributes:
    * `device_path` - The path (string) to the volume on your server (e.g. `/dev/vda`)
    * `size_gb` - The size (int) of the volume in GB. Typically matches `volume_size_gb` or `bulk_volume_size_gb`.
    * `type` - A string. Either `ssd` or `bulk`.
* `public_ipv4_address` - The first `public` IPv4 address of this server. The returned IP address may be `""` if the server does not have a public IPv4.
* `private_ipv4_address` - The first `private` IPv4 address of this server. The returned IP address may be `""` if the server does not have private networking enabled.
* `public_ipv6_address` - The first `public` IPv6 address of this server. The returned IP address may be `""` if the server does not have a public IPv6.
* `interfaces` - A list of interface objects attached to this server. Each interface object has the following attributes:
    * `network_name` - The name of the network the interface is attatched to.
    * `network_href` - The cloudscale.ch API URL of the network the interface is attached to.
    * `type` - Either `public` or `private`. Public interfaces are connected to the Internet, while private interfaces are not.
    * `addresses` - A list of address objects:
        * `gateway` - An IPv4 or IPv6 address that represents the default gateway for this interface.
        * `prefix_length` - The prefix length for this IP address, typically 24 for IPv4 and 128 for IPv6.
        * `reverse_ptr` - The PTR record (reverse DNS pointer) for this IP address. If you use an FQDN as your server name it will automatically be used here.
        * `version` - The IP version, either `4` or `6`.
        * `subnet_cidr` - The cidr of the subnet the address is part of.
        * `subnet_href` - The cloudscale.ch API URL of the subnet the address is part of.
* `interfaces` - A list of interface configuration objects (see [example](network.html)). Each interface object has the following attributes:
    * `type` - The type of the interface. Can be `public` or `private`.
    * `network_uuid` - The UUID of the private network this interface should be attached to. Must be compatible with `subnet_uuid` if both are specified.
    * `addresses` - Can be set only for `private` interfaces) A list of address objects:
        * `address` - An IP address that has been assigned to this server.
        * `subnet_uuid` - The UUID of the subnet this address should be part of. Must be compatible with `network_uuid` if both are specified.
* `status` - The state of a server.
