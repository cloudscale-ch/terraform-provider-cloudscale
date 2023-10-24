package cloudscale

import (
	"github.com/cloudscale-ch/cloudscale-go-sdk/v4"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"golang.org/x/oauth2"
)

type Config struct {
	Token string
}

func (c *Config) Client() (*cloudscale.Client, error) {
	tc := oauth2.NewClient(oauth2.NoContext, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.Token},
	))

	tc.Transport = logging.NewTransport("Cloudscale", tc.Transport)

	client := cloudscale.NewClient(tc)

	return client, nil
}
