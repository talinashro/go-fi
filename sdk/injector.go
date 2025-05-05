package sdk

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

// mu protects both maps
var mu sync.Mutex

// limits[key] = how many times to fail
// counters[key] = how many times we've been called
var (
	limits   map[string]int
	counters map[string]int
)

func init() {
	limits = parseEnv(os.Getenv("FI_FAILURE_COUNTS"))
	counters = make(map[string]int)
}

// parseEnv("EC2:1,STORAGE:0") → map{"EC2":1, "STORAGE":0}
func parseEnv(raw string) map[string]int {
	m := make(map[string]int)
	for _, part := range strings.Split(raw, ",") {
		if kv := strings.SplitN(part, ":", 2); len(kv) == 2 {
			if n, err := strconv.Atoi(kv[1]); err == nil {
				m[kv[0]] = n
			}
		}
	}
	return m
}

// Inject returns true if this call for `key` should fail (up to the configured limit).
func Inject(key string) bool {
	mu.Lock()
	defer mu.Unlock()

	// how many failures we should inject for this key
	limit, has := limits[key]
	if !has || limit <= 0 {
		return false
	}

	// how many times we already injected
	used := counters[key]
	if used < limit {
		counters[key] = used + 1
		return true
	}
	return false
}

// SetFailures overrides the failure limit for key at runtime.
func SetFailures(key string, count int) {
	mu.Lock()
	defer mu.Unlock()
	limits[key] = count
	counters[key] = 0 // reset any previous usage
}

// Reset clears all limits and counters.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	limits = make(map[string]int)
	counters = make(map[string]int)
}

// Status returns a snapshot of remaining failures per key.
func Status() map[string]int {
	mu.Lock()
	defer mu.Unlock()
	out := make(map[string]int, len(limits))
	for key, limit := range limits {
		used := counters[key]
		rem := limit - used
		if rem < 0 {
			rem = 0
		}
		out[key] = rem
	}
	return out
}
