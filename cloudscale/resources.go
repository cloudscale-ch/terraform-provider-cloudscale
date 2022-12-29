package cloudscale

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func getDeleteOperation(
	resourceType string,
	deleteFunc func(d *schema.ResourceData, meta any) error,
) schema.DeleteFunc {
	return func(d *schema.ResourceData, meta any) error {
		log.Printf("[INFO] Deleting %s: %s", resourceType, d.Id())
		err := deleteFunc(d, meta)

		if err != nil {
			return CheckDeleted(d, err, fmt.Sprintf("Error deleting %s", resourceType))
		}
		return nil
	}
}
