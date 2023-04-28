package cloudscale

import (
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

func TestAccCloudscaleLoadBalancer_DS_Basic(t *testing.T) {
	var loadBalancer cloudscale.LoadBalancer
	rInt := acctest.RandInt()
	name1 := fmt.Sprintf("terraform-%d-0", rInt)
	name2 := fmt.Sprintf("terraform-%d-1", rInt)

	config := loadBalancerConfig_baseline(2, rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				Config: config + testAccCheckCloudscaleLoadBalancerConfig_name(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists("data.cloudscale_load_balancer.foo", &loadBalancer),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_load_balancer.basic.0", "id", &loadBalancer.UUID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_load_balancer.foo", "id", &loadBalancer.UUID),
					resource.TestCheckResourceAttr(
						"data.cloudscale_load_balancer.foo", "name", name1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_load_balancer.foo", "zone_slug", "rma1"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleLoadBalancerConfig_name_and_zone(name1, "rma1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_load_balancer.foo", "name", name1),
				),
			},
			{
				Config: config + testAccCheckCloudscaleLoadBalancerConfig_name(name2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_load_balancer.foo", "name", name2),
					resource.TestCheckResourceAttr(
						"data.cloudscale_load_balancer.foo", "zone_slug", "rma1"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleLoadBalancerConfig_name_and_zone(name2, "rma1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_load_balancer.foo", "name", name2),
					resource.TestCheckResourceAttr(
						"data.cloudscale_load_balancer.foo", "zone_slug", "rma1"),
				),
			},
			{
				Config:      config + testAccCheckCloudscaleLoadBalancerConfig_name_and_zone(name1, "lpg1"),
				ExpectError: regexp.MustCompile(`Found zero load balancers`),
			},
			{

				Config: config + testAccCheckCloudscaleLoadBalancerConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_load_balancer.basic.0", "name", name1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_load_balancer.foo", "name", name1),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_load_balancer.basic.0", "id", &loadBalancer.UUID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_load_balancer.foo", "id", &loadBalancer.UUID),
				),
			},
			{
				Config:      config + "\n" + `data "cloudscale_load_balancer" "foo" {}`,
				ExpectError: regexp.MustCompile(`Found \d+ load balancers, expected one`),
			},
		},
	})
}

func loadBalancerConfig_baseline(count int, rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer" "basic" {
  count       = %d
  name        = "terraform-%d-${count.index}"
  flavor_slug = "lb-standard"
  zone_slug   = "rma1"
}`, count, rInt)
}

func testAccCheckCloudscaleLoadBalancerConfig_name(name string) string {
	return fmt.Sprintf(`
data "cloudscale_load_balancer" "foo" {
  name = "%s"
}
`, name)
}

func testAccCheckCloudscaleLoadBalancerConfig_name_and_zone(name, zone_slug string) string {
	return fmt.Sprintf(`
data "cloudscale_load_balancer" "foo" {
  name      = "%s"
  zone_slug	= "%s"
}
`, name, zone_slug)
}

func testAccCheckCloudscaleLoadBalancerConfig_id() string {
	return fmt.Sprintf(`
data "cloudscale_load_balancer" "foo" {
  id = "${cloudscale_load_balancer.basic.0.id}"
}
`)
}
