# Set the variable value in *.tfvars file
# or using the -var="cloudscale_token=..." CLI option
variable "cloudscale_token" {}

# Configure the cloudscale.ch Provider
provider "cloudscale" {
  token = "${var.cloudscale_token}"
}

# Create a New Server
resource "cloudscale_server" "web" {
  name      			= "db-master"
  flavor    			= "flex-4"
  image     			= "debian-8"
  volume_size_gb	= 10
  ssh_keys = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}