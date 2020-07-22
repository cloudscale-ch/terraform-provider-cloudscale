---
layout: "cloudscale"
page_title: "cloudscale.ch: cloudscale_objects_user"
sidebar_current: "docs-cloudscale-resource-objects-user"
description: |-
  Provides a cloudscale.ch Objects User resource. This can be used to create, modify, and delete Objects Users.
---

# cloudscale\_objects\_user

Provides a cloudscale.ch Objects User for the S3-compatible object storage.

**Hint**: When using this resource, your Terraform state will contain sensitive data, namely the objects user secret
key. Hence you should treat the Terraform state the same way as you treat the secret key itself. For more
information, see <a href="/docs/state/sensitive-data.html">here</a>.

## Example Usage

```hcl
# Create an objects user
resource "cloudscale_objects_user" "basic" {
  display_name = "donald_knuth"
}
```

## Argument Reference

The following arguments are supported when adding Objects Users:

* `display_name` - (Required) The display name of the objects user.

The following arguments are supported when updating Objects Users:

* `display_name` - (Required) The new display name of the objects user.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the current resource.
* `user_id` - The unique identifier of the objects user.
* `keys` - A list of key objects containing the access and secret key associated with the objects user. Currently, only one key object is returned. Each key object has the following attributes:
  * `access_key` - The S3 access key of the objects user.
  * `secret_key` - The S3 secret key of the objects user.

