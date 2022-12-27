package cloudscale

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudscaleLoadBalancerPool_Basic(t *testing.T) {
	var loadBalancer cloudscale.LoadBalancer
	var loadBalancerPool cloudscale.LoadBalancerPool

	rInt := acctest.RandInt()
	lbPoolName := fmt.Sprintf("terraform-%d-lb-pool", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt) + testAccCloudscaleLoadBalancerPoolConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists("cloudscale_load_balancer.lb-acc-test", &loadBalancer),
					testAccCheckCloudscaleLoadBalancerPoolExists("cloudscale_load_balancer_pool.lb-pool-acc-test", &loadBalancerPool),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer_pool.lb-pool-acc-test", "name", lbPoolName),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer_pool.lb-pool-acc-test", "algorithm", "round_robin"),
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer_pool.lb-pool-acc-test", "protocol", "tcp"),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_load_balancer_pool.lb-pool-acc-test", "load_balancer_uuid", &loadBalancer.UUID),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_load_balancer_pool.lb-pool-acc-test", "load_balancer_name", &loadBalancer.Name),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_load_balancer_pool.lb-pool-acc-test", "load_balancer_href", &loadBalancer.HREF),
				),
			},
		},
	})
}

func testAccCheckCloudscaleLoadBalancerPoolExists(n string, pool *cloudscale.LoadBalancerPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Load Balancer Pool ID is set")
		}

		client := testAccProvider.Meta().(*cloudscale.Client)

		id := rs.Primary.ID

		// Try to find the load balancer
		retrieveLoadBalancerPool, err := client.LoadBalancerPools.Get(context.Background(), id)

		if err != nil {
			return err
		}

		if retrieveLoadBalancerPool.UUID != rs.Primary.ID {
			return fmt.Errorf("Load Balancer Pool not found")
		}

		*pool = *retrieveLoadBalancerPool

		return nil
	}
}

func testAccCloudscaleLoadBalancerPoolConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer_pool" "lb-pool-acc-test" {
  name = "terraform-%d-lb-pool"
  algorithm = "round_robin"
  protocol = "tcp"
  load_balancer_uuid = cloudscale_load_balancer.lb-acc-test.id
}
`, rInt)
}
