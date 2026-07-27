package cloudscale

import (
	"context"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"golang.org/x/oauth2"
)

type Config struct {
	Token   string
	Version string
}

func (c *Config) Client() (*cloudscale.Client, error) {
	tc := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.Token},
	))

	tc.Transport = logging.NewSubsystemLoggingHTTPTransport("Cloudscale", tc.Transport)

	client := cloudscale.NewClient(tc)
	if c.Version != "" {
		client.UserAgent = client.UserAgent + " terraform-provider-cloudscale/" + c.Version
	}

	return client, nil
}
