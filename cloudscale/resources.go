package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// getCreateOperation builds a CreateFunc from discrete steps:
//   - createFunc:    the underlying create implementation
//   - mutexKeyFunc:  derives a mutex key to serialize concurrent operations; nil = no lock
func getCreateOperation(
	createFunc schema.CreateContextFunc,
	mutexKeyFunc func(d *schema.ResourceData) (string, bool),
) schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		if mutexKeyFunc != nil {
			if key, ok := mutexKeyFunc(d); ok {
				if err := globalMu.LockContext(ctx, key); err != nil {
					return diag.FromErr(err)
				}
				defer globalMu.Unlock(key)
			}
		}
		return createFunc(ctx, d, meta)
	}
}

// getReadOperation builds a ReadFunc from discrete steps:
//   - idFunc:      extracts the resource identifier from state
//   - readFunc:    fetches the resource from the API by that identifier
//   - gatherFunc:  converts the API response into flat key/value pairs for state
func getReadOperation[TResource any, TResourceID any](
	resourceHumanName string,
	idFunc func(d *schema.ResourceData) TResourceID,
	readFunc func(ctx context.Context, rID TResourceID, meta any) (*TResource, error),
	gatherFunc func(resource *TResource) ResourceDataRaw,
) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		rId := idFunc(d)
		resource, err := readFunc(ctx, rId, meta)

		if err != nil {
			return diag.FromErr(CheckDeleted(d, err, fmt.Sprintf("Error retrieving %s (%v)", resourceHumanName, rId)))
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
	updateFunc func(ctx context.Context, rId TResourceID, meta any, updateRequest *TRequest) error,
	resourceReadFunc schema.ReadContextFunc,
	gatherRequestsFunc func(d *schema.ResourceData) []*TRequest,
	mutexKeyFunc func(d *schema.ResourceData) (string, bool),
) schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		if mutexKeyFunc != nil {
			if key, ok := mutexKeyFunc(d); ok {
				if err := globalMu.LockContext(ctx, key); err != nil {
					return diag.FromErr(err)
				}
				defer globalMu.Unlock(key)
			}
		}
		rId := idFunc(d)
		updateRequests := gatherRequestsFunc(d)
		for _, request := range updateRequests {
			err := updateFunc(ctx, rId, meta, request)
			if err != nil {
				return diag.FromErr(fmt.Errorf("error updating the %s (%s) status (%s)", resourceHumanName, d.Id(), err))
			}
		}
		return resourceReadFunc(ctx, d, meta)
	}
}

// getDeleteOperation builds a DeleteFunc from discrete steps:
//   - idFunc:          extracts the resource identifier from state
//   - deleteFunc:      calls the API to delete the resource by that identifier
//   - mutexKeyFunc:    derives a mutex key to serialize concurrent operations; nil = no lock
func getDeleteOperation[TResourceID any](
	resourceHumanName string,
	idFunc func(d *schema.ResourceData) TResourceID,
	deleteFunc func(ctx context.Context, rId TResourceID, meta any) error,
	mutexKeyFunc func(d *schema.ResourceData) (string, bool),
) schema.DeleteContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		log.Printf("[INFO] Deleting %s: %s", resourceHumanName, d.Id())
		if mutexKeyFunc != nil {
			if key, ok := mutexKeyFunc(d); ok {
				if err := globalMu.LockContext(ctx, key); err != nil {
					return diag.FromErr(err)
				}
				defer globalMu.Unlock(key)
			}
		}
		rId := idFunc(d)
		err := deleteFunc(ctx, rId, meta)

		if err != nil {
			return diag.FromErr(CheckDeleted(d, err, fmt.Sprintf("Error deleting %s", resourceHumanName)))
		}
		return nil
	}
}

// uuidLockKey returns a mutexKeyFunc that derives a lock key from a UUID attribute.
// If attr is not set — which should never happen for Required attributes — it skips
// the lock rather than acquiring one on a meaningless empty key.
func uuidLockKey(attr string, keyFunc func(string) string) func(*schema.ResourceData) (string, bool) {
	return func(d *schema.ResourceData) (string, bool) {
		uuid, ok := d.GetOk(attr)
		if !ok {
			log.Printf("[WARN] uuidLockKey: %s not set for resource %s, skipping lock", attr, d.Id())
			return "", false
		}
		return keyFunc(uuid.(string)), true
	}
}

type GenericResourceIdentifier struct {
	Id string
}

func getGenericResourceIdentifierFromSchema(d *schema.ResourceData) GenericResourceIdentifier {
	return GenericResourceIdentifier{Id: d.Id()}
}
