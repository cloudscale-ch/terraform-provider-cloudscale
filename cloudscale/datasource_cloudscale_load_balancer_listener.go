package cloudscale

import (
	"context"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleLoadBalancerListener() *schema.Resource {
	recordSchema := getLoadBalancerListenerSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("load balancer listeners", recordSchema, getFetchFunc(
			listLoadBalancerListeners,
			gatherLoadBalancerListenerResourceData,
		)),
		Schema: recordSchema,
	}
}

func listLoadBalancerListeners(d *schema.ResourceData, meta any) ([]cloudscale.LoadBalancerListener, error) {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancerListeners.List(context.Background())
}
