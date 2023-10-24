package cloudscale

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v4"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudscaleFloatingIP_DS_Basic(t *testing.T) {
	var floatingIP cloudscale.FloatingIP
	rInt := acctest.RandInt()
	reverse_ptr1 := fmt.Sprintf("terraform-%d-0.test.cloudscale.ch", rInt)
	reverse_ptr2 := fmt.Sprintf("terraform-%d-1.test.cloudscale.ch", rInt)
	config := floatingIPConfig_baseline(2, rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				Config: config + testAccCheckCloudscaleFloatingIPConfig_reverse_ptr(reverse_ptr1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleFloatingIPExists("data.cloudscale_floating_ip.foo", &floatingIP),
					resource.TestCheckResourceAttrSet(
						"data.cloudscale_floating_ip.foo", "id"),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_floating_ip.foo", "network", &floatingIP.Network),
					resource.TestCheckResourceAttr(
						"data.cloudscale_floating_ip.foo", "reverse_ptr", reverse_ptr1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_floating_ip.foo", "prefix_length", "128"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_floating_ip.foo", "ip_version", "6"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_floating_ip.foo", "region_slug", "rma"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_floating_ip.foo", "type", "regional"),
					resource.TestCheckResourceAttrSet(
						"data.cloudscale_floating_ip.foo", "href"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleFloatingIPConfig_reverse_ptr_and_region(reverse_ptr1, "rma"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_floating_ip.foo", "reverse_ptr", reverse_ptr1),
				),
			},
			{
				Config: config + testAccCheckCloudscaleFloatingIPConfig_reverse_ptr(reverse_ptr2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_floating_ip.foo", "reverse_ptr", reverse_ptr2),
					resource.TestCheckResourceAttr(
						"data.cloudscale_floating_ip.foo", "region_slug", "rma"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleFloatingIPConfig_reverse_ptr_and_region(reverse_ptr2, "rma"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_floating_ip.foo", "reverse_ptr", reverse_ptr2),
					resource.TestCheckResourceAttr(
						"data.cloudscale_floating_ip.foo", "region_slug", "rma"),
				),
			},
			{
				Config:      config + testAccCheckCloudscaleFloatingIPConfig_reverse_ptr_and_region(reverse_ptr1, "lpg"),
				ExpectError: regexp.MustCompile(`Found zero Floating IPs`),
			},
			{

				Config: config + testAccCheckCloudscaleFloatingIPConfig_cidr(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.basic.0", "reverse_ptr", reverse_ptr1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_floating_ip.foo", "reverse_ptr", reverse_ptr1),
					resource.TestCheckResourceAttrSet(
						"data.cloudscale_floating_ip.foo", "id"),
				),
			},
			{
				Config:      config + testAccCheckCloudscaleFloatingIPConfig_reverse_ptr_type_ip_version(reverse_ptr1, "regional", "4"),
				ExpectError: regexp.MustCompile(`Found zero Floating IPs`),
			},
			{
				Config:      config + testAccCheckCloudscaleFloatingIPConfig_reverse_ptr_type_ip_version(reverse_ptr1, "global", "6"),
				ExpectError: regexp.MustCompile(`Found zero Floating IPs`),
			},
			{
				Config: config + testAccCheckCloudscaleFloatingIPConfig_reverse_ptr_type_ip_version(reverse_ptr1, "regional", "6"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_floating_ip.foo", "reverse_ptr", reverse_ptr1),
				),
			},
			{

				Config: config + testAccCheckCloudscaleFloatingIPConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_floating_ip.foo", "reverse_ptr", reverse_ptr1),
				),
			},
			{
				Config:      config + "\n" + `data "cloudscale_floating_ip" "foo" {}`,
				ExpectError: regexp.MustCompile(`Found \d+ Floating IPs, expected one`),
			},
		},
	})
}

func TestAccCloudscaleFloatingIP_DS_WithServer(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testFloatingIPConfig_ds_with_server(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.cloudscale_floating_ip.foo", "id", "cloudscale_floating_ip.gateway", "id"),
					resource.TestCheckResourceAttrPair(
						"data.cloudscale_floating_ip.foo", "cidr", "cloudscale_floating_ip.gateway", "cidr"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_floating_ip.foo", "reverse_ptr", "vip.web-worker01.example.com"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_floating_ip.foo", "prefix_length", "32"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_floating_ip.foo", "ip_version", "4"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_floating_ip.foo", "region_slug", "rma"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_floating_ip.foo", "type", "regional"),
					resource.TestCheckResourceAttrPair(
						"data.cloudscale_floating_ip.foo", "next_hop", "cloudscale_server.basic", "public_ipv4_address"),
					resource.TestCheckResourceAttrPair(
						"data.cloudscale_floating_ip.foo", "server", "cloudscale_server.basic", "id"),
					resource.TestCheckResourceAttrSet(
						"data.cloudscale_floating_ip.foo", "href"),
				),
			},
		},
	})
}

func TestAccCloudscaleFloatingIP_DS_NotExisting(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckCloudscaleFloatingIPConfig_reverse_ptr("terraform-unknown"),
				ExpectError: regexp.MustCompile(`Found zero Floating IPs`),
			},
		},
	})
}

func testAccCheckCloudscaleFloatingIPConfig_reverse_ptr(reverse_ptr string) string {
	return fmt.Sprintf(`
data "cloudscale_floating_ip" "foo" {
  ip_version = 6
  reverse_ptr = "%s"
}
`, reverse_ptr)
}

func testAccCheckCloudscaleFloatingIPConfig_reverse_ptr_and_region(reverse_ptr, region_slug string) string {
	return fmt.Sprintf(`
data "cloudscale_floating_ip" "foo" {
  ip_version = 6
  reverse_ptr = "%s"
  region_slug = "%s"
}
`, reverse_ptr, region_slug)
}

func testAccCheckCloudscaleFloatingIPConfig_cidr() string {
	return fmt.Sprintf(`
data "cloudscale_floating_ip" "foo" {
  network = "${cloudscale_floating_ip.basic.0.network}"
}
`)
}

func floatingIPConfig_baseline(count int, rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_floating_ip" "basic" {
  count       = "%v"
  ip_version  = 6
  region_slug = "rma"
  reverse_ptr = "terraform-%d-${count.index}.test.cloudscale.ch"
}`, count, rInt)
}

func testFloatingIPConfig_ds_with_server(rInt int) string {
	return testAccCheckCloudscaleFloatingIPConfig_server(rInt) + `
data "cloudscale_floating_ip" "foo" {
  network = "${cloudscale_floating_ip.gateway.network}"
}
	`
}

func testAccCheckCloudscaleFloatingIPConfig_reverse_ptr_type_ip_version(reverse_ptr, type_, ip_version string) string {
	return fmt.Sprintf(`
data "cloudscale_floating_ip" "foo" {
  ip_version = "%s"
  type = "%s"
  reverse_ptr = "%s"
}
`, ip_version, type_, reverse_ptr)
}

func testAccCheckCloudscaleFloatingIPConfig_id() string {
	return fmt.Sprintf(`
data "cloudscale_floating_ip" "foo" {
  id = "${cloudscale_floating_ip.basic.0.id}"
}
`)
}
