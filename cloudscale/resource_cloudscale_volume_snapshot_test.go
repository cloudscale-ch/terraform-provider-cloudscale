package cloudscale

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("cloudscale_volume_snapshot", &resource.Sweeper{
		Name: "cloudscale_volume_snapshot",
		F:    testSweepVolumeSnapshots,
	})
}

func testSweepVolumeSnapshots(region string) error {
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client := meta.(*cloudscale.Client)

	snapshots, err := client.VolumeSnapshots.List(context.Background())
	if err != nil {
		return err
	}

	foundError := error(nil)
	for _, s := range snapshots {
		if strings.HasPrefix(s.Name, "terraform-") {
			log.Printf("Destroying volume snapshot %s", s.Name)

			if err := client.VolumeSnapshots.Delete(context.Background(), s.UUID); err != nil {
				foundError = err
			}
		}
	}
	return foundError
}

func TestAccCloudscaleVolumeSnapshot_Basic(t *testing.T) {
	var sourceVolume cloudscale.Volume
	var snapshot cloudscale.VolumeSnapshot

	rInt := acctest.RandInt()
	snapName := fmt.Sprintf("terraform-%d-snap", rInt)

	resourceName := "cloudscale_volume_snapshot.basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleVolumeSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleVolumeSnapshotConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleVolumeExists("cloudscale_volume.source", &sourceVolume),
					testAccCheckCloudscaleVolumeSnapshotExists(resourceName, &snapshot),
					resource.TestCheckResourceAttrPtr(
						resourceName, "id", &snapshot.UUID),
					resource.TestCheckResourceAttrPtr(
						resourceName, "href", &snapshot.HREF),
					resource.TestCheckResourceAttr(
						resourceName, "name", snapName),
					resource.TestCheckResourceAttr(
						resourceName, "status", "available"),
					resource.TestCheckResourceAttrPair(
						resourceName, "source_volume_uuid",
						"cloudscale_volume.source", "id"),
					resource.TestCheckResourceAttr(
						resourceName, "size_gb", "50"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccCloudscaleVolumeSnapshot_UpdateName(t *testing.T) {
	var sourceVolume cloudscale.Volume
	var afterCreate, afterUpdate cloudscale.VolumeSnapshot

	rInt1 := acctest.RandInt()
	snapName := fmt.Sprintf("terraform-%d-snap", rInt1)
	rInt2 := acctest.RandInt()
	updatedName := fmt.Sprintf("terraform-%d-snap", rInt2)

	resourceName := "cloudscale_volume_snapshot.basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleVolumeSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleVolumeSnapshotConfig_basic(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleVolumeExists("cloudscale_volume.source", &sourceVolume),
					testAccCheckCloudscaleVolumeSnapshotExists(resourceName, &afterCreate),
					resource.TestCheckResourceAttrPtr(
						resourceName, "id", &afterCreate.UUID),
					resource.TestCheckResourceAttrPtr(
						resourceName, "href", &afterCreate.HREF),
					resource.TestCheckResourceAttr(
						resourceName, "name", snapName),
					resource.TestCheckResourceAttr(
						resourceName, "status", "available"),
					resource.TestCheckResourceAttrPair(
						resourceName, "source_volume_uuid",
						"cloudscale_volume.source", "id"),
					resource.TestCheckResourceAttr(
						resourceName, "size_gb", "50"),
				),
			},
			{
				Config: testAccCloudscaleVolumeSnapshotConfig_basic(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleVolumeSnapshotExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "name", updatedName),
					testAccCheckVolumeSnapshotIsSame(t, &afterCreate, &afterUpdate),
				),
			},
		},
	})
}

func TestAccCloudscaleVolumeSnapshot_import_basic(t *testing.T) {
	var afterImport, afterUpdate cloudscale.VolumeSnapshot

	rInt1 := acctest.RandInt()
	snapName := fmt.Sprintf("terraform-%d-snap", rInt1)
	rInt2 := acctest.RandInt()
	updatedName := fmt.Sprintf("terraform-%d-snap", rInt2)

	resourceName := "cloudscale_volume_snapshot.basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleVolumeSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleVolumeSnapshotConfig_basic(rInt1),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccCloudscaleVolumeSnapshotConfig_basic(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleVolumeSnapshotExists(resourceName, &afterImport),
					resource.TestCheckResourceAttr(
						resourceName, "name", snapName),
				),
			},
			{
				Config: testAccCloudscaleVolumeSnapshotConfig_basic(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleVolumeSnapshotExists(resourceName, &afterUpdate),
					resource.TestCheckResourceAttr(
						resourceName, "name", updatedName),
					testAccCheckVolumeSnapshotIsSame(t, &afterImport, &afterUpdate),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateId:     "does-not-exist",
				ExpectError:       regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

func TestAccCloudscaleVolumeSnapshot_tags(t *testing.T) {
	var sourceVolume cloudscale.Volume
	var snapshot cloudscale.VolumeSnapshot

	rInt := acctest.RandInt()

	resourceName := "cloudscale_volume_snapshot.basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleVolumeSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudscaleVolumeSnapshotConfig_withTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleVolumeExists("cloudscale_volume.source", &sourceVolume),
					testAccCheckCloudscaleVolumeSnapshotExists(resourceName, &snapshot),
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
				Config: testAccCloudscaleVolumeSnapshotConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "0"),
					testTagsMatch(resourceName),
				),
			},
			{
				Config: testAccCloudscaleVolumeSnapshotConfig_withTags(rInt),
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

func testAccCheckCloudscaleVolumeSnapshotDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudscale.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudscale_volume_snapshot" {
			continue
		}

		id := rs.Primary.ID

		snap, err := client.VolumeSnapshots.Get(context.Background(), id)
		if err == nil {
			return fmt.Errorf("The volume snapshot %v remained, even though the resource was destroyed", snap)
		} else {
			errorResponse, ok := err.(*cloudscale.ErrorResponse)
			if !ok || errorResponse.StatusCode != http.StatusNotFound {
				return fmt.Errorf(
					"Error waiting for volume snapshot (%s) to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func testAccCheckVolumeSnapshotIsSame(t *testing.T,
	before, after *cloudscale.VolumeSnapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if adr := before; adr == after {
			t.Fatalf("Passed the same instance twice, address is equal=%v",
				adr)
		}
		if before.UUID != after.UUID {
			t.Fatalf("Not expected a change of VolumeSnapshot IDs got=%s, expected=%s",
				after.UUID, before.UUID)
		}
		return nil
	}
}

func testAccCloudscaleVolumeSnapshotConfig_base(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_volume" "source" {
  name    = "terraform-%d-vol"
  size_gb = 50
  type    = "ssd"
}
`, rInt)
}

func testAccCloudscaleVolumeSnapshotConfig_basic(rInt int) string {
	return testAccCloudscaleVolumeSnapshotConfig_base(rInt) + fmt.Sprintf(`
resource "cloudscale_volume_snapshot" "basic" {
  name               = "terraform-%d-snap"
  source_volume_uuid = cloudscale_volume.source.id
}
`, rInt)
}

func testAccCloudscaleVolumeSnapshotConfig_withTags(rInt int) string {
	return testAccCloudscaleVolumeSnapshotConfig_base(rInt) + fmt.Sprintf(`
resource "cloudscale_volume_snapshot" "basic" {
  name               = "terraform-%d-snap"
  source_volume_uuid = cloudscale_volume.source.id
  tags = {
    my-foo = "foo"
    my-bar = "bar"
  }
}
`, rInt)
}