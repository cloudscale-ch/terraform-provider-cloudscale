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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists("cloudscale_load_balancer.lb-acc-test", &loadBalancer),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.lb-acc-test", "name", lbName),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.lb-acc-test", "flavor_slug", "lb-flex-4-2"),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.lb-acc-test", "zone_slug", "lpg1"),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.lb-acc-test", "status", "running"),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.lb-acc-test", "vip_addresses.#", "1"),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.lb-acc-test", "vip_addresses.0.version", "4"),
					resource.TestCheckResourceAttrSet(
						"cloudscale_load_balancer.lb-acc-test", "vip_addresses.0.address"),
					resource.TestCheckResourceAttrSet(
						"cloudscale_load_balancer.lb-acc-test", "vip_addresses.0.subnet_href"),
					resource.TestCheckResourceAttrSet(
						"cloudscale_load_balancer.lb-acc-test", "vip_addresses.0.subnet_cidr"),
					resource.TestCheckResourceAttrSet(
						"cloudscale_load_balancer.lb-acc-test", "vip_addresses.0.subnet_uuid"),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists("cloudscale_load_balancer.lb-acc-test", &afterCreate),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.lb-acc-test", "name", lbName),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists("cloudscale_load_balancer.lb-acc-test", &afterUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.lb-acc-test", "name", updatedLBName),
					testAccCheckLoadBalancerIsSame(t, &afterCreate, &afterUpdate),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancer_PrivateNetwork(t *testing.T) {
	var loadBalancer cloudscale.LoadBalancer

	rInt1, rInt2 := acctest.RandInt(), acctest.RandInt()
	cidr := "192.168.42.0/24"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_priateNetwork(rInt1, rInt2, cidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists("cloudscale_load_balancer.lb-acc-test", &loadBalancer),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.lb-acc-test", "vip_addresses.#", "1"),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.lb-acc-test", "vip_addresses.0.version", "4"),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.lb-acc-test", "vip_addresses.0.address", "192.168.42.124"),
					resource.TestCheckResourceAttrSet(
						"cloudscale_load_balancer.lb-acc-test", "vip_addresses.0.subnet_href"),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.lb-acc-test", "vip_addresses.0.subnet_cidr", cidr),
					resource.TestCheckResourceAttrSet(
						"cloudscale_load_balancer.lb-acc-test", "vip_addresses.0.subnet_uuid"),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1),
			},
			{
				ResourceName:            "cloudscale_load_balancer.lb-acc-test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists("cloudscale_load_balancer.lb-acc-test", &afterImport),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.lb-acc-test", "name", lbName),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists("cloudscale_load_balancer.lb-acc-test", &afterUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.lb-acc-test", "name", updatedName),
					testAccCheckLoadBalancerIsSame(t, &afterImport, &afterUpdate),
				),
			},
			{
				ResourceName:      "cloudscale_load_balancer.lb-acc-test",
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudscaleLoadBalancerConfigWithZoneAndTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists("cloudscale_load_balancer.tagged", &afterImport),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.tagged", "name", fmt.Sprintf("terraform-%d-lb", rInt)),
					testTagsMatch("cloudscale_load_balancer.tagged"),
				),
			},
			{
				ResourceName:      "cloudscale_load_balancer.tagged",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "cloudscale_load_balancer.tagged",
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateId:     "does-not-exist",
				ExpectError:       regexp.MustCompile(`Cannot import non-existent remote object`),
			},
			{
				Config: testAccCheckCloudscaleLoadBalancerConfigWithZone(42),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists("cloudscale_load_balancer.tagged", &afterUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.tagged", "name", "terraform-42-lb"),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.tagged", "tags.%", "0"),
					testAccCheckLoadBalancerIsSame(t, &afterImport, &afterUpdate),
					testTagsMatch("cloudscale_load_balancer.tagged"),
				),
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

func testAccCheckCloudscaleLoadBalancerExists(n string, loadBalancer *cloudscale.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Load Balancer ID is set")
		}

		client := testAccProvider.Meta().(*cloudscale.Client)

		id := rs.Primary.ID

		// Try to find the load balancer
		retrieveLoadBalancer, err := client.LoadBalancers.Get(context.Background(), id)

		if err != nil {
			return err
		}

		if retrieveLoadBalancer.UUID != rs.Primary.ID {
			return fmt.Errorf("Load Balancer not found")
		}

		*loadBalancer = *retrieveLoadBalancer

		return nil
	}
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
      flavor_slug = "lb-flex-4-2"
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
  flavor_slug = "lb-flex-4-2"
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
  flavor_slug = "lb-flex-4-2"
  zone_slug   = "lpg1"
}
`, rInt)
}

func testAccCheckCloudscaleLoadBalancerConfigWithZoneAndTags(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer" "tagged" {
  name        = "terraform-%d-lb"
  flavor_slug = "lb-flex-4-2"
  zone_slug   = "lpg1"
  tags = {
    my-foo = "foo"
    my-bar = "bar"
  }
}
`, rInt)
}
