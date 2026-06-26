package cloudscale

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ResourceDataRaw = map[string]any

func fillResourceData(d *schema.ResourceData, map_ ResourceDataRaw) {
	for k, v := range map_ {
		if k != "id" {
			d.Set(k, v)
		}
	}
}

func dataSourceResourceRead(
	name string,
	sourceSchema map[string]*schema.Schema,
	fetchFunc func(d *schema.ResourceData, meta any) ([]ResourceDataRaw, error),
) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		resources, err := fetchFunc(d, meta)
		if err != nil {
			return diag.Errorf("Issue with fetching resources: %s", err)
		}
		var foundItems []map[string]any

		// Filter resources: each set attribute must match (maps use subset semantics).
		for _, m := range resources {
			match := true
			for key, schemaEntry := range sourceSchema {
				attr, ok := d.GetOk(key)
				if !ok {
					continue // not a filter criterion
				}
				if schemaEntry.Type == schema.TypeMap {
					// Tags: all filter key-value pairs must be present in the resource (subset, not exact).
					filterMap := attr.(map[string]interface{})
					resourceMap, _ := m[key].(map[string]interface{})
					for fk, fv := range filterMap {
						if resourceMap[fk] != fv {
							match = false
							break // one tag mismatch is sufficient
						}
					}
				} else if schemaEntry.Type == schema.TypeList || schemaEntry.Type == schema.TypeSet {
					// Gather functions return []string from the SDK struct; d.GetOk returns []interface{}.
					// Normalise before comparing so reflect.DeepEqual sees the same dynamic type.
					if !reflect.DeepEqual(toInterfaceSlice(m[key]), attr) {
						match = false
					}
				} else if !reflect.DeepEqual(m[key], attr) {
					match = false
				}
				if !match {
					break // skip remaining attributes
				}
			}
			if match {
				foundItems = append(foundItems, m)
			}
		}
		if len(foundItems) > 1 {
			return diag.Errorf("Found %d %s, expected one", len(foundItems), name)
		} else if len(foundItems) == 0 {
			return diag.Errorf("Found zero %s", name)
		}
		item := foundItems[0]
		d.SetId(item["id"].(string))
		delete(item, "id")
		fillResourceData(d, item)

		return nil
	}
}

func toInterfaceSlice(v any) []interface{} {
	switch s := v.(type) {
	case []string:
		result := make([]interface{}, len(s))
		for i, str := range s {
			result[i] = str
		}
		return result
	case []interface{}:
		return s
	default:
		return nil
	}
}

func getFetchFunc[TResource any](
	listFunc func(d *schema.ResourceData, meta any) ([]TResource, error),
	gatherFunc func(resource *TResource) ResourceDataRaw,
) func(d *schema.ResourceData, meta any) ([]ResourceDataRaw, error) {
	return func(d *schema.ResourceData, meta any) ([]ResourceDataRaw, error) {
		list, err := listFunc(d, meta)
		if err != nil {
			return nil, err
		}

		var rawItems []ResourceDataRaw
		for _, resource := range list {

			rawItems = append(rawItems, gatherFunc(&resource))
		}
		return rawItems, nil
	}
}
