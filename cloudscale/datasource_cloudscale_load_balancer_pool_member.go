package cloudscale

import (
	"context"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleLoadBalancerPoolMember() *schema.Resource {
	recordSchema := getLoadBalancerPoolMemberSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("load balancer pool members", recordSchema, getFetchFunc(
			listLoadBalancerPoolMembers,
			gatherLoadBalancerPoolMemberResourceData,
		)),
		Schema: recordSchema,
	}
}

func listLoadBalancerPoolMembers(d *schema.ResourceData, meta any) ([]cloudscale.LoadBalancerPoolMember, error) {
	client := meta.(*cloudscale.Client)
	poolId := d.Get("pool_uuid").(string)
	return client.LoadBalancerPoolMembers.List(context.Background(), poolId)
}
