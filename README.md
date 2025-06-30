# go-fi

[![Go Report Card](https://goreportcard.com/badge/github.com/talinashro/go-fi)](https://goreportcard.com/report/github.com/talinashro/go-fi)
[![Go Reference](https://pkg.go.dev/badge/github.com/talinashro/go-fi.svg)](https://pkg.go.dev/github.com/talinashro/go-fi)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A lightweight, thread-safe Go library for in-process fault injection. Perfect for testing error handling, chaos engineering, and simulating real-world failure scenarios in your applications.

## Features

- **Simple API**: One function call to inject failures anywhere in your code
- **Two failure modes**: First-N failures or precise Nth failure
- **Runtime control**: Dynamically configure failures without restarting
- **YAML configuration**: Declarative failure specifications
- **HTTP control server**: Remote management via REST API
- **Thread-safe**: Safe for concurrent use
- **Zero dependencies**: Minimal footprint (only YAML parser)

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
    // Inject failure based on the key
    if faultinject.Inject("database-connect") {
        return fmt.Errorf("injected database connection failure")
    }
    
    // Your actual database connection logic here
    return nil
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

## Use Cases

### 1. Testing Error Handling

```go
func TestUserService_CreateUser(t *testing.T) {
    // Fail the first database operation
    faultinject.SetFailures("db-insert", 1)
    
    service := NewUserService()
    user, err := service.CreateUser("john@example.com")
    
    assert.Error(t, err)
    assert.Nil(t, user)
}

// In your actual UserService implementation:
func (s *UserService) CreateUser(email string) (*User, error) {
    // ... validation logic ...
    
    // Inject failure before database insert
    if faultinject.Inject("db-insert") {
        return nil, fmt.Errorf("injected database failure")
    }
    
    // Actual database insert logic
    user := &User{Email: email}
    if err := s.db.Create(user).Error; err != nil {
        return nil, err
    }
    
    return user, nil
}
```

### 2. Chaos Engineering

```go
func TestServiceResilience(t *testing.T) {
    // Simulate intermittent API failures
    faultinject.SetNthFailure("external-api", 3) // Fail only 3rd call
    
    service := NewService()
    
    // First two calls succeed
    result1, _ := service.CallExternalAPI()
    result2, _ := service.CallExternalAPI()
    
    // Third call fails
    _, err := service.CallExternalAPI()
    assert.Error(t, err)
    
    // Fourth call succeeds again
    result4, _ := service.CallExternalAPI()
}

// In your actual Service implementation:
func (s *Service) CallExternalAPI() (string, error) {
    // Inject failure before making external API call
    if faultinject.Inject("external-api") {
        return "", fmt.Errorf("injected external API failure")
    }
    
    // Actual external API call logic
    resp, err := http.Get("https://api.example.com/data")
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    // ... process response ...
    return "success", nil
}
```

### 3. Load Testing with Failures

```go
func BenchmarkServiceWithFailures(b *testing.B) {
    // Inject 10% failure rate
    faultinject.SetFailures("api-call", b.N/10)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        if faultinject.Inject("api-call") {
            // Simulate failure handling
            continue
        }
        // Normal API call
        callAPI()
    }
}

// In your actual API calling function:
func callAPI() error {
    // Inject failure before making API call
    if faultinject.Inject("api-call") {
        return fmt.Errorf("injected API failure")
    }
    
    // Actual API call logic
    resp, err := http.Post("https://api.example.com/endpoint", "application/json", nil)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

### 4. Microservices Testing

```go
func TestOrderService_Integration(t *testing.T) {
    // Simulate payment service being down
    faultinject.SetFailures("payment-service", 999) // Always fail
    
    orderService := NewOrderService()
    
    order, err := orderService.CreateOrder(Order{
        UserID: "user123",
        Items:  []Item{{ID: "item1", Quantity: 2}},
    })
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "payment service unavailable")
}

// In your actual OrderService implementation:
func (s *OrderService) CreateOrder(order Order) (*Order, error) {
    // ... order validation ...
    
    // Inject failure before calling payment service
    if faultinject.Inject("payment-service") {
        return nil, fmt.Errorf("payment service unavailable")
    }
    
    // Actual payment service call
    paymentResult, err := s.paymentService.ProcessPayment(order.TotalAmount())
    if err != nil {
        return nil, fmt.Errorf("payment failed: %w", err)
    }
    
    // ... save order to database ...
    return &order, nil
}
```

## API Reference

#### `Inject(key string) bool`
Injects a failure for the given key. Returns `true` if the operation should fail.

```go
if faultinject.Inject("my-operation") {
    return fmt.Errorf("injected failure")
}
```

#### `SetFailures(key string, count int)`
Configures the key to fail the first `count` calls.

```go
faultinject.SetFailures("database-connect", 3) // Fail first 3 calls
```

#### `SetNthFailure(key string, nth int)`
Configures the key to fail only on the Nth call.

```go
faultinject.SetNthFailure("api-call", 5) // Fail only 5th call
```

#### `Reset()`
Clears all configured failures and resets counters.

```go
faultinject.Reset() // Start fresh
```

#### `Status() map[string]int`
Returns remaining failure counts for each key.

```go
status := faultinject.Status()
// Returns: map[string]int{"database-connect": 2, "api-call": 0}
```

#### `LoadSpec(path string) error`
Loads failure configuration from a YAML file.

```go
if err := faultinject.LoadSpec("faults.yaml"); err != nil {
    log.Fatalf("Failed to load fault spec: %v", err)
}
```

#### `StartControlServer(addr string, runHandler http.HandlerFunc)`
Starts an HTTP server for remote control.

```go
faultinject.StartControlServer(":8081", nil)
```

#### `SetAllowedEnvironments(envs []string)`
Configures which environments allow fault injection.

```go
faultinject.SetAllowedEnvironments([]string{"development", "staging", "testing"})
```

#### `SetProductionEnvironments(envs []string)`
Configures which environments are considered production.

```go
faultinject.SetProductionEnvironments([]string{"production", "prod"})
```

## Environment-Based Control

go-fi automatically disables fault injection in production environments to prevent accidental failures in live systems.

### Default Behavior

- **Allowed environments**: `development`, `staging`, `testing`
- **Production environments**: `production`, `prod`
- **Default**: If environment is not explicitly allowed, fault injection is disabled

### Environment Detection

The library checks for environment variables in this order:
1. `ENVIRONMENT`
2. `ENV`
3. `GO_ENV`

### Configuration

```go
// Customize allowed environments
faultinject.SetAllowedEnvironments([]string{"dev", "test", "qa"})

// Customize production environments
faultinject.SetProductionEnvironments([]string{"prod", "live", "production"})
```

### Examples

```bash
# Development - fault injection enabled
ENVIRONMENT=development go run main.go

# Staging - fault injection enabled
ENV=staging go run main.go

# Production - fault injection disabled
GO_ENV=production go run main.go

# Unknown environment - fault injection disabled (defaults to production)
ENVIRONMENT=unknown go run main.go
```

### Production Safety

In production environments:
- `Inject()` always returns `false`
- `SetFailures()` and `SetNthFailure()` are no-ops
- All fault injection calls are safely ignored
- No runtime overhead from fault injection logic

## Simplified Fault Injection

The basic `Inject()` function is powerful enough to handle all fault injection scenarios. Here are the most common patterns:

### 1. Basic Error Injection

```go
// Simple error injection
if faultinject.Inject("db-insert") {
    return fmt.Errorf("database connection failed")
}

// With custom error messages
if faultinject.Inject("api-call") {
    return fmt.Errorf("API call failed: %s", "timeout")
}
```

### 2. Context-Aware Injection

```go
// Check context override first, then use Inject
if ctx.Value("faultinject:db-insert") == true || faultinject.Inject("db-insert") {
    return fmt.Errorf("database connection failed")
}
```

### 3. HTTP Middleware

For web applications, use middleware to automatically inject failures:

```go
// HTTP middleware with default 500 status code
mux := http.NewServeMux()
mux.Handle("/api/users", faultinject.HTTPMiddleware("user-api")(userHandler))

// HTTP middleware with custom status code
mux.Handle("/api/payments", faultinject.HTTPMiddleware("payment-api")(paymentHandler))
```

### 4. Function Decorators

Wrap functions with fault injection using decorators:

```go
// Original function
func createUser(user User) error {
    return db.Create(&user).Error
}

// Decorated with fault injection
createUserWithFaults := faultinject.WithFaultInjection("user-create", createUser)

// Usage
err := createUserWithFaults(user)
```

### 5. Complex Scenarios with Basic Inject

```go
// Complex error handling
if faultinject.Inject("payment-process") {
    log.Println("Simulating payment processing failure...")
    time.Sleep(100 * time.Millisecond)
    return fmt.Errorf("payment gateway timeout")
}

// Database operations with custom logic
if faultinject.Inject("db-query") {
    log.Println("Simulating database query failure...")
    return fmt.Errorf("database connection pool exhausted")
}

// External service calls
if faultinject.Inject("email-service") {
    log.Println("Simulating email service failure...")
    return fmt.Errorf("SMTP server unreachable")
}

// Conditional logic
if faultinject.Inject("api-call") {
    if isRetryable {
        return fmt.Errorf("retryable API failure")
    } else {
        return fmt.Errorf("permanent API failure")
    }
}
```

### 6. One-Liner Patterns

```go
// Database operations
func (s *UserService) CreateUser(user User) error {
    if faultinject.Inject("db-insert") {
        return fmt.Errorf("database insert failed")
    }
    return s.db.Create(&user).Error
}

// API calls
func (s *APIClient) CallAPI() error {
    if faultinject.Inject("api-call") {
        return fmt.Errorf("API call failed")
    }
    return s.makeAPICall()
}

// HTTP handlers
func userHandler(w http.ResponseWriter, r *http.Request) {
    if faultinject.Inject("user-handler") {
        http.Error(w, "handler failure", 500)
        return
    }
    // Normal handler logic
}
```

## Comparison of Approaches

| Approach | Code Overhead | Flexibility | Use Case |
|----------|---------------|-------------|----------|
| `Inject()` | Low | High | All scenarios |
| HTTP Middleware | Low | Medium | Web applications |
| Decorators | Low | High | Function wrapping |

The basic `Inject()` function is sufficient for most use cases and provides the most flexibility!

## HTTP Control Server

The control server is an **embedded HTTP server** that provides REST API endpoints to manage fault injection **while your application is running**. This allows you to dynamically change fault injection settings without restarting your application.

### What is it for?

The control server enables:

- **Runtime Configuration Changes**: Modify fault injection settings via HTTP requests
- **Chaos Engineering Experiments**: Dynamically inject failures during live testing
- **Load Testing with Dynamic Failures**: Adjust failure rates during load tests
- **Development and Staging Control**: Monitor and control fault injection in non-production environments

### Available Endpoints

#### `POST /set?key=<key>&count=<n>`
Sets the number of failures for a specific key.

```bash
# Set database operations to fail the first 3 times
curl -X POST "http://localhost:8081/set?key=database-connect&count=3"

# Set payment service to always fail (999 failures)
curl -X POST "http://localhost:8081/set?key=payment-service&count=999"
```

#### `GET /status`
Returns the current status of all configured failures in JSON format.

```bash
curl "http://localhost:8081/status"
```

**Response:**
```json
{
  "database-connect": 2,
  "payment-service": 0,
  "external-api": 1
}
```

#### `POST /reset`
Clears all configured failures and resets counters.

```bash
curl -X POST "http://localhost:8081/reset"
```

#### `POST /run` (Optional)
Custom endpoint for running your application logic (if provided).

### Real-World Use Cases

#### 1. Chaos Engineering Experiments

```bash
#!/bin/bash
# chaos-test.sh

# Start application with control server
./app &

# Wait for startup
sleep 5

# Simulate database outage
curl -X POST "http://localhost:8081/set?key=database-connect&count=999"

# Monitor application behavior
sleep 30

# Check remaining failures
curl "http://localhost:8081/status"

# Reset and continue
curl -X POST "http://localhost:8081/reset"
```

#### 2. Load Testing with Dynamic Failures

```python
import requests
import time

# Gradually increase failure rate
for failure_rate in [10, 25, 50, 75, 100]:
    requests.post(f"http://localhost:8081/set?key=api-call&count={failure_rate}")
    time.sleep(60)  # Test for 1 minute
    
    # Check status
    status = requests.get("http://localhost:8081/status").json()
    print(f"Failure rate {failure_rate}%: {status}")
```

#### 3. Integration Testing

```go
func TestIntegrationWithFailures(t *testing.T) {
    // Start control server
    faultinject.StartControlServer(":8081", nil)
    
    // Run your application
    go runApp()
    
    // Dynamically inject failures during test
    time.Sleep(time.Second)
    http.Post("http://localhost:8081/set?key=database&count=1", "", nil)
    
    // Continue testing...
}
```

### Benefits

- **No Restart Required**: Change fault injection settings without stopping your application
- **Remote Management**: Control fault injection from anywhere via HTTP
- **Real-time Monitoring**: Check current status and remaining failures
- **Automation Friendly**: Easy to integrate with CI/CD pipelines and testing scripts
- **Environment Safe**: Automatically disabled in production environments

The control server essentially gives you a **remote control panel** for your fault injection system, making it much more flexible and powerful for testing and chaos engineering scenarios.

## Advanced Examples

### Testing Circuit Breaker Pattern

```go
func TestCircuitBreaker(t *testing.T) {
    circuitBreaker := NewCircuitBreaker(3, time.Second)
    
    // Simulate service failures
    faultinject.SetFailures("service-call", 5)
    
    for i := 0; i < 10; i++ {
        err := circuitBreaker.Execute(func() error {
            if faultinject.Inject("service-call") {
                return fmt.Errorf("service unavailable")
            }
            return nil
        })
        
        if i < 3 {
            assert.Error(t, err) // First 3 calls fail
        } else if i < 6 {
            assert.Error(t, err) // Circuit breaker open
        } else {
            assert.NoError(t, err) // Circuit breaker closed
        }
    }
}

// In your actual service implementation:
func (s *Service) CallDownstreamService() error {
    // Inject failure before calling downstream service
    if faultinject.Inject("service-call") {
        return fmt.Errorf("service unavailable")
    }
    
    // Actual downstream service call
    resp, err := http.Get("https://downstream-service.com/api")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

### Testing Retry Logic

```go
func TestRetryWithFailures(t *testing.T) {
    // Fail first 2 calls, succeed on 3rd
    faultinject.SetFailures("api-call", 2)
    
    var attempts int
    err := retry.Do(
        func() error {
            attempts++
            if faultinject.Inject("api-call") {
                return fmt.Errorf("temporary failure")
            }
            return nil
        },
        retry.Attempts(5),
        retry.Delay(time.Millisecond*100),
    )
    
    assert.NoError(t, err)
    assert.Equal(t, 3, attempts) // Should succeed on 3rd attempt
}

// In your actual API client:
func (c *APIClient) CallAPI() error {
    // Inject failure before making API call
    if faultinject.Inject("api-call") {
        return fmt.Errorf("temporary failure")
    }
    
    // Actual API call logic
    resp, err := http.Post(c.endpoint, "application/json", c.body)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

### Testing Graceful Degradation

```go
func TestGracefulDegradation(t *testing.T) {
    // Simulate cache being down
    faultinject.SetFailures("cache-get", 999)
    
    service := NewUserService()
    
    // Should fall back to database
    user, err := service.GetUser("user123")
    
    assert.NoError(t, err)
    assert.NotNil(t, user)
    // Verify user was loaded from database, not cache
}

// In your actual UserService implementation:
func (s *UserService) GetUser(id string) (*User, error) {
    // Try cache first
    if faultinject.Inject("cache-get") {
        // Cache failure - fall back to database
        log.Printf("Cache miss for user %s, falling back to database", id)
    } else {
        // Try to get from cache
        if user, found := s.cache.Get(id); found {
            return user.(*User), nil
        }
    }
    
    // Fall back to database
    var user User
    if err := s.db.Where("id = ?", id).First(&user).Error; err != nil {
        return nil, err
    }
    
    // Cache the result for next time
    s.cache.Set(id, &user, time.Minute*5)
    
    return &user, nil
}
```

## Best Practices

### 1. Use Descriptive Keys
```go
// Good
faultinject.Inject("user-service-create")
faultinject.Inject("payment-gateway-process")

// Avoid
faultinject.Inject("fail")
faultinject.Inject("error")
```

### 2. Reset Between Tests
```go
func TestMain(m *testing.M) {
    // Reset before running tests
    faultinject.Reset()
    os.Exit(m.Run())
}
```

### 3. Use YAML for Complex Scenarios
```yaml
# test-scenarios.yaml
failures:
  database-primary: 2
  database-replica: 1
  redis-cache: 3

precise-failures:
  payment-service: 5
  email-service: 10
```

### 4. Monitor in Production
```go
// Use HTTP control server for production monitoring
faultinject.StartControlServer(":8081", nil)

// Monitor via health checks
func healthCheck() bool {
    status := faultinject.Status()
    // Alert if too many failures are configured
    return len(status) == 0
}
```

## Production Deployment Strategies

### Strategy 1: Build Tags (Recommended)

Use Go build tags to conditionally include fault injection only in test builds:

```go
//go:build testing

package main

import "github.com/talinashro/go-fi/faultinject"

func init() {
    // Only load fault injection in test builds
    faultinject.LoadSpec("faults.yaml")
}
```

**Main application code:**
```go
// user_service.go
package main

import "github.com/talinashro/go-fi/faultinject"

func (s *UserService) CreateUser(email string) (*User, error) {
    // Fault injection call - will be no-op in production
    if faultinject.Inject("db-insert") {
        return nil, fmt.Errorf("injected database failure")
    }
    
    // Actual business logic
    user := &User{Email: email}
    if err := s.db.Create(user).Error; err != nil {
        return nil, err
    }
    
    return user, nil
}
```

**Build commands:**
```bash
# Production build (no fault injection)
go build -o app

# Test build (with fault injection)
go build -tags testing -o app-test

# Run tests
go test -tags testing ./...
```

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details on how to submit pull requests, report issues, and contribute to the project.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by Netflix's Chaos Monkey
- Built for the Go community's testing needs
- Thanks to all contributors and users

**Test setup:**
```go
// test_setup.go
//go:build testing

package main

import "github.com/talinashro/go-fi/faultinject"

func init() {
    // Only runs in test builds
    faultinject.NoOpLoadSpec("test-faults.yaml")
}
```
