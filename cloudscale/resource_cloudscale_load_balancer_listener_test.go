package cloudscale

import (
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
					resource.TestCheckResourceAttrPtr(
						resourceName, "href", &loadBalancerListener.HREF),
					resource.TestCheckResourceAttrPtr(
						resourceName, "id", &loadBalancerListener.UUID),
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
					resource.TestCheckResourceAttr(
						resourceName, "timeout_client_data_ms", "50000"),
					resource.TestCheckResourceAttr(
						resourceName, "timeout_member_connect_ms", "5000"),
					resource.TestCheckResourceAttr(
						resourceName, "timeout_member_data_ms", "50000"),
					resource.TestCheckResourceAttr(
						resourceName, "timeout_tcp_inspect_ms", "0"),
					resource.TestCheckResourceAttr(
						resourceName, "allowed_cidrs.#", "0"),
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

func TestAccCloudscaleLoadBalancerListener_UpdateAllowedCidrs(t *testing.T) {
	var afterCreate, afterUpdate cloudscale.LoadBalancerListener

	rInt1 := acctest.RandInt()
	rInt2 := acctest.RandInt()

	resourceName := "cloudscale_load_balancer_listener.lb-listener-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerListenerConfig_cidrs(rInt1, `["10.0.0.0/8"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerListenerExists(resourceName, &afterCreate),
					resource.TestCheckResourceAttr(
						resourceName, "allowed_cidrs.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "allowed_cidrs.0", "10.0.0.0/8"),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerListenerConfig_cidrs(rInt2, `["172.16.0.0/12", "192.168.0.0/16"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerListenerExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "allowed_cidrs.#", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "allowed_cidrs.0", "172.16.0.0/12"),
					resource.TestCheckResourceAttr(
						resourceName, "allowed_cidrs.1", "192.168.0.0/16"),
					testAccCheckLoadBalancerListenerIsSame(t, &afterCreate, &afterUpdate, true),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerListener_UpdatePool(t *testing.T) {
	var pool1, pool2 cloudscale.LoadBalancerPool
	var afterCreate, afterUpdate cloudscale.LoadBalancerListener

	rInt := acctest.RandInt()

	resourceName := "cloudscale_load_balancer_listener.lb-listener-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerListenerConfig_multiple(rInt, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolExists("cloudscale_load_balancer_pool.lb-pool-acc-test.0", &pool1),
					testAccCheckCloudscaleLoadBalancerPoolExists("cloudscale_load_balancer_pool.lb-pool-acc-test.1", &pool2),
					testAccCheckCloudscaleLoadBalancerListenerExists(resourceName, &afterCreate),
					resource.TestCheckResourceAttrPtr(
						resourceName, "pool_uuid", &pool1.UUID),
					resource.TestCheckResourceAttrPair(
						resourceName, "pool_uuid",
						"cloudscale_load_balancer_pool.lb-pool-acc-test.0", "id"),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerListenerConfig_multiple(rInt, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerListenerExists(resourceName, &afterUpdate),
					testAccCheckLoadBalancerListenerIsSame(t, &afterCreate, &afterUpdate, false),
					resource.TestCheckResourceAttrPtr(
						resourceName, "pool_uuid", &pool2.UUID),
					resource.TestCheckResourceAttrPair(
						resourceName, "pool_uuid",
						"cloudscale_load_balancer_pool.lb-pool-acc-test.1", "id"),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerListener_import_basic(t *testing.T) {
	var pool cloudscale.LoadBalancerPool
	var beforeImport, afterImport cloudscale.LoadBalancerListener

	rInt1 := acctest.RandInt()
	lbListenerName := fmt.Sprintf("terraform-%d-lb-listener", rInt1)
	rInt2 := acctest.RandInt()
	lbListenerNameUpdated := fmt.Sprintf("terraform-%d-lb-listener", rInt2)

	poolResourceName := "cloudscale_load_balancer_pool.lb-pool-acc-test"
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
					testAccCheckCloudscaleLoadBalancerPoolExists(poolResourceName, &pool),
					testAccCheckCloudscaleLoadBalancerListenerExists(resourceName, &beforeImport),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerListenerConfig_basic(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerListenerExists(resourceName, &afterImport),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbListenerName),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerListenerConfig_basic(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerListenerExists(resourceName, &afterImport),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbListenerNameUpdated),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerListener_import_withTags(t *testing.T) {
	var pool cloudscale.LoadBalancerPool
	var beforeImport, afterUpdate cloudscale.LoadBalancerListener

	rInt := acctest.RandInt()
	lbListenerName := fmt.Sprintf("terraform-%d-lb-listener", rInt)

	poolResourceName := "cloudscale_load_balancer_pool.lb-pool-acc-test"
	resourceName := "cloudscale_load_balancer_listener.lb-listener-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerListenerConfigWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolExists(poolResourceName, &pool),
					testAccCheckCloudscaleLoadBalancerListenerExists(resourceName, &beforeImport),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbListenerName),
					testTagsMatch(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerListenerConfig_basic(42),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerListenerExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("terraform-%d-lb-listener", 42)),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "0"),
					testAccCheckLoadBalancerListenerIsSame(t, &beforeImport, &afterUpdate, true),
					testTagsMatch(resourceName),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerListener_tags(t *testing.T) {
	rInt1, rInt2, rInt3 := acctest.RandInt(), acctest.RandInt(), acctest.RandInt()

	resourceName := "cloudscale_load_balancer_listener.lb-listener-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt2) +
					testAccCloudscaleLoadBalancerListenerConfigWithTags(rInt3),
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
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt2) +
					testAccCloudscaleLoadBalancerListenerConfig_basic(rInt3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "0"),
					testTagsMatch(resourceName),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt2) +
					testAccCloudscaleLoadBalancerListenerConfigWithTags(rInt3),
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

func testAccCloudscaleLoadBalancerListenerConfig_multiple(rInt int, poolIndex int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer" "lb-acc-test" {
  name        = "terraform-%[1]d-lb"
  flavor_slug = "lb-flex-4-2"
  zone_slug   = "rma1"
}

resource "cloudscale_load_balancer_pool" "lb-pool-acc-test" {
  count              = 2
  name               = "terraform-%[1]d-lb-pool-${count.index}"
  load_balancer_uuid = cloudscale_load_balancer.lb-acc-test.id
  algorithm          = "round_robin"
  protocol           = "tcp"
}

resource "cloudscale_load_balancer_listener" "lb-listener-acc-test" {
  name          = "terraform-%[1]d-listener"
  pool_uuid     = cloudscale_load_balancer_pool.lb-pool-acc-test[%[2]d].id
  protocol      = "tcp"
  protocol_port = 80
}
`, rInt, poolIndex)

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

func testAccCloudscaleLoadBalancerListenerConfig_cidrs(rInt int, cidrs string) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer_listener" "lb-listener-acc-test" {
  name = "terraform-%d-lb-listener"
  pool_uuid = cloudscale_load_balancer_pool.lb-pool-acc-test.id
  protocol = "tcp"
  protocol_port = 80
  
  allowed_cidrs = %s
}
`, rInt, cidrs)
}

func testAccCloudscaleLoadBalancerListenerConfigWithTags(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer_listener" "lb-listener-acc-test" {
  name = "terraform-%d-lb-listener"
  pool_uuid = cloudscale_load_balancer_pool.lb-pool-acc-test.id
  protocol = "tcp"
  protocol_port = 80
  tags = {
    my-foo = "foo"
    my-bar = "bar"
  }
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
