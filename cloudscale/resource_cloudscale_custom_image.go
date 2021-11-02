package cloudscale

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"log"
	"math"
	"time"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudscaleCustomImage() *schema.Resource {
	return &schema.Resource{
		Create: resourceCustomImageCreate,
		Read:   resourceCustomImageRead,
		Update: resourceCustomImageUpdate,
		Delete: resourceCustomImageDelete,

		Schema: getCustomImageSchema(false),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
		},
	}
}

func getCustomImageSchema(isDataSource bool) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: !isDataSource,
			Optional: isDataSource,
		},
		"user_data_handling": {
			Type:     schema.TypeString,
			Required: !isDataSource,
			Computed: isDataSource,
		},
		"zone_slugs": {
			Type:     schema.TypeSet,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Required: !isDataSource,
			Computed: isDataSource,
			ForceNew: true,
		},
		"slug": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"size_gb": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"checksums": {
			Type: schema.TypeMap,
			Elem: &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			Computed: true,
		},
	}
	if isDataSource {
		m["id"] = &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		}
	} else {
		m["import_url"] = &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		}
		m["import_source_format"] = &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		}
		m["import_uuid"] = &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		}
		m["import_status"] = &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		}
		m["import_href"] = &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		}
	}
	return m
}

func resourceCustomImageCreate(d *schema.ResourceData, meta interface{}) error {
	timeout := d.Timeout(schema.TimeoutCreate)
	startTime := time.Now()

	client := meta.(*cloudscale.Client)

	opts := &cloudscale.CustomImageImportRequest{
		URL:              d.Get("import_url").(string),
		Name:             d.Get("name").(string),
		Slug:             d.Get("slug").(string),
		UserDataHandling: d.Get("user_data_handling").(string),
		SourceFormat:     d.Get("import_source_format").(string),
		Zones:            nil,
	}
	zoneSlugs := d.Get("zone_slugs").(*schema.Set).List()
	z := make([]string, len(zoneSlugs))

	for i := range zoneSlugs {
		z[i] = zoneSlugs[i].(string)
	}

	opts.Zones = z

	log.Printf("[DEBUG] CustomImage create configuration: %#v", opts)

	customImageImport, err := client.CustomImageImports.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating customImageImport: %z", err)
	}

	d.SetId(customImageImport.CustomImage.UUID)

	log.Printf("[INFO] CustomImage ID %s", d.Id())

	remainingTime := timeout - time.Since(startTime)
	_, err = waitForCustomImageImportStatus(customImageImport.UUID, d, meta, []string{"in_progress"}, "import_status", "success", remainingTime)
	if err != nil {
		return fmt.Errorf("Error waiting for custom image import status (%s) (%s) ", customImageImport.UUID, err)
	}
	customImageImport, err = client.CustomImageImports.Get(context.Background(), customImageImport.UUID)
	if err != nil {
		return fmt.Errorf("Error getting customImage: %z", err)
	}
	customImage, err := client.CustomImages.Get(context.Background(), customImageImport.CustomImage.UUID)
	if err != nil {
		return fmt.Errorf("Error getting customImage: %z", err)
	}

	err = fillCustomImageResourceData(d, customImageImport, customImage)
	if err != nil {
		return err
	}
	return nil
}

func fillCustomImageResourceData(d *schema.ResourceData, customImageImport *cloudscale.CustomImageImport, customImage *cloudscale.CustomImage) error {
	fillResourceData(d, gatherCustomImageResourceData(customImage))

	// Here we add data for resources, but not for data sources. This means
	// that data sources will not have access to this content.
	d.Set("import_href", customImageImport.HREF)
	d.Set("import_uuid", customImageImport.UUID)
	d.Set("import_status", customImageImport.Status)

	return nil
}

func gatherCustomImageResourceData(customImage *cloudscale.CustomImage) ResourceDataRaw {
	m := make(map[string]interface{})
	m["id"] = customImage.UUID
	m["href"] = customImage.HREF
	m["name"] = customImage.Name
	m["slug"] = customImage.Slug
	m["size_gb"] = customImage.SizeGB
	m["user_data_handling"] = customImage.UserDataHandling
	m["checksums"] = customImage.Checksums

	zoneSlugs := make([]string, 0, len(customImage.Zones))
	for _, zone := range customImage.Zones {
		zoneSlugs = append(zoneSlugs, zone.Slug)
	}
	m["zone_slugs"] = zoneSlugs
	return m
}


func resourceCustomImageRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	customImage, err := client.CustomImages.Get(context.Background(), d.Id())
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving customImage")
	}

	importUUID, ok := d.GetOk("import_uuid")
	if !ok {
		return fmt.Errorf("Error getting import_uuid")
	}
	customImageImport, err := client.CustomImageImports.Get(context.Background(), importUUID.(string))

	err = fillCustomImageResourceData(d, customImageImport, customImage)
	if err != nil {
		return err
	}
	return nil
}

func resourceCustomImageUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	for _, attribute := range []string{"name", "slug", "user_data_handling"} {
		// cloudscale.ch customImage attributes can only be changed one at a time.
		if d.HasChange(attribute) {
			opts := &cloudscale.CustomImageRequest{}
			if attribute == "name" {
				opts.Name = d.Get(attribute).(string)
			} else if attribute == "slug" {
				opts.Slug = d.Get(attribute).(string)
			} else if attribute == "user_data_handling" {
				opts.UserDataHandling = d.Get(attribute).(string)
			}
			err := client.CustomImages.Update(context.Background(), id, opts)
			if err != nil {
				return fmt.Errorf("Error updating the CustomImage (%s) status (%s) ", id, err)
			}
		}
	}
	return resourceCustomImageRead(d, meta)
}

func resourceCustomImageDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	log.Printf("[INFO] Deleting CustomImage: %s", d.Id())
	err := client.CustomImages.Delete(context.Background(), id)

	if err != nil {
		return CheckDeleted(d, err, "Error deleting customImage")
	}
	return nil
}

func waitForCustomImageImportStatus(uuid string, d *schema.ResourceData, meta interface{}, pending []string, attribute, target string, timeout time.Duration) (interface{}, error) {
	log.Printf(
		"[INFO] Waiting %s for custom image import (%s) to have %s of %s",
		timeout, uuid, attribute, target)

	stateConf := &resource.StateChangeConf{
		Pending:        pending,
		Target:         []string{target},
		Refresh:        newCustomImageImportRefreshFunc(uuid, d, attribute, meta),
		Timeout:        timeout,
		Delay:          10 * time.Second,
		MinTimeout:     10 * time.Second,
		NotFoundChecks: math.MaxInt32,
	}

	return stateConf.WaitForState()
}

func newCustomImageImportRefreshFunc(uuid string, d *schema.ResourceData, attribute string, meta interface{}) resource.StateRefreshFunc {
	client := meta.(*cloudscale.Client)
	return func() (interface{}, string, error) {
		customImageImport, err := client.CustomImageImports.Get(context.Background(), uuid)
		if err != nil {
			return nil, "", fmt.Errorf("Error retrieving customImageImport (refresh) %s", err)
		}

		log.Printf("[INFO] Status is %s", customImageImport.Status)

		if customImageImport.Status == "failed" {
			return nil, "", fmt.Errorf("CustomImageImport status %s, abort", customImageImport.Status)
		}

		return customImageImport, customImageImport.Status, nil
	}
}
