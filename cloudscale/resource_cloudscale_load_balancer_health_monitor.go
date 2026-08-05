package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const healthMonitorHumanName = "load balancer health monitor"

// Health monitor operations serialize on the load balancer that owns the parent pool
// via lockKeyFromPoolUUID.
var (
	resourceCloudscaleLoadBalancerHealthMonitorCreate = getCreateOperation(createLoadBalancerHealthMonitor, lockKeyFromPoolUUID)
	resourceCloudscaleLoadBalancerHealthMonitorRead   = getReadOperation(healthMonitorHumanName, getGenericResourceIdentifierFromSchema, readLoadBalancerHealthMonitor, gatherLoadBalancerHealthMonitorResourceData)
	resourceCloudscaleLoadBalancerHealthMonitorUpdate = getUpdateOperation(healthMonitorHumanName, getGenericResourceIdentifierFromSchema, updateLoadBalancerHealthMonitor, resourceCloudscaleLoadBalancerHealthMonitorRead, gatherLoadBalancerHealthMonitorUpdateRequests, lockKeyFromPoolUUID)
	resourceCloudscaleLoadBalancerHealthMonitorDelete = getDeleteOperation(healthMonitorHumanName, getGenericResourceIdentifierFromSchema, deleteLoadBalancerHealthMonitor, lockKeyFromPoolUUID)
)

func resourceCloudscaleLoadBalancerHealthMonitor() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudscaleLoadBalancerHealthMonitorCreate,
		ReadContext:   resourceCloudscaleLoadBalancerHealthMonitorRead,
		UpdateContext: resourceCloudscaleLoadBalancerHealthMonitorUpdate,
		DeleteContext: resourceCloudscaleLoadBalancerHealthMonitorDelete,

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
		"delay_s": {
			Type:     schema.TypeInt,
			Optional: t.isResource(),
			Computed: true,
		},
		"timeout_s": {
			Type:     schema.TypeInt,
			Optional: t.isResource(),
			Computed: true,
		},
		"up_threshold": {
			Type:     schema.TypeInt,
			Optional: t.isResource(),
			Computed: true,
		},
		"down_threshold": {
			Type:     schema.TypeInt,
			Optional: t.isResource(),
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
			Optional: t.isResource(),
			Computed: true,
		},
		"http_method": {
			Type:     schema.TypeString,
			Optional: t.isResource(),
			Computed: true,
		},
		"http_url_path": {
			Type:     schema.TypeString,
			Optional: t.isResource(),
			Computed: true,
		},
		"http_version": {
			Type:     schema.TypeString,
			Optional: t.isResource(),
			Computed: true,
			ForceNew: t.isResource(),
		},
		"http_host": {
			Type:     schema.TypeString,
			Optional: t.isResource(),
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

func createLoadBalancerHealthMonitor(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.LoadBalancerHealthMonitorRequest{
		Pool: d.Get("pool_uuid").(string),
		Type: d.Get("type").(string),
	}

	if attr, ok := d.GetOk("delay_s"); ok {
		opts.DelayS = attr.(int)
	}
	if attr, ok := d.GetOk("timeout_s"); ok {
		opts.TimeoutS = attr.(int)
	}
	if attr, ok := d.GetOk("up_threshold"); ok {
		opts.UpThreshold = attr.(int)
	}
	if attr, ok := d.GetOk("down_threshold"); ok {
		opts.DownThreshold = attr.(int)
	}

	if opts.Type == "http" || opts.Type == "https" {
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
			httpOpts.URLPath = attr.(string)
		}
		if attr, ok := d.GetOk("http_host"); ok {
			s := attr.(string)
			httpOpts.Host = &s
		}
		opts.HTTP = &httpOpts
	}

	opts.Tags = TagsFromState(d)

	log.Printf("[DEBUG] LoadBalancerHealthMonitor create configuration: %#v", opts)

	healthMonitor, err := client.LoadBalancerHealthMonitors.Create(ctx, opts)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating LoadBalancerHealthMonitor: %s", err))
	}

	d.SetId(healthMonitor.UUID)

	log.Printf("[INFO] LoadBalancerHealthMonitor UUID: %s", d.Id())
	return resourceCloudscaleLoadBalancerHealthMonitorRead(ctx, d, meta)
}

func getCodes(codes []any) []string {
	s := make([]string, len(codes))
	for i := range codes {
		s[i] = codes[i].(string)
	}
	return s
}

func readLoadBalancerHealthMonitor(ctx context.Context, rId GenericResourceIdentifier, meta any) (*cloudscale.LoadBalancerHealthMonitor, error) {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancerHealthMonitors.Get(ctx, rId.Id)
}

func updateLoadBalancerHealthMonitor(ctx context.Context, rId GenericResourceIdentifier, meta any, updateRequest *cloudscale.LoadBalancerHealthMonitorRequest) error {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancerHealthMonitors.Update(ctx, rId.Id, updateRequest)
}

func gatherLoadBalancerHealthMonitorUpdateRequests(d *schema.ResourceData) []*cloudscale.LoadBalancerHealthMonitorRequest {
	requests := make([]*cloudscale.LoadBalancerHealthMonitorRequest, 0)

	for _, attribute := range []string{
		"delay_s", "timeout_s", "up_threshold", "down_threshold",
		"http_expected_codes", "http_method", "http_url_path", "http_host",
		"tags",
	} {
		if d.HasChange(attribute) {
			log.Printf("[INFO] Attribute %s changed", attribute)
			opts := &cloudscale.LoadBalancerHealthMonitorRequest{}
			requests = append(requests, opts)

			if attribute == "delay_s" {
				opts.DelayS = d.Get(attribute).(int)
			} else if attribute == "timeout_s" {
				opts.TimeoutS = d.Get(attribute).(int)
			} else if attribute == "up_threshold" {
				opts.UpThreshold = d.Get(attribute).(int)
			} else if attribute == "down_threshold" {
				opts.DownThreshold = d.Get(attribute).(int)
			} else if attribute == "tags" {
				opts.Tags = TagsFromState(d)
			}

			monitorType := d.Get("type").(string)
			if monitorType == "http" || monitorType == "https" {
				httpOpts := cloudscale.LoadBalancerHealthMonitorHTTPRequest{}
				if attribute == "http_expected_codes" {
					codes := d.Get(attribute).([]any)
					s := getCodes(codes)
					httpOpts.ExpectedCodes = s
				}
				if attribute == "http_method" {
					httpOpts.Method = d.Get(attribute).(string)
				} else if attribute == "http_url_path" {
					httpOpts.URLPath = d.Get(attribute).(string)
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
	m["delay_s"] = loadBalancerHealthMonitor.DelayS
	m["timeout_s"] = loadBalancerHealthMonitor.TimeoutS
	m["up_threshold"] = loadBalancerHealthMonitor.UpThreshold
	m["down_threshold"] = loadBalancerHealthMonitor.DownThreshold
	m["type"] = loadBalancerHealthMonitor.Type
	if loadBalancerHealthMonitor.HTTP != nil {
		m["http_expected_codes"] = loadBalancerHealthMonitor.HTTP.ExpectedCodes
		m["http_method"] = loadBalancerHealthMonitor.HTTP.Method
		m["http_url_path"] = loadBalancerHealthMonitor.HTTP.URLPath
		m["http_version"] = loadBalancerHealthMonitor.HTTP.Version
		m["http_host"] = loadBalancerHealthMonitor.HTTP.Host
	} else {
		m["http_expected_codes"] = nil
	}
	m["tags"] = TagsToState(loadBalancerHealthMonitor.Tags)
	return m
}

func deleteLoadBalancerHealthMonitor(ctx context.Context, rId GenericResourceIdentifier, meta any) error {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancerHealthMonitors.Delete(ctx, rId.Id)
}
