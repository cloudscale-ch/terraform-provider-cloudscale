package cloudscale

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v10"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudscaleRouterExtraRoute_Basic(t *testing.T) {
	var router cloudscale.Router

	rInt := acctest.RandInt()
	resourceName := "cloudscale_router_extra_route.basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleRouterExtraRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: routerExtraRouteConfig_basic(rInt, "10.10.10.0/24", "192.168.99.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleRouterExists("cloudscale_router.basic", &router),
					resource.TestCheckResourceAttr(
						resourceName, "destination", "10.10.10.0/24"),
					resource.TestCheckResourceAttr(
						resourceName, "next_hop", "192.168.99.1"),
					resource.TestCheckResourceAttrSet(
						resourceName, "href"),
					resource.TestCheckResourceAttrPair(
						resourceName, "router_uuid", "cloudscale_router.basic", "id"),
				),
			},
		},
	})
}

func TestAccCloudscaleRouterExtraRoute_import_basic(t *testing.T) {
	var router cloudscale.Router
	var extraRoute *cloudscale.RouterExtraRoute

	rInt := acctest.RandInt()
	resourceName := "cloudscale_router_extra_route.basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleRouterExtraRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: routerExtraRouteConfig_basic(rInt, "10.10.10.0/24", "192.168.99.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleRouterExists("cloudscale_router.basic", &router),
					testAccCheckCloudscaleRouterExtraRouteExists(resourceName, &extraRoute),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return fmt.Sprintf("%s.%s", router.UUID, extraRoute.UUID), nil
				},
			},
		},
	})
}

func testAccCheckCloudscaleRouterExtraRouteExists(n string, extraRoute **cloudscale.RouterExtraRoute) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no router extra route ID is set")
		}

		client := testAccProvider.Meta().(*cloudscale.Client)

		routerID := rs.Primary.Attributes["router_uuid"]
		router, err := client.Routers.Get(context.Background(), routerID)
		if err != nil {
			return err
		}

		for i := range router.ExtraRoutes {
			if router.ExtraRoutes[i].UUID == rs.Primary.ID {
				*extraRoute = &router.ExtraRoutes[i]
				return nil
			}
		}

		return fmt.Errorf("extra route %s not found on router %s", rs.Primary.ID, routerID)
	}
}

func testAccCheckCloudscaleRouterExtraRouteDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudscale.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudscale_router_extra_route" {
			continue
		}

		routerID := rs.Primary.Attributes["router_uuid"]
		router, err := client.Routers.Get(context.Background(), routerID)
		if err != nil {
			// Parent router is gone -> extra route is gone too.
			if cerr, ok := errors.AsType[*cloudscale.ErrorResponse](err); ok && cerr.StatusCode == http.StatusNotFound {
				continue
			}
			return fmt.Errorf("error retrieving router %s: %s", routerID, err)
		}

		for i := range router.ExtraRoutes {
			if router.ExtraRoutes[i].UUID == rs.Primary.ID {
				return fmt.Errorf("extra route %s still exists on router %s", rs.Primary.ID, routerID)
			}
		}
	}

	return nil
}

func routerExtraRouteConfig_basic(rInt int, destination, nextHop string) string {
	return fmt.Sprintf(`
resource "cloudscale_network" "basic" {
  name                    = "terraform-%d"
  zone_slug               = "rma1"
  auto_create_ipv4_subnet = false
}

resource "cloudscale_subnet" "basic" {
  cidr         = "192.168.99.0/24"
  network_uuid = cloudscale_network.basic.id
}

resource "cloudscale_router" "basic" {
  name             = "terraform-%d"
  zone_slug        = "rma1"
  internet_gateway = true
}

resource "cloudscale_interface" "basic" {
  router_uuid  = cloudscale_router.basic.id
  network_uuid = cloudscale_network.basic.id

  addresses {
    subnet_uuid = cloudscale_subnet.basic.id
    address     = "192.168.99.10"
  }
}

resource "cloudscale_router_extra_route" "basic" {
  depends_on = [cloudscale_interface.basic]

  router_uuid = cloudscale_router.basic.id
  destination = "%s"
  next_hop    = "%s"
}
`, rInt, rInt, destination, nextHop)
}
