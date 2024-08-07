package cloudscale

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func sharedConfigForRegion(region string) (any, error) {
	if os.Getenv("CLOUDSCALE_API_TOKEN") == "" {
		return nil, fmt.Errorf("empty CLOUDSCALE_API_TOKEN")
	}

	config := Config{
		Token: os.Getenv("CLOUDSCALE_API_TOKEN"),
	}

	// configures a default client for the region, using the above env vars
	client, err := config.Client()
	if err != nil {
		return nil, fmt.Errorf("error getting cloudscale client")
	}

	return client, nil
}
