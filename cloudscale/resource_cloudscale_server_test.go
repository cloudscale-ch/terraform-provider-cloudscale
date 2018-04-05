package cloudscale

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
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

func TestAccCloudscaleServer_Basic(t *testing.T) {
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
						"cloudscale_server.basic", "flavor_slug", "flex-2"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "image_slug", "debian-8"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "interfaces.0.type", "public"),
				),
			},
		},
	})
}

func TestAccCloudscaleServer_Basic_stopped(t *testing.T) {
	var server cloudscale.Server

	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleServerConfig_basic_stopped(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &server),
					testAccCheckCloudscaleServerAttributes(&server),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "flavor_slug", "flex-2"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "image_slug", "debian-8"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "status", "stopped"),
				),
			},
		},
	})
}

func TestAccCloudscaleServer_AntiAffinity(t *testing.T) {
	var serverA, serverB cloudscale.Server

	aInt := acctest.RandInt()
	bInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleServerConfig_anti_affinity_group(aInt, bInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerExists("cloudscale_server.dbmaster", &serverA),
					testAccCheckCloudscaleServerExists("cloudscale_server.web", &serverB),
					testAccAntiAffinityGroup(t, &serverA, &serverB),
				),
			},
		},
	})
}

func TestAccCloudscaleServer_Update(t *testing.T) {
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
						"cloudscale_server.basic", "flavor_slug", "flex-2"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "image_slug", "debian-8"),
				),
			},
			{
				Config: testAccCheckCloudscaleServerConfig_update_state_stopped(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &afterUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "status", "stopped"),
					testAccCheckServerChanged(t, &afterCreate, &afterUpdate),
				),
			},
			{
				Config: testAccCheckCloudscaleServerConfig_update_state_running(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &afterUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "status", "running"),
					testAccCheckServerChanged(t, &afterCreate, &afterUpdate),
				),
			},
		},
	})
}

func TestAccCloudscaleServer_Recreated(t *testing.T) {
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
						"cloudscale_server.basic", "flavor_slug", "flex-2"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "image_slug", "debian-8"),
				),
			},
			{
				Config: testAccCheckCloudscaleServerConfig_update_recreate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &afterUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "flavor_slug", "flex-4"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "interfaces.#", "2"),
					testAccCheckServerRecreated(t, &afterCreate, &afterUpdate),
				),
			},
		},
	})
}

func TestAccCloudscaleServer_PrivateNetwork(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleServerConfig_only_private_network(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_server.private", "name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_server.private", "interfaces.#", "1"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.private", "interfaces.0.type", "private"),
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

		// Try to find the server
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

		// Try to find the server
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
			return fmt.Errorf("Bad image_slug_slug: %s", server.Image.Slug)
		}

		if server.Flavor.Slug != "flex-2" {
			return fmt.Errorf("Bad flavor_slug_slug: %s", server.Image.Slug)
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

func testAccAntiAffinityGroup(t *testing.T,
	serverA, serverB *cloudscale.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if serverA.UUID != serverB.AntiAfinityWith[0].UUID {
			t.Fatalf("Server A (%s) not in anti_affinity_with", serverB.UUID)
		}
		if serverB.UUID != serverA.AntiAfinityWith[0].UUID {
			t.Fatalf("Server B (%s) not in anti_affinity_with", serverB.UUID)
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
  name      					= "terraform-%d"
  flavor_slug    			= "flex-2"
  image_slug     			= "debian-8"
  volume_size_gb			= 10
  ssh_keys 						= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY=", "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, rInt)
}

func testAccCheckCloudscaleServerConfig_basic_stopped(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      					= "terraform-%d"
  flavor_slug    			= "flex-2"
  image_slug     			= "debian-8"
  volume_size_gb			= 10
  status							= "stopped"
  ssh_keys 						= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, rInt)
}

func testAccCheckCloudscaleServerConfig_update_state_stopped(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      					= "terraform-%d"
  flavor_slug    			= "flex-2"
  image_slug     			= "debian-8"
  volume_size_gb			= 10
	status 							= "stopped"
  ssh_keys 						= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, rInt)
}

func testAccCheckCloudscaleServerConfig_anti_affinity_group(aInt, bInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "dbmaster" {
  name      					= "terraform-%d"
  flavor_slug    			= "flex-2"
  image_slug     			= "debian-8"
  ssh_keys 						= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}
resource "cloudscale_server" "web" {
  name      					= "terraform-%d"
  flavor_slug    			= "flex-2"
  image_slug     			= "debian-8"
  ssh_keys 						= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
	anti_affinity_uuid 	= "${cloudscale_server.dbmaster.id}"
}`, aInt, bInt)
}

func testAccCheckCloudscaleServerConfig_update_state_running(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      					= "terraform-%d"
  flavor_slug    			= "flex-2"
  image_slug     			= "debian-8"
  volume_size_gb			= 10
	status 							= "running"
  ssh_keys = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, rInt)
}

func testAccCheckCloudscaleServerConfig_update_recreate(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      			    = "terraform-%d"
  flavor_slug    			= "flex-4"
  image_slug     			= "debian-8"
  use_private_network		= true
  volume_size_gb			= 10
  ssh_keys 						= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, rInt)
}

func testAccCheckCloudscaleServerConfig_only_private_network(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "private" {
  name      			    = "terraform-%d"
  flavor_slug    			= "flex-4"
  image_slug     			= "debian-8"
  use_private_network		= true
  use_public_network		= false
  volume_size_gb			= 10
  ssh_keys 						= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, rInt)
}
