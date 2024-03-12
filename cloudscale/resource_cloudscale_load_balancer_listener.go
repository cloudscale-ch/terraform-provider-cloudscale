package cloudscale

import (
	"context"
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

const listenerHumanName = "load balancer listener"

var (
	resourceCloudscaleLoadBalancerListenerRead   = getReadOperation(listenerHumanName, getGenericResourceIdentifierFromSchema, readLoadBalancerListener, gatherLoadBalancerListenerResourceData)
	resourceCloudscaleLoadBalancerListenerUpdate = getUpdateOperation(listenerHumanName, getGenericResourceIdentifierFromSchema, updateLoadBalancerListener, resourceCloudscaleLoadBalancerListenerRead, gatherLoadBalancerListenerUpdateRequest)
	resourceCloudscaleLoadBalancerListenerDelete = getDeleteOperation(listenerHumanName, deleteLoadBalancerListener)
)

func resourceCloudscaleLoadBalancerListener() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudscaleLoadBalancerListenerCreate,
		Read:   resourceCloudscaleLoadBalancerListenerRead,
		Update: resourceCloudscaleLoadBalancerListenerUpdate,
		Delete: resourceCloudscaleLoadBalancerListenerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: getLoadBalancerListenerSchema(RESOURCE),
	}
}

func getLoadBalancerListenerSchema(t SchemaType) map[string]*schema.Schema {
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
			Optional: true,
			Computed: true,
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
		"protocol": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Computed: t.isDataSource(),
		},
		"protocol_port": {
			Type:     schema.TypeInt,
			Required: t.isResource(),
			Computed: t.isDataSource(),
		},
		"timeout_client_data_ms": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"timeout_member_connect_ms": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"timeout_member_data_ms": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"allowed_cidrs": {
			Type:     schema.TypeList,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Optional: true,
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

func resourceCloudscaleLoadBalancerListenerCreate(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.LoadBalancerListenerRequest{
		Name:         d.Get("name").(string),
		Protocol:     d.Get("protocol").(string),
		ProtocolPort: d.Get("protocol_port").(int),
	}

	if attr, ok := d.GetOk("pool_uuid"); ok {
		opts.Pool = attr.(string)
	}

	if attr, ok := d.GetOk("timeout_client_data_ms"); ok {
		opts.TimeoutClientDataMS = attr.(int)
	}
	if attr, ok := d.GetOk("timeout_member_connect_ms"); ok {
		opts.TimeoutMemberConnectMS = attr.(int)
	}
	if attr, ok := d.GetOk("timeout_member_data_ms"); ok {
		opts.TimeoutMemberDataMS = attr.(int)
	}

	allowedCIDRs := d.Get("allowed_cidrs").([]any)
	s := make([]string, len(allowedCIDRs))
	for i := range allowedCIDRs {
		s[i] = allowedCIDRs[i].(string)
	}
	opts.AllowedCIDRs = s

	opts.Tags = CopyTags(d)

	log.Printf("[DEBUG] LoadBalancerListener create configuration: %#v", opts)

	loadBalancerListener, err := client.LoadBalancerListeners.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating LoadBalancerListener: %s", err)
	}

	d.SetId(loadBalancerListener.UUID)

	log.Printf("[INFO] LoadBalancerListener ID: %s", d.Id())
	err = resourceCloudscaleLoadBalancerListenerRead(d, meta)
	if err != nil {
		return fmt.Errorf("Error reading the load balancer listener (%s): %s", d.Id(), err)
	}
	return nil
}

func gatherLoadBalancerListenerResourceData(loadbalancerlistener *cloudscale.LoadBalancerListener) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = loadbalancerlistener.UUID
	m["href"] = loadbalancerlistener.HREF
	m["name"] = loadbalancerlistener.Name
	if loadbalancerlistener.Pool != nil {
		m["pool_uuid"] = loadbalancerlistener.Pool.UUID
		m["pool_name"] = loadbalancerlistener.Pool.Name
		m["pool_href"] = loadbalancerlistener.Pool.HREF
	} else {
		m["pool_uuid"] = nil
		m["pool_name"] = nil
		m["pool_href"] = nil
	}
	m["protocol"] = loadbalancerlistener.Protocol
	m["protocol_port"] = loadbalancerlistener.ProtocolPort
	m["timeout_client_data_ms"] = loadbalancerlistener.TimeoutClientDataMS
	m["timeout_member_connect_ms"] = loadbalancerlistener.TimeoutMemberConnectMS
	m["timeout_member_data_ms"] = loadbalancerlistener.TimeoutMemberDataMS
	m["allowed_cidrs"] = loadbalancerlistener.AllowedCIDRs
	m["tags"] = loadbalancerlistener.Tags
	return m
}

func readLoadBalancerListener(rId GenericResourceIdentifier, meta any) (*cloudscale.LoadBalancerListener, error) {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancerListeners.Get(context.Background(), rId.Id)
}

func updateLoadBalancerListener(rId GenericResourceIdentifier, meta any, updateRequest *cloudscale.LoadBalancerListenerRequest) error {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancerListeners.Update(context.Background(), rId.Id, updateRequest)
}

func gatherLoadBalancerListenerUpdateRequest(d *schema.ResourceData) []*cloudscale.LoadBalancerListenerRequest {
	requests := make([]*cloudscale.LoadBalancerListenerRequest, 0)

	for _, attribute := range []string{
		"name", "protocol", "protocol_port",
		"timeout_client_data_ms", "timeout_member_connect_ms", "timeout_member_data_ms",
		"allowed_cidrs",
		"tags",
	} {
		if d.HasChange(attribute) {
			log.Printf("[INFO] Attribute %s changed", attribute)
			opts := &cloudscale.LoadBalancerListenerRequest{}
			requests = append(requests, opts)

			if attribute == "name" {
				opts.Name = d.Get(attribute).(string)
			} else if attribute == "protocol_port" {
				opts.ProtocolPort = d.Get(attribute).(int)
			} else if attribute == "protocol" {
				opts.Protocol = d.Get(attribute).(string)
			} else if attribute == "timeout_client_data_ms" {
				opts.TimeoutClientDataMS = d.Get(attribute).(int)
			} else if attribute == "timeout_member_connect_ms" {
				opts.TimeoutMemberConnectMS = d.Get(attribute).(int)
			} else if attribute == "timeout_member_data_ms" {
				opts.TimeoutMemberDataMS = d.Get(attribute).(int)
			} else if attribute == "allowed_cidrs" {
				allowedCIDRs := d.Get("allowed_cidrs").([]any)
				s := make([]string, len(allowedCIDRs))
				for i := range allowedCIDRs {
					s[i] = allowedCIDRs[i].(string)
				}
				opts.AllowedCIDRs = s
			} else if attribute == "tags" {
				opts.Tags = CopyTags(d)
			}
		}
	}
	return requests
}

func deleteLoadBalancerListener(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	return client.LoadBalancerListeners.Delete(context.Background(), id)
}
