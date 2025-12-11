package cloudscale

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v6"
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

func TestAccCloudscaleServerGroup_Update(t *testing.T) {
	var beforeUpdate, afterUpdate cloudscale.ServerGroup

	rInt1 := acctest.RandInt()
	rInt2 := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleServerGroupConfigWithZone(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerGroupExists("cloudscale_server_group.servergroup", &beforeUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.servergroup", "type", "anti-affinity"),
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.servergroup", "zone_slug", "lpg1"),
				),
			},
			{
				Config: testAccCheckCloudscaleServerGroupConfigWithZone(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerGroupExists("cloudscale_server_group.servergroup", &afterUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.servergroup", "type", "anti-affinity"),
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.servergroup", "zone_slug", "lpg1"),
					testAccCheckServerGroupIsSame(t, &beforeUpdate, &afterUpdate),
				),
			},
		},
	})
}

func TestAccCloudscaleServerGroup_import_basic(t *testing.T) {
	var afterImport, afterUpdate cloudscale.ServerGroup

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleServerGroupConfigWithZone(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerGroupExists("cloudscale_server_group.servergroup", &afterImport),
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.servergroup", "name", fmt.Sprintf("terraform-%d-group", rInt)),
				),
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
			{
				Config: testAccCheckCloudscaleServerGroupConfigWithZone(42),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerGroupExists("cloudscale_server_group.servergroup", &afterUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.servergroup", "name", "terraform-42-group"),
					testAccCheckServerGroupIsSame(t, &afterImport, &afterUpdate),
				),
			},
		},
	})
}

func TestAccCloudscaleServerGroup_import_withTags(t *testing.T) {
	var afterImport, afterUpdate cloudscale.ServerGroup

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleServerGroupConfigWithZoneAndTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerGroupExists("cloudscale_server_group.servergroup", &afterImport),
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.servergroup", "name", fmt.Sprintf("terraform-%d-group", rInt)),
					testTagsMatch("cloudscale_server_group.servergroup"),
				),
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
			{
				Config: testAccCheckCloudscaleServerGroupConfigWithZone(42),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerGroupExists("cloudscale_server_group.servergroup", &afterUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.servergroup", "name", "terraform-42-group"),
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.servergroup", "tags.%", "0"),
					testAccCheckServerGroupIsSame(t, &afterImport, &afterUpdate),
					testTagsMatch("cloudscale_server_group.servergroup"),
				),
			},
		},
	})
}

func TestAccCloudscaleServerGroup_tags(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleServerGroupConfigWithZoneAndTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.servergroup", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.servergroup", "tags.my-foo", "foo"),
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.servergroup", "tags.my-bar", "bar"),
					testTagsMatch("cloudscale_server_group.servergroup"),
				),
			},
			{
				Config: testAccCheckCloudscaleServerGroupConfigWithZone(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.servergroup", "tags.%", "0"),
					testTagsMatch("cloudscale_server_group.servergroup"),
				),
			},
			{
				Config: testAccCheckCloudscaleServerGroupConfigWithZoneAndTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.servergroup", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.servergroup", "tags.my-foo", "foo"),
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.servergroup", "tags.my-bar", "bar"),
					testTagsMatch("cloudscale_server_group.servergroup"),
				),
			},
		},
	})
}

func testAccCheckServerGroupIsSame(t *testing.T, before, after *cloudscale.ServerGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if adr := before; adr == after {
			t.Fatalf("Passed the same instance twice, address is equal=%v",
				adr)
		}
		if before.UUID != after.UUID {
			t.Fatalf("Not expected a change of Server Group IDs got=%s, expected=%s",
				after.UUID, before.UUID)
		}
		return nil
	}
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
  flavor_slug               = "flex-8-4"
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
  flavor_slug               = "flex-8-4"
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

func testAccCheckCloudscaleServerGroupConfigWithZoneAndTags(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server_group" "servergroup" {
  name = "terraform-%d-group"
  type = "anti-affinity"
  zone_slug = "lpg1"
  tags = {
    my-foo = "foo"
    my-bar = "bar"
  }
}`, rInt)
}
