package cloudscale

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"math"
	"time"
)

func waitForStatus(
	pending []string,
	target string,
	timeout *time.Duration,
	refreshFunc resource.StateRefreshFunc,
) (any, error) {
	if timeout == nil {
		defaultTimeout := 5 * time.Minute
		timeout = &(defaultTimeout)
	}

	stateConf := &resource.StateChangeConf{
		Pending:        pending,
		Target:         []string{target},
		Refresh:        refreshFunc,
		Timeout:        *timeout,
		Delay:          10 * time.Second,
		MinTimeout:     3 * time.Second,
		NotFoundChecks: math.MaxInt32,
	}

	return stateConf.WaitForState()
}
