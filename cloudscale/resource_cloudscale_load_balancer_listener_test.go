package cloudscale

import (
	"context"
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccCloudscaleLoadBalancerListener_Basic(t *testing.T) {
	var loadBalancer cloudscale.LoadBalancer
	var loadBalancerPool cloudscale.LoadBalancerPool
	var loadBalancerListener cloudscale.LoadBalancerListener

	rInt := acctest.RandInt()
	lbListenerName := fmt.Sprintf("terraform-%d-lb-listener", rInt)

	resourceName := "cloudscale_load_balancer_listener.lb-listener-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerListenerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists("cloudscale_load_balancer.lb-acc-test", &loadBalancer),
					testAccCheckCloudscaleLoadBalancerPoolExists("cloudscale_load_balancer_pool.lb-pool-acc-test", &loadBalancerPool),
					testAccCheckCloudscaleLoadBalancerListenerExists(resourceName, &loadBalancerListener),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbListenerName),
					resource.TestCheckResourceAttr(
						resourceName, "protocol_port", "80"),
					resource.TestCheckResourceAttr(
						resourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttrPtr(
						resourceName, "pool_uuid", &loadBalancerPool.UUID),
					resource.TestCheckResourceAttrPtr(
						resourceName, "pool_name", &loadBalancerPool.Name),
					resource.TestCheckResourceAttrPtr(
						resourceName, "pool_href", &loadBalancerPool.HREF),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerListener_UpdateName(t *testing.T) {
	var afterCreate, afterUpdate cloudscale.LoadBalancerListener

	rInt1 := acctest.RandInt()
	lbListenerName := fmt.Sprintf("terraform-%d-lb-listener", rInt1)
	rInt2 := acctest.RandInt()
	lbListenerNameUpdated := fmt.Sprintf("terraform-%d-lb-listener", rInt2)

	resourceName := "cloudscale_load_balancer_listener.lb-listener-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerListenerConfig_basic(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerListenerExists(resourceName, &afterCreate),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbListenerName),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerListenerConfig_basic(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerListenerExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbListenerNameUpdated),
					testAccCheckLoadBalancerListenerIsSame(t, &afterCreate, &afterUpdate, true),
				),
			},
		},
	})
}

func testAccCheckCloudscaleLoadBalancerListenerExists(n string, listener *cloudscale.LoadBalancerListener) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Load Balancer Pool Member ID is set")
		}

		client := testAccProvider.Meta().(*cloudscale.Client)

		id := rs.Primary.ID

		// Try to find the load balancer
		retrieveLoadBalancerPool, err := client.LoadBalancerListeners.Get(context.Background(), id)

		if err != nil {
			return err
		}

		if retrieveLoadBalancerPool.UUID != rs.Primary.ID {
			return fmt.Errorf("Load Balancer Pool Member not found")
		}

		*listener = *retrieveLoadBalancerPool

		return nil
	}

}

func testAccCloudscaleLoadBalancerListenerConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer_listener" "lb-listener-acc-test" {
	name = "terraform-%d-lb-listener"
    pool_uuid = cloudscale_load_balancer_pool.lb-pool-acc-test.id
    protocol = "tcp"
    protocol_port = 80
}
`, rInt)
}

func testAccCheckLoadBalancerListenerIsSame(t *testing.T,
	before *cloudscale.LoadBalancerListener, after *cloudscale.LoadBalancerListener,
	expectSame bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if adr := before; adr == after {
			t.Fatalf("Passed the same instance twice, address is equal=%v",
				adr)
		}
		isSame := before.UUID == after.UUID
		if isSame != expectSame {
			t.Fatalf("Unexpected LoadBalancerPoolMember IDs got=%s, expected=%s, isSame=%t",
				after.UUID, before.UUID, isSame)
		}
		return nil
	}
}
