package cloudscale

import (
	"context"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v4"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleLoadBalancerPool() *schema.Resource {
	recordSchema := getLoadBalancerPoolSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("load balancer pools", recordSchema, getFetchFunc(
			listLoadBalancerPools,
			gatherLoadBalancerPoolResourceData,
		)),
		Schema: recordSchema,
	}
}

func listLoadBalancerPools(d *schema.ResourceData, meta any) ([]cloudscale.LoadBalancerPool, error) {
	client := meta.(*cloudscale.Client)
	return client.LoadBalancerPools.List(context.Background())
}
