package cloudscale

import (
	"context"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleLoadBalancerPool() *schema.Resource {
	recordSchema := getLoadBalancerPoolSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("load balancer pools", recordSchema, loadBalancerPoolsRead),
		Schema:      recordSchema,
	}
}

func loadBalancerPoolsRead(meta interface{}) ([]ResourceDataRaw, error) {
	client := meta.(*cloudscale.Client)
	loadBalancerPoolList, err := client.LoadBalancerPools.List(context.Background())
	if err != nil {
		return nil, err
	}
	var rawItems []ResourceDataRaw
	for _, pool := range loadBalancerPoolList {
		rawItems = append(rawItems, gatherLoadBalancerPoolResourceData(&pool))
	}
	return rawItems, nil
}
