package cloudscale

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudscaleServer_DS_Basic(t *testing.T) {
	var server cloudscale.Server
	rInt := acctest.RandInt()
	name1 := fmt.Sprintf("terraform-%d-0", rInt)
	name2 := fmt.Sprintf("terraform-%d-1", rInt)
	config := serverConfig_baseline(2, rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				Config: config + testAccCheckCloudscaleServerConfig_name(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerExists("data.cloudscale_server.foo", &server),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_server.basic.0", "id", &server.UUID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_server.foo", "id", &server.UUID),
					resource.TestCheckResourceAttr(
						"data.cloudscale_server.foo", "name", name1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_server.foo", "status", "running"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_server.foo", "flavor_slug", "flex-4-1"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_server.foo", "image_slug", DefaultImageSlug),
					resource.TestCheckResourceAttr(
						"data.cloudscale_server.foo", "interfaces.0.type", "public"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_server.foo", "ssh_host_keys.#", "4"),
					testAccCheckServerIp("data.cloudscale_server.foo"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_server.foo", "zone_slug", "rma1"),
					resource.TestCheckResourceAttrSet(
						"data.cloudscale_server.foo", "href"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleServerConfig_name_and_zone(name1, "rma1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_server.foo", "name", name1),
				),
			},
			{
				Config: config + testAccCheckCloudscaleServerConfig_name(name2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_server.foo", "name", name2),
					resource.TestCheckResourceAttr(
						"data.cloudscale_server.foo", "zone_slug", "rma1"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleServerConfig_name_and_zone(name2, "rma1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_server.foo", "name", name2),
					resource.TestCheckResourceAttr(
						"data.cloudscale_server.foo", "zone_slug", "rma1"),
				),
			},
			{
				Config:      config + testAccCheckCloudscaleServerConfig_name_and_zone(name1, "lpg1"),
				ExpectError: regexp.MustCompile(`Found zero servers`),
			},
			{

				Config: config + testAccCheckCloudscaleServerConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic.0", "name", name1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_server.foo", "name", name1),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_server.basic.0", "id", &server.UUID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_server.foo", "id", &server.UUID),
				),
			},
			{
				Config:      config + "\n" + `data "cloudscale_server" "foo" {}`,
				ExpectError: regexp.MustCompile(`Found \d+ servers, expected one`),
			},
		},
	})
}

func TestAccCloudscaleServer_DS_NotExisting(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckCloudscaleServerConfig_name("terraform-unknown"),
				ExpectError: regexp.MustCompile(`Found zero servers`),
			},
		},
	})
}

func testAccCheckCloudscaleServerConfig_name(name string) string {
	return fmt.Sprintf(`
data "cloudscale_server" "foo" {
  name = "%s"
}
`, name)
}

func testAccCheckCloudscaleServerConfig_name_and_zone(name, zone_slug string) string {
	return fmt.Sprintf(`
data "cloudscale_server" "foo" {
  name      = "%s"
  zone_slug	= "%s"
}
`, name, zone_slug)
}

func testAccCheckCloudscaleServerConfig_id() string {
	return fmt.Sprintf(`
data "cloudscale_server" "foo" {
  id = "${cloudscale_server.basic.0.id}"
}
`)
}

func serverConfig_baseline(count, rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  count                     = %d
  name                      = "terraform-%d-${count.index}"
  flavor_slug               = "flex-4-1"
  allow_stopping_for_update = true
  image_slug                = "%s"
  volume_size_gb            = 10
  zone_slug                 = "rma1"
  ssh_keys 	                = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY=", "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, count, rInt, DefaultImageSlug)
}
