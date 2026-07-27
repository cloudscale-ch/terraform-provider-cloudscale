package cloudscale

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v10"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const subnetHumanName = "subnet"

var (
	resourceCloudscaleSubnetCreate = getCreateOperation(createSubnet, nil)
	resourceCloudscaleSubnetRead   = getReadOperation(subnetHumanName, getGenericResourceIdentifierFromSchema, readSubnet, gatherSubnetResourceData)
	resourceCloudscaleSubnetUpdate = getUpdateOperation(subnetHumanName, getGenericResourceIdentifierFromSchema, updateSubnet, resourceCloudscaleSubnetRead, gatherSubnetUpdateRequests, nil)
	resourceCloudscaleSubnetDelete = getDeleteOperation(subnetHumanName, getGenericResourceIdentifierFromSchema, deleteSubnet, nil)
)

func resourceCloudscaleSubnet() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudscaleSubnetCreate,
		ReadContext:   resourceCloudscaleSubnetRead,
		UpdateContext: resourceCloudscaleSubnetUpdate,
		DeleteContext: resourceCloudscaleSubnetDelete,

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
	} else {
		m["disable_dns_servers"] = &schema.Schema{
			Type:          schema.TypeBool,
			Optional:      true,
			ConflictsWith: []string{"dns_servers"},
		}
	}
	return m
}

func createSubnet(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

	disableDnsServers := d.Get("disable_dns_servers").(bool)
	if disableDnsServers {
		opts.DNSServers = &[]string{}
	} else {
		if dnsServersRaw, ok := d.GetOk("dns_servers"); ok {
			dnsServers := dnsServersRaw.([]interface{})
			dnsServersStr := make([]string, len(dnsServers))
			for i := range dnsServers {
				dnsServersStr[i] = dnsServers[i].(string)
			}
			opts.DNSServers = &dnsServersStr
		} else {
			opts.DNSServers = &cloudscale.UseCloudscaleDefaults
		}
	}

	opts.Tags = TagsFromState(d)

	log.Printf("[DEBUG] Subnet create configuration: %#v", opts)

	subnet, err := client.Subnets.Create(ctx, opts)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error creating subnet: %s", err))
	}

	d.SetId(subnet.UUID)

	log.Printf("[INFO] Subnet ID %s", d.Id())

	return resourceCloudscaleSubnetRead(ctx, d, meta)
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
	m["tags"] = TagsToState(subnet.Tags)
	return m
}

func readSubnet(ctx context.Context, rId GenericResourceIdentifier, meta any) (*cloudscale.Subnet, error) {
	client := meta.(*cloudscale.Client)
	return client.Subnets.Get(ctx, rId.Id)
}

func updateSubnet(ctx context.Context, rId GenericResourceIdentifier, meta any, updateRequest *cloudscale.SubnetUpdateRequest) error {
	client := meta.(*cloudscale.Client)
	return client.Subnets.Update(ctx, rId.Id, updateRequest)
}

func gatherSubnetUpdateRequests(d *schema.ResourceData) []*cloudscale.SubnetUpdateRequest {
	requests := make([]*cloudscale.SubnetUpdateRequest, 0)

	for _, attribute := range []string{"gateway_address", "dns_servers", "tags", "disable_dns_servers"} {
		if d.HasChange(attribute) {
			log.Printf("[INFO] Attribute %s changed", attribute)
			opts := &cloudscale.SubnetUpdateRequest{}
			requests = append(requests, opts)

			if attribute == "gateway_address" {
				opts.GatewayAddress = d.Get(attribute).(string)
			} else if attribute == "dns_servers" || attribute == "disable_dns_servers" {
				disableDnsServers := d.Get("disable_dns_servers").(bool)
				if disableDnsServers {
					opts.DNSServers = &[]string{}
				} else {
					if dnsServersRaw, ok := d.GetOk("dns_servers"); ok {
						dnsServers := dnsServersRaw.([]interface{})
						dnsServersStr := make([]string, len(dnsServers))
						for i := range dnsServers {
							dnsServersStr[i] = dnsServers[i].(string)
						}
						opts.DNSServers = &dnsServersStr
					} else {
						opts.DNSServers = &cloudscale.UseCloudscaleDefaults
					}
				}
			} else if attribute == "tags" {
				opts.Tags = TagsFromState(d)
			}
		}
	}
	return requests
}

func deleteSubnet(ctx context.Context, rId GenericResourceIdentifier, meta any) error {
	client := meta.(*cloudscale.Client)
	// sending the next request immediately can cause errors, since the port cleanup process is still ongoing
	time.Sleep(5 * time.Second)
	return client.Subnets.Delete(ctx, rId.Id)
}
