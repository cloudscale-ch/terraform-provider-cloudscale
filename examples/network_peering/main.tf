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
  # user_data: Add static routes on jump host to reach both private networks
  jump_host_user_data = <<-EOT
    #cloud-config
    runcmd:
      - ["ip", "route", "add", "10.0.1.0/24", "via", "10.0.99.1"]
      - ["ip", "route", "add", "10.0.2.0/24", "via", "10.0.99.2"]
  EOT
}

# Network A: 10.0.1.0/24
resource "cloudscale_network" "net_a" {
  name                    = "net-a"
  zone_slug               = var.zone
  auto_create_ipv4_subnet = false
  mtu                     = 1500
}

resource "cloudscale_subnet" "subnet_a" {
  network_uuid    = cloudscale_network.net_a.id
  cidr            = "10.0.1.0/24"
  gateway_address = "10.0.1.1"
}

# Network B: 10.0.2.0/24
resource "cloudscale_network" "net_b" {
  name                    = "net-b"
  zone_slug               = var.zone
  auto_create_ipv4_subnet = false
  mtu                     = 1500
}

resource "cloudscale_subnet" "subnet_b" {
  network_uuid    = cloudscale_network.net_b.id
  cidr            = "10.0.2.0/24"
  gateway_address = "10.0.2.1"
}

# Peering network: 10.0.99.0/24
resource "cloudscale_network" "net_peering" {
  name                    = "net-peering"
  zone_slug               = var.zone
  auto_create_ipv4_subnet = false
  mtu                     = 1500
}

resource "cloudscale_subnet" "subnet_peering" {
  network_uuid = cloudscale_network.net_peering.id
  cidr         = "10.0.99.0/24"
}

# Router A: connects net_a and peering network, provides internet gateway
resource "cloudscale_router" "router_a" {
  name             = "router-a"
  zone_slug        = var.zone
  internet_gateway = true
}

resource "cloudscale_interface" "router_a_net_a" {
  router_uuid  = cloudscale_router.router_a.id
  network_uuid = cloudscale_network.net_a.id

  addresses {
    subnet_uuid = cloudscale_subnet.subnet_a.id
    address     = "10.0.1.1"
  }
}

resource "cloudscale_interface" "router_a_peering" {
  router_uuid  = cloudscale_router.router_a.id
  network_uuid = cloudscale_network.net_peering.id

  addresses {
    subnet_uuid = cloudscale_subnet.subnet_peering.id
    address     = "10.0.99.1"
  }
}

# Router B: connects net_b and peering network, provides internet gateway
resource "cloudscale_router" "router_b" {
  name             = "router-b"
  zone_slug        = var.zone
  internet_gateway = true
}

resource "cloudscale_interface" "router_b_net_b" {
  router_uuid  = cloudscale_router.router_b.id
  network_uuid = cloudscale_network.net_b.id

  addresses {
    subnet_uuid = cloudscale_subnet.subnet_b.id
    address     = "10.0.2.1"
  }
}

resource "cloudscale_interface" "router_b_peering" {
  router_uuid  = cloudscale_router.router_b.id
  network_uuid = cloudscale_network.net_peering.id

  addresses {
    subnet_uuid = cloudscale_subnet.subnet_peering.id
    address     = "10.0.99.2"
  }
}

# Extra routes for peering: each router learns how to reach the other network
resource "cloudscale_router_extra_route" "router_a_to_net_b" {
  depends_on = [cloudscale_interface.router_a_peering]

  router_uuid = cloudscale_router.router_a.id
  destination = "10.0.2.0/24"
  next_hop    = "10.0.99.2"
}

resource "cloudscale_router_extra_route" "router_b_to_net_a" {
  depends_on = [cloudscale_interface.router_b_peering]

  router_uuid = cloudscale_router.router_b.id
  destination = "10.0.1.0/24"
  next_hop    = "10.0.99.1"
}

# Server A on net_a
resource "cloudscale_server" "server_a" {
  name           = "server-a"
  zone_slug      = var.zone
  flavor_slug    = "flex-4-1"
  image_slug     = "debian-13"
  volume_size_gb = 10

  interfaces {
    type         = "private"
    network_uuid = cloudscale_network.net_a.id
  }

  ssh_keys                  = [var.ssh_public_key]
  allow_stopping_for_update = true
}

# Server B on net_b
resource "cloudscale_server" "server_b" {
  name           = "server-b"
  zone_slug      = var.zone
  flavor_slug    = "flex-4-1"
  image_slug     = "debian-13"
  volume_size_gb = 10

  interfaces {
    type         = "private"
    network_uuid = cloudscale_network.net_b.id
  }

  ssh_keys                  = [var.ssh_public_key]
  allow_stopping_for_update = true
}

# Jump host: public interface + private interface on peering network
# (Routes to net_a and net_b are injected via user_data, avoiding multiple
# default routes that would break SSH connectivity.)
resource "cloudscale_server" "jump_host" {
  name           = "jump-host"
  zone_slug      = var.zone
  flavor_slug    = "flex-4-1"
  image_slug     = "debian-13"
  volume_size_gb = 10

  interfaces {
    type         = "private"
    network_uuid = cloudscale_network.net_peering.id
  }

  interfaces {
    type = "public"
  }

  user_data                 = local.jump_host_user_data
  ssh_keys                  = [var.ssh_public_key]
  allow_stopping_for_update = true
}

# Output: jump host public IP for SSH access
output "jump_host_public_ipv4" {
  value = cloudscale_server.jump_host.public_ipv4_address
}

# Output: server private IPs (for SSH via jump host)
output "server_a_private_address" {
  value = cloudscale_server.server_a.interfaces[0].addresses[0].address
}

output "server_b_private_address" {
  value = cloudscale_server.server_b.interfaces[0].addresses[0].address
}
