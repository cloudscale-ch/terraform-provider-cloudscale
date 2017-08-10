package cloudscale

import (
	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"golang.org/x/oauth2"
)

type Config struct {
	Token string
}

func (c *Config) Client() (*cloudscale.Client, error) {

	tc := oauth2.NewClient(oauth2.NoContext, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.Token},
	))
	client := cloudscale.NewClient(tc)

	return client, nil
}
