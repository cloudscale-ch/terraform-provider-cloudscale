---
page_title: "cloudscale.ch: cloudscale_custom_image"
---

# cloudscale\_custom\_image

Get information on a cloudscale.ch custom image.

## Example Usage

```hcl
data "cloudscale_custom_image" "your_image" {
  name                 = "Your Distro 42.42"
}

# Create a Server using the custom image
resource "cloudscale_server" "your_server" {
  name           = "your-server"
  flavor_slug    = "flex-8-4"
  image_uuid     = "${data.cloudscale_custom_image.your_image.id}"
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

The following arguments can be used to look up a custom image:

* `id` - (Optional) The UUID of a custom image.
* `name` - (Optional) The human-readable name of a custom image.
* `slug` - (Optional) A string identifying a custom image.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current resource.
* `size_gb` - The size in GB of the custom image.
* `checksums` - The checksums of the custom image as map.
* `user_data_handling` - How user_data will be handled when creating a server. Options include `pass-through` and `extend-cloud-config`.
* `zone_slugs` - The zones in which the custom image will be available. Options include `lpg1` and `rma1`.
