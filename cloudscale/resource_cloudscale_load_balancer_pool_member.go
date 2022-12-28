package cloudscale

import (
	"context"
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strings"
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
	return m
}

func resourceCloudscaleLoadBalancerPoolMemberCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.LoadBalancerPoolMemberRequest{
		Name:         d.Get("name").(string),
		ProtocolPort: d.Get("protocol_port").(int),
		MonitorPort:  d.Get("monitor_port").(int),
		Address:      d.Get("address").(string),
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

func fillLoadBalancerPoolMemberSchema(d *schema.ResourceData, loadbalancerpoolMember *cloudscale.LoadBalancerPoolMember) {
	fillResourceData(d, gatherLoadBalancerPoolMemberResourceData(loadbalancerpoolMember))
}

func gatherLoadBalancerPoolMemberResourceData(loadbalancerpoolMember *cloudscale.LoadBalancerPoolMember) ResourceDataRaw {
	m := make(map[string]interface{})
	m["id"] = loadbalancerpoolMember.UUID
	m["href"] = loadbalancerpoolMember.HREF
	m["name"] = loadbalancerpoolMember.Name
	m["pool_uuid"] = loadbalancerpoolMember.Pool.UUID
	m["pool_name"] = loadbalancerpoolMember.Pool.Name
	m["pool_href"] = loadbalancerpoolMember.Pool.HREF
	m["protocol_port"] = loadbalancerpoolMember.ProtocolPort
	m["monitor_port"] = loadbalancerpoolMember.MonitorPort
	m["address"] = loadbalancerpoolMember.Address
	m["status"] = loadbalancerpoolMember.Status
	m["tags"] = loadbalancerpoolMember.Tags
	return m
}

func resourceCloudscaleLoadBalancerPoolMemberRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	poolID := d.Get("pool_uuid").(string)
	loadbalancerPoolMember, err := client.LoadBalancerPoolMembers.Get(context.Background(), poolID, d.Id())
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving load balancer pool")
	}

	fillLoadBalancerPoolMemberSchema(d, loadbalancerPoolMember)
	return nil
}

func resourceCloudscaleLoadBalancerPoolMemberUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	poolID := d.Get("pool_uuid").(string)

	for _, attribute := range []string{"name", "protocol_port", "monitor_port", "tags"} {
		if d.HasChange(attribute) {
			opts := &cloudscale.LoadBalancerPoolMemberRequest{}
			if attribute == "name" {
				opts.Name = d.Get(attribute).(string)
			} else if attribute == "protocol_port" {
				opts.ProtocolPort = d.Get(attribute).(int)
			} else if attribute == "monitor_port" {
				opts.MonitorPort = d.Get(attribute).(int)
			} else if attribute == "tags" {
				opts.Tags = CopyTags(d)
			}
			err := client.LoadBalancerPoolMembers.Update(context.Background(), poolID, id, opts)
			if err != nil {
				return fmt.Errorf("Error updating the Load Balancer Pool Member (%s): %s", id, err)
			}
		}
	}
	return resourceCloudscaleLoadBalancerPoolMemberRead(d, meta)
}

func resourceCloudscaleLoadBalancerPoolMemberDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	poolID := d.Get("pool_uuid").(string)

	log.Printf("[NFO] Deleting LoadBalancerPoolMember: %s", id)
	err := client.LoadBalancerPoolMembers.Delete(context.Background(), poolID, id)

	if err != nil {
		return CheckDeleted(d, err, "Error deleting load balancer pool")
	}
	return nil
}
