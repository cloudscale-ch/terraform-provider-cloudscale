package cloudscale

import (
	"context"
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

const healthMonitorHumanName = "load balancer health monitor"

var (
	resourceCloudscaleLoadBalancerHealthMonitorRead   = getReadOperation(healthMonitorHumanName, getGenericResourceIdentifierFromSchema, readLoadBalancerHealthMonitor, gatherLoadBalancerHealthMonitorResourceData)
	resourceCloudscaleLoadBalancerHealthMonitorUpdate = getUpdateOperation(healthMonitorHumanName, getGenericResourceIdentifierFromSchema, updateLoadBalancerHealthMonitor, resourceCloudscaleLoadBalancerHealthMonitorRead, gatherLoadBalancerHealthMonitorUpdateRequests)
	resourceCloudscaleLoadBalancerHealthMonitorDelete = getDeleteOperation(healthMonitorHumanName, deleteLoadBalancerHealthMonitor)
)

func resourceCloudscaleLoadBalancerHealthMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudscaleLoadBalancerHealthMonitorCreate,
		Read:   resourceCloudscaleLoadBalancerHealthMonitorRead,
		Update: resourceCloudscaleLoadBalancerHealthMonitorUpdate,
		Delete: resourceCloudscaleLoadBalancerHealthMonitorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: getLoadBalancerHealthMonitorSchema(RESOURCE),
	}
}

func getLoadBalancerHealthMonitorSchema(t SchemaType) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
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
		"delay": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"timeout": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"max_retries": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"max_retries_down": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"type": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Computed: t.isDataSource(),
			ForceNew: true,
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

func resourceCloudscaleLoadBalancerHealthMonitorCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.LoadBalancerHealthMonitorRequest{
		Pool: d.Get("pool_uuid").(string),
		Type: d.Get("type").(string),
	}

	if attr, ok := d.GetOk("delay"); ok {
		opts.Delay = attr.(int)
	}
	if attr, ok := d.GetOk("timeout"); ok {
		opts.Timeout = attr.(int)
	}
	if attr, ok := d.GetOk("max_retries"); ok {
		opts.MaxRetries = attr.(int)
	}
	if attr, ok := d.GetOk("max_retries_down"); ok {
		opts.MaxRetriesDown = attr.(int)
	}

	opts.Tags = CopyTags(d)

	log.Printf("[DEBUG] LoadBalancerHealthMonitor create configuration: %#v", opts)

	healthMonitor, err := client.LoadBalancerHealthMonitors.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating LoadBalancerHealthMonitor: %s", err)
	}

	d.SetId(healthMonitor.UUID)

	log.Printf("[INFO] LoadBalancerHealthMonitor UUID: %s", d.Id())
	err = resourceCloudscaleLoadBalancerHealthMonitorRead(d, meta)
	if err != nil {
		return fmt.Errorf("Error reading the load balancer health monitor (%s): %s", d.Id(), err)
	}
	return nil
}

func readLoadBalancerHealthMonitor(rId GenericResourceIdentifier, meta any) (*cloudscale.LoadBalancerHealthMonitor, error) {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancerHealthMonitors.Get(context.Background(), rId.Id)
}

func updateLoadBalancerHealthMonitor(rId GenericResourceIdentifier, meta any, updateRequest *cloudscale.LoadBalancerHealthMonitorRequest) error {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancerHealthMonitors.Update(context.Background(), rId.Id, updateRequest)
}

func gatherLoadBalancerHealthMonitorUpdateRequests(d *schema.ResourceData) []*cloudscale.LoadBalancerHealthMonitorRequest {
	requests := make([]*cloudscale.LoadBalancerHealthMonitorRequest, 0)

	for _, attribute := range []string{"delay", "timeout", "max_retries", "max_retries_down", "tags"} {
		if d.HasChange(attribute) {
			log.Printf("[INFO] Attribute %s changed", attribute)
			opts := &cloudscale.LoadBalancerHealthMonitorRequest{}
			requests = append(requests, opts)

			if attribute == "delay" {
				opts.Delay = d.Get(attribute).(int)
			} else if attribute == "timeout" {
				opts.Timeout = d.Get(attribute).(int)
			} else if attribute == "max_retries" {
				opts.MaxRetries = d.Get(attribute).(int)
			} else if attribute == "max_retries_down" {
				opts.MaxRetriesDown = d.Get(attribute).(int)
			} else if attribute == "tags" {
				opts.Tags = CopyTags(d)
			}
		}
	}

	return requests
}

func gatherLoadBalancerHealthMonitorResourceData(loadBalancerHealthMonitor *cloudscale.LoadBalancerHealthMonitor) ResourceDataRaw {
	m := make(ResourceDataRaw)
	m["id"] = loadBalancerHealthMonitor.UUID
	m["href"] = loadBalancerHealthMonitor.HREF
	m["pool_uuid"] = loadBalancerHealthMonitor.Pool.UUID
	m["pool_name"] = loadBalancerHealthMonitor.Pool.Name
	m["pool_href"] = loadBalancerHealthMonitor.Pool.HREF
	m["delay"] = loadBalancerHealthMonitor.Delay
	m["timeout"] = loadBalancerHealthMonitor.Timeout
	m["max_retries"] = loadBalancerHealthMonitor.MaxRetries
	m["max_retries_down"] = loadBalancerHealthMonitor.MaxRetriesDown
	m["type"] = loadBalancerHealthMonitor.Type
	m["tags"] = loadBalancerHealthMonitor.Tags
	return m
}

func deleteLoadBalancerHealthMonitor(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	return client.LoadBalancerHealthMonitors.Delete(context.Background(), id)
}
