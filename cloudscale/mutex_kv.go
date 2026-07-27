package cloudscale

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// MutexKV is a simple key/value store for per-key mutexes. It can be used to
// serialize changes across collaborators that share knowledge of the keys they
// must serialize on.
//
// Each key is backed by a buffered channel of size 1. Sending to the channel
// acquires the lock; receiving from it releases it. The channel size of 1 is
// the critical invariant: it allows exactly one sender to hold the lock at a
// time, while any additional senders block until the slot is freed.
//
// Inspired by the Hashicorp helper/mutexkv package (public domain,
// https://developer.hashicorp.com/terraform/plugin/sdkv2/guides/v2-upgrade-guide#removal-of-helper-mutexkv-package), but
// rewritten to use channels (the idiomatic Go approach to synchronization:
// "don't communicate by sharing memory; share memory by communicating")
// so that LockContext can respect context cancellation without polling.
type MutexKV struct {
	lock  sync.Mutex
	store map[string]chan struct{}
}

// LockContext tries to acquire the mutex for the given key, blocking until the
// lock is acquired or the context is done. Returns an error if the context is
// cancelled or its deadline is exceeded before the lock is acquired.
func (m *MutexKV) LockContext(ctx context.Context, key string) error {
	log.Printf("[DEBUG] Locking %q", key)
	c := m.get(key)
	select {
	case c <- struct{}{}:
		log.Printf("[DEBUG] Locked %q", key)
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timed out waiting to acquire lock %q: %w", key, ctx.Err())
	}
}

// Unlock the mutex for the given key. Caller must have called LockContext for the same key first
func (m *MutexKV) Unlock(key string) {
	log.Printf("[DEBUG] Unlocking %q", key)
	select {
	case <-m.get(key):
		log.Printf("[DEBUG] Unlocked %q", key)
	default:
		panic(fmt.Sprintf("Unlock of unlocked key %q", key))
	}
}

// Returns a channel for the given key, no guarantee of its lock status
func (m *MutexKV) get(key string) chan struct{} {
	m.lock.Lock()
	defer m.lock.Unlock()
	c, ok := m.store[key]
	if !ok {
		c = make(chan struct{}, 1) // size 1: allows exactly one holder at a time
		m.store[key] = c
	}
	return c
}

// Returns a properly initialized MutexKV
func NewMutexKV() *MutexKV {
	return &MutexKV{
		store: make(map[string]chan struct{}),
	}
}

// globalMu is a provider-wide MutexKV instance shared across all resources.
// Resources use namespaced keys (e.g. "cloudscale/volume-snapshot/<uuid>") to
// prevent cross-resource interference while sharing a single store.
var globalMu = NewMutexKV()
