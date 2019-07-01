## 2.0.0 (Unreleased)

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

* Initial release of the CloudScale.ch provider
