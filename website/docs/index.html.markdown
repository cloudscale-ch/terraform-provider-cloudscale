---
layout: "cloudscale"
page_title: "Provider: cloudscale.ch"
sidebar_current: "docs-cloudscale-index"
description: |-
  The cloudscale.ch provider is used to interact with the resources supported by cloudscale.ch. The provider needs to be configured with the proper credentials before it can be used.
---

# cloudscale.ch Provider

The cloudscale.ch provider is used to interact with the resources supported by cloudscale.ch. The provider needs to be configured with proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

Terraform 0.13 and later:
```hcl
terraform {
  required_providers {
    cloudscale = {
      source = "cloudscale-ch/cloudscale"
      // The version attribute can be used to pin to a specific version
      //version = "~> 3.0.0"
    }
  }
}

# Create a resource
resource "cloudscale_server" "web-worker01" {
  # ...
}
```

Terraform 0.12 and earlier:
```hcl
provider "cloudscale" {
  version = "~> 2.3.0"
}

# Create a resource
resource "cloudscale_server" "web-worker01" {
  # ...
}
```

## Authentication

Please create a cloudscale.ch API token with read/write access in
our [Cloud Control Panel](https://control.cloudscale.ch/). You can then
pass the token to the provider using one of the following methods:

### Environment variable (recommended)

Set a shell environment variable called `CLOUDSCALE_TOKEN`.

### Static credentials

Add the following configuration:

```hcl
# Set the variable value in a *.tfvars file or use 
# the -var="cloudscale_token=..." CLI option.
variable "cloudscale_token" {}

provider "cloudscale" {
  token = var.cloudscale_token
}
```
