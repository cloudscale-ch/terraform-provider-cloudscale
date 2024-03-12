package cloudscale

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudscaleSubnet_DS_Basic(t *testing.T) {
	var subnet cloudscale.Subnet
	rInt := acctest.RandInt()
	uniqueInt := 1
	cidr1 := fmt.Sprintf("10.%d.0.0/24", uniqueInt)
	cidr2 := fmt.Sprintf("10.%d.1.0/24", uniqueInt)
	config := subnetConfig_baseline(2, rInt, uniqueInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				Config: config + testAccCheckCloudscaleSubnetConfig_cidr(cidr1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleSubnetExists("data.cloudscale_subnet.foo", &subnet),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_subnet.multi-subnet.0", "id", &subnet.UUID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_subnet.foo", "id", &subnet.UUID),
					resource.TestCheckResourceAttr(
						"data.cloudscale_subnet.foo", "cidr", cidr1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_subnet.foo", "gateway_address", ""),
					resource.TestCheckResourceAttr(
						"data.cloudscale_subnet.foo", "network_name", fmt.Sprintf(`terraform-%d-0`, rInt)),
					resource.TestCheckResourceAttrSet(
						"data.cloudscale_subnet.foo", "network_uuid"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_subnet.foo", "dns_servers.#", "2"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleSubnetConfig_cidr_and_gateway(cidr1, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_subnet.foo", "cidr", cidr1),
				),
			},
			{
				Config: config + testAccCheckCloudscaleSubnetConfig_cidr(cidr2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_subnet.foo", "cidr", cidr2),
					resource.TestCheckResourceAttr(
						"data.cloudscale_subnet.foo", "gateway_address", ""),
				),
			},
			{
				Config: config + testAccCheckCloudscaleSubnetConfig_cidr_and_gateway(cidr1, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_subnet.foo", "cidr", cidr1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_subnet.foo", "gateway_address", ""),
				),
			},
			{
				Config:      config + testAccCheckCloudscaleSubnetConfig_cidr_and_gateway(cidr1, "1.1.1.1"),
				ExpectError: regexp.MustCompile(`.*Found zero subnets.*`),
			},
			{
				Config: config + testAccCheckCloudscaleSubnetConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_subnet.multi-subnet.0", "cidr", cidr1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_subnet.foo", "cidr", cidr1),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_subnet.multi-subnet.0", "id", &subnet.UUID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_subnet.foo", "id", &subnet.UUID),
				),
			},
			{
				Config: config + testAccCheckCloudscaleSubnetConfig_network_uuid(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_subnet.foo", "cidr", cidr1),
				),
			},
			{
				Config: config + testAccCheckCloudscaleSubnetConfig_network_name(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_subnet.foo", "cidr", cidr1),
				),
			},
			{
				Config:      config + "\n" + `data "cloudscale_subnet" "foo" {}`,
				ExpectError: regexp.MustCompile(`Found \d+ subnets, expected one`),
			},
		},
	})
}

func TestAccCloudscaleSubnet_DS_NotExisting(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckCloudscaleSubnetConfig_cidr("terraform-unknown-subnet"),
				ExpectError: regexp.MustCompile(`.*Found zero subnets.*`),
			},
		},
	})
}

func testAccCheckCloudscaleSubnetConfig_cidr(cidr string) string {
	return fmt.Sprintf(`
data "cloudscale_subnet" "foo" {
  cidr = "%s"
}
`, cidr)
}

func testAccCheckCloudscaleSubnetConfig_cidr_and_gateway(cidr, gateway string) string {
	return fmt.Sprintf(`
data "cloudscale_subnet" "foo" {
  cidr            = "%s"
  gateway_address = "%s"
}
`, cidr, gateway)
}

func testAccCheckCloudscaleSubnetConfig_id() string {
	return fmt.Sprintf(`
data "cloudscale_subnet" "foo" {
  id = "${cloudscale_subnet.multi-subnet.0.id}"
}
`)
}

func testAccCheckCloudscaleSubnetConfig_network_uuid() string {
	return `
data "cloudscale_subnet" "foo" {
  network_uuid = "${cloudscale_network.multi-net.0.id}"
}
`
}

func testAccCheckCloudscaleSubnetConfig_network_name() string {
	return `
data "cloudscale_subnet" "foo" {
  network_name = "${cloudscale_network.multi-net.0.name}"
}
`
}
