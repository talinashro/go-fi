// Copyright 2025 Talina Shrotriya
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
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

	// Create HTTP server with different middleware configurations
	mux := http.NewServeMux()

	// 1. Default middleware (500 error)
	mux.Handle("/api/users", faultinject.HTTPMiddleware("user-api")(userHandler))

	// 2. Custom JSON response
	mux.Handle("/api/payments", faultinject.HTTPMiddlewareWithResponse("payment-api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(503)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "payment service unavailable",
			"code":    "PAYMENT_DOWN",
			"retry":   "true",
			"timeout": "30s",
		})
	})(paymentHandler))

	// 3. Health check with custom status
	mux.Handle("/api/health", faultinject.HTTPMiddlewareWithResponse("health-check", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(503)
		w.Write([]byte("health check failed - service degraded"))
	})(healthHandler))

	// 4. Data API with retry headers
	mux.Handle("/api/data", faultinject.HTTPMiddlewareWithResponse("data-api", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Simulating data API failure...")
		w.Header().Set("Retry-After", "30")
		w.Header().Set("X-Failure-Reason", "database_connection")
		http.Error(w, "service temporarily unavailable", 503)
	})(dataHandler))

	// 5. Slow response simulation
	mux.Handle("/api/slow", faultinject.HTTPMiddlewareWithResponse("slow-api", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Simulating slow API response...")
		time.Sleep(5 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(408)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "request timeout",
			"code":  "TIMEOUT",
		})
	})(slowHandler))

	// 6. Custom error with logging
	mux.Handle("/api/critical", faultinject.HTTPMiddlewareWithResponse("critical-api", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("CRITICAL: API failure for %s", r.URL.Path)
		w.Header().Set("X-Error-ID", "CRITICAL_001")
		http.Error(w, "critical service failure", 500)
	})(criticalHandler))

	log.Println("HTTP server starting on :8080")
	log.Println("Available endpoints:")
	log.Println("  GET  /api/users     - Default 500 error")
	log.Println("  POST /api/payments  - Custom JSON response")
	log.Println("  GET  /api/health    - Health check failure")
	log.Println("  GET  /api/data      - Data API with retry headers")
	log.Println("  GET  /api/slow      - Slow response simulation")
	log.Println("  GET  /api/critical  - Critical error with logging")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

// Handler functions
func userHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "users retrieved successfully"})
}

func paymentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "payment processed successfully"})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  []string{"item1", "item2", "item3"},
		"count": 3,
	})
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "data retrieved successfully"})
}

func criticalHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "critical operation completed"})
}
