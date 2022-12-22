package cloudscale

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"log"
	"net/http"
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
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.lb-acc-test", "name", lbName),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.lb-acc-test", "flavor_slug", "lb-flex-4-2"),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.lb-acc-test", "zone_slug", "lpg1"),
					resource.TestCheckResourceAttrSet(
						"cloudscale_load_balancer.lb-acc-test", "status"),
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

func testAccCloudscaleLoadBalancerConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer" "lb-acc-test" {
	  name        = "terraform-%d-lb"
      flavor_slug = "lb-flex-4-2"
	  zone_slug   = "lpg1"
}
`, rInt)
}
