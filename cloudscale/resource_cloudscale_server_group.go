package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudscaleServerGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerGroupCreate,
		Read:   resourceServerGroupRead,
		Update: resourceServerGroupUpdate,
		Delete: resourceServerGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: getServerGroupSchema(RESOURCE),
	}
}

func getServerGroupSchema(t SchemaType) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Optional: t.isDataSource(),
		},
		"type": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Computed: t.isDataSource(),
			ForceNew: true,
		},
		"zone_slug": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"tags": &TagsSchema,
	}
	if t.isDataSource() {
		m["id"] = &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		}
	}
	return m
}

func resourceServerGroupCreate(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.ServerGroupRequest{
		Name: d.Get("name").(string),
		Type: d.Get("type").(string),
	}

	if attr, ok := d.GetOk("zone_slug"); ok {
		opts.Zone = attr.(string)
	}
	opts.Tags = CopyTags(d)

	log.Printf("[DEBUG] ServerGroup create configuration: %#v", opts)

	serverGroup, err := client.ServerGroups.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating server group: %s", err)
	}

	d.SetId(serverGroup.UUID)

	log.Printf("[INFO] ServerGroup ID %s", d.Id())

	fillServerGroupResourceData(d, serverGroup)
	return nil
}

func fillServerGroupResourceData(d *schema.ResourceData, serverGroup *cloudscale.ServerGroup) {
	fillResourceData(d, gatherServerGroupResourceData(serverGroup))
}

func gatherServerGroupResourceData(serverGroup *cloudscale.ServerGroup) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = serverGroup.UUID
	m["href"] = serverGroup.HREF
	m["name"] = serverGroup.Name
	m["type"] = serverGroup.Type
	m["zone_slug"] = serverGroup.Zone.Slug
	m["tags"] = serverGroup.Tags
	return m
}

func resourceServerGroupRead(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)

	serverGroup, err := client.ServerGroups.Get(context.Background(), d.Id())
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving server group")
	}

	fillServerGroupResourceData(d, serverGroup)
	return nil
}

func resourceServerGroupUpdate(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	for _, attribute := range []string{"name", "tags"} {
		// cloudscale.ch ServerGroup attributes can only be changed one at a time.
		if d.HasChange(attribute) {
			opts := &cloudscale.ServerGroupRequest{}
			if attribute == "name" {
				opts.Name = d.Get(attribute).(string)
			} else if attribute == "tags" {
				opts.Tags = CopyTags(d)
			}
			err := client.ServerGroups.Update(context.Background(), id, opts)
			if err != nil {
				return fmt.Errorf("Error updating the Server Group (%s) status (%s) ", id, err)
			}
		}
	}
	return resourceServerGroupRead(d, meta)
}

func resourceServerGroupDelete(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	log.Printf("[INFO] Deleting ServerGroup: %s", d.Id())
	err := client.ServerGroups.Delete(context.Background(), id)

	if err != nil {
		return CheckDeleted(d, err, "Error deleting server group")
	}
	return nil
}
