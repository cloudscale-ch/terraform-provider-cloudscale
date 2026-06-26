package cloudscale

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Source: https://developer.hashicorp.com/terraform/plugin/sdkv2/guides/v2-upgrade-guide#removal-of-helper-mutexkv-package
// License: public domain

// MutexKV is a simple key/value store for arbitrary mutexes. It can be used to
// serialize changes across arbitrary collaborators that share knowledge of the
// keys they must serialize on.
//
// The initial use case is to let aws_security_group_rule resources serialize
// their access to individual security groups based on SG ID.
type MutexKV struct {
	lock  sync.Mutex
	store map[string]*sync.Mutex
}

// LockContext tries to acquire the mutex for the given key, retrying until the
// context is done. Returns an error if the context deadline is exceeded before
// the lock is acquired.
func (m *MutexKV) LockContext(ctx context.Context, key string) error {
	log.Printf("[DEBUG] Locking %q", key)
	mu := m.get(key)
	for {
		if mu.TryLock() {
			log.Printf("[DEBUG] Locked %q", key)
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting to acquire lock %q: %w", key, ctx.Err())
		case <-time.After(50 * time.Millisecond):
		}
	}
}

// Unlock the mutex for the given key. Caller must have called LockContext for the same key first
func (m *MutexKV) Unlock(key string) {
	log.Printf("[DEBUG] Unlocking %q", key)
	m.get(key).Unlock()
	log.Printf("[DEBUG] Unlocked %q", key)
}

// Returns a mutex for the given key, no guarantee of its lock status
func (m *MutexKV) get(key string) *sync.Mutex {
	m.lock.Lock()
	defer m.lock.Unlock()
	mutex, ok := m.store[key]
	if !ok {
		mutex = &sync.Mutex{}
		m.store[key] = mutex
	}
	return mutex
}

// Returns a properly initialized MutexKV
func NewMutexKV() *MutexKV {
	return &MutexKV{
		store: make(map[string]*sync.Mutex),
	}
}

// globalMu is a provider-wide MutexKV instance shared across all resources.
// Resources use namespaced keys (e.g. "cloudscale/volume-snapshot/<uuid>") to
// prevent cross-resource interference while sharing a single store.
var globalMu = NewMutexKV()
