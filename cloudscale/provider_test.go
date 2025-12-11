package cloudscale

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"context"
	"strconv"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"cloudscale": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("CLOUDSCALE_API_TOKEN"); v == "" {
		t.Fatal("CLOUDSCALE_API_TOKEN must be set for acceptance tests")
	}
}

func testTagsMatch(resource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}

		attributes := rs.Primary.Attributes
		href, found := attributes["href"]
		if !found {
			return fmt.Errorf("No HREF found")
		}

		client := testAccProvider.Meta().(*cloudscale.Client)
		ctx := context.Background()
		req, err := client.NewRequest(ctx, http.MethodGet, href, nil)
		if err != nil {
			return err
		}

		tagged := new(cloudscale.TaggedResource)
		err = client.Do(ctx, req, tagged)
		if err != nil {
			return err
		}
		in_state := attributes["tags.%"]
		actual := strconv.Itoa(len(tagged.Tags))
		if in_state != actual {
			return fmt.Errorf("State has %s tags, API has %s tags", in_state, actual)
		}

		return nil
	}
}
