package cloudscale

import (
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk"
	"github.com/hashicorp/terraform/helper/schema"
	"net/http"
)

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
