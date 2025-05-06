// Copyright 2025 Talina Shrotriya
// SPDX-License-Identifier: Apache-2.0

package faultinject

import "sync"

var (
	mu       sync.Mutex
	limits   = make(map[string]int) // old “fail first N” behavior
	precise  = make(map[string]int) // new “fail only on Nth call” behavior
	counters = make(map[string]int)
)

// Inject returns true if this key should fail.
//   - If precise[key] > 0, it fails *only* when counters[key] == precise[key].
//   - Otherwise if limits[key] > 0, it fails while counters[key] ≤ limits[key].
func Inject(key string) bool {
	mu.Lock()
	defer mu.Unlock()

	// bump attempt count
	cnt := counters[key] + 1
	counters[key] = cnt

	// precise-nth behavior takes priority
	if nth, ok := precise[key]; ok && nth > 0 {
		return cnt == nth
	}

	// fallback: first-N failures
	if lim, ok := limits[key]; ok && lim > 0 {
		return cnt <= lim
	}

	return false
}

// SetFailures is the old API: fail the first `count` calls to key.
func SetFailures(key string, count int) {
	mu.Lock()
	defer mu.Unlock()
	limits[key] = count
	// clear any precise setting for this key
	delete(precise, key)
	counters[key] = 0
}

// SetNthFailure makes Inject(key) return true *only* on the Nth call.
func SetNthFailure(key string, nth int) {
	mu.Lock()
	defer mu.Unlock()
	precise[key] = nth
	// clear any first-N setting for this key
	delete(limits, key)
	counters[key] = 0
}

// Reset clears all configured behaviors and counters.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	limits = make(map[string]int)
	precise = make(map[string]int)
	counters = make(map[string]int)
}

// Status returns remaining "first-N" failures per key.
func Status() map[string]int {
	mu.Lock()
	defer mu.Unlock()
	out := make(map[string]int, len(limits))
	for k, lim := range limits {
		used := counters[k]
		rem := lim - used
		if rem < 0 {
			rem = 0
		}
		out[k] = rem
	}
	return out
}
