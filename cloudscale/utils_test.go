package cloudscale

import (
	"fmt"
	"net/http"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v8"
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
	testAccCheckCloudscaleVolumeSnapshotExists            = getTestAccCheckCloudscaleResourceExistsFunc(volumeSnapshotHumanName, getId, readVolumeSnapshot)

	testAccCheckCloudscaleVolumeNotExists = getTestAccCheckCloudscaleResourceNotExistsFunc(
		volumeHumanName,
		func(v *cloudscale.Volume) GenericResourceIdentifier { return GenericResourceIdentifier{Id: v.UUID} },
		readVolume,
	)
	testAccCheckCloudscaleVolumeSnapshotNotExists = getTestAccCheckCloudscaleResourceNotExistsFunc(
		volumeSnapshotHumanName,
		func(vs *cloudscale.VolumeSnapshot) GenericResourceIdentifier { return GenericResourceIdentifier{Id: vs.UUID} },
		readVolumeSnapshot,
	)
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

func getTestAccCheckCloudscaleResourceNotExistsFunc[TResource any, TResourceID any](
	resourceType string,
	resourceIdFunc func(resource *TResource) TResourceID,
	readFunc func(rId TResourceID, meta any) (*TResource, error),
) func(resource *TResource) resource.TestCheckFunc {
	return func(resource *TResource) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			resourceId := resourceIdFunc(resource)
			_, err := readFunc(resourceId, testAccProvider.Meta())
			if err == nil {
				return fmt.Errorf("%s still exists", resourceType)
			}
			errorResponse, ok := err.(*cloudscale.ErrorResponse)
			if !ok || errorResponse.StatusCode != http.StatusNotFound {
				return fmt.Errorf("error verifying %s was destroyed: %s", resourceType, err)
			}
			return nil
		}
	}
}
