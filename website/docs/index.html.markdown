---
layout: "cloudscale"
page_title: "Provider: CloudScale"
sidebar_current: "docs-cloudscale-index"
description: |-
  The CloudScale provider is used to interact with the resources supported by CloudScale. The provider needs to be configured with the proper credentials before it can be used.
---

# CloudScale Provider

The CloudScale provider is used to interact with the
resources supported by CloudScale. The provider needs to be configured
with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
# Set the variable value in *.tfvars file
# or using -var="cloudscale_token=..." CLI option
variable "cloudscale_token" {}

# Configure the CloudScale Provider
provider "cloudscale" {
  token = "${var.cloudscale_token}"
}

# Create a web server
resource "cloudscale_server" "web" {
  # ...
}
```

## Argument Reference

The following arguments are supported:

* `token` - (Required) This is the CloudScale API token. This can also be specified
  with the `CLOUDSCALE_TOKEN` shell environment variable.

