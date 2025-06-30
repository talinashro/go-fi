# go-fi

[![Go Report Card](https://goreportcard.com/badge/github.com/talinashro/go-fi)](https://goreportcard.com/report/github.com/talinashro/go-fi)
[![Go Reference](https://pkg.go.dev/badge/github.com/talinashro/go-fi.svg)](https://pkg.go.dev/github.com/talinashro/go-fi)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A lightweight, thread-safe Go library for in-process fault injection. Perfect for testing error handling, chaos engineering, and simulating real-world failure scenarios.

## Features

- **Simple API**: One function call to inject failures anywhere
- **Two failure modes**: First-N failures or precise Nth failure
- **Runtime control**: Dynamically configure failures without restarting
- **YAML configuration**: Declarative failure specifications
- **HTTP control server**: Remote management via REST API
- **Thread-safe**: Safe for concurrent use
- **Production-safe**: Automatically disabled in production environments

## Installation

```bash
go get github.com/talinashro/go-fi@latest
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "github.com/talinashro/go-fi/faultinject"
)

func main() {
    // Configure a failure: fail the first 2 calls to "database-connect"
    faultinject.SetFailures("database-connect", 2)
    
    // Simulate database operations
    for i := 0; i < 5; i++ {
        if err := connectToDatabase(); err != nil {
            log.Printf("Database connection %d failed: %v", i+1, err)
        } else {
            log.Printf("Database connection %d succeeded", i+1)
        }
    }
}

func connectToDatabase() error {
    if faultinject.Inject("database-connect") {
        return fmt.Errorf("injected database connection failure")
    }
    return nil // Your actual database connection logic
}
```

**Output:**
```
Database connection 1 failed: injected database connection failure
Database connection 2 failed: injected database connection failure
Database connection 3 succeeded
Database connection 4 succeeded
Database connection 5 succeeded
```

## API Reference

### Core Functions

```go
// Inject a failure
if faultinject.Inject("my-operation") {
    return fmt.Errorf("injected failure")
}

// Configure failures
faultinject.SetFailures("db-connect", 3)     // Fail first 3 calls
faultinject.SetNthFailure("api-call", 5)     // Fail only 5th call
faultinject.Reset()                          // Clear all failures

// Check status
status := faultinject.Status()               // Returns remaining counts
```

### Context-Aware Injection

```go
// Check context override first, then use Inject
if faultinject.InjectWithContext(ctx, "db-insert") {
    return fmt.Errorf("database connection failed")
}
```

### HTTP Middleware

```go
// Default 500 error
mux.Handle("/api/users", faultinject.HTTPMiddleware("user-api")(userHandler))

// Custom response
mux.Handle("/api/payments", faultinject.HTTPMiddlewareWithResponse("payment-api", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(503)
    w.Write([]byte(`{"error": "payment service unavailable"}`))
})(paymentHandler))
```

### Function Decorators

```go
// Wrap functions with fault injection
createUserWithFaults := faultinject.WithFaultInjection("user-create", createUser)
err := createUserWithFaults(user)
```

## Use Cases

### Testing Error Handling

```go
func TestUserService_CreateUser(t *testing.T) {
    faultinject.SetFailures("db-insert", 1)
    
    service := NewUserService()
    user, err := service.CreateUser("john@example.com")
    
    assert.Error(t, err)
    assert.Nil(t, user)
}

// In your service:
func (s *UserService) CreateUser(email string) (*User, error) {
    if faultinject.Inject("db-insert") {
        return nil, fmt.Errorf("injected database failure")
    }
    // Actual database insert logic
    return s.db.Create(&User{Email: email})
}
```

### Chaos Engineering

```go
func TestServiceResilience(t *testing.T) {
    faultinject.SetNthFailure("external-api", 3) // Fail only 3rd call
    
    service := NewService()
    
    // First two calls succeed, third fails, fourth succeeds
    result1, _ := service.CallExternalAPI() // Success
    result2, _ := service.CallExternalAPI() // Success
    _, err := service.CallExternalAPI()     // Failure
    assert.Error(t, err)
    result4, _ := service.CallExternalAPI() // Success
}
```

### Load Testing with Failures

```go
func BenchmarkServiceWithFailures(b *testing.B) {
    faultinject.SetFailures("api-call", b.N/10) // 10% failure rate
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        if faultinject.Inject("api-call") {
            continue // Simulate failure handling
        }
        callAPI() // Normal API call
    }
}
```

## YAML Configuration

```yaml
# faults.yaml
failures:
  database-connect: 3
  api-call: 1

precise-failures:
  payment-service: 5
  email-service: 10
```

```go
// Load configuration
if err := faultinject.LoadSpec("faults.yaml"); err != nil {
    log.Fatalf("Failed to load fault spec: %v", err)
}
```

## HTTP Control Server

Start a control server for runtime management:

```go
faultinject.StartControlServer(":8081", nil)
```

### Available Endpoints

```bash
# Set failures
curl -X POST "http://localhost:8081/set?key=database-connect&count=3"

# Check status
curl "http://localhost:8081/status"

# Reset all
curl -X POST "http://localhost:8081/reset"
```

## Environment-Based Control

Fault injection is automatically disabled in production environments:

```bash
# Development - enabled
ENVIRONMENT=development go run main.go

# Production - disabled
ENVIRONMENT=production go run main.go
```

### Configuration

```go
// Customize environments
faultinject.SetAllowedEnvironments([]string{"dev", "test", "qa"})
faultinject.SetProductionEnvironments([]string{"prod", "live"})
```

## Best Practices

### 1. Use Descriptive Keys
```go
faultinject.Inject("user-service-create")  // Good
faultinject.Inject("fail")                 // Avoid
```

### 2. Reset Between Tests
```go
func TestMain(m *testing.M) {
    faultinject.Reset()
    os.Exit(m.Run())
}
```

### 3. Production Deployment
Use build tags to exclude fault injection from production builds:

```go
//go:build testing
package main

import "github.com/talinashro/go-fi/faultinject"

func init() {
    faultinject.LoadSpec("faults.yaml")
}
```

```bash
# Production build (no fault injection)
go build -o app

# Test build (with fault injection)
go build -tags testing -o app-test
```

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
