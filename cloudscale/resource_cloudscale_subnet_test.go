package cloudscale

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v4"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	// it's sufficient to sweep networks
}

func TestAccCloudscaleSubnet_Minimal(t *testing.T) {
	var subnet cloudscale.Subnet
	var network cloudscale.Network

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: subnetconfigMinimal(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic", &network),
					testAccCheckCloudscaleSubnetExists("cloudscale_subnet.basic", &subnet),
					testAccCheckCloudscaleSubnetOnNetwork(&subnet, &network),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "cidr", "10.11.12.0/24"),
					resource.TestCheckResourceAttrSet("cloudscale_subnet.basic", "network_href"),
					resource.TestCheckResourceAttrSet("cloudscale_subnet.basic", "network_uuid"),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "network_name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "gateway_address", ""),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "dns_servers.#", "2"),
					resource.TestCheckResourceAttrSet("cloudscale_subnet.basic", "dns_servers.0"),
					resource.TestCheckResourceAttrSet("cloudscale_subnet.basic", "dns_servers.1"),
				),
			},
		},
	})
}

func TestAccCloudscaleSubnet_AllAttrs(t *testing.T) {
	var subnet cloudscale.Subnet
	var network cloudscale.Network

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: networkconfigMinimal(rInt, false) + "\n" + subnetconfigAllAttrs(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic", &network),
					testAccCheckCloudscaleSubnetExists("cloudscale_subnet.basic", &subnet),
					testAccCheckCloudscaleSubnetOnNetwork(&subnet, &network),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "cidr", "10.11.12.0/24"),
					resource.TestCheckResourceAttrSet("cloudscale_subnet.basic", "network_href"),
					resource.TestCheckResourceAttrSet("cloudscale_subnet.basic", "network_uuid"),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "network_name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "gateway_address", "10.11.12.99"),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "dns_servers.#", "2"),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "dns_servers.0", "8.8.4.4"),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "dns_servers.1", "8.8.8.8"),
				),
			},
		},
	})
}

func TestAccCloudscaleSubnet_Update(t *testing.T) {
	var subnet cloudscale.Subnet
	var network cloudscale.Network

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: subnetconfigMinimal(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic", &network),
					testAccCheckCloudscaleSubnetExists("cloudscale_subnet.basic", &subnet),
					testAccCheckCloudscaleSubnetOnNetwork(&subnet, &network),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "cidr", "10.11.12.0/24"),
					resource.TestCheckResourceAttrSet("cloudscale_subnet.basic", "network_href"),
					resource.TestCheckResourceAttrSet("cloudscale_subnet.basic", "network_uuid"),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "network_name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "gateway_address", ""),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "dns_servers.#", "2"),
					resource.TestCheckResourceAttrSet("cloudscale_subnet.basic", "dns_servers.0"),
					resource.TestCheckResourceAttrSet("cloudscale_subnet.basic", "dns_servers.1"),
				),
			},
			{
				Config: subnetconfigUpdated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic", &network),
					testAccCheckCloudscaleSubnetExists("cloudscale_subnet.basic", &subnet),
					testAccCheckCloudscaleSubnetOnNetwork(&subnet, &network),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "cidr", "10.11.12.0/24"),
					resource.TestCheckResourceAttrSet("cloudscale_subnet.basic", "network_href"),
					resource.TestCheckResourceAttrSet("cloudscale_subnet.basic", "network_uuid"),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "network_name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "gateway_address", "10.11.12.10"),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "dns_servers.#", "3"),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "dns_servers.0", "1.2.3.4"),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "dns_servers.1", "5.6.7.8"),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "dns_servers.2", "9.10.11.12"),
				),
			},
		},
	})
}

func TestAccCloudscaleSubnet_ServerWithPublicAndPrivate(t *testing.T) {
	var network cloudscale.Network
	var subnet cloudscale.Subnet
	var server cloudscale.Server

	rInt1 := acctest.RandInt()
	rInt2 := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: subnetconfigMinimal(rInt1) + "\n" + serverConfigWithPublicAndLayerThree(rInt2, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic", &network),
					testAccCheckCloudscaleSubnetExists("cloudscale_subnet.basic", &subnet),
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &server),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.#", "2"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.0.type", "public"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.0.addresses.#", "2"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.1.type", "private"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.1.network_name", fmt.Sprintf("terraform-%d", rInt1)),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.1.addresses.#", "1"),
					resource.TestCheckResourceAttrSet("cloudscale_server.basic", "interfaces.1.addresses.0.subnet_uuid"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.1.addresses.0.subnet_cidr", "10.11.12.0/24"),
					resource.TestCheckResourceAttrSet("cloudscale_server.basic", "interfaces.1.addresses.0.subnet_href"),
				),
			},
			{
				Config: subnetconfigMinimal(rInt1) + "\n" + serverConfigWithPublicAndLayerThree(rInt2, "10.11.12.13"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic", &network),
					testAccCheckCloudscaleSubnetExists("cloudscale_subnet.basic", &subnet),
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &server),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.#", "2"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.0.type", "public"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.0.addresses.#", "2"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.1.type", "private"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.1.network_name", fmt.Sprintf("terraform-%d", rInt1)),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.1.addresses.#", "1"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.1.addresses.0.address", "10.11.12.13"),
					resource.TestCheckResourceAttrSet("cloudscale_server.basic", "interfaces.1.addresses.0.subnet_uuid"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.1.addresses.0.subnet_cidr", "10.11.12.0/24"),
					resource.TestCheckResourceAttrSet("cloudscale_server.basic", "interfaces.1.addresses.0.subnet_href"),
				),
			},
		},
	})
}

func TestAccCloudscaleSubnet_ServerAndMultipleSubnets(t *testing.T) {
	count := 2
	networks := make([]cloudscale.Network, count, count)
	subnets := make([]cloudscale.Subnet, count, count)
	var server cloudscale.Server

	rInt1 := acctest.RandInt()
	rInt2 := acctest.RandInt()
	uniqueInt := 2

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: multipleSubnetConfig(rInt1, rInt2, 0, 0, uniqueInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.multi-net.0", &networks[0]),
					testAccCheckCloudscaleNetworkExists("cloudscale_network.multi-net.1", &networks[1]),
					testAccCheckCloudscaleSubnetExists("cloudscale_subnet.multi-subnet.0", &subnets[0]),
					testAccCheckCloudscaleSubnetExists("cloudscale_subnet.multi-subnet.1", &subnets[1]),
					testAccCheckCloudscaleSubnetOnNetwork(&subnets[0], &networks[0]),
					testAccCheckCloudscaleSubnetOnNetwork(&subnets[1], &networks[1]),

					testAccCheckCloudscaleServerExists("cloudscale_server.web-worker01", &server),
					testAccCheckCloudscaleAddressOnSubnet(&server, &subnets[0], 0, 0),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.#", "1"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.type", "private"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.network_name", fmt.Sprintf("terraform-%d-0", rInt1)),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.addresses.#", "1"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.addresses.0.address", "10.2.0.124"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.addresses.0.subnet_cidr", "10.2.0.0/24"),
				),
			},
			{
				Config: multipleSubnetConfig(rInt1, rInt2, 1, 1, uniqueInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.multi-net.0", &networks[0]),
					testAccCheckCloudscaleNetworkExists("cloudscale_network.multi-net.1", &networks[1]),
					testAccCheckCloudscaleSubnetExists("cloudscale_subnet.multi-subnet.0", &subnets[0]),
					testAccCheckCloudscaleSubnetExists("cloudscale_subnet.multi-subnet.1", &subnets[1]),
					testAccCheckCloudscaleSubnetOnNetwork(&subnets[0], &networks[0]),
					testAccCheckCloudscaleSubnetOnNetwork(&subnets[1], &networks[1]),

					testAccCheckCloudscaleServerExists("cloudscale_server.web-worker01", &server),
					testAccCheckCloudscaleAddressOnSubnet(&server, &subnets[1], 0, 0),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.#", "1"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.type", "private"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.network_name", fmt.Sprintf("terraform-%d-1", rInt1)),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.addresses.#", "1"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.addresses.0.address", "10.2.1.124"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.addresses.0.subnet_cidr", "10.2.1.0/24"),
				),
			},
			{
				Config: multipleSubnetConfig(rInt1, rInt2, 0, -100, uniqueInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.multi-net.0", &networks[0]),
					testAccCheckCloudscaleNetworkExists("cloudscale_network.multi-net.1", &networks[1]),
					testAccCheckCloudscaleSubnetExists("cloudscale_subnet.multi-subnet.0", &subnets[0]),
					testAccCheckCloudscaleSubnetExists("cloudscale_subnet.multi-subnet.1", &subnets[1]),
					testAccCheckCloudscaleSubnetOnNetwork(&subnets[0], &networks[0]),
					testAccCheckCloudscaleSubnetOnNetwork(&subnets[1], &networks[1]),

					testAccCheckCloudscaleServerExists("cloudscale_server.web-worker01", &server),
					testAccCheckCloudscaleAddressOnSubnet(&server, &subnets[0], 0, 0),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.#", "1"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.type", "private"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.network_name", fmt.Sprintf("terraform-%d-0", rInt1)),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.addresses.#", "1"),
					resource.TestCheckResourceAttrSet("cloudscale_server.web-worker01", "interfaces.0.addresses.0.address"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.addresses.0.subnet_cidr", "10.2.0.0/24"),
				),
			},
			{
				Config: multipleSubnetConfig(rInt1, rInt2, 0, 0, uniqueInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.multi-net.0", &networks[0]),
					testAccCheckCloudscaleNetworkExists("cloudscale_network.multi-net.1", &networks[1]),
					testAccCheckCloudscaleSubnetExists("cloudscale_subnet.multi-subnet.0", &subnets[0]),
					testAccCheckCloudscaleSubnetExists("cloudscale_subnet.multi-subnet.1", &subnets[1]),
					testAccCheckCloudscaleSubnetOnNetwork(&subnets[0], &networks[0]),
					testAccCheckCloudscaleSubnetOnNetwork(&subnets[1], &networks[1]),

					testAccCheckCloudscaleServerExists("cloudscale_server.web-worker01", &server),
					testAccCheckCloudscaleAddressOnSubnet(&server, &subnets[0], 0, 0),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.#", "1"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.type", "private"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.network_name", fmt.Sprintf("terraform-%d-0", rInt1)),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.addresses.#", "1"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.addresses.0.address", "10.2.0.124"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.addresses.0.subnet_cidr", "10.2.0.0/24"),
				),
			},
			{
				Config: multipleSubnetConfig(rInt1, rInt2, -100, 1, uniqueInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.multi-net.0", &networks[0]),
					testAccCheckCloudscaleNetworkExists("cloudscale_network.multi-net.1", &networks[1]),
					testAccCheckCloudscaleSubnetExists("cloudscale_subnet.multi-subnet.0", &subnets[0]),
					testAccCheckCloudscaleSubnetExists("cloudscale_subnet.multi-subnet.1", &subnets[1]),
					testAccCheckCloudscaleSubnetOnNetwork(&subnets[0], &networks[0]),
					testAccCheckCloudscaleSubnetOnNetwork(&subnets[1], &networks[1]),

					testAccCheckCloudscaleServerExists("cloudscale_server.web-worker01", &server),
					testAccCheckCloudscaleAddressOnSubnet(&server, &subnets[1], 0, 0),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.#", "1"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.type", "private"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.network_name", fmt.Sprintf("terraform-%d-1", rInt1)),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.addresses.#", "1"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.addresses.0.address", "10.2.1.124"),
					resource.TestCheckResourceAttr("cloudscale_server.web-worker01", "interfaces.0.addresses.0.subnet_cidr", "10.2.1.0/24"),
				),
			},
		},
	})
}

func TestAccCloudscaleSubnet_import_basic(t *testing.T) {
	var afterImport, afterUpdate cloudscale.Subnet

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: subnetconfigMinimal(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleSubnetExists("cloudscale_subnet.basic", &afterImport),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "gateway_address", ""),
				),
			},
			{
				ResourceName:      "cloudscale_subnet.basic",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "cloudscale_subnet.basic",
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateId:     "does-not-exist",
				ExpectError:       regexp.MustCompile(`Cannot import non-existent remote object`),
			},
			{
				Config: subnetconfigUpdated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleSubnetExists("cloudscale_subnet.basic", &afterUpdate),
					resource.TestCheckResourceAttr("cloudscale_subnet.basic", "gateway_address", "10.11.12.10"),
					testAccSubnetIsSame(t, &afterImport, &afterUpdate),
				),
			},
		},
	})
}

func TestAccCloudscaleSubnet_tags(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: subnetconfigMinimalWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_subnet.basic", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"cloudscale_subnet.basic", "tags.my-foo", "foo"),
					resource.TestCheckResourceAttr(
						"cloudscale_subnet.basic", "tags.my-bar", "bar"),
					testTagsMatch("cloudscale_subnet.basic"),
				),
			},
			{
				Config: subnetconfigMinimal(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_subnet.basic", "tags.%", "0"),
					testTagsMatch("cloudscale_subnet.basic"),
				),
			},
			{
				Config: subnetconfigMinimalWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_subnet.basic", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"cloudscale_subnet.basic", "tags.my-foo", "foo"),
					resource.TestCheckResourceAttr(
						"cloudscale_subnet.basic", "tags.my-bar", "bar"),
					testTagsMatch("cloudscale_subnet.basic"),
				),
			},
		},
	})
}

func testAccSubnetIsSame(t *testing.T,
	before, after *cloudscale.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.UUID != after.UUID {
			t.Fatalf("Not expected a change of Subnet IDs got=%s, expected=%s",
				after.UUID, before.UUID)
		}
		return nil
	}
}

func testAccCheckCloudscaleSubnetOnNetwork(subnet *cloudscale.Subnet, network *cloudscale.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if subnet.Network.UUID != network.UUID {
			return fmt.Errorf("Subnet not on expected Network got=%s, expected=%s", network.UUID, subnet.Network.UUID)
		}

		return nil
	}
}

func testAccCheckCloudscaleAddressOnSubnet(server *cloudscale.Server, subnet *cloudscale.Subnet, iIndex int, aIndex int) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if uuid := server.Interfaces[iIndex].Addresses[aIndex].Subnet.UUID; uuid != subnet.UUID {
			return fmt.Errorf("Address not on expected subnet got=%s, expected=%s", uuid, subnet.UUID)
		}

		return nil
	}
}

func testAccCheckCloudscaleSubnetDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudscale.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudscale_subnet" {
			continue
		}

		id := rs.Primary.ID

		// Try to find the subnet
		v, err := client.Subnets.Get(context.Background(), id)
		if err == nil {
			return fmt.Errorf("Subnet %v still exists", v)
		} else {
			errorResponse, ok := err.(*cloudscale.ErrorResponse)
			if !ok || errorResponse.StatusCode != http.StatusNotFound {
				return fmt.Errorf(
					"Error waiting for subnet (%s) to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func subnetconfigMinimal(rInt int) string {
	return networkconfigMinimal(rInt, false) + fmt.Sprintf(`
resource "cloudscale_subnet" "basic" {
  cidr            = "10.11.12.0/24"
  network_uuid    = cloudscale_network.basic.id
}
`)
}

func subnetconfigMinimalWithTags(rInt int) string {
	return networkconfigMinimal(rInt, false) + fmt.Sprintf(`
resource "cloudscale_subnet" "basic" {
  cidr            = "10.11.12.0/24"
  network_uuid    = cloudscale_network.basic.id
  tags = {
    my-foo = "foo"
    my-bar = "bar"
  }
}
`)
}

func subnetconfigAllAttrs() string {
	return fmt.Sprintf(`
resource "cloudscale_subnet" "basic" {
  cidr            = "10.11.12.0/24"
  network_uuid    = cloudscale_network.basic.id
  dns_servers     = ["8.8.4.4", "8.8.8.8"]
  gateway_address = "10.11.12.99"
}
`)
}

func subnetconfigUpdated(rInt int) string {
	return networkconfigMinimal(rInt, false) + "\n" + fmt.Sprintf(`
resource "cloudscale_subnet" "basic" {
  cidr         	  = "10.11.12.0/24"
  network_uuid 	  = cloudscale_network.basic.id
  gateway_address = "10.11.12.10"
  dns_servers     = ["1.2.3.4", "5.6.7.8", "9.10.11.12"]
}
`)
}

func serverConfigWithPublicAndLayerThree(rInt int, fixedAddress string) string {
	template := `
resource "cloudscale_server" "basic" {
  name      				= "terraform-%d"
  zone_slug                 = "rma1"
  flavor_slug    			= "flex-4-1"
  image_slug     			= "%s"
  interfaces                {
    type                    = "public"
  }
  interfaces                {
    type                    = "private"
    addresses {
      subnet_uuid           = "${cloudscale_subnet.basic.id}"     
      %s
    }
  }
  volume_size_gb			= 10
  ssh_keys 					= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY=", "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`
	fixedAddressLine := ""
	if fixedAddress != "" {
		fixedAddressLine = fmt.Sprintf(`address               = "%s"`, fixedAddress)
	}

	return fmt.Sprintf(template, rInt, DefaultImageSlug, fixedAddressLine)
}

func subnetConfig_baseline(count int, rInt int, uniqueSubnetIdentifier int) string {
	return fmt.Sprintf(`
resource "cloudscale_network" "multi-net" {
  count = %d
  name = "terraform-%d-${count.index}"
  auto_create_ipv4_subnet = false
}

resource "cloudscale_subnet" "multi-subnet" {
  count = 2
  cidr = "10.%d.${count.index}.0/24"
  network_uuid = cloudscale_network.multi-net[count.index].id
}`, count, rInt, uniqueSubnetIdentifier)
}

func multipleSubnetConfig(rInt1 int, rInt2 int, networkIndex int, subnetIndex int, uniqueSubnetIdentifier int) string {
	template := `
resource "cloudscale_server" "web-worker01" {
 name = "terraform-%d"
 flavor_slug = "flex-8-4"
 image_slug = "%s"
 volume_size_gb = 50
 interfaces {
   type = "private"
   %s
   %s
 }
 ssh_keys = [
   "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIL2jzgla23DfRVLQr3KT20QQYovqCCN3clHrjm2ZuQFW user@example.com"
 ]
}`

	networkTemplate := ""
	if networkIndex >= 0 {
		networkTemplate = fmt.Sprintf(`
   network_uuid = cloudscale_network.multi-net[%d].id
`, networkIndex)
	}

	addressTemplate := ""
	if subnetIndex >= 0 {
		addressTemplate = fmt.Sprintf(`
   addresses {
     address       = "10.%d.%d.124"
     subnet_uuid   = cloudscale_subnet.multi-subnet[%d].id
   }`, uniqueSubnetIdentifier, subnetIndex, subnetIndex)
	}

	return subnetConfig_baseline(2, rInt1, uniqueSubnetIdentifier) + fmt.Sprintf(template, rInt2, DefaultImageSlug, networkTemplate, addressTemplate)
}
