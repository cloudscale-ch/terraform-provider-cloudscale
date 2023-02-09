package cloudscale

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"regexp"
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

	resourceName := "cloudscale_load_balancer_pool.lb-pool-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists("cloudscale_load_balancer.lb-acc-test", &loadBalancer),
					testAccCheckCloudscaleLoadBalancerPoolExists(resourceName, &loadBalancerPool),
					resource.TestCheckResourceAttrPtr(
						resourceName, "href", &loadBalancerPool.HREF),
					resource.TestCheckResourceAttrPtr(
						resourceName, "id", &loadBalancerPool.UUID),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbPoolName),
					resource.TestCheckResourceAttr(
						resourceName, "algorithm", "round_robin"),
					resource.TestCheckResourceAttr(
						resourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttrPtr(
						resourceName, "load_balancer_uuid", &loadBalancer.UUID),
					resource.TestCheckResourceAttrPtr(
						resourceName, "load_balancer_name", &loadBalancer.Name),
					resource.TestCheckResourceAttrPtr(
						resourceName, "load_balancer_href", &loadBalancer.HREF),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerPool_UpdateName(t *testing.T) {
	var afterCreate, afterUpdate cloudscale.LoadBalancerPool

	rInt1 := acctest.RandInt()
	lbPoolName := fmt.Sprintf("terraform-%d-lb-pool", rInt1)
	rInt2 := acctest.RandInt()
	updatedLBPoolName := fmt.Sprintf("terraform-%d-lb-pool", rInt2)

	resourceName := "cloudscale_load_balancer_pool.lb-pool-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolExists(resourceName, &afterCreate),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbPoolName),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "name", updatedLBPoolName),
					testAccCheckLoadBalancerPoolIsSame(t, &afterCreate, &afterUpdate, true),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerPool_UpdateLB(t *testing.T) {
	var lb1, lb2 cloudscale.LoadBalancer
	var afterCreate, afterUpdate cloudscale.LoadBalancerPool

	rInt := acctest.RandInt()

	resourceName := "cloudscale_load_balancer_pool.lb-pool-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerPoolConfig_multiple(rInt, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists("cloudscale_load_balancer.lb-acc-test.0", &lb1),
					testAccCheckCloudscaleLoadBalancerExists("cloudscale_load_balancer.lb-acc-test.1", &lb2),
					testAccCheckCloudscaleLoadBalancerPoolExists(resourceName, &afterCreate),
					resource.TestCheckResourceAttrPtr(
						resourceName, "load_balancer_uuid", &lb1.UUID),
					resource.TestCheckResourceAttrPair(
						resourceName, "load_balancer_uuid",
						"cloudscale_load_balancer.lb-acc-test.0", "id"),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerPoolConfig_multiple(rInt, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolExists(resourceName, &afterUpdate),
					testAccCheckLoadBalancerPoolIsSame(t, &afterCreate, &afterUpdate, false),
					resource.TestCheckResourceAttrPtr(
						resourceName, "load_balancer_uuid", &lb2.UUID),
					resource.TestCheckResourceAttrPair(
						resourceName, "load_balancer_uuid",
						"cloudscale_load_balancer.lb-acc-test.1", "id"),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerPool_import_basic(t *testing.T) {
	var afterImport, afterUpdate cloudscale.LoadBalancerPool

	rInt1 := acctest.RandInt()
	lbPoolName := fmt.Sprintf("terraform-%d-lb-pool", rInt1)
	rInt2 := acctest.RandInt()
	updatedLBPoolName := fmt.Sprintf("terraform-%d-lb-pool", rInt2)

	resourceName := "cloudscale_load_balancer_pool.lb-pool-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolExists(resourceName, &afterImport),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbPoolName),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "name", updatedLBPoolName),
					testAccCheckLoadBalancerPoolIsSame(t, &afterImport, &afterUpdate, true),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateId:     "does-not-exist",
				ExpectError:       regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerPool_import_withTags(t *testing.T) {
	var beforeImport, afterUpdate cloudscale.LoadBalancerPool

	rInt := acctest.RandInt()

	resourceName := "cloudscale_load_balancer_pool.lb-pool-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerPoolConfigWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolExists(resourceName, &beforeImport),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("terraform-%d-lb-pool", rInt)),
					testTagsMatch(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateId:     "does-not-exist",
				ExpectError:       regexp.MustCompile(`Cannot import non-existent remote object`),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(42),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "name", "terraform-42-lb-pool"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "0"),
					testAccCheckLoadBalancerPoolIsSame(t, &beforeImport, &afterUpdate, true),
					testTagsMatch(resourceName),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerPool_tags(t *testing.T) {
	rInt1, rInt2 := acctest.RandInt(), acctest.RandInt()

	resourceName := "cloudscale_load_balancer_pool.lb-pool-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfigWithTags(rInt2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.my-foo", "foo"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.my-bar", "bar"),
					testTagsMatch(resourceName),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "0"),
					testTagsMatch(resourceName),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfigWithTags(rInt2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.my-foo", "foo"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.my-bar", "bar"),
					testTagsMatch(resourceName),
				),
			},
		},
	})
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

func testAccCloudscaleLoadBalancerPoolConfig_multiple(rInt int, lbIndex int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer" "lb-acc-test" {
  count       = 2
  name        = "terraform-%[1]d-lb-${count.index}"
  flavor_slug = "lb-small"
  zone_slug   = "rma1"
}

resource "cloudscale_load_balancer_pool" "lb-pool-acc-test" {
  name               = "terraform-%[1]d-lb-pool"
  load_balancer_uuid = cloudscale_load_balancer.lb-acc-test[%[2]d].id
  algorithm          = "round_robin"
  protocol           = "tcp"
}
`, rInt, lbIndex)
}

func testAccCloudscaleLoadBalancerPoolConfigWithTags(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer_pool" "lb-pool-acc-test" {
  name = "terraform-%d-lb-pool"
  algorithm = "round_robin"
  protocol = "tcp"
  load_balancer_uuid = cloudscale_load_balancer.lb-acc-test.id
  tags = {
    my-foo = "foo"
    my-bar = "bar"
  }
}
`, rInt)
}

func testAccCheckLoadBalancerPoolIsSame(t *testing.T,
	before, after *cloudscale.LoadBalancerPool,
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
