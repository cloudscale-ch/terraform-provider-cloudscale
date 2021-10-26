package cloudscale

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudscaleServerGroup_DS_Basic(t *testing.T) {
	var serverGroup cloudscale.ServerGroup
	rInt := acctest.RandInt()
	name1 := fmt.Sprintf("terraform-%d-0", rInt)
	name2 := fmt.Sprintf("terraform-%d-1", rInt)
	config := serverGroupConfig_baseline(2, rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				Config: config + testAccCheckCloudscaleServerGroupConfig_name(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleServerGroupExists("data.cloudscale_server_group.foo", &serverGroup),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_server_group.basic.0", "id", &serverGroup.UUID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_server_group.foo", "id", &serverGroup.UUID),
					resource.TestCheckResourceAttr(
						"data.cloudscale_server_group.foo", "name", name1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_server_group.foo", "type", "anti-affinity"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_server_group.foo", "zone_slug", "rma1"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleServerGroupConfig_name_and_zone(name1, "rma1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_server_group.foo", "name", name1),
				),
			},
			{
				Config: config + testAccCheckCloudscaleServerGroupConfig_name(name2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_server_group.foo", "name", name2),
					resource.TestCheckResourceAttr(
						"data.cloudscale_server_group.foo", "type", "anti-affinity"),
					resource.TestCheckResourceAttr(
						"data.cloudscale_server_group.foo", "zone_slug", "rma1"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleServerGroupConfig_name_and_zone(name2, "rma1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_server_group.foo", "name", name2),
					resource.TestCheckResourceAttr(
						"data.cloudscale_server_group.foo", "zone_slug", "rma1"),
				),
			},
			{
				Config:      config + testAccCheckCloudscaleServerGroupConfig_name_and_zone(name1, "lpg1"),
				ExpectError: regexp.MustCompile(`Found zero server groups`),
			},
			{

				Config: config + testAccCheckCloudscaleServerGroupConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_server_group.basic.0", "name", name1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_server_group.foo", "name", name1),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_server_group.basic.0", "id", &serverGroup.UUID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_server_group.foo", "id", &serverGroup.UUID),
				),
			},
			{
				Config:      config + "\n" + `data "cloudscale_server_group" "foo" {}`,
				ExpectError: regexp.MustCompile(`Found \d+ server groups, expected one`),
			},
		},
	})
}

func TestAccCloudscaleServerGroup_DS_NotExisting(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckCloudscaleServerGroupConfig_name("terraform-unknown"),
				ExpectError: regexp.MustCompile(`Found zero server groups`),
			},
		},
	})
}

func testAccCheckCloudscaleServerGroupExists(n string, server_group *cloudscale.ServerGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Server group ID is set")
		}

		client := testAccProvider.Meta().(*cloudscale.Client)

		id := rs.Primary.ID

		// Try to find the server group
		retrieveServerGroup, err := client.ServerGroups.Get(context.Background(), id)

		if err != nil {
			return err
		}

		if retrieveServerGroup.UUID != rs.Primary.ID {
			return fmt.Errorf("Server group not found")
		}

		*server_group = *retrieveServerGroup

		return nil
	}
}

func serverGroupConfig_baseline(count int, rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server_group" "basic" {
  count        = "%v"
  name         = "terraform-%d-${count.index}"
  type         = "anti-affinity"
  zone_slug    = "rma1"
}`, count, rInt)
}

func testAccCheckCloudscaleServerGroupConfig_name(name string) string {
	return fmt.Sprintf(`
data "cloudscale_server_group" "foo" {
  name               = "%s"
}
`, name)
}

func testAccCheckCloudscaleServerGroupConfig_name_and_zone(name, zone_slug string) string {
	return fmt.Sprintf(`
data "cloudscale_server_group" "foo" {
  name               = "%s"
  zone_slug			 = "%s"
}
`, name, zone_slug)
}

func testAccCheckCloudscaleServerGroupConfig_id() string {
	return fmt.Sprintf(`
data "cloudscale_server_group" "foo" {
  id               = "${cloudscale_server_group.basic.0.id}"
}
`)
}
