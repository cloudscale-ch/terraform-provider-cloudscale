package cloudscale

import (
	"fmt"
	"net/http"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	TagsSchema schema.Schema = schema.Schema{
		Type: schema.TypeMap,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Optional: true,
	}
)

// TagsFromState reads the "tags" attribute from Terraform state and converts it to the SDK type.
func TagsFromState(d *schema.ResourceData) *cloudscale.TagMap {
	newTags := make(cloudscale.TagMap)

	for k, v := range d.Get("tags").(map[string]any) {
		newTags[k] = v.(string)
	}

	return &newTags
}

// TagsToState converts SDK tags to the map type used in Terraform state.
func TagsToState(tags cloudscale.TagMap) map[string]interface{} {
	result := make(map[string]interface{}, len(tags))
	for k, v := range tags {
		result[k] = v
	}
	return result
}

// CheckDeleted checks the error to see if it's a 404 (Not Found) and, if so,
// sets the resource ID to the empty string instead of throwing an error.
func CheckDeleted(d *schema.ResourceData, err error, msg string) error {
	errorResponse, ok := err.(*cloudscale.ErrorResponse)
	if ok && errorResponse.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}
	return fmt.Errorf("%s %s: %s", msg, d.Id(), err)
}
