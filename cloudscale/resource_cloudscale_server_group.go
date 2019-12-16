package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceCloudScaleServerGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerGroupCreate,
		Read:   resourceServerGroupRead,
		Delete: resourceServerGroupDelete,

		Schema: getServerGroupSchema(),
	}
}

func getServerGroupSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{

		// Required attributes

		"name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},

		"type": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},

		// Optional attributes

		"zone_slug": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},

		// Computed attributes

		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
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

	err = fillServerGroupResourceData(d, serverGroup)
	if err != nil {
		return err
	}
	return nil
}

func fillServerGroupResourceData(d *schema.ResourceData, serverGroup *cloudscale.ServerGroup) error {
	d.Set("href", serverGroup.HREF)
	d.Set("name", serverGroup.Name)
	d.Set("type", serverGroup.Type)
	d.Set("zone_slug", serverGroup.Zone.Slug)

	return nil
}

func resourceServerGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	serverGroup, err := client.ServerGroups.Get(context.Background(), d.Id())
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving server group")
	}

	err = fillServerGroupResourceData(d, serverGroup)
	if err != nil {
		return err
	}
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
