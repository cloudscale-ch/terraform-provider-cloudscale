package cloudscale

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("cloudscale_floating_ip", &resource.Sweeper{
		Name: "cloudscale_floating_ip",
		F:    testSweepFloatingIps,
	})

}

func testSweepFloatingIps(region string) error {
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client := meta.(*cloudscale.Client)

	ips, err := client.FloatingIPs.List(context.Background())
	if err != nil {
		return err
	}

	foundError := error(nil)
	for _, ip := range ips {
		if err := client.FloatingIPs.Delete(context.Background(), ip.IP()); err != nil {
			foundError = err
		}
	}

	return foundError
}

func TestAccCloudscaleFloatingIP_Detached(t *testing.T) {
	var floatingIP cloudscale.FloatingIP

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudScaleFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudScaleFloatingIPConfig_detached(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudScaleFloatingIPExists("cloudscale_floating_ip.detached", &floatingIP),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.detached", "ip_version", "6"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.detached", "region_slug", "lpg"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.detached", "type", "regional"),
					resource.TestCheckNoResourceAttr(
						"cloudscale_floating_ip.detached", "server"),
				),
			},
		},
	})
}

func TestAccCloudscaleFloatingIP_GlobalDetached(t *testing.T) {
	var floatingIP cloudscale.FloatingIP

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudScaleFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudScaleFloatingIPConfig_globalDetached(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudScaleFloatingIPExists("cloudscale_floating_ip.global", &floatingIP),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.global", "ip_version", "6"),
					resource.TestCheckNoResourceAttr(
						"cloudscale_floating_ip.global", "region_slug"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.global", "type", "global"),
					resource.TestCheckNoResourceAttr(
						"cloudscale_floating_ip.global", "server"),
				),
			},
		},
	})
}

func TestAccCloudscaleFloatingIP_Server(t *testing.T) {
	var floatingIP cloudscale.FloatingIP
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudScaleFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudScaleFloatingIPConfig_server(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudScaleFloatingIPExists("cloudscale_floating_ip.gateway", &floatingIP),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.gateway", "ip_version", "4"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.gateway", "reverse_ptr", "vip.web-worker01.example.com"),
				),
			},
		},
	})
}

func TestAccCloudscaleFloatingIP_ServerWithZone(t *testing.T) {
	var floatingIP cloudscale.FloatingIP
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudScaleFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudScaleFloatingIPConfig_server_with_zone(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudScaleFloatingIPExists("cloudscale_floating_ip.minfloating", &floatingIP),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.minfloating", "ip_version", "6"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.minfloating", "region_slug", "lpg"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.minlpg", "zone_slug", "lpg1"),
				),
			},
		},
	})
}

func TestAccCloudscaleFloatingIP_Update(t *testing.T) {
	var beforeUpdate, afterUpdate cloudscale.FloatingIP
	rIntA := acctest.RandInt()
	rIntB := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudScaleFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudScaleFloatingIPConfig_update_first(rIntA, rIntB),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudScaleFloatingIPExists("cloudscale_floating_ip.gateway", &beforeUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.gateway", "ip_version", "4"),
				),
			},
			{
				Config: testAccCheckCloudScaleFloatingIPConfig_update_second(rIntA, rIntB),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudScaleFloatingIPExists("cloudscale_floating_ip.gateway", &afterUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.gateway", "ip_version", "4"),
					testAccCheckFloaingIPChanged(t, &beforeUpdate, &afterUpdate),
				),
			},
		},
	})
}

func testAccCheckCloudScaleFloatingIPExists(n string, floatingIP *cloudscale.FloatingIP) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*cloudscale.Client)

		id := rs.Primary.ID
		// Try to find the FloatingIP
		foundFloatingIP, err := client.FloatingIPs.Get(context.Background(), id)

		if err != nil {
			return err
		}

		if foundFloatingIP.IP() != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		*floatingIP = *foundFloatingIP

		return nil
	}
}

func testAccCheckCloudScaleFloatingIPDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudscale.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudscale_floating_ip.gateway" {
			continue
		}

		// Try to find the key
		_, err := client.FloatingIPs.Get(context.Background(), rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Floating IP still exists")
		} else {
			errorResponse, ok := err.(*cloudscale.ErrorResponse)
			if !ok || errorResponse.StatusCode != http.StatusNotFound {
				return fmt.Errorf(
					"Error waiting for floating IP (%s) to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func testAccCheckFloaingIPChanged(t *testing.T,
	before, after *cloudscale.FloatingIP) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.Server.UUID == after.Server.UUID {
			t.Fatalf("Expected a change of Server IDs got=%s",
				after.Server.UUID)
		}
		return nil
	}
}

func testAccCheckCloudScaleFloatingIPConfig_detached() string {
	return fmt.Sprintf(`
resource "cloudscale_floating_ip" "detached" {
  ip_version = 6
  region_slug = "lpg"
}`)
}

func testAccCheckCloudScaleFloatingIPConfig_globalDetached() string {
	return fmt.Sprintf(`
resource "cloudscale_floating_ip" "global" {
  ip_version = 6
  type = "global"
}`)
}

func testAccCheckCloudScaleFloatingIPConfig_server(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      					= "terraform-%d"
  flavor_slug    			= "flex-2"
  image_slug     			= "%s"
  volume_size_gb			= 10
  ssh_keys = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}
resource "cloudscale_floating_ip" "gateway" {
  server 					= "${cloudscale_server.basic.id}"
  ip_version     	= 4
	reverse_ptr 		= "vip.web-worker01.example.com"
}`, rInt, DefaultImageSlug)
}

func testAccCheckCloudScaleFloatingIPConfig_update_first(rIntA, rIntB int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      					= "terraform-%d"
  flavor_slug    			= "flex-2"
  image_slug     			= "%s"
  volume_size_gb			= 10
  ssh_keys = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}

resource "cloudscale_server" "advanced" {
  name      					= "terraform-%d"
  flavor_slug    			= "flex-2"
  image_slug     			= "%s"
  volume_size_gb			= 10
  ssh_keys = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}

resource "cloudscale_floating_ip" "gateway" {
  server 					= "${cloudscale_server.basic.id}"
  ip_version     	= 4
}`, rIntA, DefaultImageSlug, rIntB, DefaultImageSlug)
}

func testAccCheckCloudScaleFloatingIPConfig_update_second(rIntA, rIntB int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      					= "terraform-%d"
  flavor_slug    			= "flex-2"
  image_slug     			= "%s"
  volume_size_gb			= 10
  ssh_keys 						= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}

resource "cloudscale_server" "advanced" {
  name      					= "terraform-%d"
  flavor_slug    			= "flex-2"
  image_slug     			= "%s"
  volume_size_gb			= 10
  ssh_keys 						= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}

resource "cloudscale_floating_ip" "gateway" {
  server 					= "${cloudscale_server.advanced.id}"
  ip_version     	= 4
}`, rIntA, DefaultImageSlug, rIntB, DefaultImageSlug)
}

func testAccCheckCloudScaleFloatingIPConfig_server_with_zone(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "minlpg" {
  name = "terraform-%d"
  flavor_slug = "flex-2"
  image_slug = "%s"
  volume_size_gb = 10
  zone_slug = "lpg1"
  ssh_keys = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}
resource "cloudscale_floating_ip" "minfloating" {
  server = "${cloudscale_server.minlpg.id}"
  ip_version = 6
  region_slug = "lpg"
  reverse_ptr = "vip.web-worker01.example.com"
}`, rInt, DefaultImageSlug)
}
