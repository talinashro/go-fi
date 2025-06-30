// Copyright 2025 Talina Shrotriya
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/talinashro/go-fi/faultinject"
)

func main() {
	// Load fault injection configuration
	if err := faultinject.LoadSpec("faults.yaml"); err != nil {
		log.Fatalf("Failed to load fault spec: %v", err)
	}

	log.Println("=== Function-Based Fault Injection Examples ===")

	// Example 1: Simple function-based injection
	log.Println("1. Simple function-based injection:")
	if err := faultinject.InjectWithFn("db-insert", func() error {
		return fmt.Errorf("database connection failed")
	}); err != nil {
		log.Printf("   Error: %v", err)
	}

	// Example 2: Context-aware function injection
	log.Println("2. Context-aware function injection:")
	ctx := context.Background()
	if err := faultinject.InjectWithFnContext(ctx, "api-call", func() error {
		return fmt.Errorf("API call failed")
	}); err != nil {
		log.Printf("   Error: %v", err)
	}

	// Example 3: Complex error handling
	log.Println("3. Complex error handling:")
	if err := faultinject.InjectWithFn("payment-process", func() error {
		log.Println("   Simulating payment processing failure...")
		time.Sleep(100 * time.Millisecond)
		return fmt.Errorf("payment gateway timeout")
	}); err != nil {
		log.Printf("   Error: %v", err)
	}

	// Example 4: Database operations with custom logic
	log.Println("4. Database operations:")
	if err := faultinject.InjectWithFn("db-query", func() error {
		log.Println("   Simulating database query failure...")
		return fmt.Errorf("database connection pool exhausted")
	}); err != nil {
		log.Printf("   Error: %v", err)
	}

	// Example 5: External service calls
	log.Println("5. External service calls:")
	if err := faultinject.InjectWithFn("email-service", func() error {
		log.Println("   Simulating email service failure...")
		return fmt.Errorf("SMTP server unreachable")
	}); err != nil {
		log.Printf("   Error: %v", err)
	}

	// Example 6: Build tag helpers with functions
	log.Println("6. Build tag helpers with functions:")
	if err := faultinject.NoOpInjectWithFn("user-create", func() error {
		return fmt.Errorf("user creation failed")
	}); err != nil {
		log.Printf("   Error: %v", err)
	}

	log.Println("=== Examples completed ===")
}
