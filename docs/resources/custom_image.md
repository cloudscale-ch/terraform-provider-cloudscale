---
page_title: "cloudscale.ch: cloudscale_custom_image"
---

# cloudscale\_custom\_image

Provides a cloudscale.ch custom image resource. This can be used to create, modify, import, and delete custom images.

## Example Usage

```hcl
# Create a custom image
resource "cloudscale_custom_image" "your_image" {
  import_url           = "https://mirror.example.com/your-distro-12.12-openstack-amd64.raw"
  import_source_format = "raw"
  name                 = "Your Distro 12.12"
  slug                 = "your-distro-12.12"
  user_data_handling   = "extend-cloud-config"
  firmware_type        = "bios"
  zone_slugs           = ["rma1"]
  
  timeouts {
    create = "10m"
  }
}

# Create a Server using the custom image
resource "cloudscale_server" "your_server" {
  name           = "your-server"
  flavor_slug    = "flex-8-4"
  image_uuid     = "${cloudscale_custom_image.your_image.id}"
  volume_size_gb = 16
  zone_slug      = "rma1"
  ssh_keys       = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIL2jzgla23DfRVLQr3KT20QQYovqCCN3clHrjm2ZuQFW user@example.com"]

  // If your image does not print complete SSH host keys to console during initial boot in the following format
  // enable the option below.
  //  
  // -----BEGIN SSH HOST KEY KEYS-----
  // ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJIdoMOxHQZwxnthOnUpd0Wl7TPRsJdj5KvW9YdE3Pbk
  // [... more keys ...] 
  // -----END SSH HOST KEY KEYS----- 
  //
  //skip_waiting_for_ssh_host_keys = true
}
```

## Argument Reference

The following arguments are supported when creating/changing custom images:

* `import_url` - (Required) The URL used to download the image.
* `import_source_format` - (Required) The file format of the image referenced in the `import_url`. Options include `raw`.
* `name` - (Required) The human-readable name of the custom image.
* `slug` - (Optional) A string identifying the custom image for use within the API.
* `user_data_handling` - (Required) How user_data will be handled when creating a server. Options include `pass-through` and `extend-cloud-config`.
* `firmware_type` - (Optional) The firmware type that will be used for servers created with the custom image. Options include `bios` and `uefi`.
* `zone_slugs` - (Required) Specify the zones in which the custom image will be available. Options include `lpg1` and `rma1`.
* `timeouts` - (Optional) Specify how long certain operations are allowed to take before being considered to have failed. Currently, only the `create` timeout can be specified. Takes a string representation of a duration, such as `20m` for 20 minutes (default), `10s` for ten seconds, or `2h` for two hours.
* `tags` - (Optional) Tags allow you to assign custom metadata to resources:
  ```
  tags = {
    foo = "bar"
  }
  ```
  Tags are always strings (both keys and values).

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current resource.
* `size_gb` - The size in GB of the custom image.
* `checksums` - The checksums of the custom image as map.
* `import_href` - The cloudscale.ch API URL of the custom image import.
* `import_uuid` - The UUID of the custom image import.
* `import_status` - The status of the custom image import. Options include `started`, `in_progress`, `failed`, `success`.


## Import

Custom images can currently not be imported, please use a data source.
