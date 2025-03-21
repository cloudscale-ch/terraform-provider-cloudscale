## 5.0.0
* :warning: **Breaking Change**: The default value for the `status` attribute 
  of `cloudscale_server` is now `"running"` when no value is provided 
  (i.e., missing in your .tf file). If your servers are intended to be in
  a state other than `"running"`, please explicitly set the appropriate state
  before upgrading to this version.
* :warning: **Breaking Change**: The default timeout for server changes is 
  now `1h` instead of `5m`. This change is necessary because changing the 
  flavor of a GPU server with a scratch disk can take a significant 
  amount of time if data is moved to a new host during the process. 
  If you prefer to use the old timeout, you can now add the following 
  block to your `cloudscale_server` configuration:
  ```hcl
  timeouts {
    update = "5m"
  }
* Update go dependencies.

## 4.4.0
* Use `import_source_format` again. This must be specified to import custom images in formats other than `raw`. 
  See also [our blog](https://www.cloudscale.ch/en/news/2024/07/31/securing-qcow2-image-imports). 
  There is no need to change this for existing imported images.
* Update go dependencies.

## 4.3.0
* Add `disable_dns_servers` to subnet.
* Update go dependencies.

## 4.2.3
* Update go dependencies.

## 4.2.2
* Update go dependencies.

## 4.2.1
 * Update go dependencies.

## 4.2.0
 * Support for cloudscale.ch [Loads Balancers](https://www.cloudscale.ch/en/api/v1#load-balancers).
 * Ignore `import_source_format` as it has been deprecated in the cloudscale.ch API.
   You can remove the attribute from your Terraform file if you wish. The suggested
   in-place upgrades are a no-ops.

## 4.1.0
 * Add firmware_type to custom_image.
 * Update to go 1.18.

## 4.0.0
 * Implement tags for resources (#59)
 * Mark the keys attribute of `cloudscale_objects_user` as sensitive (#63)
 * Use consistent naming and usage of variables across all cloudscale.ch tools (#58)
 * Update to latest terraform-plugin-sdk to ensure compatibility with Terraform v1.1.x (#58)
 * Update to latest cloudscale-go-sdk (#58)
 * Update to latest terraform-plugin-sdk (#62)
 * :warning: **Breaking Change**: To be consistent with cloudscale.ch's other tools, the
    environment variable `CLOUDSCALE_TOKEN` has been renamed to `CLOUDSCALE_API_TOKEN`.
    Please adapt your environment accordingly. If you are configuring the token through
    some other means than an environment variable, you are not affected by this change.

## 3.2.0
* Add data sources:
  - `cloudscale_server`
  - `cloudscale_server_group`
  - `cloudscale_volume`
  - `cloudscale_network`
  - `cloudscale_subnet`
  - `cloudscale_floating_ip`
  - `cloudscale_custom_image`
  - `cloudscale_objects_user`
* Add terraform import for all resources (except Custom Images)
* Allow updating the name of server groups.
* Allow updating the PTR record (reverse DNS pointer) of Floating IPs.

## 3.1.0
* Update to go 1.16 (#48) to support Apple silicon.

## 3.0.0
* Upgrade terraform-plugin-sdk to v2 (#43)
* Add Support for Custom Images (#44)
* Add Options for SSH Host Keys (#45)
* :warning: **Breaking Change**: Terraform versions older than 0.12 are no longer supported.

## 2.3.0 (October 19, 2020)
* Allow creating Global Floating IPs (#34, #36)

## 2.2.0 (July 23, 2020)

* **New Resource**: `cloudscale_objects_user` is now available (#29)
* Allow creating unattached Floating IPs (#30)

## 2.1.2 (April 22, 2020)

FEATURES:

* Add Subnets and Addresses (#25)

## 2.1.1 (December 04, 2019)

FEATURES:

* Add Support for Networks (#20)
* Add Password Option to Server (#21)

## 2.1.0 (November 20, 2019)

FEATURES:

* Support for Terraform 0.12.x
* Add Zones/Regions to use with all resources

## 2.0.0 (July 12, 2019)

FEATURES:

* **New Resource**: `cloudscale_server_group` is now available (#16)

BACKWARDS INCOMPATIBILITIES:

* Implicit server groups are no longer supported. This means that you cannot
  just use `anti_affinity_with` anymore.

## 1.1.0 (April 11, 2019)

FEATURES:

* **New Resource**: `cloudscale_volume` is now available (#5)

ENHANCEMENTS:
* Added support for scaling servers (#13)
* Added support for scaling root volumes (#14)

IMPROVEMENTS:

* Expose the first public/private IPv4 and IPv6 addresses as string attributes `public_ipv4`,
  `public_ipv6` and `private_ipv4` (#8)

## 1.0.1 (April 06, 2018)


IMPROVEMENTS:

* `resource_cloudscale_server`: Use documented defaults for `cloudscale_server` ([#1](https://github.com/terraform-providers/terraform-provider-aws/issues/1))

## 1.0.0 (November 01, 2017)

* Initial release of the cloudscale.ch provider
