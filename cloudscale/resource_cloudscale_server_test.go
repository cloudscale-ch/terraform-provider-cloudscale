package cloudscale

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const DefaultImageSlug = "debian-9"

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

	foundError := error(nil)
	for _, s := range servers {
		if strings.HasPrefix(s.Name, "terraform-") {
			log.Printf("Destroying Server %s", s.Name)

			if err := client.Servers.Delete(context.Background(), s.UUID); err != nil {
				foundError = err
			}
		}
	}
	return foundError
}

func TestAccCloudscaleServer_Basic(t *testing.T) {
	var server cloudscale.Server

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
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
						"cloudscale_server.basic", "image_slug", DefaultImageSlug),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "interfaces.0.type", "public"),
					testAccCheckServerIp("cloudscale_server.basic"),
				),
			},
		},
	})
}

func TestAccCloudscaleServer_Basic_stopped(t *testing.T) {
	var server cloudscale.Server

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
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
						"cloudscale_server.basic", "image_slug", DefaultImageSlug),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "status", "stopped"),
				),
			},
		},
	})
}

func TestAccCloudscaleServer_UpdateStatus(t *testing.T) {
	var afterCreate, afterUpdate cloudscale.Server

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
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
						"cloudscale_server.basic", "image_slug", DefaultImageSlug),
					testAccCheckServerIp("cloudscale_server.basic"),
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
					testAccCheckServerIp("cloudscale_server.basic"),
					testAccCheckServerChanged(t, &afterCreate, &afterUpdate),
				),
			},
		},
	})
}

func TestAccCloudscaleServer_Password(t *testing.T) {
	var afterCreate cloudscale.Server

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testServerPasswordConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerExists("cloudscale_server.password", &afterCreate),
					resource.TestCheckResourceAttr("cloudscale_server.password", "ssh_keys.#", "0"),
				),
			},
		},
	})
}

func TestAccCloudscaleServer_Recreated(t *testing.T) {
	var afterCreate, afterUpdate cloudscale.Server

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
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
						"cloudscale_server.basic", "image_slug", DefaultImageSlug),
					testAccCheckServerIp("cloudscale_server.basic"),
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
					testAccCheckServerIp("cloudscale_server.basic"),
					testAccCheckServerRecreated(t, &afterCreate, &afterUpdate),
				),
			},
		},
	})
}

func TestAccCloudscaleServer_PrivateNetwork(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
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
					testAccCheckServerIp("cloudscale_server.private"),
				),
			},
		},
	})
}

func TestAccCloudscaleServer_UpdateNameAndFlavorAndVolumeSize(t *testing.T) {
	var afterCreate, afterUpdate cloudscale.Server

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
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
						"cloudscale_server.basic", "status", "running"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "volume_size_gb", "10"),
					testAccCheckServerIp("cloudscale_server.basic"),
				),
			},
			{
				Config: testAccCheckCloudscaleServerConfig_scaled_and_renamed(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerExists("cloudscale_server.basic", &afterUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "flavor_slug", "flex-4"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "name", fmt.Sprintf("terraform-%d-foobar", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "status", "running"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.basic", "volume_size_gb", "11"),
					testAccCheckServerChanged(t, &afterCreate, &afterUpdate),
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
		s, err := client.Servers.Get(context.Background(), id)

		// Wait

		if err == nil {
			return fmt.Errorf("The server %v remained, even though the resource was destoryed", s)
		} else {
			errorResponse, ok := err.(*cloudscale.ErrorResponse)
			if !ok || errorResponse.StatusCode != http.StatusNotFound {
				return fmt.Errorf(
					"Error waiting for server (%s) to be destroyed: %s",
					rs.Primary.ID, err)
			}
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

		if server.Image.Slug != DefaultImageSlug {
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

func testAccCheckServerIp(n string) resource.TestCheckFunc {
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

		for _, networkInterface := range retrieveServer.Interfaces {
			for _, ipAddress := range networkInterface.Addresses {
				if ipAddress.Version == 4 && networkInterface.Type == "public" {
					err := resource.TestCheckResourceAttr(n, "public_ipv4_address", ipAddress.Address)(s)
					if err != nil {
						return err
					}
				} else if ipAddress.Version == 4 && networkInterface.Type == "private" {
					err := resource.TestCheckResourceAttr(n, "private_ipv4_address", ipAddress.Address)(s)
					if err != nil {
						return err
					}
				} else if ipAddress.Version == 6 && networkInterface.Type == "public" {
					err := resource.TestCheckResourceAttr(n, "public_ipv6_address", ipAddress.Address)(s)
					if err != nil {
						return err
					}
				}
			}
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
  name      					= "terraform-%d"
  flavor_slug    			= "flex-2"
  allow_stopping_for_update = true
  image_slug     			= "%s"
  volume_size_gb			= 10
  ssh_keys 						= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY=", "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, rInt, DefaultImageSlug)
}

func testAccCheckCloudscaleServerConfig_basic_stopped(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      					= "terraform-%d"
  flavor_slug    			= "flex-2"
  image_slug     			= "%s"
  volume_size_gb			= 10
  status							= "stopped"
  ssh_keys 						= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, rInt, DefaultImageSlug)
}

func testAccCheckCloudscaleServerConfig_update_state_stopped(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      					= "terraform-%d"
  flavor_slug    			= "flex-2"
  image_slug     			= "%s"
  volume_size_gb			= 10
  status 							= "stopped"
  ssh_keys 						= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, rInt, DefaultImageSlug)
}

func testAccCheckCloudscaleServerConfig_update_state_running(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      					= "terraform-%d"
  flavor_slug    			= "flex-2"
  image_slug     			= "%s"
  volume_size_gb			= 10
  status 							= "running"
  ssh_keys = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, rInt, DefaultImageSlug)
}

func testAccCheckCloudscaleServerConfig_update_recreate(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      			    = "terraform-%d"
  flavor_slug    			= "flex-4"
  image_slug     			= "%s"
  use_private_network		= true
  volume_size_gb			= 10
  ssh_keys 						= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, rInt, DefaultImageSlug)
}

func testAccCheckCloudscaleServerConfig_only_private_network(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "private" {
  name      			    = "terraform-%d"
  flavor_slug    			= "flex-4"
  image_slug     			= "%s"
  use_private_network		= true
  use_public_network		= false
  volume_size_gb			= 10
  ssh_keys 						= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, rInt, DefaultImageSlug)
}

func testAccCheckCloudscaleServerConfig_scaled_and_renamed(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      					= "terraform-%d-foobar"
  flavor_slug    			= "flex-4"
  allow_stopping_for_update = true
  image_slug     			= "%s"
  volume_size_gb			= 11
  ssh_keys 						= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY=", "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}`, rInt, DefaultImageSlug)
}

func testServerPasswordConfig(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "password" {
  name                      = "terraform-%d"
  flavor_slug    			= "flex-2"
  allow_stopping_for_update = true
  image_slug     			= "pfsense-2.4.5"
  volume_size_gb			= 10
  password                  = "rivella17"
  use_private_network       = true
}`, rInt)
}
