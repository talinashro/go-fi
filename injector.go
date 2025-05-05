// Copyright 2025 Talina Shrotriya
// SPDX-License-Identifier: Apache-2.0

// Package faultinject provides in-process “fail N times” injection.
package faultinject

import "sync"

var (
	mu       sync.Mutex
	limits   = make(map[string]int)
	counters = make(map[string]int)
)

// Inject returns true if this key should fail (up to its configured limit).
func Inject(key string) bool {
	mu.Lock()
	defer mu.Unlock()

	limit, ok := limits[key]
	if !ok || limit == 0 {
		return false
	}
	if counters[key] < limit {
		counters[key]++
		return true
	}
	return false
}

// SetFailures sets the failure limit for a key and resets its counter.
func SetFailures(key string, count int) {
	mu.Lock()
	defer mu.Unlock()
	limits[key] = count
	counters[key] = 0
}

// Reset clears all failure limits and counters.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	limits = make(map[string]int)
	counters = make(map[string]int)
}

// Status returns how many failures remain per key.
func Status() map[string]int {
	mu.Lock()
	defer mu.Unlock()
	out := make(map[string]int, len(limits))
	for k, limit := range limits {
		used := counters[k]
		rem := limit - used
		if rem < 0 {
			rem = 0
		}
		out[k] = rem
	}
	return out
}
