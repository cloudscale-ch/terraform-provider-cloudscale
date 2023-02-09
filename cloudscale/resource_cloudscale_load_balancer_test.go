package cloudscale

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"log"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("cloudscale_load_balancer", &resource.Sweeper{
		Name: "cloudscale_load_balancer",
		F:    testSweepLoadBalancers,
	})
}

func testSweepLoadBalancers(region string) error {
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client := meta.(*cloudscale.Client)

	loadBalancers, err := client.LoadBalancers.List(context.Background())
	if err != nil {
		return err
	}

	foundError := error(nil)
	for _, lb := range loadBalancers {
		if strings.HasPrefix(lb.Name, "terraform-") {
			log.Printf("Destroying load balancer %s", lb.Name)

			if err := client.LoadBalancers.Delete(context.Background(), lb.UUID); err != nil {
				foundError = err
			}
		}
	}
	return foundError
}

func TestAccCloudscaleLoadBalancer_Basic(t *testing.T) {
	var loadBalancer cloudscale.LoadBalancer

	rInt := acctest.RandInt()
	lbName := fmt.Sprintf("terraform-%d-lb", rInt)

	resourceName := "cloudscale_load_balancer.lb-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists(resourceName, &loadBalancer),
					resource.TestCheckResourceAttrPtr(
						resourceName, "href", &loadBalancer.HREF),
					resource.TestCheckResourceAttrPtr(
						resourceName, "id", &loadBalancer.UUID),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbName),
					resource.TestCheckResourceAttr(
						resourceName, "flavor_slug", "lb-small"),
					resource.TestCheckResourceAttr(
						resourceName, "zone_slug", "lpg1"),
					resource.TestCheckResourceAttr(
						resourceName, "status", "running"),
					resource.TestCheckResourceAttr(
						resourceName, "vip_addresses.#", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "vip_addresses.0.version", "4"),
					resource.TestCheckResourceAttrSet(
						resourceName, "vip_addresses.0.address"),
					resource.TestCheckResourceAttrSet(
						resourceName, "vip_addresses.0.subnet_href"),
					resource.TestCheckResourceAttrSet(
						resourceName, "vip_addresses.0.subnet_cidr"),
					resource.TestCheckResourceAttrSet(
						resourceName, "vip_addresses.0.subnet_uuid"),
					resource.TestCheckResourceAttr(
						resourceName, "vip_addresses.1.version", "6"),
					resource.TestCheckResourceAttrSet(
						resourceName, "vip_addresses.1.address"),
					resource.TestCheckResourceAttrSet(
						resourceName, "vip_addresses.1.subnet_href"),
					resource.TestCheckResourceAttrSet(
						resourceName, "vip_addresses.1.subnet_cidr"),
					resource.TestCheckResourceAttrSet(
						resourceName, "vip_addresses.1.subnet_uuid"),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancer_UpdateName(t *testing.T) {
	var afterCreate, afterUpdate cloudscale.LoadBalancer

	rInt1 := acctest.RandInt()
	lbName := fmt.Sprintf("terraform-%d-lb", rInt1)
	rInt2 := acctest.RandInt()
	updatedLBName := fmt.Sprintf("terraform-%d-lb", rInt2)

	resourceName := "cloudscale_load_balancer.lb-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists(resourceName, &afterCreate),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbName),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "name", updatedLBName),
					testAccCheckLoadBalancerIsSame(t, &afterCreate, &afterUpdate),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancer_PrivateNetwork(t *testing.T) {
	var loadBalancer cloudscale.LoadBalancer
	var subnet cloudscale.Subnet

	rInt1, rInt2 := acctest.RandInt(), acctest.RandInt()
	cidr := "192.168.42.0/24"

	subnetResourceName := "cloudscale_subnet.privnet-subnet"
	resourceName := "cloudscale_load_balancer.lb-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_priateNetwork(rInt1, rInt2, cidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleSubnetExists(subnetResourceName, &subnet),
					testAccCheckCloudscaleLoadBalancerExists(resourceName, &loadBalancer),
					resource.TestCheckResourceAttr(
						resourceName, "vip_addresses.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "vip_addresses.0.version", "4"),
					resource.TestCheckResourceAttr(
						resourceName, "vip_addresses.0.address", "192.168.42.124"),
					resource.TestCheckResourceAttrPtr(
						resourceName, "vip_addresses.0.subnet_href", &subnet.HREF),
					resource.TestCheckResourceAttr(
						resourceName, "vip_addresses.0.subnet_cidr", cidr),
					resource.TestCheckResourceAttrPtr(
						resourceName, "vip_addresses.0.subnet_uuid", &subnet.UUID),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancer_import_basic(t *testing.T) {
	var afterImport, afterUpdate cloudscale.LoadBalancer

	rInt1 := acctest.RandInt()
	lbName := fmt.Sprintf("terraform-%d-lb", rInt1)
	rInt2 := acctest.RandInt()
	updatedName := fmt.Sprintf("terraform-%d-lb", rInt2)

	resourceName := "cloudscale_load_balancer.lb-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists(resourceName, &afterImport),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbName),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "name", updatedName),
					testAccCheckLoadBalancerIsSame(t, &afterImport, &afterUpdate),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateId:     "does-not-exist",
				ExpectError:       regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancer_import_withTags(t *testing.T) {
	var afterImport, afterUpdate cloudscale.LoadBalancer

	rInt := acctest.RandInt()

	resourceName := "cloudscale_load_balancer.tagged"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleLoadBalancerConfigWithZoneAndTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists(resourceName, &afterImport),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("terraform-%d-lb", rInt)),
					testTagsMatch(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateId:     "does-not-exist",
				ExpectError:       regexp.MustCompile(`Cannot import non-existent remote object`),
			},
			{
				Config: testAccCheckCloudscaleLoadBalancerConfigWithZone(42),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "name", "terraform-42-lb"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "0"),
					testAccCheckLoadBalancerIsSame(t, &afterImport, &afterUpdate),
					testTagsMatch(resourceName),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancer_tags(t *testing.T) {
	rInt := acctest.RandInt()

	resourceName := "cloudscale_load_balancer.tagged"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleLoadBalancerConfigWithZoneAndTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.my-foo", "foo"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.my-bar", "bar"),
					testTagsMatch(resourceName),
				),
			},
			{
				Config: testAccCheckCloudscaleLoadBalancerConfigWithZone(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "0"),
					testTagsMatch(resourceName),
				),
			},
			{
				Config: testAccCheckCloudscaleLoadBalancerConfigWithZoneAndTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.my-foo", "foo"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.my-bar", "bar"),
					testTagsMatch(resourceName),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerPoolMember_ScenarioSubnet(t *testing.T) {
	// this is a big test case, where we verify a scenario with
	// a server in a private network and ensure that the health monitor status
	// becomes "up".
	var loadBalancerPoolMember cloudscale.LoadBalancerPoolMember

	rInt1 := acctest.RandInt()
	rInt2 := acctest.RandInt()
	rInt3 := acctest.RandInt()
	rInt4 := acctest.RandInt()
	rInt5 := acctest.RandInt()
	rInt6 := acctest.RandInt()

	resourceName := "cloudscale_load_balancer_pool_member.basic"
	expectedMonitorStatus := "up"
	config := testAccCloudscaleLoadBalancerConfig_subnet(rInt1, rInt2, rInt3, rInt4, rInt5, rInt6)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolMemberExists(resourceName, &loadBalancerPoolMember),
					waitForMonitorStatus(&loadBalancerPoolMember, expectedMonitorStatus)),
			},
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					// this check is in a separate step to ensure the status is refreshed form the API:
					resource.TestCheckResourceAttr(
						resourceName, "monitor_status", expectedMonitorStatus)),
			},
		},
	})
}

func testAccCheckCloudscaleLoadBalancerDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudscale.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudscale_load_balancer" {
			continue
		}

		id := rs.Primary.ID

		// Try to find the load balancer
		lb, err := client.LoadBalancers.Get(context.Background(), id)

		// Wait

		if err == nil {
			return fmt.Errorf("The load balancer %v remained, even though the resource was destoryed", lb)
		} else {
			errorResponse, ok := err.(*cloudscale.ErrorResponse)
			if !ok || errorResponse.StatusCode != http.StatusNotFound {
				return fmt.Errorf(
					"Error waiting for load balancer (%s) to be destroyed: %lb",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func testAccCheckLoadBalancerIsSame(t *testing.T,
	before, after *cloudscale.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if adr := before; adr == after {
			t.Fatalf("Passed the same instance twice, address is equal=%v",
				adr)
		}
		if before.UUID != after.UUID {
			t.Fatalf("Not expected a change of LoadBalancer IDs got=%s, expected=%s",
				after.UUID, before.UUID)
		}
		return nil
	}
}

func testAccCloudscaleLoadBalancerConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer" "lb-acc-test" {
	  name        = "terraform-%d-lb"
      flavor_slug = "lb-small"
	  zone_slug   = "lpg1"
}
`, rInt)
}

func testAccCloudscaleLoadBalancerConfig_priateNetwork(rInt1, rInt2 int, cidr string) string {
	return fmt.Sprintf(`
resource "cloudscale_network" "privnet" {
  name                    = "terraform-%d-network"
  zone_slug               = "lpg1"
  mtu                     = "9000"
  auto_create_ipv4_subnet = "false"
}

resource "cloudscale_subnet" "privnet-subnet" {
  cidr               = "%s"
  network_uuid       = cloudscale_network.privnet.id
}

resource "cloudscale_load_balancer" "lb-acc-test" {
  name        = "terraform-%d-lb"
  flavor_slug = "lb-small"
  zone_slug   = "lpg1"

  vip_addresses {
    subnet_uuid = cloudscale_subnet.privnet-subnet.id
    address     = "192.168.42.124"
  }
}
`, rInt1, cidr, rInt2)
}

func testAccCheckCloudscaleLoadBalancerConfigWithZone(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer" "tagged" {
  name        = "terraform-%d-lb"
  flavor_slug = "lb-small"
  zone_slug   = "lpg1"
}
`, rInt)
}

func testAccCheckCloudscaleLoadBalancerConfigWithZoneAndTags(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer" "tagged" {
  name        = "terraform-%d-lb"
  flavor_slug = "lb-small"
  zone_slug   = "lpg1"
  tags = {
    my-foo = "foo"
    my-bar = "bar"
  }
}
`, rInt)
}

func testAccCloudscaleLoadBalancerConfig_subnet(rInt1, rInt2, rInt3, rInt4, rInt5, rInt6 int) string {
	return fmt.Sprintf(`
resource "cloudscale_network" "privnet" {
  name                    = "terraform-%1d"
  zone_slug               = "lpg1"
  mtu                     = "9000"
  auto_create_ipv4_subnet = "false"
}

resource "cloudscale_subnet" "privnet-subnet" {
  cidr               = "10.11.12.0/24"
  network_uuid       = cloudscale_network.privnet.id
}

resource "cloudscale_server" "fixed" {
  name            = "terraform-%2d"
  zone_slug       = "lpg1"
  flavor_slug     = "flex-4-1"
  image_slug      = "debian-10"
  interfaces      {
    type          = "private"
    addresses {
      subnet_uuid = "${cloudscale_subnet.privnet-subnet.id}"     
      address     = "10.11.12.13"
    }
  }
  volume_size_gb  = 10
  ssh_keys        = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY=", "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}

resource "cloudscale_load_balancer" "basic" {
  name        = "terraform-%3d"
  flavor_slug = "lb-small"
  zone_slug   = "lpg1"
}

resource "cloudscale_load_balancer_pool" "basic" {
  name               = "terraform-%4d"
  algorithm          = "round_robin"
  protocol           = "tcp"
  load_balancer_uuid = "${cloudscale_load_balancer.basic.id}"
}

resource "cloudscale_load_balancer_listener" "basic" {
  name          = "terraform-%5d"
  pool_uuid     = "${cloudscale_load_balancer_pool.basic.id}"
  protocol      = "tcp"
  protocol_port = "22"
}

resource "cloudscale_load_balancer_health_monitor" "basic" {
  pool_uuid        = "${cloudscale_load_balancer_pool.basic.id}"
  delay_s          = 10
  up_threshold     = 3
  down_threshold   = 3
  timeout_s        = 5
  type             = "tcp"
}

resource "cloudscale_load_balancer_pool_member" "basic" {
  name          = "terraform-%6d"
  pool_uuid     = "${cloudscale_load_balancer_pool.basic.id}"
  protocol_port = 22
  subnet_uuid   = "${cloudscale_subnet.privnet-subnet.id}"    
  address       = "${cloudscale_server.fixed.interfaces[0].addresses[0].address}"
}
`, rInt1, rInt2, rInt3, rInt4, rInt5, rInt6)
}
