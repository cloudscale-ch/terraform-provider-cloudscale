package cloudscale

import (
	"fmt"
	"math"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

// waitForDeleted polls existsFunc until the resource is gone.
// existsFunc must return (true, nil) while the resource exists,
// (false, nil) once it is gone, or (_, err) on unexpected errors.
func waitForDeleted(timeout time.Duration, existsFunc func() (bool, error)) error {
	deadline := time.Now().Add(timeout)
	time.Sleep(10 * time.Second)
	for {
		exists, err := existsFunc()
		if err != nil {
			return err
		}
		if !exists {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout: resource still exists after %s", timeout)
		}
		time.Sleep(10 * time.Second)
	}
}
