package cloudscale

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CLOUDSCALE_TOKEN", nil),
				Description: "The token for API operations.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"cloudscale_server":       resourceCloudScaleServer(),
			"cloudscale_server_group": resourceCloudScaleServerGroup(),
			"cloudscale_volume":       resourceCloudScaleVolume(),
			"cloudscale_network":      resourceCloudScaleNetwork(),
			"cloudscale_subnet":       resourceCloudScaleSubnet(),
			"cloudscale_floating_ip":  resourceCloudScaleFloatingIP(),
			"cloudscale_objects_user": resourceCloudScaleObjectsUser(),
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
