package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleFloatingIP() *schema.Resource {
	recordSchema := getFloatingIPSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("Floating IPs", recordSchema, getFetchFunc(
			listFloatingIPs,
			gatherFloatingIPResourceData,
		)),
		Schema: recordSchema,
	}
}

func listFloatingIPs(d *schema.ResourceData, meta any) ([]cloudscale.FloatingIP, error) {
	client := meta.(*cloudscale.Client)
	return client.FloatingIPs.List(context.Background())
}