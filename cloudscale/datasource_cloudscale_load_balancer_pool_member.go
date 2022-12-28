package cloudscale

import (
	"context"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleLoadBalancerPoolMember() *schema.Resource {
	recordSchema := getLoadBalancerPoolMemberSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("load balancer pool members", recordSchema, loadBalancerPoolMembersRead),
		Schema:      recordSchema,
	}
}

func loadBalancerPoolMembersRead(d *schema.ResourceData, meta any) ([]ResourceDataRaw, error) {
	client := meta.(*cloudscale.Client)
	poolId := d.Get("pool_uuid").(string)
	loadBalancerPoolMemberList, err := client.LoadBalancerPoolMembers.List(context.Background(), poolId)
	if err != nil {
		return nil, err
	}
	var rawItems []ResourceDataRaw
	for _, poolMember := range loadBalancerPoolMemberList {
		rawItems = append(rawItems, gatherLoadBalancerPoolMemberResourceData(&poolMember))
	}
	return rawItems, nil
}
