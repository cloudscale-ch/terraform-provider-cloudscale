package cloudscale

import (
	"context"
	"math"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func waitForStatus(
	ctx context.Context,
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

	return stateConf.WaitForStateContext(ctx)
}

// waitForDeleted polls existsFunc until the resource is gone. The deadline is
// carried by ctx: the SDK injects d.Timeout(schema.TimeoutDelete) before
// calling DeleteContext, so no separate timeout parameter is needed.
// existsFunc must return (true, nil) while the resource exists,
// (false, nil) once it is gone, or (_, err) on unexpected errors.
func waitForDeleted(ctx context.Context, existsFunc func() (bool, error)) error {
	time.Sleep(10 * time.Second)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		exists, err := existsFunc()
		if err != nil {
			return err
		}
		if !exists {
			return nil
		}
		time.Sleep(10 * time.Second)
	}
}
