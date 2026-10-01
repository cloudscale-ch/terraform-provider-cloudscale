# Set variables in a *.tfvars file, pass them via the CLI with
# -var="variable_name=value", or use TF_VAR_<variable_name>
# environment variables.
#
# The cloudscale API token can alternatively be provided through
# CLOUDSCALE_API_TOKEN. In this case, the cloudscale_api_token
# variable and provider configuration can be omitted.

variable "cloudscale_api_token" {}

terraform {
  required_providers {
    cloudscale = {
      source = "cloudscale-ch/cloudscale"
      // version = "~> x.y.z"
    }
  }
}

provider "cloudscale" {
  token = "${var.cloudscale_api_token}"
}

variable "zone" {
  type        = string
  description = "The zone in which the setup will be deployed. Currently rma1 and lpg1 are available."
  default     = "rma1"

  validation {
    condition     = contains(["rma1", "lpg1"], var.zone)
    error_message = "The zone must be rma1 or lpg1."
  }
}

variable "ssh_public_key" {
  type        = string
  description = "The public key used for accessing the servers via SSH."
}

locals {
  # Service subnet - gateway_address is set: default route pushed via DHCP
  service_subnet_cidr       = "10.10.10.0/24"
  service_gateway_ip        = "10.10.10.1"

  # Jump host subnet - gateway_address is not set: route to service subnet must be set via cloud-init
  jump_host_subnet_cidr     = "10.0.0.0/24"
  jump_host_gateway_ip      = "10.0.0.1"

  # user_data: Add static route on jump host to reach private network via router
  jump_host_user_data       = "#cloud-config\nruncmd:\n  - [\"ip\", \"route\", \"add\", \"${local.service_subnet_cidr}\", \"via\", \"${local.jump_host_gateway_ip}\"]"
}

resource "cloudscale_network" "service_network" {
  name                    = "service-network"
  zone_slug               = var.zone
  auto_create_ipv4_subnet = false
  mtu                     = 1500
}

resource "cloudscale_subnet" "service_subnet" {
  network_uuid    = cloudscale_network.service_network.id
  cidr            = local.service_subnet_cidr
  gateway_address = local.service_gateway_ip
}

resource "cloudscale_network" "jump_host_network" {
  name                    = "jump-host-network"
  zone_slug               = var.zone
  auto_create_ipv4_subnet = false
  mtu                     = 1500
}

resource "cloudscale_subnet" "jump_host_subnet" {
  network_uuid    = cloudscale_network.jump_host_network.id
  cidr            = local.jump_host_subnet_cidr
  # Do not set the the gateway_address here, so the default route of the jump host server is not overwritten
}

resource "cloudscale_router" "internet_gateway" {
  name              = "internet-gateway"
  zone_slug         = var.zone
  internet_gateway  = true
}

resource "cloudscale_interface" "internet_gateway_service_interface" {
  router_uuid  = cloudscale_router.internet_gateway.id
  network_uuid = cloudscale_network.service_network.id
  addresses {
    subnet_uuid = cloudscale_subnet.service_subnet.id
    address     = local.service_gateway_ip
  }
}

resource "cloudscale_interface" "internet_gateway_jump_host_interface" {
  router_uuid  = cloudscale_router.internet_gateway.id
  network_uuid = cloudscale_network.jump_host_network.id
  addresses {
    subnet_uuid = cloudscale_subnet.jump_host_subnet.id
    address     = local.jump_host_gateway_ip
  }
}

resource "cloudscale_server" "postgresql_db_server" {
  name           = "postgresql-db"
  zone_slug      = var.zone
  flavor_slug    = "flex-4-1"
  image_slug     = "debian-13"
  volume_size_gb = 20
  interfaces {
    type = "private"
    network_uuid = cloudscale_network.service_network.id
  }
  ssh_keys       = [var.ssh_public_key]
  allow_stopping_for_update = true
}

resource "cloudscale_server" "jump_host_server" {
  name           = "jump-host"
  zone_slug      = var.zone
  flavor_slug    = "flex-4-1"
  image_slug     = "debian-13"
  volume_size_gb = 20
  interfaces {
    type = "private"
    network_uuid = cloudscale_network.jump_host_network.id
  }
  interfaces {
    type = "public"
  }
  user_data      = local.jump_host_user_data
  ssh_keys       = [var.ssh_public_key]
  allow_stopping_for_update = true
}