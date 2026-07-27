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
	fetchFunc func(ctx context.Context, d *schema.ResourceData, meta any) ([]ResourceDataRaw, error),
) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		resources, err := fetchFunc(ctx, d, meta)
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
					filterMap := attr.(map[string]any)
					resourceMap, _ := m[key].(map[string]any)
					for fk, fv := range filterMap {
						if resourceMap[fk] != fv {
							match = false
							break // one tag mismatch is sufficient
						}
					}
				} else if schemaEntry.Type == schema.TypeList {
					// Gather functions return []string from the SDK struct; d.GetOk returns []any.
					// Normalise before comparing so reflect.DeepEqual sees the same dynamic type.
					if !reflect.DeepEqual(toAnySlice(m[key]), attr) {
						match = false
					}
				} else if schemaEntry.Type == schema.TypeSet {
					// As of this writing no data source filter field uses TypeSet, but fields
					// like ssh_keys and server_group_ids do on the resource side and are
					// candidates to be added. For those, subset semantics make sense: filtering
					// by ssh_keys = ["key-a"] should match a server that has key-a among its
					// keys, not only servers with exactly that one key.
					// d.GetOk returns *schema.Set; build a lookup from the resource slice and
					// check that every filter element is present.
					filterList := attr.(*schema.Set).List()
					resourceSlice := toAnySlice(m[key])
					resourceLookup := make(map[any]struct{}, len(resourceSlice))
					for _, v := range resourceSlice {
						resourceLookup[v] = struct{}{}
					}
					for _, v := range filterList {
						if _, ok := resourceLookup[v]; !ok {
							match = false
							break
						}
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

func toAnySlice(v any) []any {
	switch s := v.(type) {
	case []string:
		result := make([]any, len(s))
		for i, str := range s {
			result[i] = str
		}
		return result
	case []any:
		return s
	default:
		return nil
	}
}

func getFetchFunc[TResource any](
	listFunc func(ctx context.Context, d *schema.ResourceData, meta any) ([]TResource, error),
	gatherFunc func(resource *TResource) ResourceDataRaw,
) func(ctx context.Context, d *schema.ResourceData, meta any) ([]ResourceDataRaw, error) {
	return func(ctx context.Context, d *schema.ResourceData, meta any) ([]ResourceDataRaw, error) {
		list, err := listFunc(ctx, d, meta)
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
