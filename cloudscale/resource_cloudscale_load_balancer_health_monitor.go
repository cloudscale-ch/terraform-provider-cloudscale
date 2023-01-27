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
		"http_expected_codes": {
			Type:     schema.TypeList,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Optional: true,
			Computed: true,
		},
		"http_method": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"http_url_path": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"http_version": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"http_host": {
			Type:     schema.TypeString,
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

	if opts.Type == "http" {
		httpOpts := cloudscale.LoadBalancerHealthMonitorHTTPRequest{}
		if attr, ok := d.GetOk("http_expected_codes"); ok {
			codes := attr.([]any)
			s := getCodes(codes)
			httpOpts.ExpectedCodes = s
		}
		if attr, ok := d.GetOk("http_method"); ok {
			httpOpts.Method = attr.(string)
		}
		if attr, ok := d.GetOk("http_version"); ok {
			httpOpts.Version = attr.(string)
		}
		if attr, ok := d.GetOk("http_url_path"); ok {
			httpOpts.UrlPath = attr.(string)
		}
		if attr, ok := d.GetOk("http_host"); ok {
			s := attr.(string)
			httpOpts.Host = &s
		}
		opts.HTTP = &httpOpts
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

func getCodes(codes []any) []string {
	s := make([]string, len(codes))
	for i := range codes {
		s[i] = codes[i].(string)
	}
	return s
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

	for _, attribute := range []string{
		"delay", "timeout", "max_retries", "max_retries_down",
		"http_expected_codes", "http_method", "http_url_path", "http_version", "http_host",
		"tags",
	} {
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

			if d.Get("type").(string) == "http" {
				httpOpts := cloudscale.LoadBalancerHealthMonitorHTTPRequest{}
				if attribute == "http_expected_codes" {
					codes := d.Get(attribute).([]any)
					s := getCodes(codes)
					httpOpts.ExpectedCodes = s
				}
				if attribute == "http_method" {
					httpOpts.Method = d.Get(attribute).(string)
				} else if attribute == "http_url_path" {
					httpOpts.UrlPath = d.Get(attribute).(string)
				} else if attribute == "http_version" {
					httpOpts.Version = d.Get(attribute).(string)
				} else if attribute == "http_host" {
					if attr, ok := d.GetOk(attribute); ok {
						s := attr.(string)
						httpOpts.Host = &s
					}
				}
				opts.HTTP = &httpOpts
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
	if loadBalancerHealthMonitor.HTTP != nil {
		m["http_expected_codes"] = loadBalancerHealthMonitor.HTTP.ExpectedCodes
		m["http_method"] = loadBalancerHealthMonitor.HTTP.Method
		m["http_url_path"] = loadBalancerHealthMonitor.HTTP.UrlPath
		m["http_version"] = loadBalancerHealthMonitor.HTTP.Version
		m["http_host"] = loadBalancerHealthMonitor.HTTP.Host
	} else {
		m["http_expected_codes"] = nil
	}
	m["tags"] = loadBalancerHealthMonitor.Tags
	return m
}

func deleteLoadBalancerHealthMonitor(d *schema.ResourceData, meta any) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()
	return client.LoadBalancerHealthMonitors.Delete(context.Background(), id)
}
