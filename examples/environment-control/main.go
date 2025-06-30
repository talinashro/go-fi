// Copyright 2025 Talina Shrotriya
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/talinashro/go-fi/faultinject"
)

func main() {
	// Get current environment
	env := getEnvironment()
	log.Printf("Current environment: %s", env)

	// Configure fault injection
	faultinject.SetFailures("db-connect", 2)
	faultinject.SetFailures("api-call", 1)

	log.Println("=== Environment-Based Fault Injection Demo ===")

	// Test database connection
	log.Println("1. Testing database connection:")
	for i := 1; i <= 3; i++ {
		if err := connectToDatabase(); err != nil {
			log.Printf("   Attempt %d: %v", i, err)
		} else {
			log.Printf("   Attempt %d: Success", i)
		}
	}

	// Test API call
	log.Println("2. Testing API call:")
	if err := callAPI(); err != nil {
		log.Printf("   Error: %v", err)
	} else {
		log.Println("   Success")
	}

	// Show current status
	log.Printf("3. Current fault injection status: %v", faultinject.Status())
}

func getEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = os.Getenv("ENV")
	}
	if env == "" {
		env = os.Getenv("GO_ENV")
	}
	if env == "" {
		env = "unknown"
	}
	return env
}

func connectToDatabase() error {
	if faultinject.Inject("db-connect") {
		return fmt.Errorf("database connection failed")
	}
	return nil
}

func callAPI() error {
	if faultinject.Inject("api-call") {
		return fmt.Errorf("API call failed")
	}
	return nil
}
