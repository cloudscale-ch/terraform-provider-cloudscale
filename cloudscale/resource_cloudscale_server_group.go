package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudscaleServerGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerGroupCreate,
		Read:   resourceServerGroupRead,
		Delete: resourceServerGroupDelete,

		Schema: getServerGroupSchema(false),
	}
}

func getServerGroupSchema(isDataSource bool) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: !isDataSource,
			Optional: isDataSource,
			ForceNew: true,
		},
		"type": {
			Type:     schema.TypeString,
			Required: !isDataSource,
			Computed: isDataSource,
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
	}
	if isDataSource {
		m["id"] = &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		}
	}
	return m
}

func resourceServerGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.ServerGroupRequest{
		Name: d.Get("name").(string),
		Type: d.Get("type").(string),
	}

	if attr, ok := d.GetOk("zone_slug"); ok {
		opts.Zone = attr.(string)
	}

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
	m := make(map[string]interface{})
	m["id"] = serverGroup.UUID
	m["href"] = serverGroup.HREF
	m["name"] = serverGroup.Name
	m["type"] = serverGroup.Type
	m["zone_slug"] = serverGroup.Zone.Slug
	return m
}

func resourceServerGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	serverGroup, err := client.ServerGroups.Get(context.Background(), d.Id())
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving server group")
	}

	fillServerGroupResourceData(d, serverGroup)
	return nil
}

func resourceServerGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	log.Printf("[INFO] Deleting ServerGroup: %s", d.Id())
	err := client.ServerGroups.Delete(context.Background(), id)

	if err != nil {
		return CheckDeleted(d, err, "Error deleting server group")
	}
	return nil
}
