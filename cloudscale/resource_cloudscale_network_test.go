package cloudscale

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func init() {
	resource.AddTestSweepers("cloudscale_network", &resource.Sweeper{
		Name: "cloudscale_network",
		F:    testSweepNetworks,
	})
}

func testSweepNetworks(region string) error {
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client := meta.(*cloudscale.Client)

	networks, err := client.Networks.List(context.Background())
	if err != nil {
		return err
	}

	foundError := error(nil)
	for _, s := range networks {
		if strings.HasPrefix(s.Name, "terraform-") {
			log.Printf("Destroying Network %s", s.Name)

			if err := client.Networks.Delete(context.Background(), s.UUID); err != nil {
				foundError = err
			}
		}
	}
	return foundError
}

func TestAccCloudscaleNetwork_DetachedWithZone(t *testing.T) {
	var network cloudscale.Network

	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: networkconfigWithZone(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic", &network),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic", "mtu", "3421"),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic", "zone_slug", "lpg1"),
				),
			},
		},
	})
}

func TestAccCloudscaleNetwork_Change(t *testing.T) {
	var network cloudscale.Network

	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: networkConfig_baseline(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic", &network),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic", "mtu", "1500"),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic", "zone_slug", "rma1"),
				),
			},
			{
				Config: networkConfig_multiple_changes(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic", &network),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic", "name", fmt.Sprintf("terraform-%d-renamed", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic", "mtu", "9000"),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic", "zone_slug", "rma1"),
				),
			},
		},
	})
}

func testAccCheckCloudscaleNetworkExists(n string, network *cloudscale.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network ID is set")
		}

		client := testAccProvider.Meta().(*cloudscale.Client)

		id := rs.Primary.ID

		// Try to find the network
		retrieveNetwork, err := client.Networks.Get(context.Background(), id)

		if err != nil {
			return err
		}

		if retrieveNetwork.UUID != rs.Primary.ID {
			return fmt.Errorf("Network not found")
		}

		*network = *retrieveNetwork

		return nil
	}
}

func testAccCheckCloudscaleNetworkDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudscale.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudscale_network" {
			continue
		}

		id := rs.Primary.ID

		// Try to find the network
		v, err := client.Networks.Get(context.Background(), id)
		if err == nil {
			return fmt.Errorf("Network %v still exists", v)
		} else {
			errorResponse, ok := err.(*cloudscale.ErrorResponse)
			if !ok || errorResponse.StatusCode != http.StatusNotFound {
				return fmt.Errorf(
					"Error waiting for network (%s) to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func networkConfig_baseline(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_network" "basic" {
  name         = "terraform-%d"
  mtu          = "1500"
  zone_slug    = "rma1"
}`, rInt)
}

func networkConfig_multiple_changes(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_network" "basic" {
  name         = "terraform-%d-renamed"
  mtu          = "9000"
  zone_slug    = "rma1"
}`, rInt)
}

func networkconfigWithZone(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_network" "basic" {
  name         = "terraform-%d"
  mtu          = "3421"
  zone_slug    = "lpg1"
}`, rInt)
}
