package cloudscale

import (
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
)

var (
	TagsSchema schema.Schema = schema.Schema{
		Type:     schema.TypeMap,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Optional: true,
	};
)

func CopyTags(d *schema.ResourceData) *cloudscale.TagMap {
	newTags := make(cloudscale.TagMap)

	for k, v := range d.Get("tags").(map[string]interface{}) {
		newTags[k] = v.(string)
	}

	return &newTags
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
