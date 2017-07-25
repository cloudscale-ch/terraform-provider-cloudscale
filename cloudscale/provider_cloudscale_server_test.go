package cloudscale

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/cloudscale-ch/cloudscale"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("cloudscale_server", &resource.Sweeper{
		Name: "cloudscale_server",
		F:    testSweepServers,
	})

}

func testSweepServers(region string) error {
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client := meta.(*cloudscale.Client)

	servers, err := client.Servers.List(context.Background())
	if err != nil {
		return err
	}

	for _, s := range servers {
		if strings.HasPrefix(s.Name, "terraform-") {
			log.Printf("Destroying Server %s", s.Name)

			if err := client.Servers.Delete(context.Background(), s.UUID); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccCloudscale_Basic(t *testing.T) {
	var server cloudscale.Server

	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleServerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &server),
					testAccCheckCloudscaleServerAttributes(&server),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "flavor", "flex-2"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "image", "debian-8"),
				),
			},
		},
	})
}

func TestAccCloudscale_Update(t *testing.T) {
	var afterCreate, afterUpdate cloudscale.Server

	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleServerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &afterCreate),
					testAccCheckCloudscaleServerAttributes(&afterCreate),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "flavor", "flex-2"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "image", "debian-8"),
				),
			},
			{
				Config: testAccCheckCloudscaleServerConfig_update_state(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &afterUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "status", "stopped"),
					testAccCheckServerChanged(t, &afterCreate, &afterUpdate),
				),
			},
		},
	})
}

func TestAccCloudscale_Recreated(t *testing.T) {
	var afterCreate, afterUpdate cloudscale.Server

	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleServerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &afterCreate),
					testAccCheckCloudscaleServerAttributes(&afterCreate),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "flavor", "flex-2"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "image", "debian-8"),
				),
			},
			{
				Config: testAccCheckCloudscaleServerConfig_update_recreate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &afterUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "flavor", "flex-4"),
					testAccCheckServerRecreated(t, &afterCreate, &afterUpdate),
				),
			},
		},
	})
}

func testAccCheckCloudscaleServerDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudscale.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudscale_server" {
			continue
		}

		id := rs.Primary.ID

		// Try to find the Droplet
		_, err := client.Servers.Get(context.Background(), id)

		// Wait

		if err != nil && !strings.Contains(err.Error(), "Not found") {
			return fmt.Errorf(
				"Error waiting for server (%s) to be destroyed: %s",
				rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckCloudscaleServerExists(n string, server *cloudscale.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Server ID is set")
		}

		client := testAccProvider.Meta().(*cloudscale.Client)

		id := rs.Primary.ID

		// Try to find the Droplet
		retrieveServer, err := client.Servers.Get(context.Background(), id)

		if err != nil {
			return err
		}

		if retrieveServer.UUID != rs.Primary.ID {
			return fmt.Errorf("Server not found")
		}

		*server = *retrieveServer

		return nil
	}
}

func testAccCheckCloudscaleServerAttributes(server *cloudscale.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if server.Image.Slug != "debian-8" {
			return fmt.Errorf("Bad image_slug: %s", server.Image.Slug)
		}

		if server.Flavor.Slug != "flex-2" {
			return fmt.Errorf("Bad flavor_slug: %s", server.Image.Slug)
		}

		if server.Volumes[0].SizeGB != 10 {
			return fmt.Errorf("Bad volumes_size_gb: %d", server.Volumes[0].SizeGB)
		}

		return nil
	}
}

func testAccCheckServerChanged(t *testing.T,
	before, after *cloudscale.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.UUID != after.UUID {
			t.Fatalf("Not expected a change of Server IDs got=%s, expected=%s",
				after.UUID, before.UUID)
		}
		return nil
	}
}

func testAccCheckServerRecreated(t *testing.T,
	before, after *cloudscale.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.UUID == after.UUID {
			t.Fatalf("Expected change of Server IDs, but both were %v", before.UUID)
		}
		return nil
	}
}

func testAccCheckCloudscaleServerConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      			= "terraform-%d"
  flavor    			= "flex-2"
  image     			= "debian-8"
  volume_size_gb	= 10
  ssh_keys = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, rInt)
}

func testAccCheckCloudscaleServerConfig_update_state(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      			= "terraform-%d"
  flavor    			= "flex-2"
  image     			= "debian-8"
  volume_size_gb	= 10
	state 					= "stopped"
  ssh_keys = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, rInt)
}

func testAccCheckCloudscaleServerConfig_update_recreate(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      			= "terraform-%d"
  flavor    			= "flex-4"
  image     			= "debian-8"
  volume_size_gb	= 10
	state 					= "stopped"
  ssh_keys = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, rInt)
}
