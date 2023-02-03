package cloudscale

import (
	"context"
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
	"time"
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
					testAccCloudscaleLoadBalancerHealthMonitorConfig_basic(rInt, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists("cloudscale_load_balancer.lb-acc-test", &loadBalancer),
					testAccCheckCloudscaleLoadBalancerPoolExists("cloudscale_load_balancer_pool.lb-pool-acc-test", &loadBalancerPool),
					testAccCheckCloudscaleLoadBalancerHealthMonitorExists(resourceName, &loadBalancerHealthMonitor),
					resource.TestCheckResourceAttr(
						resourceName, "delay_s", "10"),
					resource.TestCheckResourceAttr(
						resourceName, "up_threshold", "3"),
					resource.TestCheckResourceAttr(
						resourceName, "down_threshold", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "timeout_s", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "type", "tcp"),

					// ensure http attrs are not set
					resource.TestCheckResourceAttr(
						resourceName, "http_expected_codes.#", "0"),
					resource.TestCheckNoResourceAttr(
						resourceName, "http_method"),
					resource.TestCheckNoResourceAttr(
						resourceName, "http_url_path"),
					resource.TestCheckNoResourceAttr(
						resourceName, "http_version"),
					resource.TestCheckNoResourceAttr(
						resourceName, "http_host"),

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

func TestAccCloudscaleLoadBalancerHealthMonitor_UpdateDelay(t *testing.T) {
	var afterCreate, afterUpdate cloudscale.LoadBalancerHealthMonitor

	rInt1 := acctest.RandInt()

	resourceName := "cloudscale_load_balancer_health_monitor.lb-health_monitor-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerHealthMonitorConfig_basic(rInt1, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerHealthMonitorExists(resourceName, &afterCreate),
					resource.TestCheckResourceAttr(
						resourceName, "delay_s", fmt.Sprintf("10")),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerHealthMonitorConfig_basic(rInt1, 15),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerHealthMonitorExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "delay_s", fmt.Sprintf("15")),
					testAccCheckLoadBalancerHealthMonitorIsSame(t, &afterCreate, &afterUpdate, true),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerHealthMonitor_UpdateHTTP(t *testing.T) {
	var afterCreate, afterUpdate cloudscale.LoadBalancerHealthMonitor

	rInt1 := acctest.RandInt()

	resourceName := "cloudscale_load_balancer_health_monitor.lb-health_monitor-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerHealthMonitorConfig_http(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerHealthMonitorExists(resourceName, &afterCreate),
					resource.TestCheckResourceAttr(
						resourceName, "type", "http"),
					resource.TestCheckResourceAttr(
						resourceName, "http_expected_codes.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "http_expected_codes.0", "200"),
					resource.TestCheckResourceAttr(
						resourceName, "http_method", "GET"),
					resource.TestCheckResourceAttr(
						resourceName, "http_url_path", "/"),
					resource.TestCheckResourceAttr(
						resourceName, "http_version", "1.1"),
					resource.TestCheckResourceAttr(
						resourceName, "http_host", "www.cloudscale.ch"),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerHealthMonitorConfig_http_modified(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerHealthMonitorExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "type", "http"),
					resource.TestCheckResourceAttr(
						resourceName, "http_expected_codes.#", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "http_expected_codes.0", "418"),
					resource.TestCheckResourceAttr(
						resourceName, "http_expected_codes.1", "425"),
					resource.TestCheckResourceAttr(
						resourceName, "http_method", "PATCH"),
					resource.TestCheckResourceAttr(
						resourceName, "http_url_path", "/fail"),
					resource.TestCheckResourceAttr(
						resourceName, "http_version", "1.1"),
					resource.TestCheckResourceAttr(
						resourceName, "http_host", "www.cloudscale-status.net"),
					testAccCheckLoadBalancerHealthMonitorIsSame(t, &afterCreate, &afterUpdate, true),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerHealthMonitor_UpdatePool(t *testing.T) {
	var pool1, pool2 cloudscale.LoadBalancerPool
	var afterCreate, afterUpdate cloudscale.LoadBalancerHealthMonitor

	resourceName := "cloudscale_load_balancer_health_monitor.lb-health_monitor-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerHealthMonitorConfig_multiple(15, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolExists("cloudscale_load_balancer_pool.lb-pool-acc-test.0", &pool1),
					testAccCheckCloudscaleLoadBalancerPoolExists("cloudscale_load_balancer_pool.lb-pool-acc-test.1", &pool2),
					testAccCheckCloudscaleLoadBalancerHealthMonitorExists(resourceName, &afterCreate),
					resource.TestCheckResourceAttrPtr(
						resourceName, "pool_uuid", &pool1.UUID),
					resource.TestCheckResourceAttrPair(
						resourceName, "pool_uuid",
						"cloudscale_load_balancer_pool.lb-pool-acc-test.0", "id"),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerHealthMonitorConfig_multiple(15, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerHealthMonitorExists(resourceName, &afterUpdate),
					testAccCheckLoadBalancerHealthMonitorIsSame(t, &afterCreate, &afterUpdate, false),
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

func TestAccCloudscaleLoadBalancerHealthMonitor_import_basic(t *testing.T) {
	var pool cloudscale.LoadBalancerPool
	var beforeImport, afterImport, afterUpdate cloudscale.LoadBalancerHealthMonitor

	rInt1 := acctest.RandInt()

	poolResourceName := "cloudscale_load_balancer_pool.lb-pool-acc-test"
	resourceName := "cloudscale_load_balancer_health_monitor.lb-health_monitor-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerHealthMonitorConfig_basic(rInt1, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolExists(poolResourceName, &pool),
					testAccCheckCloudscaleLoadBalancerHealthMonitorExists(resourceName, &beforeImport),
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
					testAccCloudscaleLoadBalancerHealthMonitorConfig_basic(rInt1, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerHealthMonitorExists(resourceName, &afterImport),
					resource.TestCheckResourceAttr(
						resourceName, "delay_s", "10"),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerHealthMonitorConfig_basic(rInt1, 15),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerHealthMonitorExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "delay_s", "15"),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerHealthMonitor_import_withTags(t *testing.T) {
	var pool cloudscale.LoadBalancerPool
	var beforeImport, afterUpdate cloudscale.LoadBalancerHealthMonitor

	rInt := acctest.RandInt()

	poolResourceName := "cloudscale_load_balancer_pool.lb-pool-acc-test"
	resourceName := "cloudscale_load_balancer_health_monitor.lb-health_monitor-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerHealthMonitorConfigWithTags(10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolExists(poolResourceName, &pool),
					testAccCheckCloudscaleLoadBalancerHealthMonitorExists(resourceName, &beforeImport),
					resource.TestCheckResourceAttr(
						resourceName, "delay_s", fmt.Sprintf("%v", 1010)),
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
					testAccCloudscaleLoadBalancerHealthMonitorConfig_basic(rInt, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerHealthMonitorExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "delay_s", "10"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "0"),
					testAccCheckLoadBalancerHealthMonitorIsSame(t, &beforeImport, &afterUpdate, true),
					testTagsMatch(resourceName),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerHealthMonitor_tags(t *testing.T) {
	rInt1, rInt2, rInt3 := acctest.RandInt(), acctest.RandInt(), acctest.RandInt()

	resourceName := "cloudscale_load_balancer_health_monitor.lb-health_monitor-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt2) +
					testAccCloudscaleLoadBalancerHealthMonitorConfigWithTags(rInt3),
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
					testAccCloudscaleLoadBalancerHealthMonitorConfig_basic(rInt3, 10),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "0"),
					testTagsMatch(resourceName),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt2) +
					testAccCloudscaleLoadBalancerHealthMonitorConfigWithTags(rInt3),
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

func TestAccCloudscaleLoadBalancerHealthMonitor_MemberStatus(t *testing.T) {
	var loadBalancerPoolMember cloudscale.LoadBalancerPoolMember

	rInt := acctest.RandInt()

	memberResourceName := "cloudscale_load_balancer_pool_member.lb-pool-member-acc-test"

	basicConfig := testAccCloudscaleLoadBalancerConfig_basic(rInt) +
		testAccCloudscaleLoadBalancerPoolConfig_basic(rInt) +
		testAccCloudscaleLoadBalancerPoolMemberConfig_basic(rInt, true) +
		testAccCloudscaleLoadBalancerListenerConfig_basic(rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: basicConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolMemberExists(memberResourceName, &loadBalancerPoolMember),
					waitForMonitorStatus(&loadBalancerPoolMember, "no_monitor"),
					resource.TestCheckResourceAttr(memberResourceName,
						"monitor_status", "no_monitor"),
				),
			},
			{
				Config: basicConfig +
					testAccCloudscaleLoadBalancerHealthMonitorConfig_basic(rInt, 10),
				Check: resource.ComposeTestCheckFunc(
					waitForMonitorStatus(&loadBalancerPoolMember, "up"),
				),
			},
			{
				Config: basicConfig +
					testAccCloudscaleLoadBalancerHealthMonitorConfig_basic(rInt, 10),
				Check: resource.ComposeTestCheckFunc(
					// this check is in a separate step to ensure the status is refreshed form the API:
					resource.TestCheckResourceAttr(memberResourceName,
						"monitor_status", "up"),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerHealthMonitorHTTP_MemberStatus(t *testing.T) {
	var loadBalancerPoolMember cloudscale.LoadBalancerPoolMember

	rInt := acctest.RandInt()

	memberResourceName := "cloudscale_load_balancer_pool_member.lb-pool-member-acc-test"

	basicConfig := testAccCloudscaleLoadBalancerConfig_basic(rInt) +
		testAccCloudscaleLoadBalancerPoolConfig_basic(rInt) +
		testAccCloudscaleLoadBalancerPoolMemberConfig_basic(rInt, true) +
		testAccCloudscaleLoadBalancerListenerConfig_basic(rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: basicConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolMemberExists(memberResourceName, &loadBalancerPoolMember),
					waitForMonitorStatus(&loadBalancerPoolMember, "no_monitor"),
					resource.TestCheckResourceAttr(memberResourceName,
						"monitor_status", "no_monitor"),
				),
			},
			{
				Config: basicConfig +
					testAccCloudscaleLoadBalancerHealthMonitorConfig_http(10),
				Check: resource.ComposeTestCheckFunc(
					waitForMonitorStatus(&loadBalancerPoolMember, "up"),
				),
			},
			{
				Config: basicConfig +
					testAccCloudscaleLoadBalancerHealthMonitorConfig_http(10),
				Check: resource.ComposeTestCheckFunc(
					// this check is in a separate step to ensure the status is refreshed form the API:
					resource.TestCheckResourceAttr(memberResourceName,
						"monitor_status", "up"),
				),
			},
			{
				Config: basicConfig +
					testAccCloudscaleLoadBalancerHealthMonitorConfig_http_modified(10),
				Check: resource.ComposeTestCheckFunc(
					waitForMonitorStatus(&loadBalancerPoolMember, "error"),
				),
			},
		},
	})
}

func testAccCheckLoadBalancerHealthMonitorIsSame(t *testing.T,
	before *cloudscale.LoadBalancerHealthMonitor, after *cloudscale.LoadBalancerHealthMonitor,
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

func testAccCloudscaleLoadBalancerHealthMonitorConfig_multiple(rInt int, poolIndex int) string {
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

resource "cloudscale_load_balancer_health_monitor" "lb-health_monitor-acc-test" {
  pool_uuid        = cloudscale_load_balancer_pool.lb-pool-acc-test[%[2]d].id
  delay            = %[1]d
  type             = "tcp"
}
`, rInt, poolIndex)

}

func testAccCloudscaleLoadBalancerHealthMonitorConfigWithTags(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer_health_monitor" "lb-health_monitor-acc-test" {
  pool_uuid = cloudscale_load_balancer_pool.lb-pool-acc-test.id
  type             = "tcp"
  tags = {
    my-foo = "foo"
    my-bar = "bar"
  }
}
`)
}

func testAccCloudscaleLoadBalancerHealthMonitorConfig_basic(rInt int, delay int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer_health_monitor" "lb-health_monitor-acc-test" {
  pool_uuid        = cloudscale_load_balancer_pool.lb-pool-acc-test.id
  delay            = %v
  type             = "tcp"
}
`, delay)
}

func testAccCloudscaleLoadBalancerHealthMonitorConfig_http(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer_health_monitor" "lb-health_monitor-acc-test" {
  pool_uuid        = cloudscale_load_balancer_pool.lb-pool-acc-test.id
  type             = "http"
  http_url_path    = "/"
  http_version     = "1.1"
  http_host        = "www.cloudscale.ch"
}
`)
}

func testAccCloudscaleLoadBalancerHealthMonitorConfig_http_modified(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_load_balancer_health_monitor" "lb-health_monitor-acc-test" {
  pool_uuid           = cloudscale_load_balancer_pool.lb-pool-acc-test.id
  delay               = 10
  type                = "http"
  http_expected_codes = ["418", "425"]
  http_method         = "PATCH"
  http_url_path       = "/fail"
  http_version        = "1.1"
  http_host           = "www.cloudscale-status.net"
}
`)
}

func waitForMonitorStatus(member *cloudscale.LoadBalancerPoolMember, status string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*cloudscale.Client)

		var retrievedPoolMember *cloudscale.LoadBalancerPoolMember
		var err error

		for i := 0; i < 30; i++ {
			retrievedPoolMember, err = client.LoadBalancerPoolMembers.Get(
				context.Background(), member.Pool.UUID, member.UUID,
			)
			if err != nil {
				return err
			}
			if retrievedPoolMember.MonitorStatus == status {
				return nil
			}
			time.Sleep(2 * time.Second)
		}
		return fmt.Errorf(
			"expeted MonitorStatus to become '%s', but it's still: '%s'",
			status, retrievedPoolMember.MonitorStatus,
		)
	}
}
