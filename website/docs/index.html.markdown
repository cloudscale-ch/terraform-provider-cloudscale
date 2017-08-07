---
layout: "cloudscale"
page_title: "Provider: cloudscale.ch"
sidebar_current: "docs-cloudscale-index"
description: |-
  The cloudscale.ch provider is used to interact with the resources supported by cloudscale.ch. The provider needs to be configured with the proper credentials before it can be used.
---

# cloudscale.ch Provider

The cloudscale.ch provider is used to interact with the resources supported by
cloudscale.ch. The provider needs to be configured with proper credentials
before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
# Set the variable value in *.tfvars file
# or using the -var="cloudscale_token=..." CLI option
variable "cloudscale_token" {}

# Configure the CloudScale Provider
provider "cloudscale" {
  token = "${var.cloudscale_token}"
}

# Create a New Server
resource "cloudscale_server" "web" {
  # ...
}
```

## Argument Reference

The following arguments are supported:

* `token` - (Required) This is the cloudscale.ch API token. It can also be
  specified as a shell environment variable called `CLOUDSCALE_TOKEN`. It can
  be generated in the cloudscale control panel.
