package cloudscale

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("cloudscale_server_group", &resource.Sweeper{
		Name: "cloudscale_server_group",
		F:    testSweepServerGroups,
	})
}

func testSweepServerGroups(region string) error {
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client := meta.(*cloudscale.Client)

	serverGroups, err := client.ServerGroups.List(context.Background())
	if err != nil {
		return err
	}

	foundError := error(nil)
	for _, s := range serverGroups {
		if strings.HasPrefix(s.Name, "terraform-") {
			log.Printf("Destroying server group %s", s.Name)

			if err := client.ServerGroups.Delete(context.Background(), s.UUID); err != nil {
				foundError = err
			}
		}
	}
	return foundError
}

func TestAccCloudscaleServerGroup_Basic(t *testing.T) {
	rInt := acctest.RandInt()
	groupName := fmt.Sprintf("terraform-%d-group", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleServerGroupConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_server.some_server", "server_groups.#", "1"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.some_server", "server_groups.0.name", groupName),
				),
			},
			{
				Config: testAccCheckCloudscaleServerGroupConfigAddServer(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_server.some_server", "server_groups.#", "1"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.some_server", "server_groups.0.name", groupName),
					resource.TestCheckResourceAttr(
						"cloudscale_server.some_server2", "server_groups.#", "1"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.some_server2", "server_groups.0.name", groupName),
				),
			},
		},
	})
}

func TestAccCloudscaleServerGroup_WithZone(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleServerGroupConfigWithZone(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.servergroup", "type", "anti-affinity"),
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.servergroup", "zone_slug", "lpg1"),
				),
			},
		},
	})
}

func TestAccCloudscaleServerGroup_import_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleServerGroupConfigWithZone(acctest.RandInt()),
			},
			{
				ResourceName:      "cloudscale_server_group.servergroup",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "cloudscale_server_group.servergroup",
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateId:     "does-not-exist",
				ExpectError:       regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

func testAccCheckCloudscaleServerGroupDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudscale.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudscale_server_group" {
			continue
		}

		id := rs.Primary.ID

		// Try to find the server group
		s, err := client.ServerGroups.Get(context.Background(), id)

		// Wait

		if err == nil {
			return fmt.Errorf("The server group %v remained, even though the resource was destoryed", s)
		} else {
			errorResponse, ok := err.(*cloudscale.ErrorResponse)
			if !ok || errorResponse.StatusCode != http.StatusNotFound {
				return fmt.Errorf(
					"Error waiting for server group (%s) to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func testAccCheckCloudscaleServerGroupConfig(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server_group" "ayyy" {
  name = "terraform-%d-group"
  type = "anti-affinity"
}

resource "cloudscale_server" "some_server" {
  name                      = "terraform-%d-foobar"
  server_group_ids          = ["${cloudscale_server_group.ayyy.id}"]
  flavor_slug               = "flex-4"
  image_slug                = "%s"
  ssh_keys                  = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY=", "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}
`, rInt, rInt, DefaultImageSlug)
}

func testAccCheckCloudscaleServerGroupConfigAddServer(rInt int) string {
	return testAccCheckCloudscaleServerGroupConfig(rInt) + fmt.Sprintf(`
resource "cloudscale_server" "some_server2" {
  name                      = "terraform-%d-foobar2"
  server_group_ids          = ["${cloudscale_server_group.ayyy.id}"]
  flavor_slug               = "flex-4"
  image_slug                = "%s"
  ssh_keys                  = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY=", "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, rInt, DefaultImageSlug)
}

func testAccCheckCloudscaleServerGroupConfigWithZone(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server_group" "servergroup" {
  name = "terraform-%d-group"
  type = "anti-affinity"
  zone_slug = "lpg1"
}`, rInt)
}
