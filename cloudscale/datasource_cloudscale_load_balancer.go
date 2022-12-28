package cloudscale

import (
	"context"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleLoadBalancer() *schema.Resource {
	recordSchema := getLoadBalancerSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("load balancers", recordSchema, loadBalancersRead),
		Schema:      recordSchema,
	}
}

func loadBalancersRead(d *schema.ResourceData, meta any) ([]ResourceDataRaw, error) {
	client := meta.(*cloudscale.Client)
	loadBalancerList, err := client.LoadBalancers.List(context.Background())
	if err != nil {
		return nil, err
	}
	var rawItems []ResourceDataRaw
	for _, loadBalancer := range loadBalancerList {
		rawItems = append(rawItems, gatherLoadBalancerResourceData(&loadBalancer))
	}
	return rawItems, nil
}
