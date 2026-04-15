package cloudscale

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudscaleVolumeSnapshot_DS_Basic(t *testing.T) {
	var snap cloudscale.VolumeSnapshot
	rInt := acctest.RandInt()
	name1 := fmt.Sprintf("terraform-%d-0", rInt)
	name2 := fmt.Sprintf("terraform-%d-1", rInt)

	config := volumeSnapshotConfig_baseline(2, rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleVolumeSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				Config: config + testAccCheckCloudscaleVolumeSnapshotConfig_name(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleVolumeSnapshotExists("data.cloudscale_volume_snapshot.foo", &snap),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_volume_snapshot.basic.0", "id", &snap.UUID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_volume_snapshot.foo", "id", &snap.UUID),
					resource.TestCheckResourceAttr(
						"data.cloudscale_volume_snapshot.foo", "name", name1),
				),
			},
			{
				Config: config + testAccCheckCloudscaleVolumeSnapshotConfig_name_and_source_volume(name1, "${cloudscale_volume.source.id}"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_volume_snapshot.foo", "name", name1),
				),
			},
			{
				Config: config + testAccCheckCloudscaleVolumeSnapshotConfig_name(name2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_volume_snapshot.foo", "name", name2),
				),
			},
			{
				Config: config + testAccCheckCloudscaleVolumeSnapshotConfig_name_and_source_volume(name2, "${cloudscale_volume.source.id}"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudscale_volume_snapshot.foo", "name", name2),
				),
			},
			{
				Config:      config + testAccCheckCloudscaleVolumeSnapshotConfig_name_and_source_volume("nonexistent-name", "${cloudscale_volume.source.id}"),
				ExpectError: regexp.MustCompile(`Found zero volume snapshots`),
			},
			{
				Config: config + testAccCheckCloudscaleVolumeSnapshotConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_volume_snapshot.basic.0", "name", name1),
					resource.TestCheckResourceAttr(
						"data.cloudscale_volume_snapshot.foo", "name", name1),
					resource.TestCheckResourceAttrPtr(
						"cloudscale_volume_snapshot.basic.0", "id", &snap.UUID),
					resource.TestCheckResourceAttrPtr(
						"data.cloudscale_volume_snapshot.foo", "id", &snap.UUID),
				),
			},
			{
				Config:      config + "\n" + `data "cloudscale_volume_snapshot" "foo" {}`,
				ExpectError: regexp.MustCompile(`Found \d+ volume snapshots, expected one`),
			},
		},
	})
}

func volumeSnapshotConfig_baseline(count int, rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_volume" "source" {
  name    = "terraform-%d-vol"
  size_gb = 50
  type    = "ssd"
}

resource "cloudscale_volume_snapshot" "basic" {
  count              = %d
  name               = "terraform-%d-${count.index}"
  source_volume_uuid = cloudscale_volume.source.id
}`, rInt, count, rInt)
}

func testAccCheckCloudscaleVolumeSnapshotConfig_name(name string) string {
	return fmt.Sprintf(`
data "cloudscale_volume_snapshot" "foo" {
  name = "%s"
}
`, name)
}

func testAccCheckCloudscaleVolumeSnapshotConfig_name_and_source_volume(name, volumeRef string) string {
	return fmt.Sprintf(`
data "cloudscale_volume_snapshot" "foo" {
  name               = "%s"
  source_volume_uuid = "%s"
}
`, name, volumeRef)
}

func testAccCheckCloudscaleVolumeSnapshotConfig_id() string {
	return `
data "cloudscale_volume_snapshot" "foo" {
  id = "${cloudscale_volume_snapshot.basic.0.id}"
}
`
}
