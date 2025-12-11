package cloudscale

import (
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"regexp"
	"testing"
)

var TestAddress = "10.100.10.100"

func TestAccCloudscaleLoadBalancerPoolMember_Basic(t *testing.T) {
	var loadBalancer cloudscale.LoadBalancer
	var loadBalancerPool cloudscale.LoadBalancerPool
	var loadBalancerPoolMember cloudscale.LoadBalancerPoolMember

	rInt := acctest.RandInt()
	lbPoolName := fmt.Sprintf("terraform-%d-lb-pool-member", rInt)

	resourceName := "cloudscale_load_balancer_pool_member.lb-pool-member-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerPoolMemberConfig_basic(rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists("cloudscale_load_balancer.lb-acc-test", &loadBalancer),
					testAccCheckCloudscaleLoadBalancerPoolExists("cloudscale_load_balancer_pool.lb-pool-acc-test", &loadBalancerPool),
					testAccCheckCloudscaleLoadBalancerPoolMemberExists(resourceName, &loadBalancerPoolMember),
					waitForMonitorStatus(&loadBalancerPoolMember, "no_monitor"),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerPoolMemberConfig_basic(rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerExists("cloudscale_load_balancer.lb-acc-test", &loadBalancer),
					testAccCheckCloudscaleLoadBalancerPoolExists("cloudscale_load_balancer_pool.lb-pool-acc-test", &loadBalancerPool),
					testAccCheckCloudscaleLoadBalancerPoolMemberExists(resourceName, &loadBalancerPoolMember),
					resource.TestCheckResourceAttrPtr(
						resourceName, "href", &loadBalancerPoolMember.HREF),
					resource.TestCheckResourceAttrPtr(
						resourceName, "id", &loadBalancerPoolMember.UUID),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbPoolName),
					resource.TestCheckResourceAttr(
						resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "protocol_port", "80"),
					resource.TestCheckResourceAttr(
						resourceName, "monitor_port", "0"),
					resource.TestCheckResourceAttr(
						resourceName, "address", TestAddress),
					resource.TestCheckResourceAttr(
						resourceName, "monitor_status", "no_monitor"),
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

func TestAccCloudscaleLoadBalancerPoolMember_UpdateName(t *testing.T) {
	var afterCreate, afterUpdate cloudscale.LoadBalancerPoolMember

	rInt1 := acctest.RandInt()
	lbPoolMemberName := fmt.Sprintf("terraform-%d-lb-pool-member", rInt1)
	rInt2 := acctest.RandInt()
	lbPoolMemberNameUpdated := fmt.Sprintf("terraform-%d-lb-pool-member", rInt2)

	resourceName := "cloudscale_load_balancer_pool_member.lb-pool-member-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolMemberConfig_basic(rInt1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolMemberExists(resourceName, &afterCreate),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbPoolMemberName),
					resource.TestCheckResourceAttr(
						resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolMemberConfig_basic(rInt2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolMemberExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbPoolMemberNameUpdated),
					resource.TestCheckResourceAttr(
						resourceName, "enabled", "true"),
					testAccCheckLoadBalancerPoolMemberIsSame(t, &afterCreate, &afterUpdate, true),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolMemberConfig_basic(rInt2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolMemberExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbPoolMemberNameUpdated),
					resource.TestCheckResourceAttr(
						resourceName, "enabled", "false"),
					testAccCheckLoadBalancerPoolMemberIsSame(t, &afterCreate, &afterUpdate, true),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerPoolMember_UpdatePool(t *testing.T) {
	var pool1, pool2 cloudscale.LoadBalancerPool
	var afterCreate, afterUpdate cloudscale.LoadBalancerPoolMember

	rInt := acctest.RandInt()

	resourceName := "cloudscale_load_balancer_pool_member.lb-pool-member-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerPoolMemberConfig_multiple(rInt, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolExists("cloudscale_load_balancer_pool.lb-pool-acc-test.0", &pool1),
					testAccCheckCloudscaleLoadBalancerPoolExists("cloudscale_load_balancer_pool.lb-pool-acc-test.1", &pool2),
					testAccCheckCloudscaleLoadBalancerPoolMemberExists(resourceName, &afterCreate),
					resource.TestCheckResourceAttrPtr(
						resourceName, "pool_uuid", &pool1.UUID),
					resource.TestCheckResourceAttrPair(
						resourceName, "pool_uuid",
						"cloudscale_load_balancer_pool.lb-pool-acc-test.0", "id"),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerPoolMemberConfig_multiple(rInt, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolMemberExists(resourceName, &afterUpdate),
					testAccCheckLoadBalancerPoolMemberIsSame(t, &afterCreate, &afterUpdate, false),
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

func TestAccCloudscaleLoadBalancerPoolMember_UpdateEnabled(t *testing.T) {
	var afterCreate, afterUpdate cloudscale.LoadBalancerPoolMember

	rInt1 := acctest.RandInt()

	resourceName := "cloudscale_load_balancer_pool_member.lb-pool-member-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolMemberConfig_basic(rInt1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolMemberExists(resourceName, &afterCreate),
					resource.TestCheckResourceAttr(
						resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolMemberConfig_disabled(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolMemberExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "enabled", "false"),
					testAccCheckLoadBalancerPoolMemberIsSame(t, &afterCreate, &afterUpdate, true),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerPoolMember_import_basic(t *testing.T) {
	var pool cloudscale.LoadBalancerPool
	var beforeImport, afterImport cloudscale.LoadBalancerPoolMember

	rInt1 := acctest.RandInt()
	lbPoolMemberName := fmt.Sprintf("terraform-%d-lb-pool-member", rInt1)
	rInt2 := acctest.RandInt()
	lbPoolMemberNameUpdated := fmt.Sprintf("terraform-%d-lb-pool-member", rInt2)

	poolResourceName := "cloudscale_load_balancer_pool.lb-pool-acc-test"
	resourceName := "cloudscale_load_balancer_pool_member.lb-pool-member-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolMemberConfig_basic(rInt1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolExists(poolResourceName, &pool),
					testAccCheckCloudscaleLoadBalancerPoolMemberExists(resourceName, &beforeImport),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return fmt.Sprintf("%s.%s", pool.UUID, beforeImport.UUID), nil
				},
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolMemberConfig_basic(rInt1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolMemberExists(resourceName, &afterImport),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbPoolMemberName),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolMemberConfig_basic(rInt2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolMemberExists(resourceName, &afterImport),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbPoolMemberNameUpdated),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerPoolMember_import_error_cases(t *testing.T) {
	rInt := acctest.RandInt()

	resourceName := "cloudscale_load_balancer_pool_member.lb-pool-member-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerPoolMemberConfig_basic(rInt, true),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateId:     "does-not-exist",
				ExpectError:       regexp.MustCompile(`invalid import id "does-not-exist". Expecting {pool_uuid}.{member_uuid}`),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateId:     "cb9f381c-50b8-43e7-a192-ef72e43a5cb5.38632c78-8cbd-4f66-b7d8-43d359aaac87",
				ExpectError:       regexp.MustCompile(`Cannot import non-existent remote object`),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateId:     ".",
				ExpectError:       regexp.MustCompile(`invalid import id ".". Could not parse {pool_uuid}.{member_uuid}`),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerPoolMember_import_withTags(t *testing.T) {
	var pool cloudscale.LoadBalancerPool
	var beforeImport, afterUpdate cloudscale.LoadBalancerPoolMember

	rInt := acctest.RandInt()
	lbPoolMemberName := fmt.Sprintf("terraform-%d-lb-pool-member", rInt)

	poolResourceName := "cloudscale_load_balancer_pool.lb-pool-acc-test"
	resourceName := "cloudscale_load_balancer_pool_member.lb-pool-member-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerPoolMemberConfigWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolExists(poolResourceName, &pool),
					testAccCheckCloudscaleLoadBalancerPoolMemberExists(resourceName, &beforeImport),
					resource.TestCheckResourceAttr(
						resourceName, "name", lbPoolMemberName),
					testTagsMatch(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return fmt.Sprintf("%s.%s", pool.UUID, beforeImport.UUID), nil
				},
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt) +
					testAccCloudscaleLoadBalancerPoolMemberConfig_basic(42, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleLoadBalancerPoolMemberExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("terraform-%d-lb-pool-member", 42)),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "0"),
					testAccCheckLoadBalancerPoolMemberIsSame(t, &beforeImport, &afterUpdate, true),
					testTagsMatch(resourceName),
				),
			},
		},
	})
}

func TestAccCloudscaleLoadBalancerPoolMember_tags(t *testing.T) {
	rInt1, rInt2, rInt3 := acctest.RandInt(), acctest.RandInt(), acctest.RandInt()

	resourceName := "cloudscale_load_balancer_pool_member.lb-pool-member-acc-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt2) +
					testAccCloudscaleLoadBalancerPoolMemberConfigWithTags(rInt3),
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
					testAccCloudscaleLoadBalancerPoolMemberConfig_basic(rInt3, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "0"),
					testTagsMatch(resourceName),
				),
			},
			{
				Config: testAccCloudscaleLoadBalancerConfig_basic(rInt1) +
					testAccCloudscaleLoadBalancerPoolConfig_basic(rInt2) +
					testAccCloudscaleLoadBalancerPoolMemberConfigWithTags(rInt3),
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

func testAccCloudscaleLoadBalancerSubnet(rInt int) string {
	return fmt.Sprintf(
		`
resource "cloudscale_network" "lb-net" {
  name                    = "terraform-%[1]d-network"
  zone_slug               = "rma1"
  mtu                     = "9000"
  auto_create_ipv4_subnet = "false"
}

resource "cloudscale_subnet" "lb-subnet" {
  cidr               = "10.100.10.0/24"
  network_uuid       = cloudscale_network.lb-net.id
}`, rInt)
}

func testAccCloudscaleLoadBalancerPoolMemberConfig_basic(rInt int, enabled bool) string {
	return fmt.Sprintf(`
%[4]s

resource "cloudscale_load_balancer_pool_member" "lb-pool-member-acc-test" {
  name          = "terraform-%[1]d-lb-pool-member"
  pool_uuid     = cloudscale_load_balancer_pool.lb-pool-acc-test.id
  protocol_port = 80
  address       = "%[2]s"
  enabled       = %[3]t
  subnet_uuid   = cloudscale_subnet.lb-subnet.id
}
`, rInt, TestAddress, enabled, testAccCloudscaleLoadBalancerSubnet(rInt))
}

func testAccCloudscaleLoadBalancerPoolMemberConfig_disabled(rInt int) string {
	return fmt.Sprintf(`
%s

resource "cloudscale_load_balancer_pool_member" "lb-pool-member-acc-test" {
  name          = "terraform-%d-lb-pool-member"
  pool_uuid     = cloudscale_load_balancer_pool.lb-pool-acc-test.id
  protocol_port = 80
  address       = "%s"
  enabled       = false
  subnet_uuid   = cloudscale_subnet.lb-subnet.id
}
`, testAccCloudscaleLoadBalancerSubnet(rInt), rInt, TestAddress)
}

func testAccCloudscaleLoadBalancerPoolMemberConfigWithTags(rInt int) string {
	return fmt.Sprintf(`
%s

resource "cloudscale_load_balancer_pool_member" "lb-pool-member-acc-test" {
  name          = "terraform-%d-lb-pool-member"
  pool_uuid     = cloudscale_load_balancer_pool.lb-pool-acc-test.id
  protocol_port = 80
  address       = "%s"
  subnet_uuid   = cloudscale_subnet.lb-subnet.id
  tags = {
    my-foo = "foo"
    my-bar = "bar"
  }
}
`, testAccCloudscaleLoadBalancerSubnet(rInt), rInt, TestAddress)
}

func testAccCloudscaleLoadBalancerPoolMemberConfig_multiple(rInt int, poolIndex int) string {
	return fmt.Sprintf(`
%[4]s

resource "cloudscale_load_balancer" "lb-acc-test" {
  name        = "terraform-%[1]d-lb"
  flavor_slug = "lb-standard"
  zone_slug   = "rma1"
}

resource "cloudscale_load_balancer_pool" "lb-pool-acc-test" {
  count              = 2
  name               = "terraform-%[1]d-lb-pool-${count.index}"
  load_balancer_uuid = cloudscale_load_balancer.lb-acc-test.id
  algorithm          = "round_robin"
  protocol           = "tcp"
}

resource "cloudscale_load_balancer_pool_member" "lb-pool-member-acc-test" {
  name          = "terraform-%[1]d-lb-pool-member"
  pool_uuid     = cloudscale_load_balancer_pool.lb-pool-acc-test[%[2]d].id
  address       = "%[3]s"
  subnet_uuid   = cloudscale_subnet.lb-subnet.id
  protocol_port = 80
}
`, rInt, poolIndex, TestAddress, testAccCloudscaleLoadBalancerSubnet(rInt))
}

func testAccCheckLoadBalancerPoolMemberIsSame(t *testing.T,
	before *cloudscale.LoadBalancerPoolMember, after *cloudscale.LoadBalancerPoolMember,
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
