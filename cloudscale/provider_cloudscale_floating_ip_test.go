package cloudscale

import (
	"context"
	"fmt"
	"testing"

	"github.com/cloudscale-ch/cloudscale"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("cloudscale_floating_ip", &resource.Sweeper{
		Name: "cloudscale_floating_ip",
		F:    testSweepFloatingIps,
	})

}

func testSweepFloatingIps(region string) error {
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client := meta.(*cloudscale.Client)

	ips, err := client.FloatingIPs.List(context.Background())
	if err != nil {
		return err
	}

	for _, ip := range ips {
		if err := client.FloatingIPs.Delete(context.Background(), ip.IP()); err != nil {
			return err
		}
	}

	return nil
}

func TestAccCloudscaleFloatingIP_Server(t *testing.T) {
	var floatingIP cloudscale.FloatingIP
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudScaleFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudScaleFloatingIPConfig_Server(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudScaleFloatingIPExists("cloudscale_floating_ip.gateway", &floatingIP),
					resource.TestCheckResourceAttr(
						"cloudscale_floating_ip.gateway", "ip_version", "4"),
				),
			},
		},
	})
}

func testAccCheckCloudScaleFloatingIPExists(n string, floatingIP *cloudscale.FloatingIP) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*cloudscale.Client)

		fmt.Println(rs.Primary.ID)

		id := rs.Primary.ID
		// Try to find the FloatingIP
		foundFloatingIP, err := client.FloatingIPs.Get(context.Background(), id)

		if err != nil {
			return err
		}

		if foundFloatingIP.IP() != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		*floatingIP = *foundFloatingIP

		return nil
	}
}

func testAccCheckCloudScaleFloatingIPDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudscale.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudscale_floating_ip.gateway" {
			continue
		}

		// Try to find the key
		_, err := client.FloatingIPs.Get(context.Background(), rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Floating IP still exists")
		}
	}

	return nil
}

func testAccCheckCloudScaleFloatingIPConfig_Server(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_server" "basic" {
  name      			= "terraform-%d"
  flavor    			= "flex-2"
  image     			= "debian-8"
  volume_size_gb	= 10
  ssh_keys = ["ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFEepRNW5hDct4AdJ8oYsb4lNP5E9XY5fnz3ZvgNCEv7m48+bhUjJXUPuamWix3zigp2lgJHC6SChI/okJ41GUY="]
}
resource "cloudscale_floating_ip" "gateway" {
  server 					= "${cloudscale_server.basic.id}"
  ip_version     	= 4
}`, rInt)
}
