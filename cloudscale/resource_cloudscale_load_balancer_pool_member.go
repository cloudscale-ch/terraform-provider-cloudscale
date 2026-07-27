package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v10"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const poolMemberHumanName = "load balancer pool member"

// Pool member operations serialize on the load balancer that owns the parent pool
// via lockKeyFromPoolUUID.
var (
	resourceCloudscaleLoadBalancerPoolMemberCreate = getCreateOperation(createLoadBalancerPoolMember, lockKeyFromPoolUUID)
	resourceCloudscaleLoadBalancerPoolMemberRead   = getReadOperation(poolMemberHumanName, getLoadBalancerResourceIdentifierFromSchema, readLoadBalancerPoolMember, gatherLoadBalancerPoolMemberResourceData)
	resourceCloudscaleLoadBalancerPoolMemberUpdate = getUpdateOperation(poolMemberHumanName, getLoadBalancerResourceIdentifierFromSchema, updateLoadBalancerPoolMember, resourceCloudscaleLoadBalancerPoolMemberRead, gatherLoadBalancerPoolMemberUpdateRequest, lockKeyFromPoolUUID)
	resourceCloudscaleLoadBalancerPoolMemberDelete = getDeleteOperation(poolMemberHumanName, getLoadBalancerResourceIdentifierFromSchema, deleteLoadBalancerPoolMember, lockKeyFromPoolUUID)
)

func resourceCloudscaleLoadBalancerPoolMembers() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudscaleLoadBalancerPoolMemberCreate,
		ReadContext:   resourceCloudscaleLoadBalancerPoolMemberRead,
		UpdateContext: resourceCloudscaleLoadBalancerPoolMemberUpdate,
		DeleteContext: resourceCloudscaleLoadBalancerPoolMemberDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(
				ctx context.Context,
				d *schema.ResourceData,
				m any,
			) ([]*schema.ResourceData, error) {
				poolID, id, err := splitImportID(d.Id(), "pool_uuid", "member_uuid")
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
			ForceNew: true,
		},
		"monitor_port": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: t.isDataSource(),
			ForceNew: true,
		},
		"address": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Computed: t.isDataSource(),
			ForceNew: true,
		},
		"subnet_uuid": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Computed: t.isDataSource(),
			ForceNew: true,
		},
		"subnet_cidr": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"subnet_href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"monitor_status": {
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

func createLoadBalancerPoolMember(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.LoadBalancerPoolMemberRequest{
		Name:         d.Get("name").(string),
		ProtocolPort: d.Get("protocol_port").(int),
		MonitorPort:  d.Get("monitor_port").(int),
		Address:      d.Get("address").(string),
		Subnet:       d.Get("subnet_uuid").(string),
	}
	if attr, ok := d.GetOkExists("enabled"); ok {
		val := attr.(bool)
		opts.Enabled = &val
	}
	opts.Tags = TagsFromState(d)

	log.Printf("[DEBUG] LoadBalancerPoolMember create configuration: %#v", opts)

	poolID := d.Get("pool_uuid").(string)
	poolMember, err := client.LoadBalancerPoolMembers.Create(ctx, poolID, opts)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error creating LoadBalancerPoolMember: %s", err))
	}

	d.SetId(poolMember.UUID)

	log.Printf("[INFO] LoadBalancerPoolMember ID: %s", d.Id())
	return resourceCloudscaleLoadBalancerPoolMemberRead(ctx, d, meta)
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
	m["subnet_uuid"] = loadbalancerPoolMember.Subnet.UUID
	m["subnet_cidr"] = loadbalancerPoolMember.Subnet.CIDR
	m["subnet_href"] = loadbalancerPoolMember.Subnet.HREF
	m["protocol_port"] = loadbalancerPoolMember.ProtocolPort
	m["monitor_port"] = loadbalancerPoolMember.MonitorPort
	m["address"] = loadbalancerPoolMember.Address
	m["monitor_status"] = loadbalancerPoolMember.MonitorStatus
	m["tags"] = TagsToState(loadbalancerPoolMember.Tags)
	return m
}

func readLoadBalancerPoolMember(ctx context.Context, rId LoadBalancerPoolMemberResourceIdentifier, meta any) (*cloudscale.LoadBalancerPoolMember, error) {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancerPoolMembers.Get(ctx, rId.PoolID, rId.Id)
}

func updateLoadBalancerPoolMember(ctx context.Context, rId LoadBalancerPoolMemberResourceIdentifier, meta any, updateRequest *cloudscale.LoadBalancerPoolMemberRequest) error {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancerPoolMembers.Update(ctx, rId.PoolID, rId.Id, updateRequest)
}

func gatherLoadBalancerPoolMemberUpdateRequest(d *schema.ResourceData) []*cloudscale.LoadBalancerPoolMemberRequest {
	requests := make([]*cloudscale.LoadBalancerPoolMemberRequest, 0)

	for _, attribute := range []string{"name", "enabled", "tags"} {
		if d.HasChange(attribute) {
			log.Printf("[INFO] Attribute %s changed", attribute)
			opts := &cloudscale.LoadBalancerPoolMemberRequest{}
			requests = append(requests, opts)

			if attribute == "name" {
				opts.Name = d.Get(attribute).(string)
			} else if attribute == "enabled" {
				v := d.Get(attribute).(bool)
				opts.Enabled = &v
			} else if attribute == "tags" {
				opts.Tags = TagsFromState(d)
			}
		}
	}
	return requests
}

func deleteLoadBalancerPoolMember(ctx context.Context, rId LoadBalancerPoolMemberResourceIdentifier, meta any) error {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancerPoolMembers.Delete(ctx, rId.PoolID, rId.Id)
}
