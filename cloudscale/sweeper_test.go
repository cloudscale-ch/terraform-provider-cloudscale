package cloudscale

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func sharedConfigForRegion(region string) (interface{}, error) {
	if os.Getenv("CLOUDSCALE_TOKEN") == "" {
		return nil, fmt.Errorf("empty CLOUDSCALE_TOKEN")
	}

	config := Config{
		Token: os.Getenv("CLOUDSCALE_TOKEN"),
	}

	// configures a default client for the region, using the above env vars
	client, err := config.Client()
	if err != nil {
		return nil, fmt.Errorf("error getting cloudscale client")
	}

	return client, nil
}
