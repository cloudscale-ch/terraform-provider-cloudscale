package cloudscale

import (
	"context"
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
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
			Required: t.isResource(),
			Computed: t.isDataSource(),
		},
		"timeout": {
			Type:     schema.TypeInt,
			Required: t.isResource(),
			Computed: t.isDataSource(),
		},
		"max_retries": {
			Type:     schema.TypeInt,
			Required: t.isResource(),
			Computed: t.isDataSource(),
		},
		"max_retries_down": {
			Type:     schema.TypeInt,
			Required: t.isResource(),
			Computed: t.isDataSource(),
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
		Pool:           d.Get("pool_uuid").(string),
		Delay:          d.Get("delay").(int),
		Timeout:        d.Get("timeout").(int),
		MaxRetries:     d.Get("max_retries").(int),
		MaxRetriesDown: d.Get("max_retries_down").(int),
		Type:           d.Get("type").(string),
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

func resourceCloudscaleLoadBalancerHealthMonitorRead(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)

	loadBalancerHealthMonitor, err := client.LoadBalancerHealthMonitors.Get(context.Background(), d.Id())
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving load balancer health monitor")
	}

	fillLoadBalancerHealthMonitorSchema(d, loadBalancerHealthMonitor)
	return nil
}

func fillLoadBalancerHealthMonitorSchema(d *schema.ResourceData, loadBalancerHealthMonitor *cloudscale.LoadBalancerHealthMonitor) {
	fillResourceData(d, gatherLoadBalancerHealthMonitorResourceData(loadBalancerHealthMonitor))
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

func resourceCloudscaleLoadBalancerHealthMonitorUpdate(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	for _, attribute := range []string{"delay", "timeout", "max_retries", "max_retries_down", "tags"} {
		if d.HasChange(attribute) {
			opts := &cloudscale.LoadBalancerHealthMonitorRequest{}
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
			err := client.LoadBalancerHealthMonitors.Update(context.Background(), id, opts)
			if err != nil {
				return fmt.Errorf("Error updating the Load Balancer Health Monitor (%s): %s", id, err)
			}
		}
	}

	return resourceCloudscaleLoadBalancerRead(d, meta)
}

func resourceCloudscaleLoadBalancerHealthMonitorDelete(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	log.Printf("[INFO] Deleting LoadBalancerHealthMonitor: %s", id)
	err := client.LoadBalancerHealthMonitors.Delete(context.Background(), id)

	if err != nil {
		return CheckDeleted(d, err, "Error deleting load balancer health monitor")
	}

	return nil
}
