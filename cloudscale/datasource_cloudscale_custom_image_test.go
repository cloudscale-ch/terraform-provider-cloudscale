package cloudscale

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudscaleCustomImage_DS_Basic(t *testing.T) {
	var customImage cloudscale.CustomImage
	rInt := acctest.RandInt()
	name1 := fmt.Sprintf("terraform-%d-0", rInt)
	name2 := fmt.Sprintf("terraform-%d-1", rInt)
	config := customImageConfig_baseline(2, rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleCustomImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				Config: config + testAccCheckCloudscaleCustomImageConfig_name(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleCustomImageExists("data.cloudscale_custom_image.foo", &customImage),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_custom_image.basic.0", "id", &customImage.UUID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_custom_image.foo", "id", &customImage.UUID),
					resource.TestCheckResourceAttr(
						"data.cloudscale_custom_image.foo", "name", name1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_custom_image.foo", "slug", "terra-0"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_custom_image.foo", "user_data_handling", "pass-through"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_custom_image.foo", "size_gb", "1"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_custom_image.foo", "zone_slugs.#", "2"),
					resource.TestCheckResourceAttrSet(
						"data.cloudscale_custom_image.foo", "checksums.sha256"),
					resource.TestCheckResourceAttrSet(
						"data.cloudscale_custom_image.foo", "href"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleCustomImageConfig_name_and_slug(name1, "terra-0"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_custom_image.foo", "name", name1),
				),
			},
			{
				Config: config + testAccCheckCloudscaleCustomImageConfig_name(name2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_custom_image.foo", "name", name2),
					resource.TestCheckResourceAttr(
						"data.cloudscale_custom_image.foo", "slug", "terra-1"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleCustomImageConfig_name_and_slug(name2, "terra-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_custom_image.foo", "name", name2),
					resource.TestCheckResourceAttr(
						"data.cloudscale_custom_image.foo", "slug", "terra-1"),
				),
			},
			{
				Config:      config + testAccCheckCloudscaleCustomImageConfig_name_and_slug(name1, "terra-1"),
				ExpectError: regexp.MustCompile(`Found zero Custom Images`),
			},
			{

				Config: config + testAccCheckCloudscaleCustomImageConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic.0", "name", name1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_custom_image.foo", "name", name1),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_custom_image.basic.0", "id", &customImage.UUID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_custom_image.foo", "id", &customImage.UUID),
				),
			},
			{
				Config:      config + "\n" + `data "cloudscale_custom_image" "foo" {}`,
				ExpectError: regexp.MustCompile(`Found \d+ Custom Images, expected one`),
			},
		},
	})
}

func TestAccCloudscaleCustomImage_DS_NotExisting(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckCloudscaleCustomImageConfig_name("terraform-unknown"),
				ExpectError: regexp.MustCompile(`Found zero Custom Images`),
			},
		},
	})
}

func testAccCheckCloudscaleCustomImageConfig_name(name string) string {
	return fmt.Sprintf(`
data "cloudscale_custom_image" "foo" {
  name = "%s"
}
`, name)
}

func testAccCheckCloudscaleCustomImageConfig_name_and_slug(name, slug string) string {
	return fmt.Sprintf(`
data "cloudscale_custom_image" "foo" {
  name      = "%s"
  slug	= "%s"
}
`, name, slug)
}

func testAccCheckCloudscaleCustomImageConfig_id() string {
	return fmt.Sprintf(`
data "cloudscale_custom_image" "foo" {
  id = "${cloudscale_custom_image.basic.0.id}"
}
`)
}

func customImageConfig_baseline(count int, rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_custom_image" "basic" {
  count        = "%v"
  import_url         = "%s"
  import_source_format      = "raw"
  name               = "terraform-%d-${count.index}"
  slug               = "terra-${count.index}"
  user_data_handling = "pass-through"
  zone_slugs         = ["lpg1", "rma1"]
}`, count, smallImageDownloadURL, rInt)
}
