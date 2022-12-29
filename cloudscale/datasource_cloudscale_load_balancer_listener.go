package cloudscale

import (
	"context"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleLoadBalancerListener() *schema.Resource {
	recordSchema := getLoadBalancerListenerSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("load balancer listeners", recordSchema, loadBalancerListenersRead),
		Schema:      recordSchema,
	}
}

func loadBalancerListenersRead(d *schema.ResourceData, meta any) ([]ResourceDataRaw, error) {
	client := meta.(*cloudscale.Client)
	loadBalancerListenerList, err := client.LoadBalancerListeners.List(context.Background())
	if err != nil {
		return nil, err
	}
	var rawItems []ResourceDataRaw
	for _, Listener := range loadBalancerListenerList {
		rawItems = append(rawItems, gatherLoadBalancerListenerResourceData(&Listener))
	}
	return rawItems, nil
}
