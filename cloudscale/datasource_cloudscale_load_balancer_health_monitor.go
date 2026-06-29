package cloudscale

import (
	"context"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleLoadBalancerHealthMonitor() *schema.Resource {
	recordSchema := getLoadBalancerHealthMonitorSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("load balancer health monitors", recordSchema, getFetchFunc(
			listLoadBalancerHealthMonitors,
			gatherLoadBalancerHealthMonitorResourceData,
		)),
		Schema: recordSchema,
	}
}

func listLoadBalancerHealthMonitors(ctx context.Context, d *schema.ResourceData, meta any) ([]cloudscale.LoadBalancerHealthMonitor, error) {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancerHealthMonitors.List(ctx)
}
