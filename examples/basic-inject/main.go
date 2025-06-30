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

	log.Println("=== Basic Inject() Function Examples ===")

	// Example 1: Simple error injection
	log.Println("1. Simple error injection:")
	if err := createUser("john@example.com"); err != nil {
		log.Printf("   Error: %v", err)
	}

	// Example 2: Database operations
	log.Println("2. Database operations:")
	if err := connectToDatabase(); err != nil {
		log.Printf("   Error: %v", err)
	}

	// Example 3: API calls
	log.Println("3. API calls:")
	if err := callExternalAPI(); err != nil {
		log.Printf("   Error: %v", err)
	}

	// Example 4: Complex error handling
	log.Println("4. Complex error handling:")
	if err := processPayment(); err != nil {
		log.Printf("   Error: %v", err)
	}

	// Example 5: Context-aware injection
	log.Println("5. Context-aware injection:")
	ctx := context.WithValue(context.Background(), "faultinject:email-send", true)
	if err := sendEmail(ctx, "test@example.com"); err != nil {
		log.Printf("   Error: %v", err)
	}

	// Example 6: HTTP handler
	log.Println("6. HTTP handler:")
	startHTTPServer()

	time.Sleep(2 * time.Second)
}

// Example 1: Simple error injection
func createUser(email string) error {
	if faultinject.Inject("user-create") {
		return fmt.Errorf("user creation failed")
	}
	log.Println("   User created successfully")
	return nil
}

// Example 2: Database operations
func connectToDatabase() error {
	if faultinject.Inject("db-connect") {
		return fmt.Errorf("database connection failed")
	}
	log.Println("   Database connected successfully")
	return nil
}

// Example 3: API calls
func callExternalAPI() error {
	if faultinject.Inject("api-call") {
		return fmt.Errorf("API call failed: timeout")
	}
	log.Println("   API call successful")
	return nil
}

// Example 4: Complex error handling
func processPayment() error {
	if faultinject.Inject("payment-process") {
		log.Println("   Simulating payment processing failure...")
		time.Sleep(100 * time.Millisecond)
		return fmt.Errorf("payment gateway timeout")
	}
	log.Println("   Payment processed successfully")
	return nil
}

// Example 5: Context-aware injection
func sendEmail(ctx context.Context, email string) error {
	// Check context override first, then use Inject
	if ctx.Value("faultinject:email-send") == true || faultinject.Inject("email-send") {
		return fmt.Errorf("email sending failed")
	}
	log.Printf("   Email sent to %s successfully", email)
	return nil
}

// Example 6: HTTP handler
func startHTTPServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/users", userHandler)

	go func() {
		log.Println("   HTTP server starting on :8080")
		if err := http.ListenAndServe(":8080", mux); err != nil {
			log.Printf("   HTTP server error: %v", err)
		}
	}()
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	if faultinject.Inject("user-handler") {
		http.Error(w, "handler failure", 500)
		return
	}
	w.Write([]byte(`{"message": "user handler success"}`))
}
