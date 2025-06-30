package faultinject

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestStartControlServer(t *testing.T) {
	// Reset state before each test
	resetState()

	t.Run("start server on default port", func(t *testing.T) {
		resetState()

		// Start server
		StartControlServer(":0", nil)
		// Note: We can't easily test the server is running since it's in a goroutine
		// and doesn't return a server object
	})

	t.Run("start server with run handler", func(t *testing.T) {
		resetState()

		// Start server with a run handler
		runHandler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("run handler"))
		}
		StartControlServer(":0", runHandler)
	})
}

func TestServerEndpoints(t *testing.T) {
	// Reset state before each test
	resetState()

	t.Run("GET /status endpoint", func(t *testing.T) {
		resetState()

		// Set up some faults
		SetFailures("test-fault", 5)
		SetNthFailure("precise-fault", 3)

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/status" {
				json.NewEncoder(w).Encode(Status())
			} else {
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		// Make request
		resp, err := http.Get(server.URL + "/status")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check response
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse response
		var response map[string]int
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Check if faults are in response
		if response["test-fault"] != 5 {
			t.Errorf("Expected test-fault to be 5, got %d", response["test-fault"])
		}
	})

	t.Run("POST /set endpoint - add fault", func(t *testing.T) {
		resetState()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/set" {
				if r.Method != http.MethodPost {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				k := r.URL.Query().Get("key")
				c, _ := strconv.Atoi(r.URL.Query().Get("count"))
				SetFailures(k, c)
				w.Write([]byte("OK"))
			} else {
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		// Make request
		resp, err := http.Post(server.URL+"/set?key=new-fault&count=10", "text/plain", nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check response
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Check if fault was added
		status := Status()
		if status["new-fault"] != 10 {
			t.Errorf("Expected new-fault to be 10, got %d", status["new-fault"])
		}
	})

	t.Run("POST /reset endpoint", func(t *testing.T) {
		resetState()

		// Set up some faults first
		SetFailures("test-fault", 5)

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/reset" {
				Reset()
				w.Write([]byte("OK"))
			} else {
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		// Make request
		resp, err := http.Post(server.URL+"/reset", "text/plain", nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check response
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Check if faults were reset
		status := Status()
		if len(status) != 0 {
			t.Errorf("Expected no faults after reset, got %d", len(status))
		}
	})
}

func TestServerConcurrency(t *testing.T) {
	// Reset state before each test
	resetState()

	t.Run("concurrent fault additions", func(t *testing.T) {
		resetState()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/set" {
				if r.Method != http.MethodPost {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				k := r.URL.Query().Get("key")
				c, _ := strconv.Atoi(r.URL.Query().Get("count"))
				SetFailures(k, c)
				w.Write([]byte("OK"))
			} else {
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		// Add faults concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(id int) {
				url := fmt.Sprintf("%s/set?key=concurrent-fault-%d&count=%d", server.URL, id, id+1)
				resp, err := http.Post(url, "text/plain", nil)
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
		status := Status()
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("concurrent-fault-%d", i)
			if status[key] != i+1 {
				t.Errorf("Expected %s to be %d, got %d", key, i+1, status[key])
			}
		}
	})
}

func TestServerErrorHandling(t *testing.T) {
	// Reset state before each test
	resetState()

	t.Run("invalid count parameter", func(t *testing.T) {
		resetState()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/set" {
				if r.Method != http.MethodPost {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				k := r.URL.Query().Get("key")
				c, _ := strconv.Atoi(r.URL.Query().Get("count"))
				SetFailures(k, c)
				w.Write([]byte("OK"))
			} else {
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		// Make request with invalid count
		resp, err := http.Post(server.URL+"/set?key=test&count=invalid", "text/plain", nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Should still return OK (count becomes 0)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("unsupported HTTP method", func(t *testing.T) {
		resetState()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/set" {
				if r.Method != http.MethodPost {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				k := r.URL.Query().Get("key")
				c, _ := strconv.Atoi(r.URL.Query().Get("count"))
				SetFailures(k, c)
				w.Write([]byte("OK"))
			} else {
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		// Create PUT request (unsupported)
		req, err := http.NewRequest("PUT", server.URL+"/set", nil)
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
