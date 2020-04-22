package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceCloudScaleNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkCreate,
		Read:   resourceNetworkRead,
		Update: resourceNetworkUpdate,
		Delete: resourceNetworkDelete,

		Schema: getNetworkSchema(),
	}
}

func getNetworkSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// Required attributes

		"name": {
			Type:     schema.TypeString,
			Required: true,
		},

		// Optional attributes

		"zone_slug": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"mtu": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"auto_create_ipv4_subnet": {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
		},
		"subnets": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"href": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"uuid": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"cidr": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
			Computed: true,
		},

		// Computed attributes

		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func resourceNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.NetworkCreateRequest{
		Name: d.Get("name").(string),
	}

	if attr, ok := d.GetOk("zone_slug"); ok {
		opts.Zone = attr.(string)
	}
	if attr, ok := d.GetOk("mtu"); ok {
		opts.MTU = attr.(int)
	}
	if attr, ok := d.GetOkExists("auto_create_ipv4_subnet"); ok {
		val := attr.(bool)
		opts.AutoCreateIPV4Subnet = &val
	}

	log.Printf("[DEBUG] Network create configuration: %#v", opts)

	network, err := client.Networks.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating network: %s", err)
	}

	d.SetId(network.UUID)

	log.Printf("[INFO] Network ID %s", d.Id())

	err = fillNetworkResourceData(d, network)
	if err != nil {
		return err
	}
	return nil
}

func fillNetworkResourceData(d *schema.ResourceData, network *cloudscale.Network) error {
	d.Set("href", network.HREF)
	d.Set("name", network.Name)
	d.Set("mtu", network.MTU)
	d.Set("zone_slug", network.Zone.Slug)

	subnets := make([]map[string]interface{}, 0, len(network.Subnets))
	for _, subnet := range network.Subnets {
		g := make(map[string]interface{})
		g["uuid"] = subnet.UUID
		g["cidr"] = subnet.CIDR
		g["href"] = subnet.HREF
		subnets = append(subnets, g)
	}
	err := d.Set("subnets", subnets)
	if err != nil {
		log.Printf("[DEBUG] Error setting subnets attribute: %#v, error: %#v", subnets, err)
		return fmt.Errorf("Error setting subnets attribute: %#v, error: %#v", subnets, err)
	}

	return nil
}

func resourceNetworkRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	network, err := client.Networks.Get(context.Background(), d.Id())
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving network")
	}

	err = fillNetworkResourceData(d, network)
	if err != nil {
		return err
	}
	return nil
}

func resourceNetworkUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	for _, attribute := range []string{"name", "mtu"} {
		// cloudscale.ch network attributes can only be changed one at a time.
		if d.HasChange(attribute) {
			opts := &cloudscale.NetworkUpdateRequest{}
			if attribute == "name" {
				opts.Name = d.Get(attribute).(string)
			} else if attribute == "mtu" {
				opts.MTU = d.Get(attribute).(int)
			}
			err := client.Networks.Update(context.Background(), id, opts)
			if err != nil {
				return fmt.Errorf("Error updating the Network (%s) status (%s) ", id, err)
			}
		}
	}
	return resourceNetworkRead(d, meta)
}

func resourceNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	log.Printf("[INFO] Deleting Network: %s", d.Id())
	err := client.Networks.Delete(context.Background(), id)

	if err != nil {
		return CheckDeleted(d, err, "Error deleting network")
	}
	return nil
}
