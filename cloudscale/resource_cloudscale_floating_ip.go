package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceCloudScaleFloatingIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceFloatingIPCreate,
		Read:   resourceFloatingIPRead,
		Update: resourceFloatingIPUpdate,
		Delete: resourceFloatingIPDelete,

		Schema: getFloatingIPSchema(),
	}
}

func getFloatingIPSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{

		// Required attributes

		"ip_version": {
			Type:     schema.TypeInt,
			Required: true,
			ForceNew: true,
		},
		"server": {
			Type:     schema.TypeString,
			Optional: true,
		},

		// Optional attributes

		"region_slug": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"type": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"reverse_ptr": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"prefix_length": {
			Type:     schema.TypeInt,
			ForceNew: true,
			Optional: true,
		},

		// Computed attributes

		"network": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"next_hop": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func resourceFloatingIPCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.FloatingIPCreateRequest{
		IPVersion: d.Get("ip_version").(int),
	}

	if attr, ok := d.GetOk("server"); ok {
		opts.Server = attr.(string)
	}

	if attr, ok := d.GetOk("prefix_length"); ok {
		opts.PrefixLength = attr.(int)
	}

	if attr, ok := d.GetOk("reverse_ptr"); ok {
		opts.ReversePointer = attr.(string)
	}

	if attr, ok := d.GetOk("region_slug"); ok {
		opts.Region = attr.(string)
	}

	if attr, ok := d.GetOk("type"); ok {
		opts.Type = attr.(string)
	}

	log.Printf("[DEBUG] FloatingIP create configuration: %#v", opts)

	floatingIP, err := client.FloatingIPs.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating FloatingIP: %s", err)
	}

	d.SetId(floatingIP.IP())

	return resourceFloatingIPRead(d, meta)
}
func resourceFloatingIPRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	id := d.Id()

	floatingIP, err := client.FloatingIPs.Get(context.Background(), id)
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving FloatingIP")
	}

	d.Set("href", floatingIP.HREF)
	d.Set("network", floatingIP.Network)
	d.Set("next_hop", floatingIP.NextHop)
	d.Set("reverse_ptr", floatingIP.ReversePointer)
	d.Set("type", floatingIP.Type)
	if floatingIP.Server != nil {
		d.Set("server", floatingIP.Server.UUID)
	}
	if floatingIP.Region != nil {
		d.Set("region_slug", floatingIP.Region.Slug)
	}

	return nil

}
func resourceFloatingIPUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	if d.HasChange("server") {
		serverUUID := d.Get("server").(string)

		id := d.Id()

		log.Printf("[INFO] Assigning the Floating IP %s to the Server %s", d.Id(), serverUUID)

		opts := &cloudscale.FloatingIPUpdateRequest{
			Server: serverUUID,
		}

		err := client.FloatingIPs.Update(context.Background(), id, opts)
		if err != nil {
			return fmt.Errorf("Error assigning FloatingIP (%s) to Server: %s", id, err)
		}
	}
	return resourceFloatingIPRead(d, meta)
}
func resourceFloatingIPDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	log.Printf("[INFO] Deleting FloatingIP: %s", d.Id())
	err := client.FloatingIPs.Delete(context.Background(), id)

	if err != nil {
		return CheckDeleted(d, err, "Error deleting floating IP")
	}

	return nil
}
