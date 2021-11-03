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
* Add terraform import for all resources

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
