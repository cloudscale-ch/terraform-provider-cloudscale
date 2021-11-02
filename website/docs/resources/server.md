---
page_title: "cloudscale.ch: cloudscale_server"
---

# cloudscale\_server

Provides a cloudscale.ch server resource. This can be used to create, modify, import, and delete servers. 

## Example Usage

```hcl
# Create a new server
resource "cloudscale_server" "web-worker01" {
  name                = "web-worker01"
  flavor_slug         = "flex-4"
  image_slug          = "debian-9"
  volume_size_gb      = 10
  ssh_keys            = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIL2jzgla23DfRVLQr3KT20QQYovqCCN3clHrjm2ZuQFW user@example.com"]
  
  timeouts {
    create = "10m"
  }
}
```

## Argument Reference

The following arguments are supported when creating new servers:

* `name` - (Required) Name of the new server. The name has to be a valid host name or a fully qualified domain name (FQDN).
* `flavor_slug` - (Required) The slug (name) of the flavor to use for the new server. Possible values can be found in our [API documentation](https://www.cloudscale.ch/en/api/v1#flavors).
    **Note:** If you want to update this value after initial creation, you must set [`allow_stopping_for_update`](#allow_stopping_for_update) to `true`.
* `image_slug` - (Required, if `image_uuid` not set) The slug (name) of the image (or custom image) to use for the new server. Possible values can be found in our [API documentation](https://www.cloudscale.ch/en/api/v1#images).
* `image_uuid` - (Required, if `image_slug` not set) The UUID of the custom image to use for the new server. **Note:** This is the recommended approach for custom images.
* `ssh_keys` - (Optional) A list of SSH public keys. Use the full content of your \*.pub file here.
* `password` - (Optional) The password of the default user of the new server. When omitted, no password will be set.
* `zone_slug` - (Optional) The slug of the zone in which the new server will be created. Options include `lpg1` and `rma1`.
* `volume_size_gb` - (Optional) The size in GB of the SSD root volume of the new server. If this parameter is not specified, the value will be set to 10. The minimum value is 10.
* `bulk_volume_size_gb` - (Optional, Deprecated) The size in GB of the bulk storage volume of the new server. If this parameter is not specified, no bulk storage volume will be attached to the server. Valid values are multiples of 100.
* `use_public_network` - (Optional) Attach the public network interface to the new server. Can be `true` (default) or `false`. Use [`interfaces`](#interfaces) option for advanced setups.
* `use_private_network` - (Optional) Attach the `default` private network interface to the new server. Can be `true` or `false` (default). Use [`interfaces`](#interfaces) option for advanced setups.
* `use_ipv6` - (Optional) Enable/disable IPv6 on the public network interface of the new server. Can be `true` (default) or `false`.
* `interfaces` - (Optional) A list of interface configuration objects (see [example](network.html)). Each interface object has the following attributes:
    * `type` - (Required) The type of the interface. Can be `public` or `private`.
    * `network_uuid` - (Optional, can be set only for `private` interfaces) The UUID of the private network this interface should be attached to. Must be compatible with `subnet_uuid` if both are specified.
    * `addresses` - (Optional, can be set only for `private` interfaces) A list of address objects:
        * `address` - (Optional) An IP address that has been assigned to this server.
        * `subnet_uuid` - (Optional) The UUID of the subnet this address should be part of. Must be compatible with `network_uuid` if both are specified.
    * `no_address` - (Optional, can be set only for `private` interfaces) You neet to set this to `true` if no address should be configured, e.g. if you want to attach to a network without a subnet. 
* `server_group_ids` - (Optional) A list of server group UUIDs to which the server should be added. Default to an empty list.
* `user_data` - (Optional) User data (custom cloud-config settings) to use for the new server. Needs to be valid YAML. A default configuration will be used if this parameter is not specified or set to null. Use only if you are an advanced user with knowledge of cloud-config and cloud-init.
* `status` - (Optional) The desired state of a server. Can be `running` (default) or `stopped`.
* `allow_stopping_for_update` - (Optional) If true, allows Terraform to stop the instance to update its properties. If you try to update a property that requires stopping the instance without setting this field, the update will fail.
* `skip_waiting_for_ssh_host_keys` - (Optional) If set to `true`, do not wait until SSH host keys become available.
* `timeouts` - (Optional) Specify how long certain operations are allowed to take before being considered to have failed. Currently, only the `create` timeout can be specified. Takes a string representation of a duration such as `5m` for 5 minutes (default), `10s` for ten seconds, or `2h` for two hours.

The following arguments are supported when updating servers:

* `name` - Name of the new server. The name has to be a valid host name or a fully qualified domain name (FQDN).
* `volume_size_gb` - The size in GB of the SSD root volume of the new server.
* `interfaces` - A list of interface configuration objects. Each interface object has the following attributes:
    * `type` - (Required) The type of the iinterface. Can be `public` or `private`.
    * `network_uuid` (Required for `private` interfaces) The UUID of the private network this interface should be attached to.
* `status` - The desired state of a server. Can be `running` (default) or `stopped`.


## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The UUID of this server.
* `href` - The cloudscale.ch API URL of the current resource.
* `ssh_fingerprints` - A list of SSH host key fingerprints (strings) of this server.
* `ssh_host_keys` - A list of SSH host keys (strings) of this server.
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


## Import

Volumes can be imported using the server's UUID:

```
terraform import cloudscale_volume.server 48151623-42aa-aaaa-bbbb-caffeeeeeeee
```
