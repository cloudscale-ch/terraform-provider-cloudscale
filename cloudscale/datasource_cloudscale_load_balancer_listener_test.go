package cloudscale

import (
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

func TestAccCloudscaleLoadBalancerListener_DS_Basic(t *testing.T) {
	var loadBalancerListener cloudscale.LoadBalancerListener
	rInt := acctest.RandInt()
	name1 := fmt.Sprintf("terraform-%d-lb-listener-0", rInt)
	name2 := fmt.Sprintf("terraform-%d-lb-listener-1", rInt)

	config := loadBalancerConfig_baseline(1, rInt) +
		loadBalancerPoolConfig_baseline(1, rInt) +
		loadBalancerListenerConfig_baseline(2, rInt)

	poolResourceName1 := "cloudscale_load_balancer_pool.basic.0"
	resourceName1 := "cloudscale_load_balancer_listener.basic.0"
	resourceName2 := "cloudscale_load_balancer_listener.basic.1"
	dataSourceName := "data.cloudscale_load_balancer_listener.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				Config: config + testAccCheckCloudscaleLoadBalancerListenerConfig_name(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerListenerExists(dataSourceName, &loadBalancerListener),
					resource.TestCheckResourceAttrPtr(
						resourceName1, "id", &loadBalancerListener.UUID),
					resource.TestCheckResourceAttrPtr(
						dataSourceName, "id", &loadBalancerListener.UUID),
					resource.TestCheckResourceAttr(
						dataSourceName, "name", name1),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "name", resourceName1, "name"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "pool_uuid", resourceName1, "pool_uuid"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "pool_uuid", poolResourceName1, "id"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleLoadBalancerListenerConfig_name(name2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						dataSourceName, "name", name2),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "name", resourceName2, "name"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleLoadBalancerListenerConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName1, "name", name1),
					resource.TestCheckResourceAttr(
						dataSourceName, "name", name1),
					resource.TestCheckResourceAttrPtr(
						resourceName1, "id", &loadBalancerListener.UUID),
					resource.TestCheckResourceAttrPtr(
						dataSourceName, "id", &loadBalancerListener.UUID),
				),
			},
			{
				Config:      config + `data "cloudscale_load_balancer_listener" "foo" {}`,
				ExpectError: regexp.MustCompile(`Found \d+ load balancer listeners, expected one`),
			},
		},
	})
}

func testAccCheckCloudscaleLoadBalancerListenerConfig_name(name string) string {
	return fmt.Sprintf(`
data "cloudscale_load_balancer_listener" "foo" {
  name = "%s"
}
`, name)
}

func testAccCheckCloudscaleLoadBalancerListenerConfig_id() string {
	return fmt.Sprintf(`
data "cloudscale_load_balancer_listener" "foo" {
  id = cloudscale_load_balancer_listener.basic.0.id
}
`)
}

func loadBalancerListenerConfig_baseline(count int, rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer_listener" "basic" {
  count         = %[1]d
  name          = "terraform-%[2]d-lb-listener-${count.index}"
  pool_uuid     = cloudscale_load_balancer_pool.basic.0.id
  protocol      = "tcp"
  protocol_port = "8${count.index}"
}
`, count, rInt)
}
