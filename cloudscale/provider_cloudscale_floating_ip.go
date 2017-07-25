package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCloudScaleFloatingIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceFloatingIPCreate,
		Read:   resourceFloatingIPRead,
		Update: resourceFloatingIPUpdate,
		Delete: resourceFloatingIPDelete,

		Schema: map[string]*schema.Schema{
			"ip_version": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"server": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"reverse_prt": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"network": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"next_hop": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"href": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceFloatingIPCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.FloatingIPCreateRequest{
		IPVersion: d.Get("ip_version").(int),
		Server:    d.Get("server").(string),
	}

	if attr, ok := d.GetOk("reverse_prt"); ok {
		opts.ReversePointer = attr.(string)
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
		if err.Error() == "detail: Not Found." {
			log.Printf("[WARN] Cloudscale FloatingIP (%s) not found", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving FloatingIP: %s", err)
	}

	d.Set("href", floatingIP.HREF)
	d.Set("network", floatingIP.Network)
	d.Set("next_hop", floatingIP.NextHop)
	d.Set("reverse_ptr", floatingIP.ReversePointer)
	d.Set("server", floatingIP.Server.UUID)

	return nil

}
func resourceFloatingIPUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}
func resourceFloatingIPDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	id := d.Id()

	err := client.FloatingIPs.Delete(context.Background(), id)
	if err != nil {
		return fmt.Errorf("Error deleting FloatingIP: %s", err)
	}

	return nil
}
