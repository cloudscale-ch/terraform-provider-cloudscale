package cloudscale

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func getReadOperation[TResource any, TResourceID any](
	resourceType string,
	idFunc func(d *schema.ResourceData) TResourceID,
	readFunc func(rID TResourceID, meta any) (*TResource, error),
	gatherFunc func(resource *TResource) ResourceDataRaw,
) schema.ReadFunc {
	return func(d *schema.ResourceData, meta any) error {
		rId := idFunc(d)
		resource, err := readFunc(rId, meta)

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

type GenericResourceIdentifier struct {
	Id string
}

func getGenericResourceIdentifierFromSchema(d *schema.ResourceData) GenericResourceIdentifier {
	return GenericResourceIdentifier{Id: d.Id()}
}
