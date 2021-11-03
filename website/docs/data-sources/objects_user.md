---
page_title: "cloudscale.ch: cloudscale_objects_user"
---

# cloudscale\_objects\_user

Provides access to cloudscale.ch private networks that are not managed by terraform.

**Hint**: When using this data source, your Terraform state will contain
sensitive data, namely the Objects User secret key. Hence you should treat the
Terraform state the same way as you treat the secret key itself. For more
information, see <a href="/docs/state/sensitive-data.html">here</a>.

## Example Usage

```hcl
data "cloudscale_objects_user" "basic" {
  display_name = "donald_knuth"
}
```

## Argument Reference

The following arguments can be used to look up an Objects User:

* `id` - (Optional) The unique identifier of the Objects User.
* `display_name` - (Optional) The display name of the Objects User.
* `user_id` - (Optional) The unique identifier of the Objects User. (Exactly the same as `id`)

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `href` - The cloudscale.ch API URL of the resource.
* `keys` - A list of key objects containing the access and secret key associated with the Objects User. Currently, only one key object is returned. Each key object has the following attributes:
  * `access_key` - The S3 access key of the Objects User.
  * `secret_key` - The S3 secret key of the Objects User.
