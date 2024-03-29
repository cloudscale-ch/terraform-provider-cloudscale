package cloudscale

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"log"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

func TestAccCloudscaleNetwork_DetachedMinimal(t *testing.T) {
	var network cloudscale.Network

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: networkconfigMinimal(rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic", &network),
					testAccCheckCloudscaleNetworkSubnetCount("cloudscale_network.basic", &network, 1),
					resource.TestCheckResourceAttr("cloudscale_network.basic", "mtu", "9000"),
				),
			},
		},
	})
}

func TestAccCloudscaleNetwork_DetachedNoSubnet(t *testing.T) {
	var network cloudscale.Network

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: networkconfigMinimal(rInt, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic", &network),
					testAccCheckCloudscaleNetworkSubnetCount("cloudscale_network.basic", &network, 0),
					resource.TestCheckResourceAttr("cloudscale_network.basic", "mtu", "9000"),
				),
			},
		},
	})
}

func TestAccCloudscaleNetwork_DetachedWithZone(t *testing.T) {
	var network cloudscale.Network

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: networkconfigWithZone(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic", &network),
					testAccCheckCloudscaleNetworkSubnetCount("cloudscale_network.basic", &network, 1),
					resource.TestCheckResourceAttrSet(
						"cloudscale_network.basic", "href"),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: networkConfig_baseline(1, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic.0", &network),
					testAccCheckCloudscaleNetworkSubnetCount("cloudscale_network.basic", &network, 1),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic.0", "name", fmt.Sprintf("terraform-%d-0", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic.0", "mtu", "1500"),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic.0", "zone_slug", "rma1"),
				),
			},
			{
				Config: networkConfig_multiple_changes(1, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic.0", &network),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic.0", "name", fmt.Sprintf("terraform-%d-0-renamed", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic.0", "mtu", "9000"),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic.0", "zone_slug", "rma1"),
				),
			},
		},
	})
}

func TestAccCloudscaleNetwork_Attach(t *testing.T) {
	var network cloudscale.Network
	var server cloudscale.Server

	rInt1 := acctest.RandInt()
	rInt2 := acctest.RandInt()

	networkConfig := networkConfig_baseline(1, rInt1)
	serverConfig := serverConfigWithPrivateNetwork(rInt2, 0)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: networkConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic.0", &network),
				),
			},
			{
				Config: networkConfig + "\n" + serverConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic.0", &network),
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &server),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic",
						"interfaces.0.network_name",
						fmt.Sprintf("terraform-%d-0", rInt1),
					),
				),
			},
		},
	})
}

func TestAccCloudscaleNetwork_Reattach(t *testing.T) {
	var network0, network1 cloudscale.Network
	var server cloudscale.Server

	rInt1 := acctest.RandInt()
	rInt2 := acctest.RandInt()

	networkConfig := networkConfig_baseline(2, rInt1)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: networkConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic.0", &network0),
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic.1", &network1),
				),
			},
			{
				Config: networkConfig + "\n" + serverConfigWithPrivateNetwork(rInt2, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic.0", &network0),
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic.1", &network1),
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &server),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.#", "1"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.0.network_name", fmt.Sprintf("terraform-%d-0", rInt1)),
				),
			},
			{
				Config: networkConfig + "\n" + serverConfigWithPrivateNetwork(rInt2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic.0", &network0),
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic.1", &network1),
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &server),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.#", "1"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.0.network_name", fmt.Sprintf("terraform-%d-1", rInt1)),
				),
			},
			{
				Config: networkConfig + "\n" + serverConfigWithPrivateNetwork(rInt2, 1, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic.0", &network0),
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic.1", &network1),
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &server),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.#", "2"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.0.network_name", fmt.Sprintf("terraform-%d-1", rInt1)),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.1.network_name", fmt.Sprintf("terraform-%d-0", rInt1)),
				),
			},
			{
				Config: networkConfig + "\n" + serverConfigWithPrivateNetwork(rInt2, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic.0", &network0),
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic.1", &network1),
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &server),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.#", "1"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.0.network_name", fmt.Sprintf("terraform-%d-0", rInt1)),
				),
			},
		},
	})
}

func TestAccCloudscaleNetwork_ServerWithPublicAndPrivate(t *testing.T) {
	var network cloudscale.Network
	var server cloudscale.Server

	rInt1 := acctest.RandInt()
	rInt2 := acctest.RandInt()

	networkConfig := networkconfigWithZone(rInt1)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: networkConfig + "\n" + serverConfigWithPublicAndPrivate(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic", &network),
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &server),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.#", "2"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.0.type", "public"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.0.addresses.#", "2"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.1.type", "private"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.1.network_name", fmt.Sprintf("terraform-%d", rInt1)),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.1.addresses.#", "1"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.1.no_address", "false"),
				),
			},
		},
	})
}

func TestAccCloudscaleNetwork_ServerWithPublicAndPrivateWithoutAddress(t *testing.T) {
	var network cloudscale.Network
	var server cloudscale.Server

	rInt1 := acctest.RandInt()
	rInt2 := acctest.RandInt()

	networkConfig := networkconfigNoSubnet(rInt1)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: networkConfig + "\n" + serverConfigWithPublicAndPrivateNoAddress(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic", &network),
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &server),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.#", "2"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.0.type", "public"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.0.addresses.#", "2"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.1.type", "private"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.1.network_name", fmt.Sprintf("terraform-%d", rInt1)),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.1.addresses.#", "0"),
					resource.TestCheckResourceAttr("cloudscale_server.basic", "interfaces.1.no_address", "true"),
				),
			},
		},
	})
}

func TestAccCloudscaleNetwork_IdInput(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config:      networkConfigIdInput(),
				ExpectError: regexp.MustCompile(`Invalid or unknown key`),
			},
		},
	})
}

func TestAccCloudscaleNetwork_import_basic(t *testing.T) {
	var afterImport, afterUpdate cloudscale.Network

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: networkconfigWithZone(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic", &afterImport),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
				),
			},
			{
				ResourceName:      "cloudscale_network.basic",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "cloudscale_network.basic",
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateId:     "does-not-exist",
				ExpectError:       regexp.MustCompile(`Cannot import non-existent remote object`),
			},
			{
				Config: networkconfigWithZone(42),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("cloudscale_network.basic", &afterUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic", "name", "terraform-42"),
					testAccNetworkIsSame(t, &afterImport, &afterUpdate),
				),
			},
		},
	})
}

func TestAccCloudscaleNetwork_tags(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: networkconfigNoSubnetWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic", "tags.my-foo", "foo"),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic", "tags.my-bar", "bar"),
					testTagsMatch("cloudscale_network.basic"),
				),
			},
			{
				Config: networkconfigNoSubnet(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic", "tags.%", "0"),
					testTagsMatch("cloudscale_network.basic"),
				),
			},
			{
				Config: networkconfigNoSubnetWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic", "tags.my-foo", "foo"),
					resource.TestCheckResourceAttr(
						"cloudscale_network.basic", "tags.my-bar", "bar"),
					testTagsMatch("cloudscale_network.basic"),
				),
			},
		},
	})
}

func testAccNetworkIsSame(
	t *testing.T,
	before, after *cloudscale.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.UUID != after.UUID {
			t.Fatalf("Not expected a change of Network IDs got=%s, expected=%s",
				after.UUID, before.UUID)
		}
		return nil
	}
}

func testAccCheckCloudscaleNetworkSubnetCount(n string, network *cloudscale.Network, expectedCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if actualSubnetCount := len(network.Subnets); actualSubnetCount != expectedCount {
			return fmt.Errorf("Subnet count does not match, got=%#v, want=%#v.", actualSubnetCount, expectedCount)
		}
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

func networkconfigMinimal(rInt int, autoCreateSubnet bool) string {
	return fmt.Sprintf(`
resource "cloudscale_network" "basic" {
  name                    = "terraform-%d"
  auto_create_ipv4_subnet = "%t"
}`, rInt, autoCreateSubnet)
}

func networkConfig_baseline(count int, rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_network" "basic" {
  count        = "%v"
  name         = "terraform-%d-${count.index}"
  mtu          = "1500"
  zone_slug    = "rma1"
}`, count, rInt)
}

func networkConfig_multiple_changes(count int, rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_network" "basic" {
  count        = "%v"
  name         = "terraform-%d-${count.index}-renamed"
  mtu          = "9000"
  zone_slug    = "rma1"
}`, count, rInt)
}

func networkconfigWithZone(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_network" "basic" {
  name                    = "terraform-%d"
  mtu                     = "3421"
  zone_slug               = "lpg1"
}`, rInt)
}

func networkconfigNoSubnet(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_network" "basic" {
  name                    = "terraform-%d"
  zone_slug               = "lpg1"
  auto_create_ipv4_subnet = false
}`, rInt)
}

func networkconfigNoSubnetWithTags(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_network" "basic" {
  name                    = "terraform-%d"
  zone_slug               = "lpg1"
  auto_create_ipv4_subnet = false
  tags = {
    my-foo = "foo"
    my-bar = "bar"
  }
}`, rInt)
}

func serverConfigWithPrivateNetwork(rInt int, networkIndexes ...int) string {
	template := `
resource "cloudscale_server" "basic" {
  name      				= "terraform-%d"
  zone_slug                 = "rma1"
  flavor_slug    			= "flex-4-1"
  image_slug     			= "%s"
  %s
  volume_size_gb			= 10
  ssh_keys 					= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY=", "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`

	var interfaceConfigs strings.Builder
	for _, networkIndex := range networkIndexes {
		interfaceConfigs.WriteString(
			fmt.Sprintf(`
interfaces                {
  type                    = "private"
  network_uuid            = "${cloudscale_network.basic.%v.id}"
}`, networkIndex))
	}

	result := fmt.Sprintf(template, rInt, DefaultImageSlug, interfaceConfigs.String())
	return result
}

func serverConfigWithPublicAndPrivate(rInt int) string {
	template := `
resource "cloudscale_server" "basic" {
  name      				= "terraform-%d"
  zone_slug                 = "lpg1"
  flavor_slug    			= "flex-4-1"
  image_slug     			= "%s"
  interfaces                {
    type                    = "public"
  }
  interfaces                {
    type                    = "private"
    network_uuid            = "${cloudscale_network.basic.id}"
  }
  volume_size_gb			= 10
  ssh_keys 					= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY=", "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`
	return fmt.Sprintf(template, rInt, DefaultImageSlug)
}

func serverConfigWithPublicAndPrivateNoAddress(rInt int) string {
	template := `
resource "cloudscale_server" "basic" {
  name      				= "terraform-%d"
  zone_slug                 = "lpg1"
  flavor_slug    			= "flex-4-1"
  image_slug     			= "%s"
  interfaces                {
    type                    = "public"
  }
  interfaces                {
    type                    = "private"
    network_uuid            = "${cloudscale_network.basic.id}"
    no_address              = true
  }
  volume_size_gb			= 10
  ssh_keys 					= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY=", "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`
	return fmt.Sprintf(template, rInt, DefaultImageSlug)
}

func networkConfigIdInput() string {
	return fmt.Sprintf(`
resource "cloudscale_network" "basic" {
  name = "terraform-0"
  id   = "1"
}`)
}
