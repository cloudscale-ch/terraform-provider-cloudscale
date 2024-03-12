package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const networkHumanName = "network"

var (
	resourceCloudscaleNetworkRead   = getReadOperation(networkHumanName, getGenericResourceIdentifierFromSchema, readNetwork, gatherNetworkResourceData)
	resourceCloudscaleNetworkUpdate = getUpdateOperation(networkHumanName, getGenericResourceIdentifierFromSchema, updateNetwork, resourceCloudscaleNetworkRead, gatherNetworkUpdateRequest)
	resourceCloudscaleNetworkDelete = getDeleteOperation(networkHumanName, deleteNetwork)
)

func resourceCloudscaleNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudscaleNetworkCreate,
		Read:   resourceCloudscaleNetworkRead,
		Update: resourceCloudscaleNetworkUpdate,
		Delete: resourceCloudscaleNetworkDelete,

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

func resourceCloudscaleNetworkCreate(d *schema.ResourceData, meta any) error {
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
	err = resourceCloudscaleNetworkRead(d, meta)
	if err != nil {
		return fmt.Errorf("Error reading the network (%s): %s", d.Id(), err)
	}
	return nil
}

func gatherNetworkResourceData(network *cloudscale.Network) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = network.UUID
	m["href"] = network.HREF
	m["name"] = network.Name
	m["mtu"] = network.MTU
	m["zone_slug"] = network.Zone.Slug

	subnets := make([]map[string]any, 0, len(network.Subnets))
	for _, subnet := range network.Subnets {
		g := make(map[string]any)
		g["uuid"] = subnet.UUID
		g["cidr"] = subnet.CIDR
		g["href"] = subnet.HREF
		subnets = append(subnets, g)
	}
	m["subnets"] = subnets
	m["tags"] = network.Tags
	return m
}

func readNetwork(rId GenericResourceIdentifier, meta any) (*cloudscale.Network, error) {
	client := meta.(*cloudscale.Client)
	return client.Networks.Get(context.Background(), rId.Id)
}

func updateNetwork(rId GenericResourceIdentifier, meta any, updateRequest *cloudscale.NetworkUpdateRequest) error {
	client := meta.(*cloudscale.Client)
	return client.Networks.Update(context.Background(), rId.Id, updateRequest)
}

func gatherNetworkUpdateRequest(d *schema.ResourceData) []*cloudscale.NetworkUpdateRequest {
	requests := make([]*cloudscale.NetworkUpdateRequest, 0)

	for _, attribute := range []string{"name", "mtu", "tags"} {
		if d.HasChange(attribute) {
			log.Printf("[INFO] Attribute %s changed", attribute)
			opts := &cloudscale.NetworkUpdateRequest{}
			requests = append(requests, opts)

			if attribute == "name" {
				opts.Name = d.Get(attribute).(string)
			} else if attribute == "mtu" {
				opts.MTU = d.Get(attribute).(int)
			} else if attribute == "tags" {
				opts.Tags = CopyTags(d)
			}
		}
	}
	return requests
}

func deleteNetwork(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	return client.Networks.Delete(context.Background(), id)
}
