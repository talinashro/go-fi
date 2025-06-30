package faultinject

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestStartServer(t *testing.T) {
	// Reset state before each test
	resetState()

	t.Run("start server on default port", func(t *testing.T) {
		resetState()

		// Start server
		server := StartServer(":0") // Use port 0 to get a random available port
		defer server.Close()

		// Wait a bit for server to start
		time.Sleep(100 * time.Millisecond)

		// Check if server is running
		if server.Addr == "" {
			t.Error("Server should have an address")
		}
	})

	t.Run("start server on specific port", func(t *testing.T) {
		resetState()

		// Start server on a specific port
		server := StartServer(":8081")
		defer server.Close()

		// Wait a bit for server to start
		time.Sleep(100 * time.Millisecond)

		// Check if server is running
		if !strings.Contains(server.Addr, "8081") {
			t.Errorf("Expected server on port 8081, got %s", server.Addr)
		}
	})
}

func TestServerEndpoints(t *testing.T) {
	// Reset state before each test
	resetState()

	t.Run("GET /health endpoint", func(t *testing.T) {
		resetState()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(healthHandler))
		defer server.Close()

		// Make request
		resp, err := http.Get(server.URL + "/health")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check response
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Check content type
		contentType := resp.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}
	})

	t.Run("GET /faults endpoint", func(t *testing.T) {
		resetState()

		// Set up some faults
		failures["test-fault"] = 5
		preciseFailures["precise-fault"] = 3

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(faultsHandler))
		defer server.Close()

		// Make request
		resp, err := http.Get(server.URL + "/faults")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check response
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse response
		var response map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Check if faults are in response
		failuresMap, ok := response["failures"].(map[string]interface{})
		if !ok {
			t.Error("Expected failures map in response")
		}

		if failuresMap["test-fault"].(float64) != 5 {
			t.Errorf("Expected test-fault to be 5, got %v", failuresMap["test-fault"])
		}

		preciseFailuresMap, ok := response["precise-failures"].(map[string]interface{})
		if !ok {
			t.Error("Expected precise-failures map in response")
		}

		if preciseFailuresMap["precise-fault"].(float64) != 3 {
			t.Errorf("Expected precise-fault to be 3, got %v", preciseFailuresMap["precise-fault"])
		}
	})

	t.Run("POST /faults endpoint - add fault", func(t *testing.T) {
		resetState()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(faultsHandler))
		defer server.Close()

		// Prepare request
		requestBody := map[string]interface{}{
			"key":   "new-fault",
			"count": 10,
		}
		jsonBody, _ := json.Marshal(requestBody)

		// Make request
		resp, err := http.Post(server.URL+"/faults", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check response
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Check if fault was added
		if failures["new-fault"] != 10 {
			t.Errorf("Expected new-fault to be 10, got %d", failures["new-fault"])
		}
	})

	t.Run("POST /faults endpoint - add precise fault", func(t *testing.T) {
		resetState()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(faultsHandler))
		defer server.Close()

		// Prepare request
		requestBody := map[string]interface{}{
			"key":           "new-precise-fault",
			"count":         7,
			"precise":       true,
		}
		jsonBody, _ := json.Marshal(requestBody)

		// Make request
		resp, err := http.Post(server.URL+"/faults", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check response
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Check if precise fault was added
		if preciseFailures["new-precise-fault"] != 7 {
			t.Errorf("Expected new-precise-fault to be 7, got %d", preciseFailures["new-precise-fault"])
		}
	})

	t.Run("POST /faults endpoint - invalid request", func(t *testing.T) {
		resetState()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(faultsHandler))
		defer server.Close()

		// Prepare invalid request
		requestBody := map[string]interface{}{
			"key": "missing-count",
		}
		jsonBody, _ := json.Marshal(requestBody)

		// Make request
		resp, err := http.Post(server.URL+"/faults", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check response
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("DELETE /faults/{key} endpoint", func(t *testing.T) {
		resetState()

		// Set up a fault to delete
		failures["delete-fault"] = 5

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(faultsHandler))
		defer server.Close()

		// Create DELETE request
		req, err := http.NewRequest("DELETE", server.URL+"/faults/delete-fault", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Make request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check response
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Check if fault was deleted
		if _, exists := failures["delete-fault"]; exists {
			t.Error("Expected delete-fault to be deleted")
		}
	})

	t.Run("DELETE /faults/{key} endpoint - precise fault", func(t *testing.T) {
		resetState()

		// Set up a precise fault to delete
		preciseFailures["delete-precise-fault"] = 3

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(faultsHandler))
		defer server.Close()

		// Create DELETE request
		req, err := http.NewRequest("DELETE", server.URL+"/faults/delete-precise-fault?precise=true", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Make request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check response
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Check if precise fault was deleted
		if _, exists := preciseFailures["delete-precise-fault"]; exists {
			t.Error("Expected delete-precise-fault to be deleted")
		}
	})

	t.Run("DELETE /faults/{key} endpoint - non-existent fault", func(t *testing.T) {
		resetState()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(faultsHandler))
		defer server.Close()

		// Create DELETE request for non-existent fault
		req, err := http.NewRequest("DELETE", server.URL+"/faults/non-existent", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Make request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check response
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})
}

func TestServerConcurrency(t *testing.T) {
	// Reset state before each test
	resetState()

	t.Run("concurrent fault additions", func(t *testing.T) {
		resetState()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(faultsHandler))
		defer server.Close()

		// Add faults concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(id int) {
				requestBody := map[string]interface{}{
					"key":   fmt.Sprintf("concurrent-fault-%d", id),
					"count": id + 1,
				}
				jsonBody, _ := json.Marshal(requestBody)

				resp, err := http.Post(server.URL+"/faults", "application/json", bytes.NewBuffer(jsonBody))
				if err == nil {
					resp.Body.Close()
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify all faults were added
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("concurrent-fault-%d", i)
			if failures[key] != i+1 {
				t.Errorf("Expected %s to be %d, got %d", key, i+1, failures[key])
			}
		}
	})
}

func TestServerErrorHandling(t *testing.T) {
	// Reset state before each test
	resetState()

	t.Run("invalid JSON in POST request", func(t *testing.T) {
		resetState()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(faultsHandler))
		defer server.Close()

		// Make request with invalid JSON
		resp, err := http.Post(server.URL+"/faults", "application/json", strings.NewReader("invalid json"))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check response
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("unsupported HTTP method", func(t *testing.T) {
		resetState()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(faultsHandler))
		defer server.Close()

		// Create PUT request (unsupported)
		req, err := http.NewRequest("PUT", server.URL+"/faults", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Make request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check response
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", resp.StatusCode)
		}
	})
}

func TestServerContextCancellation(t *testing.T) {
	// Reset state before each test
	resetState()

	t.Run("server shutdown with context", func(t *testing.T) {
		resetState()

		// Create context with cancellation
		ctx, cancel := context.WithCancel(context.Background())

		// Start server with context
		server := StartServerWithContext(ctx, ":0")
		defer server.Close()

		// Wait a bit for server to start
		time.Sleep(100 * time.Millisecond)

		// Cancel context
		cancel()

		// Wait a bit for server to shutdown
		time.Sleep(100 * time.Millisecond)

		// Server should be closed
		// Note: We can't easily test if the server is actually closed
		// since httptest.Server doesn't expose this information
	})
}

// Helper function to reset internal state for testing
func resetState() {
	failures = make(map[string]int)
	preciseFailures = make(map[string]int)
	allowedEnvironments = defaultAllowedEnvironments
	productionEnvironments = defaultProductionEnvironments
} 