package cloudscale

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudscaleInterface_Basic(t *testing.T) {
	var iface cloudscale.RouterInterface

	rInt := acctest.RandInt()
	resourceName := "cloudscale_interface.basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: interfaceConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleInterfaceExists(resourceName, &iface),
					resource.TestCheckResourceAttrSet(
						resourceName, "router_uuid"),
					resource.TestCheckResourceAttrSet(
						resourceName, "network_uuid"),
					resource.TestCheckResourceAttrSet(
						resourceName, "mac_address"),
					resource.TestCheckResourceAttr(
						resourceName, "addresses.#", "1"),
					resource.TestCheckResourceAttrSet(
						resourceName, "addresses.0.address"),
				),
			},
			{
				// Re-plan with the same config to prove the addresses list is
				// stable: because addresses is Computed+ForceNew, an API that
				// returned more addresses than configured would show a perpetual
				// (replacing) diff here.
				Config:             interfaceConfig_basic(rInt),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccCloudscaleInterface_import_basic(t *testing.T) {
	var router cloudscale.Router
	var iface cloudscale.RouterInterface

	rInt := acctest.RandInt()
	resourceName := "cloudscale_interface.basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: interfaceConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleRouterExists("cloudscale_router.basic", &router),
					testAccCheckCloudscaleInterfaceExists(resourceName, &iface),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return fmt.Sprintf("%s.%s", router.UUID, iface.UUID), nil
				},
			},
		},
	})
}

func testAccCheckCloudscaleInterfaceExists(n string, iface *cloudscale.RouterInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no interface ID is set")
		}

		client := testAccProvider.Meta().(*cloudscale.Client)

		routerID := rs.Primary.Attributes["router_uuid"]
		router, err := client.Routers.Get(context.Background(), routerID)
		if err != nil {
			return err
		}

		for i := range router.Interfaces {
			if router.Interfaces[i].UUID == rs.Primary.ID {
				*iface = router.Interfaces[i]
				return nil
			}
		}

		return fmt.Errorf("interface %s not found on router %s", rs.Primary.ID, routerID)
	}
}

func testAccCheckCloudscaleInterfaceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudscale.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudscale_interface" {
			continue
		}

		routerID := rs.Primary.Attributes["router_uuid"]
		router, err := client.Routers.Get(context.Background(), routerID)
		if err != nil {
			// Parent router is gone -> interface is gone too.
			if cerr, ok := errors.AsType[*cloudscale.ErrorResponse](err); ok && cerr.StatusCode == http.StatusNotFound {
				continue
			}
			return fmt.Errorf("error retrieving router %s: %s", routerID, err)
		}

		for i := range router.Interfaces {
			if router.Interfaces[i].UUID == rs.Primary.ID {
				return fmt.Errorf("interface %s still exists on router %s", rs.Primary.ID, routerID)
			}
		}
	}

	return nil
}

func interfaceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_network" "basic" {
  name                    = "terraform-%d"
  zone_slug               = "rma1"
  auto_create_ipv4_subnet = false
}

resource "cloudscale_subnet" "basic" {
  cidr         = "10.11.12.0/24"
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
    address     = "10.11.12.10"
  }
}`, rInt, rInt)
}
