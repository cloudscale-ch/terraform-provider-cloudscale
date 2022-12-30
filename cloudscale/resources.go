package cloudscale

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func getReadOperation[TResource any, TResourceID any](
	resourceHumanName string,
	idFunc func(d *schema.ResourceData) TResourceID,
	readFunc func(rID TResourceID, meta any) (*TResource, error),
	gatherFunc func(resource *TResource) ResourceDataRaw,
) schema.ReadFunc {
	return func(d *schema.ResourceData, meta any) error {
		rId := idFunc(d)
		resource, err := readFunc(rId, meta)

		if err != nil {
			return CheckDeleted(d, err, fmt.Sprintf("Error retrieving %s", resourceHumanName))
		}

		fillResourceData(d, gatherFunc(resource))
		return nil
	}
}

func getUpdateOperation[TResourceID any, TRequest any](
	resourceHumanName string,
	idFunc func(d *schema.ResourceData) TResourceID,
	updateFunc func(rId TResourceID, meta any, updateRequest *TRequest) error,
	resourceReadFunc schema.ReadFunc,
	gatherRequestsFunc func(d *schema.ResourceData) []*TRequest,
) schema.UpdateFunc {
	return func(d *schema.ResourceData, meta any) error {
		rId := idFunc(d)
		updateRequests := gatherRequestsFunc(d)
		for _, request := range updateRequests {
			err := updateFunc(rId, meta, request)
			if err != nil {
				return fmt.Errorf("error updating the %s (%s) status (%s)", resourceHumanName, d.Id(), err)
			}
		}
		return resourceReadFunc(d, meta)
	}
}

func getDeleteOperation(
	resourceHumanName string,
	deleteFunc func(d *schema.ResourceData, meta any) error,
) schema.DeleteFunc {
	return func(d *schema.ResourceData, meta any) error {
		log.Printf("[INFO] Deleting %s: %s", resourceHumanName, d.Id())
		err := deleteFunc(d, meta)

		if err != nil {
			return CheckDeleted(d, err, fmt.Sprintf("Error deleting %s", resourceHumanName))
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
