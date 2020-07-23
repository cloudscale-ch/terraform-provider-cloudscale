package cloudscale

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func init() {
	resource.AddTestSweepers("cloudscale_objects_user", &resource.Sweeper{
		Name: "cloudscale_objects_user",
		F:    testSweepObjectsUsers,
	})
}

func testSweepObjectsUsers(region string) error {
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client := meta.(*cloudscale.Client)

	ObjectsUsers, err := client.ObjectsUsers.List(context.Background())
	if err != nil {
		return err
	}

	foundError := error(nil)
	for _, u := range ObjectsUsers {
		if strings.HasPrefix(u.DisplayName, "terraform-") {
			log.Printf("Destroying ObjectsUser %#v", u.DisplayName)

			if err := client.ObjectsUsers.Delete(context.Background(), u.ID); err != nil {
				foundError = err
			}
		}
	}
	return foundError
}

func TestAccCloudscaleObjectsUser_Minimal(t *testing.T) {
	var objectsUser cloudscale.ObjectsUser

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleObjectsUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: objectsUserConfigMinimal(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleObjectsUserExists("cloudscale_objects_user.basic", &objectsUser),
					resource.TestCheckResourceAttr("cloudscale_objects_user.basic", "display_name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "href"),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "user_id"),
					resource.TestCheckResourceAttr("cloudscale_objects_user.basic", "keys.#", "1"),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "keys.0.access_key"),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "keys.0.secret_key"),
				),
			},
		},
	})
}

func TestAccCloudscaleObjectsUser_Rename(t *testing.T) {
	var objectsUser cloudscale.ObjectsUser

	rInt1 := acctest.RandInt()
	rInt2 := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleObjectsUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: objectsUserConfigMinimal(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleObjectsUserExists("cloudscale_objects_user.basic", &objectsUser),
					resource.TestCheckResourceAttr("cloudscale_objects_user.basic", "display_name", fmt.Sprintf("terraform-%d", rInt1)),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "href"),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "user_id"),
					resource.TestCheckResourceAttr("cloudscale_objects_user.basic", "keys.#", "1"),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "keys.0.access_key"),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "keys.0.secret_key"),
				),
			},
			{
				Config: objectsUserConfigMinimal(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleObjectsUserExists("cloudscale_objects_user.basic", &objectsUser),
					resource.TestCheckResourceAttr("cloudscale_objects_user.basic", "display_name", fmt.Sprintf("terraform-%d", rInt2)),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "href"),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "user_id"),
					resource.TestCheckResourceAttr("cloudscale_objects_user.basic", "keys.#", "1"),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "keys.0.access_key"),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "keys.0.secret_key"),
				),
			},
		},
	})
}

func testAccCheckCloudscaleObjectsUserExists(n string, objectsUser *cloudscale.ObjectsUser) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No objects user ID is set")
		}

		client := testAccProvider.Meta().(*cloudscale.Client)

		id := rs.Primary.ID

		// Try to find the objectsUser
		retrieveObjectsUser, err := client.ObjectsUsers.Get(context.Background(), id)

		if err != nil {
			return err
		}

		if retrieveObjectsUser.ID != rs.Primary.ID {
			return fmt.Errorf("ObjectsUser not found")
		}

		*objectsUser = *retrieveObjectsUser

		return nil
	}
}

func testAccCheckCloudscaleObjectsUserDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudscale.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudscale_objects_user" {
			continue
		}

		id := rs.Primary.ID

		// Try to find the objectsUser
		v, err := client.ObjectsUsers.Get(context.Background(), id)
		if err == nil {
			return fmt.Errorf("objects user %v still exists", v)
		} else {
			errorResponse, ok := err.(*cloudscale.ErrorResponse)
			if !ok || errorResponse.StatusCode != http.StatusNotFound {
				return fmt.Errorf(
					"Error waiting for objectsUser (%s) to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func objectsUserConfigMinimal(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_objects_user" "basic" {
  display_name    = "terraform-%d"
}
`, rInt)
}
