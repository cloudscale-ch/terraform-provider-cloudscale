package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const floatingIPHumanName = "Floating IP"

var (
	resourceFloatingIPRead   = getReadOperation(floatingIPHumanName, getGenericResourceIdentifierFromSchema, readFloatingIP, gatherFloatingIPResourceData)
	resourceFloatingIPUpdate = getUpdateOperation(floatingIPHumanName, getGenericResourceIdentifierFromSchema, updateFloatingIP, resourceFloatingIPRead, gatherFloatingIPUpdateRequest)
	resourceFloatingIPDelete = getDeleteOperation(floatingIPHumanName, deleteFloatingIP)
)

func resourceCloudscaleFloatingIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceFloatingIPCreate,
		Read:   resourceFloatingIPRead,
		Update: resourceFloatingIPUpdate,
		Delete: resourceFloatingIPDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: getFloatingIPSchema(RESOURCE),
	}
}

func getFloatingIPSchema(t SchemaType) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
		"ip_version": {
			Type:     schema.TypeInt,
			Required: t.isResource(),
			Optional: t.isDataSource(),
			ForceNew: true,
		},
		"server": {
			Type:     schema.TypeString,
			Optional: t.isResource(),
			Computed: t.isDataSource(),
		},
		"load_balancer": {
			Type:     schema.TypeString,
			Optional: t.isResource(),
			Computed: t.isDataSource(),
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
		},
		"prefix_length": {
			Type:     schema.TypeInt,
			ForceNew: true,
			Optional: t.isResource(),
			Computed: true,
		},
		"network": {
			Type:     schema.TypeString,
			Optional: t.isDataSource(),
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

func resourceFloatingIPCreate(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.FloatingIPCreateRequest{
		IPVersion: d.Get("ip_version").(int),
	}

	if attr, ok := d.GetOk("server"); ok {
		opts.Server = attr.(string)
	}
	if attr, ok := d.GetOk("load_balancer"); ok {
		opts.LoadBalancer = attr.(string)
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
	opts.Tags = CopyTags(d)

	log.Printf("[DEBUG] FloatingIP create configuration: %#v", opts)

	floatingIP, err := client.FloatingIPs.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating FloatingIP: %s", err)
	}

	d.SetId(floatingIP.IP())

	err = resourceFloatingIPRead(d, meta)
	if err != nil {
		return fmt.Errorf("Error reading the floating IP (%s): %s", d.Id(), err)
	}
	return nil
}

func gatherFloatingIPResourceData(floatingIP *cloudscale.FloatingIP) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = floatingIP.IP()
	m["href"] = floatingIP.HREF
	m["ip_version"] = floatingIP.IPVersion
	m["prefix_length"] = floatingIP.PrefixLength()
	m["network"] = floatingIP.Network
	m["next_hop"] = floatingIP.NextHop
	m["reverse_ptr"] = floatingIP.ReversePointer
	m["type"] = floatingIP.Type
	m["tags"] = floatingIP.Tags
	if floatingIP.Server != nil {
		m["server"] = floatingIP.Server.UUID
	} else {
		m["server"] = nil
	}
	if floatingIP.LoadBalancer != nil {
		m["load_balancer"] = floatingIP.LoadBalancer.UUID
	} else {
		m["load_balancer"] = nil
	}
	if floatingIP.Region != nil {
		m["region_slug"] = floatingIP.Region.Slug
	}

	return m
}

func readFloatingIP(rId GenericResourceIdentifier, meta any) (*cloudscale.FloatingIP, error) {
	client := meta.(*cloudscale.Client)
	return client.FloatingIPs.Get(context.Background(), rId.Id)

}

func updateFloatingIP(rId GenericResourceIdentifier, meta any, updateRequest *cloudscale.FloatingIPUpdateRequest) error {
	client := meta.(*cloudscale.Client)
	return client.FloatingIPs.Update(context.Background(), rId.Id, updateRequest)
}

func gatherFloatingIPUpdateRequest(d *schema.ResourceData) []*cloudscale.FloatingIPUpdateRequest {
	requests := make([]*cloudscale.FloatingIPUpdateRequest, 0)

	for _, attribute := range []string{"server", "load_balancer", "tags", "reverse_ptr"} {
		if d.HasChange(attribute) {
			log.Printf("[INFO] Attribute %s changed", attribute)
			opts := &cloudscale.FloatingIPUpdateRequest{}
			requests = append(requests, opts)

			if attribute == "reverse_ptr" {
				opts.ReversePointer = d.Get(attribute).(string)
			} else if attribute == "server" || attribute == "load_balancer" {
				serverUUID := d.Get("server").(string)
				if serverUUID != "" {
					log.Printf("[INFO] Assigning the Floating IP %s to the Server %s", d.Id(), serverUUID)
					opts.Server = serverUUID
				}
				loadBalancerUUID := d.Get("load_balancer").(string)
				if loadBalancerUUID != "" {
					log.Printf("[INFO] Assigning the Floating IP %s to the LB %s", d.Id(), loadBalancerUUID)
					opts.LoadBalancer = loadBalancerUUID
				}
			} else if attribute == "tags" {
				opts.Tags = CopyTags(d)
			}
		}
	}
	return requests
}

func deleteFloatingIP(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	return client.FloatingIPs.Delete(context.Background(), id)
}
