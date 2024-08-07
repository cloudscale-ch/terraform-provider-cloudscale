package cloudscale

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"log"
	"math"
	"time"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const customImageHumanName = "custom image"

var (
	resourceCustomImageRead   = getReadOperation(customImageHumanName, getGenericResourceIdentifierFromSchema, readCustomImage, gatherCustomImageResourceData)
	resourceCustomImageUpdate = getUpdateOperation(customImageHumanName, getGenericResourceIdentifierFromSchema, updateCustomImage, resourceCustomImageRead, gatherCustomImageUpdateRequest)
	resourceCustomImageDelete = getDeleteOperation(customImageHumanName, deleteCustomImage)
)

func resourceCloudscaleCustomImage() *schema.Resource {
	return &schema.Resource{
		Create: resourceCustomImageCreate,
		Read:   resourceCustomImageRead,
		Update: resourceCustomImageUpdate,
		Delete: resourceCustomImageDelete,

		Schema: getCustomImageSchema(RESOURCE),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
		},
	}
}

func getCustomImageSchema(t SchemaType) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Optional: t.isDataSource(),
		},
		"user_data_handling": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Computed: t.isDataSource(),
		},
		"firmware_type": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"zone_slugs": {
			Type:     schema.TypeSet,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Required: t.isResource(),
			Computed: t.isDataSource(),
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
		"tags": &TagsSchema,
	}
	if t.isDataSource() {
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
			Optional: true,
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

func resourceCustomImageCreate(d *schema.ResourceData, meta any) error {
	timeout := d.Timeout(schema.TimeoutCreate)
	startTime := time.Now()

	client := meta.(*cloudscale.Client)

	opts := &cloudscale.CustomImageImportRequest{
		URL:              d.Get("import_url").(string),
		Name:             d.Get("name").(string),
		Slug:             d.Get("slug").(string),
		UserDataHandling: d.Get("user_data_handling").(string),
		Zones:            nil,
	}
	opts.Tags = CopyTags(d)
	zoneSlugs := d.Get("zone_slugs").(*schema.Set).List()
	z := make([]string, len(zoneSlugs))
	for i := range zoneSlugs {
		z[i] = zoneSlugs[i].(string)
	}
	opts.Zones = z

	if attr, ok := d.GetOk("firmware_type"); ok {
		opts.FirmwareType = attr.(string)
	}
	if attr, ok := d.GetOk("import_source_format"); ok {
		opts.SourceFormat = attr.(string)
	}

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

	fillCustomImageResourceData(d, customImageImport, customImage)
	return nil
}

func fillCustomImageResourceData(d *schema.ResourceData, customImageImport *cloudscale.CustomImageImport, customImage *cloudscale.CustomImage) {
	fillResourceData(d, gatherCustomImageResourceData(customImage))

	// Here we add data for resources, but not for data sources. This means
	// that data sources will not have access to this content.
	d.Set("import_href", customImageImport.HREF)
	d.Set("import_uuid", customImageImport.UUID)
	d.Set("import_status", customImageImport.Status)
}

func gatherCustomImageResourceData(customImage *cloudscale.CustomImage) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = customImage.UUID
	m["href"] = customImage.HREF
	m["name"] = customImage.Name
	m["slug"] = customImage.Slug
	m["size_gb"] = customImage.SizeGB
	m["user_data_handling"] = customImage.UserDataHandling
	m["firmware_type"] = customImage.FirmwareType
	m["checksums"] = customImage.Checksums
	m["tags"] = customImage.Tags

	zoneSlugs := make([]string, 0, len(customImage.Zones))
	for _, zone := range customImage.Zones {
		zoneSlugs = append(zoneSlugs, zone.Slug)
	}
	m["zone_slugs"] = zoneSlugs
	return m
}

func readCustomImage(rId GenericResourceIdentifier, meta any) (*cloudscale.CustomImage, error) {
	client := meta.(*cloudscale.Client)
	return client.CustomImages.Get(context.Background(), rId.Id)
}

func updateCustomImage(rId GenericResourceIdentifier, meta any, updateRequest *cloudscale.CustomImageRequest) error {
	client := meta.(*cloudscale.Client)
	return client.CustomImages.Update(context.Background(), rId.Id, updateRequest)
}

func gatherCustomImageUpdateRequest(d *schema.ResourceData) []*cloudscale.CustomImageRequest {
	requests := make([]*cloudscale.CustomImageRequest, 0)

	for _, attribute := range []string{"name", "slug", "user_data_handling", "tags"} {
		if d.HasChange(attribute) {
			log.Printf("[INFO] Attribute %s changed", attribute)
			opts := &cloudscale.CustomImageRequest{}
			requests = append(requests, opts)

			if attribute == "name" {
				opts.Name = d.Get(attribute).(string)
			} else if attribute == "slug" {
				opts.Slug = d.Get(attribute).(string)
			} else if attribute == "user_data_handling" {
				opts.UserDataHandling = d.Get(attribute).(string)
			} else if attribute == "tags" {
				opts.Tags = CopyTags(d)
			}
		}
	}
	return requests
}

func deleteCustomImage(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	return client.CustomImages.Delete(context.Background(), id)
}

func waitForCustomImageImportStatus(uuid string, d *schema.ResourceData, meta any, pending []string, attribute, target string, timeout time.Duration) (any, error) {
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

func newCustomImageImportRefreshFunc(uuid string, d *schema.ResourceData, attribute string, meta any) resource.StateRefreshFunc {
	client := meta.(*cloudscale.Client)
	return func() (any, string, error) {
		customImageImport, err := client.CustomImageImports.Get(context.Background(), uuid)
		if err != nil {
			return nil, "", fmt.Errorf("Error retrieving customImageImport (%s) (refresh) %s", uuid, err)
		}

		log.Printf("[INFO] Status is %s", customImageImport.Status)

		if customImageImport.Status == "failed" {
			return nil, "", fmt.Errorf("CustomImageImport status %s, abort", customImageImport.Status)
		}

		return customImageImport, customImageImport.Status, nil
	}
}
