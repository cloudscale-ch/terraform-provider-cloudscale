package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const serverGroupHumanName = "server group"

var resourceCloudscaleServerGroupRead = getReadOperation(serverGroupHumanName, getGenericResourceIdentifierFromSchema, readServerGroup, gatherServerGroupResourceData)
var resourceCloudscaleServerGroupDelete = getDeleteOperation(serverGroupHumanName, deleteServerGroup)

func resourceCloudscaleServerGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudscaleServerGroupCreate,
		Read:   resourceCloudscaleServerGroupRead,
		Update: resourceCloudscaleServerGroupUpdate,
		Delete: resourceCloudscaleServerGroupDelete,

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

func resourceCloudscaleServerGroupCreate(d *schema.ResourceData, meta any) error {
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

	err = resourceCloudscaleServerGroupRead(d, meta)
	if err != nil {
		return fmt.Errorf("Error reading the server group (%s): %s", d.Id(), err)
	}
	return nil
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

func readServerGroup(rId GenericResourceIdentifier, meta any) (*cloudscale.ServerGroup, error) {
	client := meta.(*cloudscale.Client)
	return client.ServerGroups.Get(context.Background(), rId.Id)
}

func resourceCloudscaleServerGroupUpdate(d *schema.ResourceData, meta any) error {
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
	return resourceCloudscaleServerGroupRead(d, meta)
}

func deleteServerGroup(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	return client.ServerGroups.Delete(context.Background(), id)
}
