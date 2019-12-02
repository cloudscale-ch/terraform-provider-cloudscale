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

```hcl
# Set the variable value in a *.tfvars file or use 
# the -var="cloudscale_token=..." CLI option.
#
# You can omit both the variable and provider if you
# choose to set a shell environment variable called
# `CLOUDSCALE_TOKEN` instead.

variable "cloudscale_token" {}

provider "cloudscale" {
  token = var.cloudscale_token
}

# Create a new Server
resource "cloudscale_server" "web-worker01" {
  # ...
}

# Add a Volume
resource "cloudscale_volume" "web-worker01-volume" {
  server_uuids = [cloudscale_server.web-worker01.id]
  # ...
}

# Add a Floating IP
resource "cloudscale_floating_ip" "web-worker01-vip" {
  server = cloudscale_server.web-worker01.id
  # ...
}
```

## Argument Reference

The following arguments are supported:

* `token` - (Required) Your cloudscale.ch API token. It can also be specified as a shell environment variable called `CLOUDSCALE_TOKEN`. Please create a cloudscale.ch API token with read/write access in our [Cloud Control Panel](https://control.cloudscale.ch/).
