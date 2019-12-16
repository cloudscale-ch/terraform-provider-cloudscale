package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/terraform-providers/terraform-provider-cloudscale/cloudscale"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: cloudscale.Provider,
	})
}
