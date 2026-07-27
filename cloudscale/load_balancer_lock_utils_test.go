package cloudscale

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v10"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testClient returns a cloudscale client whose requests are served by handler
// instead of the real API.
func testClient(t *testing.T, handler http.Handler) *cloudscale.Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	// The trailing slash matters: the SDK builds request URLs with
	// BaseURL.ResolveReference(), which would otherwise drop the last segment.
	baseURL, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatalf("parsing test server URL: %s", err)
	}

	client := cloudscale.NewClient(nil)
	client.BaseURL = baseURL
	return client
}

// poolHandler serves one pool at the path the SDK uses for pool reads.
func poolHandler(t *testing.T, poolUUID string, pool cloudscale.LoadBalancerPool) http.Handler {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/load-balancers/pools/"+poolUUID, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(pool); err != nil {
			t.Errorf("encoding pool response: %s", err)
		}
	})
	return mux
}

func TestLbLockKey(t *testing.T) {
	if got, want := lbLockKey("abc"), "cloudscale/load-balancer/abc"; got != want {
		t.Errorf("lbLockKey() = %q, want %q", got, want)
	}
}

func TestUuidLockKey(t *testing.T) {
	keyFunc := uuidLockKey("load_balancer_uuid", lbLockKey)
	poolSchema := getLoadBalancerPoolSchema(RESOURCE)

	t.Run("derives the key from the attribute", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, poolSchema, map[string]any{
			"load_balancer_uuid": "lb-1",
		})

		// meta is intentionally nil: reading an attribute must not need a client.
		key, err := keyFunc(context.Background(), d, nil)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if want := "cloudscale/load-balancer/lb-1"; key != want {
			t.Errorf("key = %q, want %q", key, want)
		}
	})

	t.Run("errors when the attribute is unset", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, poolSchema, map[string]any{})

		// The attribute is Required, so an unset value is a coding error: fail rather than
		// lock on a meaningless key.
		if _, err := keyFunc(context.Background(), d, nil); err == nil {
			t.Error("expected an error when the attribute is unset")
		}
	})
}

func TestLockKeyFromPoolUUID(t *testing.T) {
	const (
		poolUUID = "pool-1"
		lbUUID   = "lb-1"
	)
	memberSchema := getLoadBalancerPoolMemberSchema(RESOURCE)

	t.Run("resolves the load balancer that owns the pool", func(t *testing.T) {
		client := testClient(t, poolHandler(t, poolUUID, cloudscale.LoadBalancerPool{
			UUID:         poolUUID,
			LoadBalancer: cloudscale.LoadBalancerStub{UUID: lbUUID},
		}))
		d := schema.TestResourceDataRaw(t, memberSchema, map[string]any{"pool_uuid": poolUUID})

		key, err := lockKeyFromPoolUUID(context.Background(), d, client)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if want := "cloudscale/load-balancer/lb-1"; key != want {
			t.Errorf("key = %q, want %q", key, want)
		}
	})

	t.Run("errors when pool_uuid is unset", func(t *testing.T) {
		client := testClient(t, http.NewServeMux())
		d := schema.TestResourceDataRaw(t, memberSchema, map[string]any{})

		if _, err := lockKeyFromPoolUUID(context.Background(), d, client); err == nil {
			t.Error("expected an error when pool_uuid is unset")
		}
	})

	// The load balancer can't be determined, so the operation must fail rather than run
	// unserialized: we can't tell which load balancer it would mutate.
	t.Run("errors when the pool cannot be read", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, `{"detail": "boom"}`, http.StatusInternalServerError)
		})
		client := testClient(t, mux)
		d := schema.TestResourceDataRaw(t, memberSchema, map[string]any{"pool_uuid": poolUUID})

		if _, err := lockKeyFromPoolUUID(context.Background(), d, client); err == nil {
			t.Error("expected an error when the pool cannot be read")
		}
	})

	// Every pool has a load balancer, so an empty UUID here is a malformed response
	// rather than a pool without one: the same failed resolution as above, just
	// detected in the payload instead of the transport.
	t.Run("errors when the pool has no load balancer", func(t *testing.T) {
		client := testClient(t, poolHandler(t, poolUUID, cloudscale.LoadBalancerPool{
			UUID: poolUUID,
		}))
		d := schema.TestResourceDataRaw(t, memberSchema, map[string]any{"pool_uuid": poolUUID})

		if _, err := lockKeyFromPoolUUID(context.Background(), d, client); err == nil {
			t.Error("expected an error when the pool has no load balancer")
		}
	})
}

// TestLoadBalancerSubResourcesShareLockKey is the regression test for the bug
// where pool members and listeners serialized on their pool while pools
// serialized on the load balancer, so operations on one load balancer never
// blocked each other. Every load balancer sub-resource must derive the same key.
func TestLoadBalancerSubResourcesShareLockKey(t *testing.T) {
	const (
		poolUUID = "pool-1"
		lbUUID   = "lb-1"
	)

	client := testClient(t, poolHandler(t, poolUUID, cloudscale.LoadBalancerPool{
		UUID:         poolUUID,
		LoadBalancer: cloudscale.LoadBalancerStub{UUID: lbUUID},
	}))

	tests := []struct {
		name        string
		resSchema   map[string]*schema.Schema
		raw         map[string]any
		lockKeyFunc mutexKeyFunc
	}{
		{
			name:        "pool",
			resSchema:   getLoadBalancerPoolSchema(RESOURCE),
			raw:         map[string]any{"load_balancer_uuid": lbUUID},
			lockKeyFunc: lockKeyFromLoadBalancerUUID,
		},
		{
			name:        "pool member",
			resSchema:   getLoadBalancerPoolMemberSchema(RESOURCE),
			raw:         map[string]any{"pool_uuid": poolUUID},
			lockKeyFunc: lockKeyFromPoolUUID,
		},
		{
			name:        "listener",
			resSchema:   getLoadBalancerListenerSchema(RESOURCE),
			raw:         map[string]any{"pool_uuid": poolUUID},
			lockKeyFunc: lockKeyFromPoolUUID,
		},
		{
			name:        "health monitor",
			resSchema:   getLoadBalancerHealthMonitorSchema(RESOURCE),
			raw:         map[string]any{"pool_uuid": poolUUID},
			lockKeyFunc: lockKeyFromPoolUUID,
		},
	}

	const want = "cloudscale/load-balancer/lb-1"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, tt.resSchema, tt.raw)

			key, err := tt.lockKeyFunc(context.Background(), d, client)
			if err != nil {
				t.Fatalf("%s: %s", tt.name, err)
			}
			if key != want {
				t.Errorf("%s locks on %q, want %q: every load balancer sub-resource must serialize on the same key",
					tt.name, key, want)
			}
		})
	}
}
