package cloudscale

import (
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccCloudscaleLoadBalancerHealthMonitor_Basic(t *testing.T) {
	var loadBalancer cloudscale.LoadBalancer
	var loadBalancerPool cloudscale.LoadBalancerPool
	var loadBalancerHealthMonitor cloudscale.LoadBalancerHealthMonitor

	rInt := acctest.RandInt()

	resourceName := "cloudscale_load_balancer_health_monitor.lb-health_monitor-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerHealthMonitorConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists("cloudscale_load_balancer.lb-acc-test", &loadBalancer),
					testAccCheckCloudscaleLoadBalancerPoolExists("cloudscale_load_balancer_pool.lb-pool-acc-test", &loadBalancerPool),
					testAccCheckCloudscaleLoadBalancerHealthMonitorExists(resourceName, &loadBalancerHealthMonitor),
					resource.TestCheckResourceAttr(
						resourceName, "delay", "10"),
					resource.TestCheckResourceAttr(
						resourceName, "max_retries", "3"),
					resource.TestCheckResourceAttr(
						resourceName, "max_retries_down", "3"),
					resource.TestCheckResourceAttr(
						resourceName, "timeout", "5"),
					resource.TestCheckResourceAttr(
						resourceName, "type", "tcp"),
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

func testAccCloudscaleLoadBalancerHealthMonitorConfig_basic() string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer_health_monitor" "lb-health_monitor-acc-test" {
  pool_uuid        = cloudscale_load_balancer_pool.lb-pool-acc-test.id
  delay            = 10
  max_retries      = 3
  max_retries_down = 3
  timeout          = 5
  type             = "tcp"
}
`)
}
