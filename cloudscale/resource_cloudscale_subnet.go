package cloudscale

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const subnetHumanName = "subnet"

var (
	resourceCloudscaleSubnetRead   = getReadOperation(subnetHumanName, getGenericResourceIdentifierFromSchema, readSubnet, gatherSubnetResourceData)
	resourceCloudscaleSubnetUpdate = getUpdateOperation(subnetHumanName, getGenericResourceIdentifierFromSchema, updateSubnet, resourceCloudscaleSubnetRead, gatherSubnetUpdateRequests)
	resourceCloudscaleSubnetDelete = getDeleteOperation(subnetHumanName, deleteSubnet)
)

func resourceCloudscaleSubnet() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudscaleSubnetCreate,
		Read:   resourceCloudscaleSubnetRead,
		Update: resourceCloudscaleSubnetUpdate,
		Delete: resourceCloudscaleSubnetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: getSubnetSchema(RESOURCE),
	}
}

func getSubnetSchema(t SchemaType) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
		"cidr": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Optional: t.isDataSource(),
			ForceNew: true,
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
			Optional: t.isResource(),
		},
		"network_name": {
			Type:     schema.TypeString,
			Computed: true,
			Optional: t.isDataSource(),
		},
		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"network_href": {
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
	}
	return m
}

func resourceCloudscaleSubnetCreate(d *schema.ResourceData, meta any) error {
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

	dnsServers := d.Get("dns_servers").([]any)
	s := make([]string, len(dnsServers))
	for i := range dnsServers {
		s[i] = dnsServers[i].(string)
	}
	opts.DNSServers = s
	opts.Tags = CopyTags(d)

	log.Printf("[DEBUG] Subnet create configuration: %#v", opts)

	subnet, err := client.Subnets.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating subnet: %s", err)
	}

	d.SetId(subnet.UUID)

	log.Printf("[INFO] Subnet ID %s", d.Id())

	err = resourceCloudscaleSubnetRead(d, meta)
	if err != nil {
		return fmt.Errorf("Error reading the subnet (%s): %s", d.Id(), err)
	}

	return nil
}

func gatherSubnetResourceData(subnet *cloudscale.Subnet) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = subnet.UUID
	m["href"] = subnet.HREF
	m["cidr"] = subnet.CIDR
	m["network_href"] = subnet.Network.HREF
	m["network_uuid"] = subnet.Network.UUID
	m["network_name"] = subnet.Network.Name
	m["gateway_address"] = subnet.GatewayAddress
	m["dns_servers"] = subnet.DNSServers
	m["tags"] = subnet.Tags
	return m
}

func readSubnet(rId GenericResourceIdentifier, meta any) (*cloudscale.Subnet, error) {
	client := meta.(*cloudscale.Client)
	return client.Subnets.Get(context.Background(), rId.Id)
}

func updateSubnet(rId GenericResourceIdentifier, meta any, updateRequest *cloudscale.SubnetUpdateRequest) error {
	client := meta.(*cloudscale.Client)
	return client.Subnets.Update(context.Background(), rId.Id, updateRequest)
}

func gatherSubnetUpdateRequests(d *schema.ResourceData) []*cloudscale.SubnetUpdateRequest {
	requests := make([]*cloudscale.SubnetUpdateRequest, 0)

	for _, attribute := range []string{"gateway_address", "dns_servers", "tags"} {
		if d.HasChange(attribute) {
			log.Printf("[INFO] Attribute %s changed", attribute)
			opts := &cloudscale.SubnetUpdateRequest{}
			requests = append(requests, opts)

			if attribute == "gateway_address" {
				opts.GatewayAddress = d.Get(attribute).(string)
			} else if attribute == "dns_servers" {
				dnsServers := d.Get("dns_servers").([]any)
				s := make([]string, len(dnsServers))
				for i := range dnsServers {
					s[i] = dnsServers[i].(string)
				}
				opts.DNSServers = s
			} else if attribute == "tags" {
				opts.Tags = CopyTags(d)
			}
		}
	}
	return requests
}

func deleteSubnet(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	// sending the next request immediately can cause errors, since the port cleanup process is still ongoing
	time.Sleep(5 * time.Second)
	return client.Subnets.Delete(context.Background(), id)
}
