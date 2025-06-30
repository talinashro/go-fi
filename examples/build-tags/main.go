// Copyright 2025 Talina Shrotriya
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/talinashro/go-fi/faultinject"
)

func main() {
	log.Println("Starting application...")

	// Simulate some operations
	for i := 1; i <= 5; i++ {
		log.Printf("Operation %d:", i)

		if err := createUser(fmt.Sprintf("user%d@example.com", i)); err != nil {
			log.Printf("  Failed: %v", err)
		} else {
			log.Printf("  Succeeded")
		}

		time.Sleep(500 * time.Millisecond)
	}

	log.Println("Application completed.")
}

func createUser(email string) error {
	// Use basic Inject() - will be no-op in production builds
	if faultinject.Inject("user-create") {
		return fmt.Errorf("injected user creation failure")
	}

	// Simulate actual user creation logic
	log.Printf("  Creating user: %s", email)
	return nil
}
