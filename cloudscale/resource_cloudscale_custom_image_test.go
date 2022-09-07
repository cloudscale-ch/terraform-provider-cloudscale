package cloudscale

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var smallImageDownloadURL string = "https://at-images.objects.lpg.cloudscale.ch/alpine"
var bootImageDownloadURL string = "https://acc-test-images.objects.lpg.cloudscale.ch/debian-10-openstack-amd64.raw"

func init() {
	resource.AddTestSweepers("cloudscale_custom_image", &resource.Sweeper{
		Name: "cloudscale_custom_image",
		F:    testSweepCustomImages,
	})

}

func testSweepCustomImages(region string) error {
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client := meta.(*cloudscale.Client)

	customImages, err := client.CustomImages.List(context.Background())
	if err != nil {
		return err
	}

	foundError := error(nil)
	for _, s := range customImages {
		if strings.HasPrefix(s.Name, "terraform-") {
			log.Printf("Destroying CustomImage %s", s.Name)

			if err := client.CustomImages.Delete(context.Background(), s.UUID); err != nil {
				foundError = err
			}
		}
	}
	return foundError
}

func TestAccCloudscaleCustomImage_Import(t *testing.T) {
	var customImage cloudscale.CustomImage

	rInt := acctest.RandInt()
	md5sum := getExpectedChecksum("md5", t)
	sha256sum := getExpectedChecksum("sha256", t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleCustomImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: customImageConfig_config("basic", smallImageDownloadURL, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleCustomImageExists("cloudscale_custom_image.basic", &customImage),
					resource.TestCheckResourceAttrSet("cloudscale_custom_image.basic", "href"),
					resource.TestCheckResourceAttrSet("cloudscale_custom_image.basic", "id"),
					resource.TestCheckResourceAttrSet("cloudscale_custom_image.basic", "import_href"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "import_status", "success"),
					resource.TestCheckResourceAttrSet("cloudscale_custom_image.basic", "import_uuid"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "slug", "terra-test-slug"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "import_source_format", "raw"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "import_url", smallImageDownloadURL),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "user_data_handling", "extend-cloud-config"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "zone_slugs.#", "2"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "checksums.md5", md5sum),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "checksums.sha256", sha256sum),
				),
			},
		},
	})
}

func TestAccCloudscaleCustomImage_Update(t *testing.T) {
	var customImage cloudscale.CustomImage

	rInt := acctest.RandInt()
	md5sum := getExpectedChecksum("md5", t)
	sha256sum := getExpectedChecksum("sha256", t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleCustomImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: customImageConfig_config("basic", smallImageDownloadURL, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleCustomImageExists("cloudscale_custom_image.basic", &customImage),
					resource.TestCheckResourceAttrSet("cloudscale_custom_image.basic", "href"),
					resource.TestCheckResourceAttrSet("cloudscale_custom_image.basic", "id"),
					resource.TestCheckResourceAttrSet("cloudscale_custom_image.basic", "import_href"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "import_status", "success"),
					resource.TestCheckResourceAttrSet("cloudscale_custom_image.basic", "import_uuid"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "slug", "terra-test-slug"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "import_source_format", "raw"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "import_url", smallImageDownloadURL),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "user_data_handling", "extend-cloud-config"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "zone_slugs.#", "2"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "checksums.md5", md5sum),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "checksums.sha256", sha256sum),
				),
			},
			{
				Config: customImageConfig_changed("basic", smallImageDownloadURL, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleCustomImageExists("cloudscale_custom_image.basic", &customImage),
					resource.TestCheckResourceAttrSet("cloudscale_custom_image.basic", "href"),
					resource.TestCheckResourceAttrSet("cloudscale_custom_image.basic", "id"),
					resource.TestCheckResourceAttrSet("cloudscale_custom_image.basic", "import_href"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "import_status", "success"),
					resource.TestCheckResourceAttrSet("cloudscale_custom_image.basic", "import_uuid"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "name", fmt.Sprintf("terraform-%d-renamed", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "slug", "terra-test-slug-changed"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "import_source_format", "raw"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "import_url", smallImageDownloadURL),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "user_data_handling", "pass-through"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "zone_slugs.#", "2"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "checksums.md5", md5sum),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "checksums.sha256", sha256sum),
				),
			},
		},
	})
}

func TestAccCloudscaleCustomImage_tags(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleCustomImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: customImageConfig_tags("basic", smallImageDownloadURL, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "tags.my-foo", "foo"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "tags.my-bar", "bar"),
					testTagsMatch("cloudscale_custom_image.basic"),
				),
			},
			{
				Config: customImageConfig_config("basic", smallImageDownloadURL, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "tags.%", "0"),
					testTagsMatch("cloudscale_custom_image.basic"),
				),
			},
			{
				Config: customImageConfig_tags("basic", smallImageDownloadURL, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "tags.my-foo", "foo"),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.basic", "tags.my-bar", "bar"),
					testTagsMatch("cloudscale_custom_image.basic"),
				),
			},
		},
	})
}

func TestAccCloudscaleCustomImage_Boot(t *testing.T) {
	var customImage cloudscale.CustomImage
	var server cloudscale.Server

	rInt1, rInt2 := acctest.RandInt(), acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleCustomImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: customImageConfig_config("debian", bootImageDownloadURL, rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleCustomImageExists("cloudscale_custom_image.debian", &customImage),
					resource.TestCheckResourceAttr(
						"cloudscale_custom_image.debian", "import_status", "success"),
				),
			},
			{
				Config: customImageConfig_config("debian", bootImageDownloadURL, rInt1) +
					"\n" + serverConfig_customImage("debian-server", "debian", rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleCustomImageExists("cloudscale_custom_image.debian", &customImage),
					testAccCheckCloudscaleServerExists("cloudscale_server.debian-server", &server),
					testAccCheckCloudscaleServerUsagesImage("cloudscale_server.debian-server", "cloudscale_custom_image.debian"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.debian-server", "status", "running"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.debian-server", "image_slug", "custom:terra-test-slug"),
					resource.TestCheckResourceAttrSet(
						"cloudscale_server.debian-server", "image_uuid"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.debian-server", "ssh_fingerprints.#", "6"),
				),
			},
		},
	})
}

/*
func TestAccCloudscaleCustomImage_import_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleCustomImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: customImageConfig_config("basic", smallImageDownloadURL, acctest.RandInt()),
			},
			{
				ResourceName:      "cloudscale_custom_image.basic",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"import_source_format", "import_url"},
			},
			{
				ResourceName:      "cloudscale_custom_image.basic",
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateId:     "does-not-exist",
				ExpectError:       regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}
*/

func testAccCheckCloudscaleCustomImageExists(n string, customImage *cloudscale.CustomImage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CustomImage ID is set")
		}

		client := testAccProvider.Meta().(*cloudscale.Client)

		id := rs.Primary.ID

		// Try to find the customImage
		retrieveCustomImage, err := client.CustomImages.Get(context.Background(), id)

		if err != nil {
			return err
		}

		if retrieveCustomImage.UUID != rs.Primary.ID {
			return fmt.Errorf("CustomImage not found")
		}

		*customImage = *retrieveCustomImage

		return nil
	}
}

func testAccCheckCloudscaleServerUsagesImage(nServer string, nCustomImage string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceServer, ok := s.RootModule().Resources[nServer]
		if !ok {
			return fmt.Errorf("Not found: %s", nServer)
		}

		resourceCustomImage, ok := s.RootModule().Resources[nCustomImage]
		if !ok {
			return fmt.Errorf("Not found: %s", nCustomImage)
		}

		customImageUUID := resourceCustomImage.Primary.ID
		serverImageUUID := resourceServer.Primary.Attributes["image_uuid"]
		if customImageUUID != serverImageUUID {
			return fmt.Errorf("%v != %v", customImageUUID, serverImageUUID)
		}

		return nil
	}
}

func testAccCheckCloudscaleCustomImageDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudscale.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudscale_custom_image" {
			continue
		}

		id := rs.Primary.ID

		// Try to find the customImage
		v, err := client.CustomImages.Get(context.Background(), id)
		if err == nil {
			return fmt.Errorf("CustomImage %v still exists", v)
		} else {
			errorResponse, ok := err.(*cloudscale.ErrorResponse)
			if !ok || errorResponse.StatusCode != http.StatusNotFound {
				return fmt.Errorf(
					"Error waiting for custom image (%s) to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func customImageConfig_config(name string, imageDownloadURL string, rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_custom_image" "%s" {
  import_url         = "%s"
  import_source_format      = "raw"
  name               = "terraform-%d"
  slug               = "terra-test-slug"
  user_data_handling = "extend-cloud-config"
  zone_slugs         = ["lpg1", "rma1"]
}`, name, imageDownloadURL, rInt)
}

func customImageConfig_tags(name string, imageDownloadURL string, rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_custom_image" "%s" {
  import_url         = "%s"
  import_source_format      = "raw"
  name               = "terraform-%d"
  slug               = "terra-test-slug"
  user_data_handling = "extend-cloud-config"
  zone_slugs         = ["lpg1", "rma1"]
  tags = {
    my-foo = "foo"
    my-bar = "bar"
  }
}`, name, imageDownloadURL, rInt)
}

func customImageConfig_changed(name string, imageDownloadURL string, rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_custom_image" "%s" {
  import_url         = "%s"
  import_source_format      = "raw"
  name               = "terraform-%d-renamed"
  slug               = "terra-test-slug-changed"
  user_data_handling = "pass-through"
  zone_slugs         = ["lpg1", "rma1"]
}`, name, imageDownloadURL, rInt)
}

func serverConfig_customImage(name string, imageName string, rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "%s" {
  flavor_slug    = "flex-4-1"
  image_uuid     = "${cloudscale_custom_image.%s.id}"
  name           = "terraform-%d"
  volume_size_gb = 10
  ssh_keys       = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY=", "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, name, imageName, rInt)
}

func getExpectedChecksum(algo string, t *testing.T) string {
	checksumURL := fmt.Sprintf("%s.%s", smallImageDownloadURL, algo)
	resp, err := http.Get(checksumURL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatal(fmt.Sprintf("Wrong http status code\n got=%#v\nwant=%#v", resp.Status, http.StatusOK))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return string(body)
}
