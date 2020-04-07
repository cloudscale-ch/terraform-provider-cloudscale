package cloudscale

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
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
				Config: networkconfigMinimal(rInt, false) + "\n" + subnetconfigMinimal(),
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
				Config: networkconfigMinimal(rInt, false) + "\n" + subnetconfigMinimal(),
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
				Config: networkconfigMinimal(rInt, false) + "\n" + subnetconfigUpdated(),
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

func testAccCheckCloudscaleSubnetOnNetwork(subnet *cloudscale.Subnet, network *cloudscale.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if subnet.Network.UUID != network.UUID {
			return fmt.Errorf("Subnet not on expected Network got=%s, expected=%s", network.UUID, subnet.Network.UUID)
		}

		return nil
	}
}

func testAccCheckCloudscaleSubnetExists(n string, subnet *cloudscale.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Subnet ID is set")
		}

		client := testAccProvider.Meta().(*cloudscale.Client)

		id := rs.Primary.ID

		// Try to find the subnet
		retrieveSubnet, err := client.Subnets.Get(context.Background(), id)

		if err != nil {
			return err
		}

		if retrieveSubnet.UUID != rs.Primary.ID {
			return fmt.Errorf("Subnet not found")
		}

		*subnet = *retrieveSubnet

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

func subnetconfigMinimal() string {
	return fmt.Sprintf(`
resource "cloudscale_subnet" "basic" {
  cidr            = "10.11.12.0/24"
  network_uuid    = cloudscale_network.basic.id
}
`)
}

func subnetconfigUpdated() string {
	return fmt.Sprintf(`
resource "cloudscale_subnet" "basic" {
  cidr         	  = "10.11.12.0/24"
  network_uuid 	  = cloudscale_network.basic.id
  gateway_address = "10.11.12.10"
  dns_servers     = ["1.2.3.4", "5.6.7.8", "9.10.11.12"]
}
`)
}
