package cloudscale

import (
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

func TestAccCloudscaleLoadBalancerPoolMember_DS_Basic(t *testing.T) {
	var loadBalancerPoolMember cloudscale.LoadBalancerPoolMember
	rInt := acctest.RandInt()
	name1 := fmt.Sprintf("terraform-%d-lb-pool-member-0", rInt)
	name2 := fmt.Sprintf("terraform-%d-lb-pool-member-1", rInt)

	config := loadBalancerConfig_baseline(1, rInt) +
		loadBalancerPoolConfig_baseline(1, rInt) +
		loadBalancerPoolMemberConfig_baseline(2, rInt)

	poolResourceName1 := "cloudscale_load_balancer_pool.basic.0"
	resourceName1 := "cloudscale_load_balancer_pool_member.basic.0"
	resourceName2 := "cloudscale_load_balancer_pool_member.basic.1"
	dataSourceName := "data.cloudscale_load_balancer_pool_member.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				Config: config + testAccCheckCloudscaleLoadBalancerPoolMemberConfig_name(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolMemberExists(dataSourceName, &loadBalancerPoolMember),
					resource.TestCheckResourceAttrPtr(
						resourceName1, "id", &loadBalancerPoolMember.UUID),
					resource.TestCheckResourceAttrPtr(
						dataSourceName, "id", &loadBalancerPoolMember.UUID),
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
				Config: config + testAccCheckCloudscaleLoadBalancerPoolMemberConfig_name(name2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						dataSourceName, "name", name2),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "name", resourceName2, "name"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleLoadBalancerPoolMemberConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName1, "name", name1),
					resource.TestCheckResourceAttr(
						dataSourceName, "name", name1),
					resource.TestCheckResourceAttrPtr(
						resourceName1, "id", &loadBalancerPoolMember.UUID),
					resource.TestCheckResourceAttrPtr(
						dataSourceName, "id", &loadBalancerPoolMember.UUID),
				),
			},
			{
				Config:      config + testAccCheckCloudscaleLoadBalancerPoolMemberConfig_pool(),
				ExpectError: regexp.MustCompile(`Found \d+ load balancer pool members, expected one`),
			},
		},
	})
}

func testAccCheckCloudscaleLoadBalancerPoolMemberConfig_name(name string) string {
	return fmt.Sprintf(`
data "cloudscale_load_balancer_pool_member" "foo" {
  name      = "%s"
  pool_uuid = cloudscale_load_balancer_pool.basic.0.id
}
`, name)
}

func testAccCheckCloudscaleLoadBalancerPoolMemberConfig_id() string {
	return fmt.Sprintf(`
data "cloudscale_load_balancer_pool_member" "foo" {
  id        = cloudscale_load_balancer_pool_member.basic.0.id
  pool_uuid = cloudscale_load_balancer_pool.basic.0.id
}
`)
}

func loadBalancerPoolMemberConfig_baseline(count int, rInt int) string {
	return fmt.Sprintf(`
%s

resource "cloudscale_load_balancer_pool_member" "basic" {
  count         = %d
  name          = "terraform-%d-lb-pool-member-${count.index}"
  pool_uuid     = cloudscale_load_balancer_pool.basic.0.id
  protocol_port = 80
  address       = "10.0.0.${count.index}"
  subnet_uuid   = cloudscale_subnet.lb-subnet.id
}
`, testAccCloudscaleLoadBalancerSubnet(rInt), count, rInt)
}

func testAccCheckCloudscaleLoadBalancerPoolMemberConfig_pool() string {
	return fmt.Sprintf(`
data "cloudscale_load_balancer_pool_member" "foo" {
  pool_uuid = cloudscale_load_balancer_pool.basic.0.id
}
`)
}
