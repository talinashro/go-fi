package faultinject

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestInject(t *testing.T) {
	// Reset state before each test
	resetState()

	tests := []struct {
		name     string
		key      string
		expected bool
		setup    func()
	}{
		{
			name:     "no fault configured",
			key:      "nonexistent",
			expected: false,
		},
		{
			name:     "fault configured with count 1",
			key:      "test-fault",
			expected: true,
			setup: func() {
				failures["test-fault"] = 1
			},
		},
		{
			name:     "fault configured with count 0",
			key:      "zero-fault",
			expected: false,
			setup: func() {
				failures["zero-fault"] = 0
			},
		},
		{
			name:     "fault configured with negative count",
			key:      "negative-fault",
			expected: false,
			setup: func() {
				failures["negative-fault"] = -1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetState()
			if tt.setup != nil {
				tt.setup()
			}

			result := Inject(tt.key)
			if result != tt.expected {
				t.Errorf("Inject(%q) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestInjectWithContext(t *testing.T) {
	// Reset state before each test
	resetState()

	tests := []struct {
		name     string
		key      string
		ctx      context.Context
		expected bool
		setup    func()
	}{
		{
			name:     "no fault configured",
			key:      "nonexistent",
			ctx:      context.Background(),
			expected: false,
		},
		{
			name:     "fault configured with count 1",
			key:      "test-fault",
			ctx:      context.Background(),
			expected: true,
			setup: func() {
				failures["test-fault"] = 1
			},
		},
		{
			name:     "context with timeout",
			key:      "timeout-fault",
			ctx:      context.Background(),
			expected: true,
			setup: func() {
				failures["timeout-fault"] = 1
			},
		},
		{
			name:     "cancelled context",
			key:      "cancelled-fault",
			ctx:      func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			expected: false,
			setup: func() {
				failures["cancelled-fault"] = 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetState()
			if tt.setup != nil {
				tt.setup()
			}

			result := InjectWithContext(tt.ctx, tt.key)
			if result != tt.expected {
				t.Errorf("InjectWithContext(ctx, %q) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestPreciseInject(t *testing.T) {
	// Reset state before each test
	resetState()

	tests := []struct {
		name     string
		key      string
		expected bool
		setup    func()
	}{
		{
			name:     "no precise fault configured",
			key:      "nonexistent",
			expected: false,
		},
		{
			name:     "precise fault configured with count 1",
			key:      "precise-fault",
			expected: true,
			setup: func() {
				preciseFailures["precise-fault"] = 1
			},
		},
		{
			name:     "precise fault configured with count 0",
			key:      "zero-precise",
			expected: false,
			setup: func() {
				preciseFailures["zero-precise"] = 0
			},
		},
		{
			name:     "precise fault configured with negative count",
			key:      "negative-precise",
			expected: false,
			setup: func() {
				preciseFailures["negative-precise"] = -1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetState()
			if tt.setup != nil {
				tt.setup()
			}

			result := PreciseInject(tt.key)
			if result != tt.expected {
				t.Errorf("PreciseInject(%q) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestPreciseInjectWithContext(t *testing.T) {
	// Reset state before each test
	resetState()

	tests := []struct {
		name     string
		key      string
		ctx      context.Context
		expected bool
		setup    func()
	}{
		{
			name:     "no precise fault configured",
			key:      "nonexistent",
			ctx:      context.Background(),
			expected: false,
		},
		{
			name:     "precise fault configured with count 1",
			key:      "precise-fault",
			ctx:      context.Background(),
			expected: true,
			setup: func() {
				preciseFailures["precise-fault"] = 1
			},
		},
		{
			name:     "cancelled context",
			key:      "cancelled-precise",
			ctx:      func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			expected: false,
			setup: func() {
				preciseFailures["cancelled-precise"] = 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetState()
			if tt.setup != nil {
				tt.setup()
			}

			result := PreciseInjectWithContext(tt.ctx, tt.key)
			if result != tt.expected {
				t.Errorf("PreciseInjectWithContext(ctx, %q) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestEnvironmentControl(t *testing.T) {
	// Reset state before each test
	resetState()

	tests := []struct {
		name           string
		environment    string
		expectedResult bool
		setup          func()
		cleanup        func()
	}{
		{
			name:           "production environment - fault injection disabled",
			environment:    "production",
			expectedResult: false,
			setup: func() {
				os.Setenv("ENVIRONMENT", "production")
				failures["test-fault"] = 1
			},
			cleanup: func() {
				os.Unsetenv("ENVIRONMENT")
			},
		},
		{
			name:           "development environment - fault injection enabled",
			environment:    "development",
			expectedResult: true,
			setup: func() {
				os.Setenv("ENVIRONMENT", "development")
				failures["test-fault"] = 1
			},
			cleanup: func() {
				os.Unsetenv("ENVIRONMENT")
			},
		},
		{
			name:           "no environment set - fault injection enabled",
			environment:    "",
			expectedResult: true,
			setup: func() {
				failures["test-fault"] = 1
			},
		},
		{
			name:           "custom production environment - fault injection disabled",
			environment:    "prod",
			expectedResult: false,
			setup: func() {
				os.Setenv("ENVIRONMENT", "prod")
				allowedEnvironments = []string{"dev", "staging", "test"}
				failures["test-fault"] = 1
			},
			cleanup: func() {
				os.Unsetenv("ENVIRONMENT")
				allowedEnvironments = defaultAllowedEnvironments
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetState()
			if tt.setup != nil {
				tt.setup()
			}
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			result := Inject("test-fault")
			if result != tt.expectedResult {
				t.Errorf("Inject() in %s environment = %v, want %v", tt.environment, result, tt.expectedResult)
			}
		})
	}
}

func TestFaultCounting(t *testing.T) {
	// Reset state before each test
	resetState()

	t.Run("fault count decreases with each call", func(t *testing.T) {
		resetState()
		failures["count-test"] = 3

		// First call should succeed (inject fault)
		if !Inject("count-test") {
			t.Error("First call should inject fault")
		}

		// Second call should succeed
		if !Inject("count-test") {
			t.Error("Second call should inject fault")
		}

		// Third call should succeed
		if !Inject("count-test") {
			t.Error("Third call should inject fault")
		}

		// Fourth call should fail (no more faults)
		if Inject("count-test") {
			t.Error("Fourth call should not inject fault")
		}
	})

	t.Run("precise fault count decreases with each call", func(t *testing.T) {
		resetState()
		preciseFailures["precise-count-test"] = 2

		// First call should succeed
		if !PreciseInject("precise-count-test") {
			t.Error("First call should inject fault")
		}

		// Second call should succeed
		if !PreciseInject("precise-count-test") {
			t.Error("Second call should inject fault")
		}

		// Third call should fail
		if PreciseInject("precise-count-test") {
			t.Error("Third call should not inject fault")
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	// Reset state before each test
	resetState()

	t.Run("concurrent fault injection", func(t *testing.T) {
		resetState()
		failures["concurrent-test"] = 100

		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				Inject("concurrent-test")
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Should have injected exactly 100 faults
		if failures["concurrent-test"] != 0 {
			t.Errorf("Expected 0 remaining faults, got %d", failures["concurrent-test"])
		}
	})

	t.Run("concurrent precise fault injection", func(t *testing.T) {
		resetState()
		preciseFailures["concurrent-precise"] = 50

		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				PreciseInject("concurrent-precise")
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Should have injected exactly 50 faults
		if preciseFailures["concurrent-precise"] != 0 {
			t.Errorf("Expected 0 remaining precise faults, got %d", preciseFailures["concurrent-precise"])
		}
	})
}

func TestContextTimeout(t *testing.T) {
	// Reset state before each test
	resetState()

	t.Run("context with timeout", func(t *testing.T) {
		resetState()
		failures["timeout-test"] = 1

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Should inject fault before timeout
		if !InjectWithContext(ctx, "timeout-test") {
			t.Error("Should inject fault before timeout")
		}
	})

	t.Run("context already cancelled", func(t *testing.T) {
		resetState()
		failures["cancelled-test"] = 1

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Should not inject fault due to cancelled context
		if InjectWithContext(ctx, "cancelled-test") {
			t.Error("Should not inject fault with cancelled context")
		}
	})
}

// Helper function to reset internal state for testing
func resetState() {
	failures = make(map[string]int)
	preciseFailures = make(map[string]int)
	allowedEnvironments = defaultAllowedEnvironments
	productionEnvironments = defaultProductionEnvironments
} 