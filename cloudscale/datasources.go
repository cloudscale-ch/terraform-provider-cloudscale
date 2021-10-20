package cloudscale

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ResourceDataRaw = map[string]interface{}

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
	fetch func(meta interface{}) ([]ResourceDataRaw, error),
) schema.ReadContextFunc {
    return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		resources, err := fetch(meta)
		if err != nil {
			return diag.Errorf("Issue with fetching resources: %s", err)
		}
		var foundItems []map[string]interface{}

		for _, m := range resources {
			match := true
			for key := range sourceSchema {
				if attr, ok := d.GetOk(key); ok {
					if m[key] != attr {
						match = false
						break
					}
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
