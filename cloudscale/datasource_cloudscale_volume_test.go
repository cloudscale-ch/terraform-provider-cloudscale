package cloudscale

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudscaleVolume_DS_Basic(t *testing.T) {
	var volume cloudscale.Volume
	rInt := acctest.RandInt()
	name1 := fmt.Sprintf("terraform-%d-0", rInt)
	name2 := fmt.Sprintf("terraform-%d-1", rInt)
	config := volumeConfig_baseline(2, rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				Config: config + testAccCheckCloudscaleVolumeConfig_name(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleVolumeExists("data.cloudscale_volume.foo", &volume),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_volume.basic.0", "id", &volume.UUID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_volume.foo", "id", &volume.UUID),
					resource.TestCheckResourceAttr(
						"data.cloudscale_volume.foo", "name", name1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_volume.foo", "size_gb", "1"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_volume.foo", "type", "ssd"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_volume.foo", "zone_slug", "rma1"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_volume.foo", "server_uuids.#", "0"),
					resource.TestCheckResourceAttrSet(
						"data.cloudscale_volume.foo", "href"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleVolumeConfig_name_and_zone(name1, "rma1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_volume.foo", "name", name1),
				),
			},
			{
				Config: config + testAccCheckCloudscaleVolumeConfig_name(name2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_volume.foo", "name", name2),
					resource.TestCheckResourceAttr(
						"data.cloudscale_volume.foo", "size_gb", "1"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_volume.foo", "zone_slug", "rma1"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_volume.foo", "server_uuids.#", "0"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleVolumeConfig_name_and_zone(name2, "rma1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_volume.foo", "name", name2),
					resource.TestCheckResourceAttr(
						"data.cloudscale_volume.foo", "zone_slug", "rma1"),
				),
			},
			{
				Config:      config + testAccCheckCloudscaleVolumeConfig_name_and_zone(name1, "lpg1"),
				ExpectError: regexp.MustCompile(`.*Found zero volumes.*`),
			},
			{

				Config: config + testAccCheckCloudscaleVolumeConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic.0", "name", name1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_volume.foo", "name", name1),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_volume.basic.0", "id", &volume.UUID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_volume.foo", "id", &volume.UUID),
				),
			},
			{
				Config:      config + "\n" + `data "cloudscale_volume" "foo" {}`,
				ExpectError: regexp.MustCompile(`Found \d+ volumes, expected one`),
			},
		},
	})
}

func TestAccCloudscaleVolume_DS_NotExisting(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckCloudscaleVolumeConfig_name("terraform-unknown-volume"),
				ExpectError: regexp.MustCompile(`Found zero volumes`),
			},
		},
	})
}

func volumeConfig_baseline(count int, rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_volume" "basic" {
  count     = "%v"
  name      = "terraform-%d-${count.index}"
  size_gb   = 1
  type      = "ssd"
  zone_slug = "rma1"
}`, count, rInt)
}

func testAccCheckCloudscaleVolumeConfig_name(name string) string {
	return fmt.Sprintf(`
data "cloudscale_volume" "foo" {
  name               = "%s"
}
`, name)
}

func testAccCheckCloudscaleVolumeConfig_name_and_zone(name, zone_slug string) string {
	return fmt.Sprintf(`
data "cloudscale_volume" "foo" {
  name               = "%s"
  zone_slug			 = "%s"
}
`, name, zone_slug)
}

func testAccCheckCloudscaleVolumeConfig_id() string {
	return fmt.Sprintf(`
data "cloudscale_volume" "foo" {
  id               = "${cloudscale_volume.basic.0.id}"
}
`)
}
