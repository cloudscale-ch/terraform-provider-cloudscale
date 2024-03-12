package cloudscale

import (
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

func TestAccCloudscaleLoadBalancerPool_DS_Basic(t *testing.T) {
	var loadBalancerPool cloudscale.LoadBalancerPool
	rInt := acctest.RandInt()
	name1 := fmt.Sprintf("terraform-%d-lb-pool-0", rInt)
	name2 := fmt.Sprintf("terraform-%d-lb-pool-1", rInt)

	config := loadBalancerConfig_baseline(2, rInt) +
		loadBalancerPoolConfig_baseline(2, rInt)

	lbResourceName1 := "cloudscale_load_balancer.basic.0"
	resourceName1 := "cloudscale_load_balancer_pool.basic.0"
	resourceName2 := "cloudscale_load_balancer_pool.basic.1"
	dataSourceName := "data.cloudscale_load_balancer_pool.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				Config: config + testAccCheckCloudscaleLoadBalancerPoolConfig_name(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolExists(dataSourceName, &loadBalancerPool),
					resource.TestCheckResourceAttrPtr(
						resourceName1, "id", &loadBalancerPool.UUID),
					resource.TestCheckResourceAttrPtr(
						dataSourceName, "id", &loadBalancerPool.UUID),
					resource.TestCheckResourceAttr(
						dataSourceName, "name", name1),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "name", resourceName1, "name"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "load_balancer_uuid", resourceName1, "load_balancer_uuid"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "load_balancer_uuid", lbResourceName1, "id"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleLoadBalancerPoolConfig_name(name2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						dataSourceName, "name", name2),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "name", resourceName2, "name"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleLoadBalancerPoolConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName1, "name", name1),
					resource.TestCheckResourceAttr(
						dataSourceName, "name", name1),
					resource.TestCheckResourceAttrPtr(
						resourceName1, "id", &loadBalancerPool.UUID),
					resource.TestCheckResourceAttrPtr(
						dataSourceName, "id", &loadBalancerPool.UUID),
				),
			},
			{
				Config:      config + "\n" + `data "cloudscale_load_balancer_pool" "foo" {}`,
				ExpectError: regexp.MustCompile(`Found \d+ load balancer pools, expected one`),
			},
		},
	})
}

func loadBalancerPoolConfig_baseline(count int, rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer_pool" "basic" {
  count              = %d
  name               = "terraform-%d-lb-pool-${count.index}"
  algorithm          = "round_robin"
  protocol           = "tcp"
  load_balancer_uuid = cloudscale_load_balancer.basic[count.index].id
}`, count, rInt)
}

func testAccCheckCloudscaleLoadBalancerPoolConfig_name(name string) string {
	return fmt.Sprintf(`
data "cloudscale_load_balancer_pool" "foo" {
 name = "%s"
}
`, name)
}

func testAccCheckCloudscaleLoadBalancerPoolConfig_id() string {
	return fmt.Sprintf(`
data "cloudscale_load_balancer_pool" "foo" {
 id = "${cloudscale_load_balancer_pool.basic.0.id}"
}
`)
}
