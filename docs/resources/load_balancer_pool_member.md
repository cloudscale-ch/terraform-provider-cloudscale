---
page_title: "cloudscale.ch: cloudscale_load_balancer_pool_member"
---

# cloudscale\_load\_balancer\_member

Provides a cloudscale.ch load balancer pool member resource. This can be used to create, modify, import, and delete load balancer pool members. 

## Example Usage

```hcl
# Create a new network
resource "cloudscale_network" "backend" {
  name                    = "backend"
  zone_slug               = "lpg1"
  auto_create_ipv4_subnet = "false"
}

# Create a new subnet
resource "cloudscale_subnet" "backend-subnet" {
  cidr         = "10.11.12.0/24"
  network_uuid = cloudscale_network.backend.id
}

# Create new workers
resource "cloudscale_server" "web-worker" {
  count       = 2
  name        = "web-worker${count.index}"
  flavor_slug = "flex-4-2"
  image_slug  = "ubuntu-22.04"
  zone_slug   = "lpg1"
  ssh_keys    = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIL2jzgla23DfRVLQr3KT20QQYovqCCN3clHrjm2ZuQFW user@example.com"]


  interfaces {
    type = "public"
  }

  interfaces {
    type = "private"
    addresses {
      subnet_uuid = cloudscale_subnet.backend-subnet.id
    }
  }
}

# Create a new load balancer
resource "cloudscale_load_balancer" "lb1" {
  name        = "web-lb1"
  flavor_slug = "lb-standard"
  zone_slug   = "lpg1"
}

# Create a new load balancer pool
resource "cloudscale_load_balancer_pool" "lb1-pool" {
  name               = "web-lb1-pool"
  algorithm          = "round_robin"
  protocol           = "tcp"
  load_balancer_uuid = cloudscale_load_balancer.lb1.id
}

# Create a new load balancer pool member
resource "cloudscale_load_balancer_pool_member" "lb1-pool-member" {
  count         = 2
  name          = "web-lb1-pool-member-${count.index}"
  pool_uuid     = cloudscale_load_balancer_pool.lb1-pool.id
  protocol_port = 80
  address       = cloudscale_server.web-worker[count.index].interfaces[1].addresses[0].address
  subnet_uuid   = cloudscale_subnet.backend-subnet.id
}
```

## Argument Reference

The following arguments are supported when creating new load balancer pool:

* `name` - (Required) Name of the new load balancer pool member.
* `enabled` - (Optional) Pool member will not receive traffic if `false`. Default is `true`.
* `pool_uuid` - (Required) The load balancer pool of the member.
* `protocol_port` - (Required) The port to which actual traffic is sent.
* `monitor_port` - (Optional) The port to which health monitor checks are sent. If not specified, `protocol_port` will be used.
* `address` - (Required) The IP address to which traffic is sent.
* `subnet_uuid` - (Required) The subnet UUID of the address must be specified here.
* `tags` - (Optional) Tags allow you to assign custom metadata to resources:
  ```hcl
  tags = {
    foo = "bar"
  }
  ```
  Tags are always strings (both keys and values).

The following arguments are supported when updating load balancer pool members:

* `name` - New name of the load balancer pool.
* `enabled` - Pool member will not receive traffic if `false`.
* `tags` - Change tags (see documentation above)

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The UUID of this load balancer pool member.
* `href` - The cloudscale.ch API URL of the current resource.
* `monitor_status` - The status of the pool's health monitor check for this member. Can be `"up"`, `"down"`, `"changing"`, `"no_monitor"` and `"unknown"`.
* `pool_name` - The load balancer pool name of the member.
* `pool_href` - The cloudscale.ch API URL of the member's load balancer pool.
* `subnet_cidr` - The CIDR of the member's address subnet.
* `subnet_href` - The cloudscale.ch API URL of the member's address subnet.


## Import

Load balancer pool members can be imported using the load balancer pool member's UUID and the pool UUID
using this schema `{pool_uuid}.{member_uuid}`:

```
terraform import cloudscale_load_balancer_pool_member.lb1-pool-member 48151623-42aa-aaaa-bbbb-caffeeeeeeee.6a18a377-9977-4cd0-b1fa-70908356efaa
```
