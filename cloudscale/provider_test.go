package cloudscale

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider(testAccProviderVersion())
	testAccProviders = map[string]*schema.Provider{
		"cloudscale": testAccProvider,
	}
}

// testAccProviderVersion identifies acceptance-test requests in the SDK
// User-Agent by the commit under test, so requests hitting the real API can
// be traced back to the CI run that made them. GITHUB_SHA is set
// automatically by GitHub Actions; it falls back to "test" for local runs.
func testAccProviderVersion() string {
	sha := os.Getenv("GITHUB_SHA")
	if sha == "" {
		return "test"
	}
	if len(sha) > 6 {
		sha = sha[:6]
	}
	return "acctest-" + sha
}

func TestProvider(t *testing.T) {
	if err := Provider("test").InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider("test")
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
