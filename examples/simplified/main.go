// Copyright 2025 Talina Shrotriya
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/talinashro/go-fi/faultinject"
)

func main() {
	// Load fault injection configuration
	if err := faultinject.LoadSpec("faults.yaml"); err != nil {
		log.Fatalf("Failed to load fault spec: %v", err)
	}

	// Start HTTP server with fault injection middleware
	startHTTPServer()

	// Demonstrate different simplified approaches
	demonstrateSimplifiedApproaches()
}

func startHTTPServer() {
	mux := http.NewServeMux()

	// Use HTTP middleware for automatic fault injection
	mux.Handle("/api/users", faultinject.HTTPMiddleware("user-api")(userHandler))
	mux.Handle("/api/payments", faultinject.HTTPMiddlewareWithStatus("payment-api", 503)(paymentHandler))

	go func() {
		log.Println("HTTP server starting on :8080")
		if err := http.ListenAndServe(":8080", mux); err != nil {
			log.Fatal(err)
		}
	}()
}

func demonstrateSimplifiedApproaches() {
	log.Println("\n=== Simplified Fault Injection Examples ===")

	// 1. Error-returning functions
	log.Println("1. Error-returning functions:")
	if err := faultinject.InjectWithError("demo-error", "demonstration error"); err != nil {
		log.Printf("   Error: %v", err)
	}

	// 2. Database helpers
	log.Println("2. Database helpers:")
	if err := faultinject.PostgresInjector.InjectConnectionFailure(); err != nil {
		log.Printf("   Error: %v", err)
	}

	// 3. Function decorators
	log.Println("3. Function decorators:")
	createUserWithFaults := faultinject.WithFaultInjection("user-create", createUser)
	if err := createUserWithFaults(User{Name: "John"}); err != nil {
		log.Printf("   Error: %v", err)
	}

	// 4. Context-based overrides
	log.Println("4. Context-based overrides:")
	ctx := context.WithValue(context.Background(), "faultinject:db-insert", true)
	if err := faultinject.InjectWithContextError(ctx, "db-insert", "database failure"); err != nil {
		log.Printf("   Error: %v", err)
	}

	// 5. One-liner patterns
	log.Println("5. One-liner patterns:")
	if err := faultinject.PostgresInjector.WithFaultInjection("insert", func() error {
		return fmt.Errorf("simulated database insert")
	}); err != nil {
		log.Printf("   Error: %v", err)
	}

	time.Sleep(2 * time.Second)
}

// HTTP handlers
func userHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("User handler called")
	w.Write([]byte(`{"message": "user created"}`))
}

func paymentHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Payment handler called")
	w.Write([]byte(`{"message": "payment processed"}`))
}

// Example functions
type User struct {
	Name string
}

func createUser(user User) error {
	log.Printf("Creating user: %s", user.Name)
	return nil
} 