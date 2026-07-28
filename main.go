package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/terraform-providers/terraform-provider-cloudscale/cloudscale"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary, via
	// -ldflags "-X main.version={{.Version}} -X main.commit={{.Commit}}"
	// (see .goreleaser.yml). version is forwarded into the provider and
	// appended to the cloudscale-go-sdk User-Agent. Pattern follows
	// HashiCorp's provider scaffold:
	// https://github.com/hashicorp/terraform-provider-scaffolding/blob/main/main.go
	version string = "dev"

	// goreleaser can also pass the specific commit if you want
	commit string = "none"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return cloudscale.Provider(version)
		},
	})
}
