package cloudscale

import (
	"context"
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strings"
)

const poolMemberHumanName = "load balancer pool member"

var (
	resourceCloudscaleLoadBalancerPoolMemberRead   = getReadOperation(poolMemberHumanName, getLoadBalancerResourceIdentifierFromSchema, readLoadBalancerPoolMember, gatherLoadBalancerPoolMemberResourceData)
	resourceCloudscaleLoadBalancerPoolMemberUpdate = getUpdateOperation(poolMemberHumanName, getLoadBalancerResourceIdentifierFromSchema, updateLoadBalancerPoolMember, resourceCloudscaleLoadBalancerPoolMemberRead, gatherLoadBalancerPoolMemberUpdateRequest)
	resourceCloudscaleLoadBalancerPoolMemberDelete = getDeleteOperation(poolMemberHumanName, deleteLoadBalancerPoolMember)
)

func resourceCloudscaleLoadBalancerPoolMembers() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudscaleLoadBalancerPoolMemberCreate,
		Read:   resourceCloudscaleLoadBalancerPoolMemberRead,
		Update: resourceCloudscaleLoadBalancerPoolMemberUpdate,
		Delete: resourceCloudscaleLoadBalancerPoolMemberDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(
				ctx context.Context,
				d *schema.ResourceData,
				m any,
			) ([]*schema.ResourceData, error) {
				poolID, id, err := splitImportID(d.Id())
				if err != nil {
					return nil, err
				}
				err = d.Set("pool_uuid", poolID)
				if err != nil {
					return nil, err
				}
				d.SetId(id)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: getLoadBalancerPoolMemberSchema(RESOURCE),
	}
}

type LoadBalancerPoolMemberResourceIdentifier struct {
	Id     string
	PoolID string
}

func getLoadBalancerResourceIdentifierFromSchema(d *schema.ResourceData) LoadBalancerPoolMemberResourceIdentifier {
	return LoadBalancerPoolMemberResourceIdentifier{
		Id:     d.Id(),
		PoolID: d.Get("pool_uuid").(string),
	}
}

func splitImportID(id string) (string, string, error) {
	parts := strings.Split(id, ".")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid import id %q. Expecting {pool_uuid}.{member_uuid}", id)
	}
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return "", "", fmt.Errorf("invalid import id %q. Could not parse {pool_uuid}.{member_uuid}", id)
	}
	return parts[0], parts[1], nil
}

func getLoadBalancerPoolMemberSchema(t SchemaType) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Optional: t.isDataSource(),
		},
		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"enabled": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"pool_uuid": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"pool_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"pool_href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"protocol_port": {
			Type:     schema.TypeInt,
			Required: t.isResource(),
			Computed: t.isDataSource(),
		},
		"monitor_port": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: t.isDataSource(),
		},
		"address": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Computed: t.isDataSource(),
			ForceNew: true,
		},
		"status": {
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

func resourceCloudscaleLoadBalancerPoolMemberCreate(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.LoadBalancerPoolMemberRequest{
		Name:         d.Get("name").(string),
		ProtocolPort: d.Get("protocol_port").(int),
		MonitorPort:  d.Get("monitor_port").(int),
		Address:      d.Get("address").(string),
	}
	if attr, ok := d.GetOkExists("enabled"); ok {
		val := attr.(bool)
		opts.Enabled = &val
	}
	opts.Tags = CopyTags(d)

	log.Printf("[DEBUG] LoadBalancerPoolMember create configuration: %#v", opts)

	poolID := d.Get("pool_uuid").(string)
	poolMember, err := client.LoadBalancerPoolMembers.Create(context.Background(), poolID, opts)
	if err != nil {
		return fmt.Errorf("Error creating LoadBalancerPoolMember: %s", err)
	}

	d.SetId(poolMember.UUID)

	log.Printf("[INFO] LoadBalancerPoolMember ID: %s", d.Id())
	err = resourceCloudscaleLoadBalancerPoolMemberRead(d, meta)
	if err != nil {
		return fmt.Errorf("Error reading the load balancer pool member (%s): %s", d.Id(), err)
	}
	return nil
}

func gatherLoadBalancerPoolMemberResourceData(loadbalancerPoolMember *cloudscale.LoadBalancerPoolMember) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = loadbalancerPoolMember.UUID
	m["href"] = loadbalancerPoolMember.HREF
	m["name"] = loadbalancerPoolMember.Name
	m["enabled"] = loadbalancerPoolMember.Enabled
	m["pool_uuid"] = loadbalancerPoolMember.Pool.UUID
	m["pool_name"] = loadbalancerPoolMember.Pool.Name
	m["pool_href"] = loadbalancerPoolMember.Pool.HREF
	m["protocol_port"] = loadbalancerPoolMember.ProtocolPort
	m["monitor_port"] = loadbalancerPoolMember.MonitorPort
	m["address"] = loadbalancerPoolMember.Address
	m["status"] = loadbalancerPoolMember.Status
	m["tags"] = loadbalancerPoolMember.Tags
	return m
}

func readLoadBalancerPoolMember(rId LoadBalancerPoolMemberResourceIdentifier, meta any) (*cloudscale.LoadBalancerPoolMember, error) {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancerPoolMembers.Get(context.Background(), rId.PoolID, rId.Id)
}

func updateLoadBalancerPoolMember(rId LoadBalancerPoolMemberResourceIdentifier, meta any, updateRequest *cloudscale.LoadBalancerPoolMemberRequest) error {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancerPoolMembers.Update(context.Background(), rId.PoolID, rId.Id, updateRequest)
}

func gatherLoadBalancerPoolMemberUpdateRequest(d *schema.ResourceData) []*cloudscale.LoadBalancerPoolMemberRequest {
	requests := make([]*cloudscale.LoadBalancerPoolMemberRequest, 0)

	for _, attribute := range []string{"name", "enabled", "protocol_port", "monitor_port", "tags"} {
		if d.HasChange(attribute) {
			log.Printf("[INFO] Attribute %s changed", attribute)
			opts := &cloudscale.LoadBalancerPoolMemberRequest{}
			requests = append(requests, opts)

			if attribute == "name" {
				opts.Name = d.Get(attribute).(string)
			} else if attribute == "enabled" {
				v := d.Get(attribute).(bool)
				opts.Enabled = &v
			} else if attribute == "protocol_port" {
				opts.ProtocolPort = d.Get(attribute).(int)
			} else if attribute == "monitor_port" {
				opts.MonitorPort = d.Get(attribute).(int)
			} else if attribute == "tags" {
				opts.Tags = CopyTags(d)
			}
		}
	}
	return requests
}

func deleteLoadBalancerPoolMember(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	poolID := d.Get("pool_uuid").(string)
	return client.LoadBalancerPoolMembers.Delete(context.Background(), poolID, id)
}
