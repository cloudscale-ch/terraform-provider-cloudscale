package cloudscale

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func getReadOperation[TResource any](
	resourceType string,
	readFunc func(d *schema.ResourceData, meta any) (*TResource, error),
	gatherFunc func(resource *TResource) ResourceDataRaw,
) schema.ReadFunc {
	return func(d *schema.ResourceData, meta any) error {
		resource, err := readFunc(d, meta)

		if err != nil {
			return CheckDeleted(d, err, fmt.Sprintf("Error retrieving %s", resourceType))
		}

		fillResourceData(d, gatherFunc(resource))
		return nil
	}
}

func getDeleteOperation(
	resourceType string,
	deleteFunc func(d *schema.ResourceData, meta any) error,
) schema.DeleteFunc {
	return func(d *schema.ResourceData, meta any) error {
		log.Printf("[INFO] Deleting %s: %s", resourceType, d.Id())
		err := deleteFunc(d, meta)

		if err != nil {
			return CheckDeleted(d, err, fmt.Sprintf("Error deleting %s", resourceType))
		}
		return nil
	}
}
