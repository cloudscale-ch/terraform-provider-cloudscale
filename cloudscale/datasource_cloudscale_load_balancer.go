package cloudscale

import (
	"context"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleLoadBalancer() *schema.Resource {
	recordSchema := getLoadBalancerSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("load balancers", recordSchema, getFetchFunc(
			listLoadBalancers,
			gatherLoadBalancerResourceData,
		)),
		Schema: recordSchema,
	}
}

func listLoadBalancers(d *schema.ResourceData, meta any) ([]cloudscale.LoadBalancer, error) {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancers.List(context.Background())
}
