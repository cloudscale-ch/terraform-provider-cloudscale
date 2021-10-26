package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudscaleFloatingIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceFloatingIPCreate,
		Read:   resourceFloatingIPRead,
		Update: resourceFloatingIPUpdate,
		Delete: resourceFloatingIPDelete,

		Schema: getFloatingIPSchema(false),
	}
}

func getFloatingIPSchema(isDataSource bool) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
		"ip_version": {
			Type:     schema.TypeInt,
			Required: !isDataSource,
			Optional: isDataSource,
			ForceNew: true,
		},
		"server": {
			Type:     schema.TypeString,
			Optional: !isDataSource,
			Computed: isDataSource,
		},
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
			Optional: !isDataSource,
			Computed: true,
		},
		"network": {
			Type:     schema.TypeString,
			Optional:   isDataSource,
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
	return m
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

	err = fillFloatingIPResourceData(d, floatingIP)
	if err != nil {
		return err
	}
	return nil
}

func fillFloatingIPResourceData(d *schema.ResourceData, floatingIP *cloudscale.FloatingIP) error {
	fillResourceData(d, gatherFloatingIPResourceData(floatingIP))
	return nil
}

func gatherFloatingIPResourceData(floatingIP *cloudscale.FloatingIP) ResourceDataRaw {
	m := make(map[string]interface{})
	m["id"] = floatingIP.IP()
	m["href"] = floatingIP.HREF
	m["ip_version"] = floatingIP.IPVersion
	m["prefix_length"] = floatingIP.PrefixLength()
	m["network"] = floatingIP.Network
	m["next_hop"] = floatingIP.NextHop
	m["reverse_ptr"] = floatingIP.ReversePointer
	m["type"] = floatingIP.Type
	if floatingIP.Server != nil {
		m["server"] = floatingIP.Server.UUID
	}
	if floatingIP.Region != nil {
		m["region_slug"] = floatingIP.Region.Slug
	}

	return m
}

func resourceFloatingIPRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	id := d.Id()

	floatingIP, err := client.FloatingIPs.Get(context.Background(), id)
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving FloatingIP")
	}

	err = fillFloatingIPResourceData(d, floatingIP)
	if err != nil {
		return err
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
