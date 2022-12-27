package cloudscale

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CLOUDSCALE_API_TOKEN", nil),
				Description: "The token for API operations.",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"cloudscale_server":             dataSourceCloudscaleServer(),
			"cloudscale_server_group":       dataSourceCloudscaleServerGroup(),
			"cloudscale_volume":             dataSourceCloudscaleVolume(),
			"cloudscale_network":            dataSourceCloudscaleNetwork(),
			"cloudscale_subnet":             dataSourceCloudscaleSubnet(),
			"cloudscale_floating_ip":        dataSourceCloudscaleFloatingIP(),
			"cloudscale_objects_user":       dataSourceCloudscaleObjectsUser(),
			"cloudscale_custom_image":       dataSourceCloudscaleCustomImage(),
			"cloudscale_load_balancer":      dataSourceCloudscaleLoadBalancer(),
			"cloudscale_load_balancer_pool": dataSourceCloudscaleLoadBalancerPool(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"cloudscale_server":                    resourceCloudscaleServer(),
			"cloudscale_server_group":              resourceCloudscaleServerGroup(),
			"cloudscale_volume":                    resourceCloudscaleVolume(),
			"cloudscale_network":                   resourceCloudscaleNetwork(),
			"cloudscale_subnet":                    resourceCloudscaleSubnet(),
			"cloudscale_floating_ip":               resourceCloudscaleFloatingIP(),
			"cloudscale_objects_user":              resourceCloudscaleObjectsUser(),
			"cloudscale_custom_image":              resourceCloudscaleCustomImage(),
			"cloudscale_load_balancer":             resourceCloudscaleLoadBalancer(),
			"cloudscale_load_balancer_pool":        resourceCloudscaleLoadBalancerPools(),
			"cloudscale_load_balancer_pool_member": resourceCloudscaleLoadBalancerPoolMembers(),
		},
		ConfigureFunc: providerConfigureClient,
	}
}

func providerConfigureClient(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Token: d.Get("token").(string),
	}
	return config.Client()
}
