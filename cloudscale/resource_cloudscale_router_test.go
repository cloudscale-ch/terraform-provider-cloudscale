package cloudscale

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("cloudscale_router", &resource.Sweeper{
		Name: "cloudscale_router",
		F:    testSweepRouters,
	})
}

func testSweepRouters(region string) error {
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client := meta.(*cloudscale.Client)

	routers, err := client.Routers.List(context.Background())
	if err != nil {
		return err
	}

	var errs []error
	for _, s := range routers {
		if strings.HasPrefix(s.Name, "terraform-") {
			tflog.Info(context.Background(), "Destroying Router", map[string]any{"name": s.Name})

			if err := client.Routers.Delete(context.Background(), s.UUID); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func TestAccCloudscaleRouter_Basic(t *testing.T) {
	var router cloudscale.Router

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: routerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleRouterExists("cloudscale_router.basic", &router),
					resource.TestCheckResourceAttr(
						"cloudscale_router.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"cloudscale_router.basic", "internet_gateway", "true"),
					resource.TestCheckResourceAttrSet(
						"cloudscale_router.basic", "href"),
					resource.TestCheckResourceAttrSet(
						"cloudscale_router.basic", "status"),
					resource.TestCheckResourceAttrSet(
						"cloudscale_router.basic", "zone_slug"),
				),
			},
		},
	})
}

func TestAccCloudscaleRouter_import_basic(t *testing.T) {
	var router cloudscale.Router

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: routerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleRouterExists("cloudscale_router.basic", &router),
					resource.TestCheckResourceAttr(
						"cloudscale_router.basic", "name", fmt.Sprintf("terraform-%d", rInt)),
				),
			},
			{
				ResourceName:      "cloudscale_router.basic",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckCloudscaleRouterDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudscale.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudscale_router" {
			continue
		}

		id := rs.Primary.ID

		// Try to find the router
		v, err := client.Routers.Get(context.Background(), id)
		if err == nil {
			return fmt.Errorf("router %v still exists", v)
		}

		if cerr, ok := errors.AsType[*cloudscale.ErrorResponse](err); !ok || cerr.StatusCode != http.StatusNotFound {
			return fmt.Errorf(
				"error waiting for router (%s) to be destroyed: %s",
				rs.Primary.ID, err)
		}
	}

	return nil
}

func routerConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_router" "basic" {
  name             = "terraform-%d"
  zone_slug        = "rma1"
  internet_gateway = true
}`, rInt)
}

func routerConfig_baseline(count int, rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_router" "basic" {
  count            = "%v"
  name             = "terraform-%d-${count.index}"
  zone_slug        = "rma1"
  internet_gateway = true
}`, count, rInt)
}
