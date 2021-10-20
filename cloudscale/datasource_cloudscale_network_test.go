package cloudscale

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudScaleNetwork_DS_Basic(t *testing.T) {
	var network cloudscale.Network
	rInt := acctest.RandInt()
	name1 := fmt.Sprintf("terraform-%d-0", rInt)
	name2 := fmt.Sprintf("terraform-%d-1", rInt)
	config := networkConfig_baseline(2, rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				Config: config + testAccCheckCloudScaleNetworkConfig_name(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleNetworkExists("data.cloudscale_network.foo", &network),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_network.foo", "id", &network.UUID),
					resource.TestCheckResourceAttr(
						"data.cloudscale_network.foo", "name", name1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_network.foo", "mtu", "1500"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_network.foo", "zone_slug", "rma1"),
					resource.TestCheckResourceAttrSet(
						"data.cloudscale_network.foo", "subnets.0.uuid"),
					resource.TestCheckResourceAttrSet(
						"data.cloudscale_network.foo", "subnets.0.cidr"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_network.foo", "subnets.#", "1"),
				),
			},
			{
				Config: config + testAccCheckCloudScaleNetworkConfig_name_and_zone(name1, "rma1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_network.foo", "name", name1),
				),
			},
			{
				Config: config + testAccCheckCloudScaleNetworkConfig_name(name2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_network.foo", "name", name2),
					resource.TestCheckResourceAttr(
						"data.cloudscale_network.foo", "mtu", "1500"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_network.foo", "zone_slug", "rma1"),
					resource.TestCheckResourceAttrSet(
						"data.cloudscale_network.foo", "subnets.0.cidr"),
					resource.TestCheckResourceAttrSet(
						"data.cloudscale_network.foo", "subnets.0.cidr"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_network.foo", "subnets.#", "1"),
				),
			},
			{
				Config: config + testAccCheckCloudScaleNetworkConfig_name_and_zone(name1, "rma1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_network.foo", "name", name1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_network.foo", "zone_slug", "rma1"),
				),
			},
			{
				Config: config + testAccCheckCloudScaleNetworkConfig_name_and_zone(name1, "lpg1"),
				ExpectError: regexp.MustCompile(`.*Found zero networks.*`),
			},
			{
				Config: config + "\n" + `data "cloudscale_network" "foo" {}`,
				ExpectError: regexp.MustCompile(`Found \d+ networks, expected one`),
			},
		},
	})
}

func TestAccCloudScaleNetwork_DS_NotExisting(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudScaleNetworkConfig_name("terraform-unknown-network"),
				ExpectError: regexp.MustCompile(`.*Found zero networks.*`),
			},
		},
	})
}

func testAccCheckCloudScaleNetworkConfig_name(name string) string {
	return fmt.Sprintf(`
data "cloudscale_network" "foo" {
  name               = "%s"
}
`, name)
}

func testAccCheckCloudScaleNetworkConfig_name_and_zone(name, zone_slug string) string {
	return fmt.Sprintf(`
data "cloudscale_network" "foo" {
  name               = "%s"
  zone_slug			 = "%s"
}
`, name, zone_slug)
}
