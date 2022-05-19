package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudscaleNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkCreate,
		Read:   resourceNetworkRead,
		Update: resourceNetworkUpdate,
		Delete: resourceNetworkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: getNetworkSchema(RESOURCE),
	}
}

func getNetworkSchema(t SchemaType) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Optional: t.isDataSource(),
			Computed: t.isDataSource(),
		},
		"zone_slug": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"mtu": {
			Type:     schema.TypeInt,
			Optional: t.isResource(),
			Computed: true,
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
	} else {
		m["auto_create_ipv4_subnet"] = &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
		}
	}
	return m
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
	opts.Tags = CopyTags(d)

	log.Printf("[DEBUG] Network create configuration: %#v", opts)

	network, err := client.Networks.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating network: %s", err)
	}

	d.SetId(network.UUID)

	log.Printf("[INFO] Network ID %s", d.Id())

	fillNetworkResourceData(d, network)
	return nil
}

func fillNetworkResourceData(d *schema.ResourceData, network *cloudscale.Network) {
	fillResourceData(d, gatherNetworkResourceData(network))
}

func gatherNetworkResourceData(network *cloudscale.Network) ResourceDataRaw {
	m := make(map[string]interface{})
	m["id"] = network.UUID
	m["href"] = network.HREF
	m["name"] = network.Name
	m["mtu"] = network.MTU
	m["zone_slug"] = network.Zone.Slug

	subnets := make([]map[string]interface{}, 0, len(network.Subnets))
	for _, subnet := range network.Subnets {
		g := make(map[string]interface{})
		g["uuid"] = subnet.UUID
		g["cidr"] = subnet.CIDR
		g["href"] = subnet.HREF
		subnets = append(subnets, g)
	}
	m["subnets"] = subnets
	m["tags"] = network.Tags
	return m
}

func resourceNetworkRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	network, err := client.Networks.Get(context.Background(), d.Id())
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving network")
	}

	fillNetworkResourceData(d, network)
	return nil
}

func resourceNetworkUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	for _, attribute := range []string{"name", "mtu", "tags"} {
		// cloudscale.ch network attributes can only be changed one at a time.
		if d.HasChange(attribute) {
			opts := &cloudscale.NetworkUpdateRequest{}
			if attribute == "name" {
				opts.Name = d.Get(attribute).(string)
			} else if attribute == "mtu" {
				opts.MTU = d.Get(attribute).(int)
			} else if attribute == "tags" {
				opts.Tags = CopyTags(d)
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
