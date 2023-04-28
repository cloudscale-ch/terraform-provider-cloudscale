package cloudscale

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	testAccCheckCloudscaleCustomImageExists               = getTestAccCheckCloudscaleResourceExistsFunc(customImageHumanName, getId, readCustomImage)
	testAccCheckCloudscaleFloatingIPExists                = getTestAccCheckCloudscaleResourceExistsFunc(floatingIPHumanName, getId, readFloatingIP)
	testAccCheckCloudscaleLoadBalancerExists              = getTestAccCheckCloudscaleResourceExistsFunc(loadBalancerHumanName, getId, readLoadBalancer)
	testAccCheckCloudscaleLoadBalancerHealthMonitorExists = getTestAccCheckCloudscaleResourceExistsFunc(healthMonitorHumanName, getId, readLoadBalancerHealthMonitor)
	testAccCheckCloudscaleLoadBalancerListenerExists      = getTestAccCheckCloudscaleResourceExistsFunc(listenerHumanName, getId, readLoadBalancerListener)
	testAccCheckCloudscaleLoadBalancerPoolExists          = getTestAccCheckCloudscaleResourceExistsFunc(poolHumanName, getId, readLoadBalancerPool)
	testAccCheckCloudscaleLoadBalancerPoolMemberExists    = getTestAccCheckCloudscaleResourceExistsFunc(poolMemberHumanName, getPoolId, readLoadBalancerPoolMember)
	testAccCheckCloudscaleNetworkExists                   = getTestAccCheckCloudscaleResourceExistsFunc(networkHumanName, getId, readNetwork)
	testAccCheckCloudscaleObjectsUserExists               = getTestAccCheckCloudscaleResourceExistsFunc(objectsUserHumanName, getId, readObjectsUser)
	testAccCheckCloudscaleServerExists                    = getTestAccCheckCloudscaleResourceExistsFunc(serverHumanName, getId, readServer)
	testAccCheckCloudscaleServerGroupExists               = getTestAccCheckCloudscaleResourceExistsFunc(serverGroupHumanName, getId, readServerGroup)
	testAccCheckCloudscaleSubnetExists                    = getTestAccCheckCloudscaleResourceExistsFunc(subnetHumanName, getId, readSubnet)
	testAccCheckCloudscaleVolumeExists                    = getTestAccCheckCloudscaleResourceExistsFunc(volumeHumanName, getId, readVolume)
)

func getId(rs *terraform.ResourceState) GenericResourceIdentifier {
	return GenericResourceIdentifier{
		Id: rs.Primary.ID,
	}
}

func getPoolId(rs *terraform.ResourceState) LoadBalancerPoolMemberResourceIdentifier {
	return LoadBalancerPoolMemberResourceIdentifier{
		Id:     rs.Primary.ID,
		PoolID: rs.Primary.Attributes["pool_uuid"],
	}
}

func getTestAccCheckCloudscaleResourceExistsFunc[TResource any, TResourceID any](
	resourceType string,
	idFunc func(d *terraform.ResourceState) TResourceID,
	readFunc func(rId TResourceID, meta any,
) (*TResource, error)) func(n string, resource *TResource) resource.TestCheckFunc {
	return func(
		n string,
		resource *TResource,
	) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources[n]
			if !ok {
				return fmt.Errorf("not found: %s", n)
			}
			if rs.Primary.ID == "" {
				return fmt.Errorf("no %s ID is set", resourceType)
			}

			resourceId := idFunc(rs)
			retrievedResource, err := readFunc(resourceId, testAccProvider.Meta())

			if err != nil {
				return err
			}

			*resource = *retrievedResource

			return nil
		}
	}
}
