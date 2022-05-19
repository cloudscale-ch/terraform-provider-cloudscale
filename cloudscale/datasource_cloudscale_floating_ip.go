package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudscaleFloatingIP() *schema.Resource {
	recordSchema := getFloatingIPSchema(DATA_SOURCE)

	return &schema.Resource{
		ReadContext: dataSourceResourceRead("Floating IPs", recordSchema, floatingIPsRead),
		Schema:      recordSchema,
	}
}

func floatingIPsRead(meta interface{}) ([]ResourceDataRaw, error) {
	client := meta.(*cloudscale.Client)
	floatingIPList, err := client.FloatingIPs.List(context.Background())
	if err != nil {
		return nil, err
	}
	var rawItems []ResourceDataRaw
	for _, floating_ip := range floatingIPList {
		rawItems = append(rawItems, gatherFloatingIPResourceData(&floating_ip))
	}
	return rawItems, nil
}
