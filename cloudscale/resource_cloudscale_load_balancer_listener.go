package cloudscale

import (
	"context"
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func resourceCloudscaleLoadBalancerListener() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudscaleLoadBalancerListenerCreate,
		Read:   resourceCloudscaleLoadBalancerListenerRead,
		Update: resourceCloudscaleLoadBalancerListenerUpdate,
		Delete: getDeleteOperation("load balancer listener", deleteLoadBalancerListener),

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
			Required: t.isResource(),
			Optional: t.isDataSource(),
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
		Pool:         d.Get("pool_uuid").(string),
		Protocol:     d.Get("protocol").(string),
		ProtocolPort: d.Get("protocol_port").(int),
	}

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

func fillLoadBalancerListenerSchema(d *schema.ResourceData, loadbalancerlistener *cloudscale.LoadBalancerListener) {
	fillResourceData(d, gatherLoadBalancerListenerResourceData(loadbalancerlistener))
}

func gatherLoadBalancerListenerResourceData(loadbalancerlistener *cloudscale.LoadBalancerListener) ResourceDataRaw {
	m := make(map[string]any)
	m["id"] = loadbalancerlistener.UUID
	m["href"] = loadbalancerlistener.HREF
	m["name"] = loadbalancerlistener.Name
	m["pool_uuid"] = loadbalancerlistener.Pool.UUID
	m["pool_name"] = loadbalancerlistener.Pool.Name
	m["pool_href"] = loadbalancerlistener.Pool.HREF
	m["protocol"] = loadbalancerlistener.Protocol
	m["protocol_port"] = loadbalancerlistener.ProtocolPort
	m["tags"] = loadbalancerlistener.Tags
	return m
}

func resourceCloudscaleLoadBalancerListenerRead(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)

	loadbalancerListener, err := client.LoadBalancerListeners.Get(context.Background(), d.Id())
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving load balancer listener")
	}

	fillLoadBalancerListenerSchema(d, loadbalancerListener)
	return nil
}

func resourceCloudscaleLoadBalancerListenerUpdate(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	for _, attribute := range []string{"name", "protocol", "protocol_port", "tags"} {
		if d.HasChange(attribute) {
			opts := &cloudscale.LoadBalancerListenerRequest{}
			if attribute == "name" {
				opts.Name = d.Get(attribute).(string)
			} else if attribute == "protocol_port" {
				opts.ProtocolPort = d.Get(attribute).(int)
			} else if attribute == "protocol" {
				opts.Protocol = d.Get(attribute).(string)
			} else if attribute == "tags" {
				opts.Tags = CopyTags(d)
			}
			err := client.LoadBalancerListeners.Update(context.Background(), id, opts)
			if err != nil {
				return fmt.Errorf("Error updating the Load Balancer Listener (%s): %s", id, err)
			}
		}
	}
	return resourceCloudscaleLoadBalancerListenerRead(d, meta)
}

func deleteLoadBalancerListener(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	return client.LoadBalancerListeners.Delete(context.Background(), id)
}
