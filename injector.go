// Copyright 2025 Talina Shrotriya
// SPDX-License-Identifier: Apache-2.0

package faultinject

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	mu       sync.Mutex
	limits   = make(map[string]int) // old "fail first N" behavior
	precise  = make(map[string]int) // new "fail only on Nth call" behavior
	counters = make(map[string]int)
	
	// Environment control
	allowedEnvironments = []string{"development", "staging", "testing"}
	productionEnvironments = []string{"production", "prod"}
)

// SetAllowedEnvironments configures which environments allow fault injection
func SetAllowedEnvironments(envs []string) {
	mu.Lock()
	defer mu.Unlock()
	allowedEnvironments = envs
}

// SetProductionEnvironments configures which environments are considered production
func SetProductionEnvironments(envs []string) {
	mu.Lock()
	defer mu.Unlock()
	productionEnvironments = envs
}

// isProductionEnvironment checks if the current environment is production
func isProductionEnvironment() bool {
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	if env == "" {
		env = strings.ToLower(os.Getenv("ENV"))
	}
	if env == "" {
		env = strings.ToLower(os.Getenv("GO_ENV"))
	}
	
	// Check if it's explicitly marked as production
	for _, prodEnv := range productionEnvironments {
		if env == prodEnv {
			return true
		}
	}
	
	// Check if it's in allowed environments
	for _, allowedEnv := range allowedEnvironments {
		if env == allowedEnv {
			return false
		}
	}
	
	// Default to production if environment is not explicitly allowed
	return true
}

// Inject returns true if this key should fail.
//   - If precise[key] > 0, it fails *only* when counters[key] == precise[key].
//   - Otherwise if limits[key] > 0, it fails while counters[key] ≤ limits[key].
//   - Fault injection is disabled in production environments.
func Inject(key string) bool {
	// Disable fault injection in production
	if isProductionEnvironment() {
		return false
	}
	
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

// InjectWithFn executes the provided function if fault injection should occur
func InjectWithFn(key string, fn func() error) error {
	if Inject(key) {
		return fn()
	}
	return nil
}

// InjectWithFnContext executes the provided function if fault injection should occur (context-aware)
func InjectWithFnContext(ctx context.Context, key string, fn func() error) error {
	if InjectWithContext(ctx, key) {
		return fn()
	}
	return nil
}

// InjectWithError is a convenience function that returns an error if injection should occur
func InjectWithError(key string, message string) error {
	if Inject(key) {
		return fmt.Errorf("injected failure: %s", message)
	}
	return nil
}

// InjectWithErrorf is a convenience function that returns a formatted error if injection should occur
func InjectWithErrorf(key string, format string, args ...interface{}) error {
	if Inject(key) {
		return fmt.Errorf("injected failure: %s", fmt.Sprintf(format, args...))
	}
	return nil
}

// InjectWithContext checks for fault injection override in context
func InjectWithContext(ctx context.Context, key string) bool {
	// Check if context has fault injection override
	if ctx != nil {
		if ctx.Err() != nil {
			return false // Do not inject if context is cancelled
		}
		if override, ok := ctx.Value("faultinject:" + key).(bool); ok {
			return override
		}
	}
	return Inject(key)
}

// InjectWithContextError combines context checking with error return
func InjectWithContextError(ctx context.Context, key string, message string) error {
	if InjectWithContext(ctx, key) {
		return fmt.Errorf("injected failure: %s", message)
	}
	return nil
}

// SetFailures is the old API: fail the first `count` calls to key.
// Fault injection is disabled in production environments.
func SetFailures(key string, count int) {
	// Disable fault injection in production
	if isProductionEnvironment() {
		return
	}
	
	mu.Lock()
	defer mu.Unlock()
	limits[key] = count
	// clear any precise setting for this key
	delete(precise, key)
	counters[key] = 0
}

// SetNthFailure makes Inject(key) return true *only* on the Nth call.
// Fault injection is disabled in production environments.
func SetNthFailure(key string, nth int) {
	// Disable fault injection in production
	if isProductionEnvironment() {
		return
	}
	
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
