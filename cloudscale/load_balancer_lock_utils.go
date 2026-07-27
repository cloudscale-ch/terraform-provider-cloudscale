package cloudscale

import (
	"context"
	"fmt"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// The cloudscale API applies one change at a time to a load balancer, and every
// sub-resource operation (pools, pool members, listeners, health monitors)
// mutates its load balancer. Concurrency buys nothing here: the API serializes
// the requests anyway; so we serialize them ourselves. This file derives the
// lock keys for that: one per load balancer.

// lbLockKey returns the mutex key identifying a single load balancer. Every
// operation that mutates the load balancer or any of its sub-resources locks on
// this key, so they all have to derive it from the same load balancer UUID.
func lbLockKey(lbUUID string) string {
	return fmt.Sprintf("cloudscale/load-balancer/%s", lbUUID)
}

// lockKeyFromLoadBalancerUUID derives the lock key for the load balancer named by the
// load_balancer_uuid attribute. It is for resources that carry that UUID in their own schema, so
// the key comes straight from state without a lookup.
var lockKeyFromLoadBalancerUUID = uuidLockKey("load_balancer_uuid", lbLockKey)

// lockKeyFromPoolUUID derives the lock key for the load balancer that owns the pool named by
// pool_uuid. It is for resources that carry pool_uuid but not load_balancer_uuid, so the load
// balancer is resolved via the API. It returns an error when the load balancer can't be determined,
// e.g.: pool_uuid unset, the lookup fails, or the pool has no load balancer.
func lockKeyFromPoolUUID(ctx context.Context, d *schema.ResourceData, meta any) (string, error) {
	poolUUID, ok := d.GetOk("pool_uuid")
	if !ok {
		return "", fmt.Errorf("cannot determine the load balancer to lock: pool_uuid is not set")
	}
	client := meta.(*cloudscale.Client)
	pool, err := client.LoadBalancerPools.Get(ctx, poolUUID.(string))
	if err != nil {
		return "", fmt.Errorf("cannot determine the load balancer to lock for pool %s: %w", poolUUID, err)
	}
	if pool.LoadBalancer.UUID == "" {
		return "", fmt.Errorf("cannot determine the load balancer to lock: pool %s has no load balancer", poolUUID)
	}
	return lbLockKey(pool.LoadBalancer.UUID), nil
}
