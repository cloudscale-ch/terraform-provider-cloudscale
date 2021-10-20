package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func customImageSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type: schema.TypeString,
		},
		"name": {
			Type: schema.TypeString,
		},
		"user_data_handling": {
			Type: schema.TypeString,
		},
		"zone_slugs": {
			Type:     schema.TypeSet,
			Elem:     &schema.Schema{Type: schema.TypeString},
			ForceNew: true,
		},
		"slug": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"href": {
			Type: schema.TypeString,
			//Computed: true,
		},
		"size_gb": {
			Type: schema.TypeInt,
			//Computed: true,
		},
		"checksums": {
			Type: schema.TypeMap,
			Elem: &schema.Schema{
				Type: schema.TypeString,
				//Computed: true,
			},
			//Computed: true,
		},
	}
}

func dataSourceCloudScaleCustomImage() *schema.Resource {
	recordSchema := customImageSchema()

	for _, f := range recordSchema {
		f.Computed = true
	}

	return &schema.Resource{
		ReadContext: customImageRead,
		Schema:      recordSchema,
	}
}

func customImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*cloudscale.Client)

	imageList, err := client.CustomImages.List(ctx)
	if err != nil {
		return diag.Errorf("Fetch error %s", err)
	}
	var foundImages []map[string]interface{}

	for _, image := range imageList {
		m := make(map[string]interface{})
		m["id"] = image.UUID
		m["name"] = image.Name
		m["href"] = image.HREF
		m["slug"] = image.Slug
		m["size_gb"] = image.SizeGB
		m["user_data_handling"] = image.UserDataHandling
		m["checksums"] = image.Checksums

		zoneSlugs := make([]string, 0, len(image.Zones))
		for _, zone := range image.Zones {
			zoneSlugs = append(zoneSlugs, zone.Slug)
		}
		err := d.Set("zone_slugs", zoneSlugs)
		if err != nil {
			return diag.Errorf("Error setting zone_slugs attribute: %#v, error: %#v", image.Zones, err)
		}

		match := true
		for key := range customImageSchema() {
			if attr, ok := d.GetOk(key); ok {
				if m[key] != attr {
					match = false
				}
			}
		}
		if match {
			foundImages = append(foundImages, m)
		}
	}
	if len(foundImages) > 1 {
		return diag.Errorf("Found %d cccustom images, expected one", len(foundImages))
	} else if len(foundImages) == 0 {
		return diag.Errorf("Found zero cccustom images")
	}

	foundImage := foundImages[0]
	for k, v := range foundImage {
		d.Set(k, v)
	}
	d.SetId(foundImage["id"].(string))

	return nil
}
