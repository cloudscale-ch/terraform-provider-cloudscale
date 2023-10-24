package cloudscale

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v4"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudscaleObjectsUser_DS_Basic(t *testing.T) {
	var objectsUser cloudscale.ObjectsUser
	rInt := acctest.RandInt()
	name1 := fmt.Sprintf("terraform-%d-0", rInt)
	name2 := fmt.Sprintf("terraform-%d-1", rInt)
	config := objectsUserConfig_baseline(2, rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleObjectsUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				Config: config + testAccCheckCloudscaleObjectsUserConfig_name(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleObjectsUserExists("data.cloudscale_objects_user.foo", &objectsUser),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_objects_user.basic.0", "id", &objectsUser.ID),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_objects_user.basic.0", "user_id", &objectsUser.ID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_objects_user.foo", "id", &objectsUser.ID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_objects_user.foo", "user_id", &objectsUser.ID),
					resource.TestCheckResourceAttr(
						"data.cloudscale_objects_user.foo", "display_name", name1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_objects_user.foo", "keys.#", "1"),
					resource.TestCheckResourceAttrSet(
						"data.cloudscale_objects_user.foo", "href"),
				),
			},
			{
				Config: config + testAccCheckCloudscaleObjectsUserConfig_name(name2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_objects_user.foo", "display_name", name2),
				),
			},
			{

				Config: config + testAccCheckCloudscaleObjectsUserConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_objects_user.basic.0", "display_name", name1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_objects_user.foo", "display_name", name1),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_objects_user.basic.0", "id", &objectsUser.ID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_objects_user.foo", "id", &objectsUser.ID),
				),
			},
			{

				Config: config + testAccCheckCloudscaleObjectsUserConfig_user_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_objects_user.basic.0", "display_name", name1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_objects_user.foo", "display_name", name1),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_objects_user.basic.0", "id", &objectsUser.ID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_objects_user.foo", "id", &objectsUser.ID),
				),
			},
			{
				Config:      config + "\n" + `data "cloudscale_objects_user" "foo" {}`,
				ExpectError: regexp.MustCompile(`Found \d+ Objects Users, expected one`),
			},
		},
	})
}

func TestAccCloudscaleObjectsUser_DS_NotExisting(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckCloudscaleObjectsUserConfig_name("terraform-unknown"),
				ExpectError: regexp.MustCompile(`Found zero Objects Users`),
			},
		},
	})
}

func testAccCheckCloudscaleObjectsUserConfig_name(display_name string) string {
	return fmt.Sprintf(`
data "cloudscale_objects_user" "foo" {
  display_name = "%s"
}
`, display_name)
}

func testAccCheckCloudscaleObjectsUserConfig_id() string {
	return fmt.Sprintf(`
data "cloudscale_objects_user" "foo" {
  id = "${cloudscale_objects_user.basic.0.id}"
}
`)
}

func testAccCheckCloudscaleObjectsUserConfig_user_id() string {
	return fmt.Sprintf(`
data "cloudscale_objects_user" "foo" {
  user_id = "${cloudscale_objects_user.basic.0.user_id}"
}
`)
}

func objectsUserConfig_baseline(count int, rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_objects_user" "basic" {
  count        = "%v"
  display_name = "terraform-%d-${count.index}"
}`, count, rInt)
}
