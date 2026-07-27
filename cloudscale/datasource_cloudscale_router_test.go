package cloudscale

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudscaleRouter_DS_Basic(t *testing.T) {
	var router cloudscale.Router
	rInt := acctest.RandInt()
	name1 := fmt.Sprintf("terraform-%d-0", rInt)
	config := routerConfig_baseline(2, rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				Config: config + testAccCheckCloudscaleRouterConfig_name(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleRouterExists("data.cloudscale_router.foo", &router),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_router.basic.0", "id", &router.UUID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_router.foo", "id", &router.UUID),
					resource.TestCheckResourceAttr(
						"data.cloudscale_router.foo", "name", name1),
				),
			},
			{
				Config: config + testAccCheckCloudscaleRouterConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr(
						"cloudscale_router.basic.0", "id", &router.UUID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_router.foo", "id", &router.UUID),
				),
			},
		},
	})
}

func TestAccCloudscaleRouter_DS_NotExisting(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckCloudscaleRouterConfig_name("terraform-unknown-router"),
				ExpectError: regexp.MustCompile(`.*Found zero routers.*`),
			},
		},
	})
}

func testAccCheckCloudscaleRouterConfig_name(name string) string {
	return fmt.Sprintf(`
data "cloudscale_router" "foo" {
  name = "%s"
}
`, name)
}

func testAccCheckCloudscaleRouterConfig_id() string {
	return `
data "cloudscale_router" "foo" {
  id = "${cloudscale_router.basic.0.id}"
}
`
}
