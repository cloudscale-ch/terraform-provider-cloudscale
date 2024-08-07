package cloudscale

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("cloudscale_volume", &resource.Sweeper{
		Name: "cloudscale_volume",
		F:    testSweepVolumes,
	})

}

func testSweepVolumes(region string) error {
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client := meta.(*cloudscale.Client)

	volumes, err := client.Volumes.List(context.Background())
	if err != nil {
		return err
	}

	foundError := error(nil)
	for _, s := range volumes {
		if strings.HasPrefix(s.Name, "terraform-") {
			log.Printf("Destroying Volume %s", s.Name)

			if err := client.Volumes.Delete(context.Background(), s.UUID); err != nil {
				foundError = err
			}
		}
	}
	return foundError
}

func TestAccCloudscaleVolume_DetachedWithZone(t *testing.T) {
	var volume cloudscale.Volume

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: volumeConfig_detached_with_zone(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleVolumeExists("cloudscale_volume.basic", &volume),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "type", "bulk"),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "size_gb", "100"),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "zone_slug", "lpg1"),
				),
			},
		},
	})
}

func TestAccCloudscaleVolume_Change(t *testing.T) {
	var volume cloudscale.Volume

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: volumeConfig_detached(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleVolumeExists("cloudscale_volume.basic", &volume),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "type", "ssd"),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "size_gb", "50"),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "server_uuids.#", "0"),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "zone_slug", "rma1"),
				),
			},
			{
				Config: volumeConfig_multiple_changes(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleVolumeExists("cloudscale_volume.basic", &volume),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "name", fmt.Sprintf("terraform-%d-renamed", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "type", "ssd"),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "size_gb", "100"),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "server_uuids.#", "0"),
				),
			},
		},
	})
}

func TestAccCloudscaleVolume_Detach(t *testing.T) {
	var server cloudscale.Server
	var volume cloudscale.Volume

	rInt1 := acctest.RandInt()
	rInt2 := acctest.RandInt()

	serverConfig := testAccCheckCloudscaleServerConfig_basic(rInt1)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: serverConfig + "\n" + volumeConfig_attached(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &server),
					testAccCheckCloudscaleVolumeExists("cloudscale_volume.basic", &volume),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "server_uuids.#", "1"),
					assertVolumeAttached(&server, &volume),
				),
			},
			{
				Config: serverConfig + "\n" + volumeConfig_detached(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleVolumeExists("cloudscale_volume.basic", &volume),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "server_uuids.#", "0"),
				),
			},
		},
	})
}

func TestAccCloudscaleVolume_Reattach(t *testing.T) {
	var server1, server2 cloudscale.Server
	var volume cloudscale.Volume

	rInt1 := acctest.RandInt()
	rInt2 := acctest.RandInt()
	rInt3 := acctest.RandInt()

	serverConfig := testAccCheckCloudscaleServerConfig_basic(rInt1)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: serverConfig + "\n" + volumeConfig_attached(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &server1),
					testAccCheckCloudscaleVolumeExists("cloudscale_volume.basic", &volume),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "server_uuids.#", "1"),
					assertVolumeAttached(&server1, &volume),
				),
			},
			{
				Config: serverConfig + "\n" + volumeConfig_reattached_volume(rInt3, rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerExists("cloudscale_server.reattach_server", &server2),
					testAccCheckCloudscaleVolumeExists("cloudscale_volume.basic", &volume),
					assertVolumeAttached(&server2, &volume),
				),
			},
		},
	})
}

func TestAccCloudscaleVolume_import_basic(t *testing.T) {
	var afterImport, afterUpdate cloudscale.Volume

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: volumeConfig_detached(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleVolumeExists("cloudscale_volume.basic", &afterImport),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
				),
			},
			{
				ResourceName:      "cloudscale_volume.basic",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "cloudscale_volume.basic",
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateId:     "does-not-exist",
				ExpectError:       regexp.MustCompile(`Cannot import non-existent remote object`),
			},
			{
				Config: volumeConfig_detached(42),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleVolumeExists("cloudscale_volume.basic", &afterUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "name", "terraform-42"),
					testAccCheckCloudscaleVolumeIsSame(t, &afterImport, &afterUpdate),
				),
			},
		},
	})
}

func TestAccCloudscaleVolume_tags(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: volumeConfig_detached_withTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "tags.my-foo", "foo"),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "tags.my-bar", "bar"),
					testTagsMatch("cloudscale_volume.basic"),
				),
			},
			{
				Config: volumeConfig_detached(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "tags.%", "0"),
					testTagsMatch("cloudscale_volume.basic"),
				),
			},
			{
				Config: volumeConfig_detached_withTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "tags.my-foo", "foo"),
					resource.TestCheckResourceAttr(
						"cloudscale_volume.basic", "tags.my-bar", "bar"),
					testTagsMatch("cloudscale_volume.basic"),
				),
			},
		},
	})
}

func testAccCheckCloudscaleVolumeIsSame(t *testing.T,
	before, after *cloudscale.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if adr := before; adr == after {
			t.Fatalf("Passed the same instance twice, address is equal=%v",
				adr)
		}
		if before.UUID != after.UUID {
			t.Fatalf("Not expected a change of Volume IDs got=%s, expected=%s",
				after.UUID, before.UUID)
		}
		return nil
	}
}

func testAccCheckCloudscaleVolumeDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudscale.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudscale_volume" {
			continue
		}

		id := rs.Primary.ID

		// Try to find the volume
		v, err := client.Volumes.Get(context.Background(), id)
		if err == nil {
			return fmt.Errorf("Volume %v still exists", v)
		} else {
			errorResponse, ok := err.(*cloudscale.ErrorResponse)
			if !ok || errorResponse.StatusCode != http.StatusNotFound {
				return fmt.Errorf(
					"Error waiting for volume (%s) to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func assertVolumeAttached(server *cloudscale.Server, volume *cloudscale.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if server.UUID != (*volume.ServerUUIDs)[0] {
			return fmt.Errorf("Volume is not attached to %s but to %v", server.UUID, volume.ServerUUIDs)
		}
		return nil
	}
}

func volumeConfig_detached(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_volume" "basic" {
  name         = "terraform-%d"
  size_gb      = 50
  type         = "ssd"
}`, rInt)
}

func volumeConfig_detached_withTags(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_volume" "basic" {
  name         = "terraform-%d"
  size_gb      = 50
  type         = "ssd"
  tags = {
    my-foo = "foo"
    my-bar = "bar"
  }
}`, rInt)
}

func volumeConfig_multiple_changes(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_volume" "basic" {
  name         = "terraform-%d-renamed"
  size_gb      = 100
  type         = "ssd"
}`, rInt)
}

func volumeConfig_attached(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_volume" "basic" {
  name         = "terraform-%d"
  size_gb      = 50
  server_uuids = ["${cloudscale_server.basic.id}"]
  type         = "ssd"
}`, rInt)
}

func volumeConfig_reattached_volume(serverInt int, volumeInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "reattach_server" {
  name        = "terraform-%d"
  flavor_slug = "flex-4-1"
  image_slug  = "%s"
  ssh_keys    = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}
resource "cloudscale_volume" "basic" {
  name         = "terraform-%d"
  size_gb      = 50
  server_uuids = ["${cloudscale_server.reattach_server.id}"]
  type         = "ssd"
}`, serverInt, DefaultImageSlug, volumeInt)
}

func volumeConfig_detached_with_zone(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_volume" "basic" {
  name         = "terraform-%d"
  size_gb      = 100
  zone_slug    = "lpg1"
  type         = "bulk"
}`, rInt)
}
