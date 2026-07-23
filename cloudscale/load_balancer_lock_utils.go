package cloudscale

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// The cloudscale API applies one change at a time to a load balancer. Creating,
// updating or deleting any of its sub-resources (pools, pool members, listeners
// and health monitors) mutates that same load balancer, so every such operation
// has to serialize on the load balancer it belongs to. This file holds the lock
// keys that make all of them agree on one key per load balancer.

// lbLockKey returns the mutex key identifying a single load balancer. Every
// operation that mutates the load balancer or any of its sub-resources locks on
// this key, so they all have to derive it from the same load balancer UUID.
func lbLockKey(lbUUID string) string {
	return fmt.Sprintf("cloudscale/load-balancer/%s", lbUUID)
}

// lbGlobalLockUUID stands in for the load balancer UUID when the load balancer an
// operation belongs to cannot be resolved. Locking on the resulting key serializes
// the operation against every load balancer: slower than necessary, but correct,
// whereas not locking at all would let it race with the load balancer it mutates.
// The angle brackets keep it from ever colliding with a real UUID.
const lbGlobalLockUUID = "<<global>>"

// lockKeyFromLoadBalancerUUID serializes operations on the load balancer named by
// the load_balancer_uuid attribute. It is meant for resources that carry that UUID
// in their own schema, so the key is read straight from state without a lookup.
var lockKeyFromLoadBalancerUUID = uuidLockKey("load_balancer_uuid", lbLockKey)

// lockKeyFromPoolUUID serializes operations on the load balancer that owns the pool
// named by the pool_uuid attribute. It is meant for resources whose schema has a
// pool_uuid but no load_balancer_uuid, so their load balancer is not in state and has
// to be resolved via the API to obtain the lock key. Such operations still mutate the
// parent load balancer and must therefore serialize on it, not on the pool.
//
// The lookup is a read and runs before the lock is taken, so it cannot deadlock. If the
// pool is there but its load balancer cannot be determined, the operation falls back to
// lbGlobalLockUUID instead of running unserialized.
func lockKeyFromPoolUUID(ctx context.Context, d *schema.ResourceData, meta any) (string, bool) {
	poolUUID, ok := d.GetOk("pool_uuid")
	if !ok {
		log.Printf("[WARN] lockKeyFromPoolUUID: pool_uuid not set for resource %s, skipping lock", d.Id())
		return "", false
	}
	client := meta.(*cloudscale.Client)
	pool, err := client.LoadBalancerPools.Get(ctx, poolUUID.(string))
	if err != nil {
		log.Printf("[WARN] lockKeyFromPoolUUID: could not resolve load balancer for pool %s: %s; locking globally", poolUUID, err)
		return lbLockKey(lbGlobalLockUUID), true
	}
	if pool.LoadBalancer.UUID == "" {
		log.Printf("[WARN] lockKeyFromPoolUUID: pool %s has no load balancer, locking globally", poolUUID)
		return lbLockKey(lbGlobalLockUUID), true
	}
	return lbLockKey(pool.LoadBalancer.UUID), true
}
