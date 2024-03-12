package cloudscale

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
		CheckDestroy: testAccCheckCloudscaleFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleFloatingIPConfig_detached("reverse.ptr"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleFloatingIPExists("cloudscale_floating_ip.detached", &floatingIP),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.detached", "ip_version", "6"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.detached", "region_slug", "lpg"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.detached", "type", "regional"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.detached", "server", ""),
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
		CheckDestroy: testAccCheckCloudscaleFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleFloatingIPConfig_globalDetached(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleFloatingIPExists("cloudscale_floating_ip.global", &floatingIP),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.global", "ip_version", "6"),
					resource.TestCheckNoResourceAttr(
						"cloudscale_floating_ip.global", "region_slug"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.global", "type", "global"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.global", "server", ""),
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
		CheckDestroy: testAccCheckCloudscaleFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleFloatingIPConfig_server(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleFloatingIPExists("cloudscale_floating_ip.gateway", &floatingIP),
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
		CheckDestroy: testAccCheckCloudscaleFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleFloatingIPConfig_server_with_zone(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleFloatingIPExists("cloudscale_floating_ip.minfloating", &floatingIP),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.minfloating", "ip_version", "6"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.minfloating", "region_slug", "lpg"),
					resource.TestCheckResourceAttr(
						"cloudscale_server.minlpg", "zone_slug", "lpg1"),
					resource.TestCheckResourceAttrSet(
						"cloudscale_floating_ip.minfloating", "server"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.minfloating", "load_balancer", ""),
				),
			},
		},
	})
}

func TestAccCloudscaleFloatingIP_LoadBalancer(t *testing.T) {
	var floatingIP cloudscale.FloatingIP
	rInt1 := acctest.RandInt()
	rInt2 := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleFloatingIPConfig_lb_and_server(rInt1, rInt2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleFloatingIPExists("cloudscale_floating_ip.lbfloating", &floatingIP),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.lbfloating", "ip_version", "6"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.lbfloating", "server", ""),
					resource.TestCheckResourceAttrSet(
						"cloudscale_floating_ip.lbfloating", "load_balancer"),
				),
			},
			{
				Config: testAccCheckCloudscaleFloatingIPConfig_lb_and_server(rInt1, rInt2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleFloatingIPExists("cloudscale_floating_ip.lbfloating", &floatingIP),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.lbfloating", "ip_version", "6"),
					resource.TestCheckResourceAttrSet(
						"cloudscale_floating_ip.lbfloating", "server"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.lbfloating", "load_balancer", ""),
				),
			},
			{
				Config: testAccCheckCloudscaleFloatingIPConfig_lb_and_server(rInt1, rInt2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleFloatingIPExists("cloudscale_floating_ip.lbfloating", &floatingIP),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.lbfloating", "ip_version", "6"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.lbfloating", "server", ""),
					resource.TestCheckResourceAttrSet(
						"cloudscale_floating_ip.lbfloating", "load_balancer"),
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
		CheckDestroy: testAccCheckCloudscaleFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleFloatingIPConfig_update_first(rIntA, rIntB),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleFloatingIPExists("cloudscale_floating_ip.gateway", &beforeUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.gateway", "ip_version", "4"),
				),
			},
			{
				Config: testAccCheckCloudscaleFloatingIPConfig_update_second(rIntA, rIntB),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleFloatingIPExists("cloudscale_floating_ip.gateway", &afterUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.gateway", "ip_version", "4"),
					testAccCheckFloatingIPChanged(t, &beforeUpdate, &afterUpdate),
				),
			},
		},
	})
}

func TestAccCloudscaleFloatingIP_import_basic(t *testing.T) {
	var afterImport, afterUpdate cloudscale.FloatingIP

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleFloatingIPConfig_detached("cartman.ptr"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleFloatingIPExists("cloudscale_floating_ip.detached", &afterImport),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.detached", "reverse_ptr", "cartman.ptr"),
				),
			},
			{
				ResourceName:      "cloudscale_floating_ip.detached",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "cloudscale_floating_ip.detached",
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateId:     "does-not-exist",
				ExpectError:       regexp.MustCompile(`Cannot import non-existent remote object`),
			},
			{
				Config: testAccCheckCloudscaleFloatingIPConfig_detached("respect.my.authoritaaa"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleFloatingIPExists("cloudscale_floating_ip.detached", &afterUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.detached", "reverse_ptr", "respect.my.authoritaaa"),
					testAccCheckFloatingIPIsSame(t, &afterImport, &afterUpdate),
				),
			},
		},
	})
}

func TestAccCloudscaleFloatingIP_tags(t *testing.T) {
	reversePtr := "cartman.ptr"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleFloatingIPConfig_detached_withTags(reversePtr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.detached", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.detached", "tags.my-foo", "foo"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.detached", "tags.my-bar", "bar"),
					testTagsMatch("cloudscale_floating_ip.detached"),
				),
			},
			{
				Config: testAccCheckCloudscaleFloatingIPConfig_detached(reversePtr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.detached", "tags.%", "0"),
					testTagsMatch("cloudscale_floating_ip.detached"),
				),
			},
			{
				Config: testAccCheckCloudscaleFloatingIPConfig_detached_withTags(reversePtr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.detached", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.detached", "tags.my-foo", "foo"),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.detached", "tags.my-bar", "bar"),
					testTagsMatch("cloudscale_floating_ip.detached"),
				),
			},
		},
	})
}

func testAccCheckCloudscaleFloatingIPDestroy(s *terraform.State) error {
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

func testAccCheckFloatingIPIsSame(t *testing.T,
	before, after *cloudscale.FloatingIP) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if adr := before; adr == after {
			t.Fatalf("Passed the same instance twice, address is equal=%v",
				adr)
		}
		if before.Network != after.Network {
			t.Fatalf("Not expected a change of Network got=%s, expected=%s",
				after.Network, before.Network)
		}
		return nil
	}
}

func testAccCheckFloatingIPChanged(t *testing.T,
	before, after *cloudscale.FloatingIP) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.Server.UUID == after.Server.UUID {
			t.Fatalf("Expected a change of Server IDs got=%s",
				after.Server.UUID)
		}
		return nil
	}
}

func testAccCheckCloudscaleFloatingIPConfig_detached(reversePtr string) string {
	return fmt.Sprintf(`
resource "cloudscale_floating_ip" "detached" {
  ip_version = 6
  region_slug = "lpg"
  reverse_ptr = "%s"
}`, reversePtr)
}

func testAccCheckCloudscaleFloatingIPConfig_detached_withTags(reversePtr string) string {
	return fmt.Sprintf(`
resource "cloudscale_floating_ip" "detached" {
  ip_version = 6
  region_slug = "lpg"
  reverse_ptr = "%s"
  tags = {
    my-foo = "foo"
    my-bar = "bar"
  }
}`, reversePtr)
}

func testAccCheckCloudscaleFloatingIPConfig_globalDetached() string {
	return fmt.Sprintf(`
resource "cloudscale_floating_ip" "global" {
  ip_version = 6
  type = "global"
}`)
}

func testAccCheckCloudscaleFloatingIPConfig_server(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      					= "terraform-%d"
  flavor_slug    			= "flex-4-1"
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

func testAccCheckCloudscaleFloatingIPConfig_update_first(rIntA, rIntB int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      					= "terraform-%d"
  flavor_slug    			= "flex-4-1"
  image_slug     			= "%s"
  volume_size_gb			= 10
  ssh_keys = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}

resource "cloudscale_server" "advanced" {
  name      					= "terraform-%d"
  flavor_slug    			= "flex-4-1"
  image_slug     			= "%s"
  volume_size_gb			= 10
  ssh_keys = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}

resource "cloudscale_floating_ip" "gateway" {
  server 					= "${cloudscale_server.basic.id}"
  ip_version     	= 4
}`, rIntA, DefaultImageSlug, rIntB, DefaultImageSlug)
}

func testAccCheckCloudscaleFloatingIPConfig_update_second(rIntA, rIntB int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      					= "terraform-%d"
  flavor_slug    			= "flex-4-1"
  image_slug     			= "%s"
  volume_size_gb			= 10
  ssh_keys 						= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}

resource "cloudscale_server" "advanced" {
  name      					= "terraform-%d"
  flavor_slug    			= "flex-4-1"
  image_slug     			= "%s"
  volume_size_gb			= 10
  ssh_keys 						= ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}

resource "cloudscale_floating_ip" "gateway" {
  server 					= "${cloudscale_server.advanced.id}"
  ip_version     	= 4
}`, rIntA, DefaultImageSlug, rIntB, DefaultImageSlug)
}

func testAccCheckCloudscaleFloatingIPConfig_server_with_zone(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "minlpg" {
  name = "terraform-%d"
  flavor_slug = "flex-4-1"
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

func testAccCheckCloudscaleFloatingIPConfig_lb_and_server(rInt1, rInt2 int, assignLB bool) string {
	assignment := `load_balancer = "${cloudscale_load_balancer.lb1.id}"`
	if !assignLB {
		assignment = `server        = "${cloudscale_server.minlpg.id}"`
	}
	return fmt.Sprintf(`
resource "cloudscale_load_balancer" "lb1" {
  name        = "terraform-%d"
  flavor_slug = "lb-standard"
  zone_slug   = "lpg1"
}
resource "cloudscale_server" "minlpg" {
  name = "terraform-%d"
  flavor_slug = "flex-4-1"
  image_slug = "%s"
  volume_size_gb = 10
  zone_slug = "lpg1"
  ssh_keys = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}

resource "cloudscale_floating_ip" "lbfloating" {
  %s
  ip_version = 6
  region_slug = "lpg"
}`, rInt1, rInt2, DefaultImageSlug, assignment)
}
