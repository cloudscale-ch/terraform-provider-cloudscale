package cloudscale

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudscaleSubnet() *schema.Resource {
	return &schema.Resource{
		Create: resourceSubnetCreate,
		Read:   resourceSubnetRead,
		Update: resourceSubnetUpdate,
		Delete: resourceSubnetDelete,

		Schema: getSubnetSchema(false),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func getSubnetSchema(isDataSource bool) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
		"cidr": {
			Type:     schema.TypeString,
			Required: !isDataSource,
			Optional: isDataSource,
		},
		"network_uuid": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
			Computed: true,
		},
		"gateway_address": {
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
		},
		"dns_servers": {
			Type:     schema.TypeList,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Computed: true,
			Optional: !isDataSource,
		},
		"network_name": {
			Type:     schema.TypeString,
			Computed: true,
			Optional: isDataSource,
		},
		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"network_href": {
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
	fillResourceData(d, gatherSubnetResourceData(subnet))
	return nil
}

func gatherSubnetResourceData(subnet *cloudscale.Subnet) ResourceDataRaw {
	m := make(map[string]interface{})
	m["id"] = subnet.UUID
	m["href"] = subnet.HREF
	m["cidr"] = subnet.CIDR
	m["network_href"] = subnet.Network.HREF
	m["network_uuid"] = subnet.Network.UUID
	m["network_name"] = subnet.Network.Name
	m["gateway_address"] = subnet.GatewayAddress
	m["dns_servers"] = subnet.DNSServers
	return m
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
