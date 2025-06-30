package faultinject

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPMiddleware(t *testing.T) {
	// Reset state before each test
	resetState()

	tests := []struct {
		name           string
		faultKey       string
		faultCount     int
		expectedStatus int
		expectedBody   string
		setup          func()
	}{
		{
			name:           "no fault configured",
			faultKey:       "nonexistent",
			expectedStatus: 200,
			expectedBody:   "success",
		},
		{
			name:           "fault configured - injects failure",
			faultKey:       "api-fault",
			faultCount:     1,
			expectedStatus: 500,
			expectedBody:   "Injected failure",
			setup: func() {
				SetFailures("api-fault", 1)
			},
		},
		{
			name:           "fault configured with count 0",
			faultKey:       "zero-fault",
			faultCount:     0,
			expectedStatus: 200,
			expectedBody:   "success",
			setup: func() {
				SetFailures("zero-fault", 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetState()
			if tt.setup != nil {
				tt.setup()
			}

			// Create test handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Write([]byte("success"))
			})

			// Create middleware
			middleware := HTTPMiddleware(tt.faultKey)
			wrappedHandler := middleware(handler)

			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			// Execute request
			wrappedHandler.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response body
			body := strings.TrimSpace(w.Body.String())
			if body != tt.expectedBody {
				t.Errorf("Expected body '%s', got '%s'", tt.expectedBody, body)
			}
		})
	}
}

func TestHTTPMiddlewareWithResponse(t *testing.T) {
	// Reset state before each test
	resetState()

	tests := []struct {
		name           string
		faultKey       string
		faultCount     int
		responseFn     func(http.ResponseWriter, *http.Request)
		expectedStatus int
		expectedBody   string
		expectedHeader string
		setup          func()
	}{
		{
			name:           "no fault configured",
			faultKey:       "nonexistent",
			expectedStatus: 200,
			expectedBody:   "success",
			responseFn: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(503)
				w.Write([]byte("custom error"))
			},
		},
		{
			name:           "fault configured - custom JSON response",
			faultKey:       "api-fault",
			faultCount:     1,
			expectedStatus: 503,
			expectedBody:   `{"error":"service unavailable"}`,
			responseFn: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(503)
				w.Write([]byte(`{"error":"service unavailable"}`))
			},
			setup: func() {
				SetFailures("api-fault", 1)
			},
		},
		{
			name:           "fault configured - custom headers",
			faultKey:       "retry-fault",
			faultCount:     1,
			expectedStatus: 503,
			expectedBody:   "retry later",
			expectedHeader: "30",
			responseFn: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Retry-After", "30")
				w.WriteHeader(503)
				w.Write([]byte("retry later"))
			},
			setup: func() {
				SetFailures("retry-fault", 1)
			},
		},
		{
			name:           "fault configured - timeout simulation",
			faultKey:       "timeout-fault",
			faultCount:     1,
			expectedStatus: 408,
			expectedBody:   "request timeout",
			responseFn: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(408)
				w.Write([]byte("request timeout"))
			},
			setup: func() {
				SetFailures("timeout-fault", 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetState()
			if tt.setup != nil {
				tt.setup()
			}

			// Create test handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Write([]byte("success"))
			})

			// Create middleware with custom response
			middleware := HTTPMiddlewareWithResponse(tt.faultKey, tt.responseFn)
			wrappedHandler := middleware(handler)

			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			// Execute request
			wrappedHandler.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response body
			body := strings.TrimSpace(w.Body.String())
			if body != tt.expectedBody {
				t.Errorf("Expected body '%s', got '%s'", tt.expectedBody, body)
			}

			// Check custom header if expected
			if tt.expectedHeader != "" {
				headerValue := w.Header().Get("Retry-After")
				if headerValue != tt.expectedHeader {
					t.Errorf("Expected Retry-After header '%s', got '%s'", tt.expectedHeader, headerValue)
				}
			}
		})
	}
}

func TestWithFaultInjection(t *testing.T) {
	// Reset state before each test
	resetState()

	tests := []struct {
		name       string
		faultKey   string
		faultCount int
		input      string
		expected   error
		setup      func()
	}{
		{
			name:       "no fault configured",
			faultKey:   "nonexistent",
			input:      "test input",
			expected:   nil,
		},
		{
			name:       "fault configured - injects failure",
			faultKey:   "func-fault",
			faultCount: 1,
			input:      "test input",
			expected:   fmt.Errorf("injected failure"),
			setup: func() {
				SetFailures("func-fault", 1)
			},
		},
		{
			name:       "fault configured with count 0",
			faultKey:   "zero-fault",
			faultCount: 0,
			input:      "test input",
			expected:   nil,
			setup: func() {
				SetFailures("zero-fault", 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetState()
			if tt.setup != nil {
				tt.setup()
			}

			// Create test function
			testFn := func(input string) error {
				return nil // Success
			}

			// Create decorated function
			decoratedFn := WithFaultInjection(tt.faultKey, testFn)

			// Execute decorated function
			err := decoratedFn(tt.input)

			// Check result
			if (err == nil && tt.expected != nil) || (err != nil && tt.expected == nil) {
				t.Errorf("Expected error %v, got %v", tt.expected, err)
			} else if err != nil && tt.expected != nil && err.Error() != tt.expected.Error() {
				t.Errorf("Expected error %v, got %v", tt.expected, err)
			}
		})
	}
}

func TestWithFaultInjectionContext(t *testing.T) {
	// Reset state before each test
	resetState()

	tests := []struct {
		name       string
		faultKey   string
		faultCount int
		input      string
		ctx        context.Context
		expected   error
		setup      func()
	}{
		{
			name:       "no fault configured",
			faultKey:   "nonexistent",
			input:      "test input",
			ctx:        context.Background(),
			expected:   nil,
		},
		{
			name:       "fault configured - injects failure",
			faultKey:   "ctx-fault",
			faultCount: 1,
			input:      "test input",
			ctx:        context.Background(),
			expected:   fmt.Errorf("injected failure"),
			setup: func() {
				SetFailures("ctx-fault", 1)
			},
		},
		{
			name:       "cancelled context - no fault injection",
			faultKey:   "ctx-fault",
			faultCount: 1,
			input:      "test input",
			ctx:        func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			expected:   nil,
			setup: func() {
				SetFailures("ctx-fault", 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetState()
			if tt.setup != nil {
				tt.setup()
			}

			// Create test function
			testFn := func(input string) error {
				return nil // Success
			}

			// Create decorated function
			decoratedFn := WithFaultInjectionContext(tt.faultKey, testFn)

			// Execute decorated function
			err := decoratedFn(tt.ctx, tt.input)

			// Check result
			if (err == nil && tt.expected != nil) || (err != nil && tt.expected == nil) {
				t.Errorf("Expected error %v, got %v", tt.expected, err)
			} else if err != nil && tt.expected != nil && err.Error() != tt.expected.Error() {
				t.Errorf("Expected error %v, got %v", tt.expected, err)
			}
		})
	}
}

func TestMiddlewareChaining(t *testing.T) {
	// Reset state before each test
	resetState()

	t.Run("multiple middleware layers", func(t *testing.T) {
		resetState()
		SetFailures("outer-fault", 1)
		SetFailures("inner-fault", 1)

		// Create test handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("success"))
		})

		// Create multiple middleware layers
		innerMiddleware := HTTPMiddleware("inner-fault")
		outerMiddleware := HTTPMiddleware("outer-fault")

		// Chain middleware
		wrappedHandler := outerMiddleware(innerMiddleware(handler))

		// Create test request
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Execute request
		wrappedHandler.ServeHTTP(w, req)

		// Should fail on first middleware (outer)
		if w.Code != 500 {
			t.Errorf("Expected status 500, got %d", w.Code)
		}

		if !strings.Contains(w.Body.String(), "Injected failure") {
			t.Errorf("Expected 'Injected failure' in response, got '%s'", w.Body.String())
		}
	})

	t.Run("mixed middleware types", func(t *testing.T) {
		resetState()
		SetFailures("default-fault", 1)

		// Create test handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("success"))
		})

		// Create mixed middleware
		defaultMiddleware := HTTPMiddleware("default-fault")
		customMiddleware := HTTPMiddlewareWithResponse("custom-fault", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(503)
			w.Write([]byte(`{"error":"custom"}`))
		})

		// Chain middleware
		wrappedHandler := defaultMiddleware(customMiddleware(handler))

		// Create test request
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Execute request
		wrappedHandler.ServeHTTP(w, req)

		// Should fail on first middleware (default)
		if w.Code != 500 {
			t.Errorf("Expected status 500, got %d", w.Code)
		}

		if !strings.Contains(w.Body.String(), "Injected failure") {
			t.Errorf("Expected 'Injected failure' in response, got '%s'", w.Body.String())
		}
	})
}

func TestMiddlewareRequestContext(t *testing.T) {
	// Reset state before each test
	resetState()

	t.Run("request context is passed through", func(t *testing.T) {
		resetState()
		SetFailures("context-fault", 1)

		// Create test handler that checks context
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add a value to context
			ctx := context.WithValue(r.Context(), "test-key", "test-value")
			r = r.WithContext(ctx)
			
			w.WriteHeader(200)
			w.Write([]byte("success"))
		})

		// Create middleware
		middleware := HTTPMiddleware("context-fault")
		wrappedHandler := middleware(handler)

		// Create test request with context
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := context.WithValue(context.Background(), "original-key", "original-value")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		// Execute request
		wrappedHandler.ServeHTTP(w, req)

		// Should fail due to fault injection
		if w.Code != 500 {
			t.Errorf("Expected status 500, got %d", w.Code)
		}
	})
}

 