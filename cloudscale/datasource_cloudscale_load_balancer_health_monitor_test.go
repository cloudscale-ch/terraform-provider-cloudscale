package cloudscale

import (
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v4"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

func TestAccCloudscaleLoadBalancerHealthMonitor_DS_Basic(t *testing.T) {
	var loadBalancerHealthMonitor cloudscale.LoadBalancerHealthMonitor
	rInt := acctest.RandInt()

	config := loadBalancerConfig_baseline(2, rInt) +
		loadBalancerPoolConfig_baseline(2, rInt) +
		loadBalancerHealthMonitorConfig_baseline(2, rInt)

	resourceName1 := "cloudscale_load_balancer_health_monitor.basic.0"
	dataSourceName := "data.cloudscale_load_balancer_health_monitor.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				Config: config + testAccCheckCloudscaleLoadBalancerHealthMonitorConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerHealthMonitorExists(resourceName1, &loadBalancerHealthMonitor),
					resource.TestCheckResourceAttr(
						resourceName1, "delay_s", fmt.Sprintf("8%d", 0)),
					resource.TestCheckResourceAttr(
						dataSourceName, "delay_s", fmt.Sprintf("8%d", 0)),
					resource.TestCheckResourceAttrPtr(
						resourceName1, "id", &loadBalancerHealthMonitor.UUID),
					resource.TestCheckResourceAttrPtr(
						dataSourceName, "id", &loadBalancerHealthMonitor.UUID),
				),
			},
			{
				Config:      config + `data "cloudscale_load_balancer_health_monitor" "foo" {}`,
				ExpectError: regexp.MustCompile(`Found \d+ load balancer health monitors, expected one`),
			},
		},
	})
}

func testAccCheckCloudscaleLoadBalancerHealthMonitorConfig_id() string {
	return fmt.Sprintf(`
data "cloudscale_load_balancer_health_monitor" "foo" {
  id = cloudscale_load_balancer_health_monitor.basic.0.id
}
`)
}

func loadBalancerHealthMonitorConfig_baseline(count int, rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer_health_monitor" "basic" {
  pool_uuid        = cloudscale_load_balancer_pool.basic[count.index].id
  delay_s          = "8${count.index}"
  up_threshold     = 3
  down_threshold   = 3
  timeout_s        = 5
  count		       =  %d
  type             = "tcp"
}
`, count)
}
