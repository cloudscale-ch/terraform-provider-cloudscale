package cloudscale

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceCloudScaleSubnet() *schema.Resource {
	return &schema.Resource{
		Create: resourceSubnetCreate,
		Read:   resourceSubnetRead,
		Update: resourceSubnetUpdate,
		Delete: resourceSubnetDelete,

		Schema: getSubnetSchema(),
	}
}

func getSubnetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// Required attributes

		"cidr": {
			Type:     schema.TypeString,
			Required: true,
		},
		"network_uuid": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},

		// Optional attributes
		"gateway_address": {
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
		},
		"dns_servers": {
			Type:     schema.TypeList,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Computed: true,
			Optional: true,
		},

		// Computed attributes

		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"network_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"network_href": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func resourceSubnetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.SubnetCreateRequest{
		CIDR: d.Get("cidr").(string),
	}

	if attr, ok := d.GetOk("network_uuid"); ok {
		opts.Network = attr.(string)
	}
	if attr, ok := d.GetOk("gateway_address"); ok {
		opts.GatewayAddress = attr.(string)
	}

	dnsServers := d.Get("dns_servers").([]interface{})
	s := make([]string, len(dnsServers))
	for i := range dnsServers {
		s[i] = dnsServers[i].(string)
	}
	opts.DNSServers = s

	log.Printf("[DEBUG] Subnet create configuration: %#v", opts)

	subnet, err := client.Subnets.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating subnet: %s", err)
	}

	d.SetId(subnet.UUID)

	log.Printf("[INFO] Subnet ID %s", d.Id())

	err = fillSubnetResourceData(d, subnet)
	if err != nil {
		return err
	}
	return nil
}

func fillSubnetResourceData(d *schema.ResourceData, subnet *cloudscale.Subnet) error {
	d.Set("href", subnet.HREF)
	d.Set("cidr", subnet.CIDR)
	d.Set("network_href", subnet.Network.HREF)
	d.Set("network_uuid", subnet.Network.UUID)
	d.Set("network_name", subnet.Network.Name)
	d.Set("gateway_address", subnet.GatewayAddress)
	d.Set("dns_servers", subnet.DNSServers)
	return nil
}

func resourceSubnetRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	subnet, err := client.Subnets.Get(context.Background(), d.Id())
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving subnet")
	}

	err = fillSubnetResourceData(d, subnet)
	if err != nil {
		return err
	}
	return nil
}

func resourceSubnetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	for _, attribute := range []string{"gateway_address", "dns_servers"} {
		// cloudscale.ch subnet attributes can only be changed one at a time.
		if d.HasChange(attribute) {
			opts := &cloudscale.SubnetUpdateRequest{}
			if attribute == "gateway_address" {
				opts.GatewayAddress = d.Get(attribute).(string)
			} else if attribute == "dns_servers" {
				dnsServers := d.Get("dns_servers").([]interface{})
				s := make([]string, len(dnsServers))

				for i := range dnsServers {
					s[i] = dnsServers[i].(string)
				}
				opts.DNSServers = s
			}
			err := client.Subnets.Update(context.Background(), id, opts)
			if err != nil {
				return fmt.Errorf("Error updating the Subnet (%s) status (%s) ", id, err)
			}
		}
	}
	return resourceSubnetRead(d, meta)
}

func resourceSubnetDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	log.Printf("[INFO] Deleting Subnet: %s", d.Id())
	// sending the next request immediately can cause errors, since the port cleanup process is still ongoing
	time.Sleep(5 * time.Second)
	err := client.Subnets.Delete(context.Background(), id)

	if err != nil {
		return CheckDeleted(d, err, "Error deleting subnet")
	}
	return nil
}
