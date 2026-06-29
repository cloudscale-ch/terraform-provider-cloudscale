package cloudscale

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// getReadOperation builds a ReadFunc from discrete steps:
//   - idFunc:      extracts the resource identifier from state
//   - readFunc:    fetches the resource from the API by that identifier
//   - gatherFunc:  converts the API response into flat key/value pairs for state
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
			return CheckDeleted(d, err, fmt.Sprintf("Error retrieving %s (%v)", resourceHumanName, rId))
		}

		fillResourceData(d, gatherFunc(resource))
		return nil
	}
}

// getUpdateOperation builds an UpdateFunc from discrete steps:
//   - idFunc:              extracts the resource identifier from state
//   - updateFunc:          sends one update request to the API (called once per request)
//   - resourceReadFunc:    refreshes state after all updates complete
//   - gatherRequestsFunc:  builds the list of update requests from changed state
//   - mutexKeyFunc:        derives a mutex key to serialize concurrent operations; nil = no lock
func getUpdateOperation[TResourceID any, TRequest any](
	resourceHumanName string,
	idFunc func(d *schema.ResourceData) TResourceID,
	updateFunc func(rId TResourceID, meta any, updateRequest *TRequest) error,
	resourceReadFunc schema.ReadFunc,
	gatherRequestsFunc func(d *schema.ResourceData) []*TRequest,
	mutexKeyFunc func(d *schema.ResourceData) (string, bool),
) schema.UpdateFunc {
	return func(d *schema.ResourceData, meta any) error {
		if mutexKeyFunc != nil {
			if key, ok := mutexKeyFunc(d); ok {
				globalMu.Lock(key)
				defer globalMu.Unlock(key)
			}
		}
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

// getDeleteOperation builds a DeleteFunc from discrete steps:
//   - idFunc:          extracts the resource identifier from state
//   - deleteFunc:      calls the API to delete the resource by that identifier
//   - mutexKeyFunc:    derives a mutex key to serialize concurrent operations; nil = no lock
func getDeleteOperation[TResourceID any](
	resourceHumanName string,
	idFunc func(d *schema.ResourceData) TResourceID,
	deleteFunc func(rId TResourceID, meta any) error,
	mutexKeyFunc func(d *schema.ResourceData) (string, bool),
) schema.DeleteFunc {
	return func(d *schema.ResourceData, meta any) error {
		log.Printf("[INFO] Deleting %s: %s", resourceHumanName, d.Id())
		if mutexKeyFunc != nil {
			if key, ok := mutexKeyFunc(d); ok {
				globalMu.Lock(key)
				defer globalMu.Unlock(key)
			}
		}
		rId := idFunc(d)
		err := deleteFunc(rId, meta)

		if err != nil {
			return CheckDeleted(d, err, fmt.Sprintf("Error deleting %s", resourceHumanName))
		}
		return nil
	}
}

// withLock wraps a Create/Update/Delete operation so it is serialized on a key derived
// from the resource data. The cloudscale API rejects many concurrent operations that
// share some resource (e.g. a parent), so Terraform's parallelism must be serialized on
// that shared key.
//
// mutexKeyFunc derives the lock key from the resource data and returns ok=false when no lock
// is needed.
func withLock(
	mutexKeyFunc func(d *schema.ResourceData) (string, bool),
	op func(d *schema.ResourceData, meta any) error,
) func(d *schema.ResourceData, meta any) error {
	return func(d *schema.ResourceData, meta any) error {
		if key, ok := mutexKeyFunc(d); ok {
			globalMu.Lock(key)
			defer globalMu.Unlock(key)
		}
		return op(d, meta)
	}
}

type GenericResourceIdentifier struct {
	Id string
}

func getGenericResourceIdentifierFromSchema(d *schema.ResourceData) GenericResourceIdentifier {
	return GenericResourceIdentifier{Id: d.Id()}
}
