// Copyright 2025 Talina Shrotriya
// SPDX-License-Identifier: Apache-2.0

package faultinject

import (
	"context"
	"fmt"
	"net/http"
)

// HTTPMiddleware creates middleware that injects failures for HTTP requests
// Returns 500 status code by default when fault injection triggers
func HTTPMiddleware(key string) func(http.Handler) http.Handler {
	return HTTPMiddlewareWithResponse(key, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Injected failure", http.StatusInternalServerError)
	})
}

// HTTPMiddlewareWithResponse creates middleware with custom response handling
func HTTPMiddlewareWithResponse(key string, responseFn func(http.ResponseWriter, *http.Request)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if Inject(key) {
				responseFn(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Decorator is a generic function decorator that injects failures
type Decorator[T any] func(T) error

// WithFaultInjection decorates a function with fault injection
func WithFaultInjection[T any](key string, fn func(T) error) Decorator[T] {
	return func(input T) error {
		if Inject(key) {
			return fmt.Errorf("injected failure")
		}
		return fn(input)
	}
}

// WithFaultInjectionContext decorates a function with context-aware fault injection
func WithFaultInjectionContext[T any](key string, fn func(T) error) func(context.Context, T) error {
	return func(ctx context.Context, input T) error {
		if InjectWithContext(ctx, key) {
			return fmt.Errorf("injected failure")
		}
		return fn(input)
	}
} 