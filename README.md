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

### Core Functions

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

## Simplified Fault Injection

To reduce the overhead of function calls, go-fi provides several convenience approaches:

### 1. Function-Based Injection (Recommended)

Execute custom functions when faults occur:

```go
// Execute a custom function when fault injection triggers
if err := faultinject.InjectWithFn("db-insert", func() error {
    return fmt.Errorf("database connection failed")
}); err != nil {
    return err
}

// Context-aware function execution
if err := faultinject.InjectWithFnContext(ctx, "api-call", func() error {
    return fmt.Errorf("API call failed")
}); err != nil {
    return err
}
```

### 2. Error-Returning Functions

Instead of checking the return value and creating an error manually:

```go
// Before: Manual check and error creation
if faultinject.Inject("db-insert") {
    return fmt.Errorf("injected database failure")
}

// After: Direct error return
if err := faultinject.InjectWithError("db-insert", "database failure"); err != nil {
    return err
}
```

**Available functions:**
```go
// Simple error with message
faultinject.InjectWithError("key", "failure message")

// Formatted error with arguments
faultinject.InjectWithErrorf("key", "failed to %s: %v", "operation", err)

// Context-aware error injection
faultinject.InjectWithContextError(ctx, "key", "failure message")
```

### 3. HTTP Middleware

For web applications, use middleware to automatically inject failures:

```go
// Simple middleware
mux := http.NewServeMux()
mux.Handle("/api/users", faultinject.HTTPMiddleware("user-api")(userHandler))

// Middleware with custom status code
mux.Handle("/api/payments", faultinject.HTTPMiddlewareWithStatus("payment-api", 503)(paymentHandler))
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

### 5. Database-Specific Helpers

Use database-specific injectors for common database operations:

```go
// PostgreSQL operations
if err := faultinject.PostgresInjector.InjectConnectionFailure(); err != nil {
    return err
}

if err := faultinject.PostgresInjector.InjectQueryFailure(); err != nil {
    return err
}

// MySQL operations
if err := faultinject.MySQLInjector.InjectTransactionFailure(); err != nil {
    return err
}

// Custom database type
redisInjector := faultinject.NewDatabaseInjector("redis")
if err := redisInjector.InjectTimeoutFailure(); err != nil {
    return err
}
```

### 6. Context-Based Overrides

Override fault injection behavior using context:

```go
// Override fault injection for specific requests
ctx := context.WithValue(context.Background(), "faultinject:db-insert", true)

// Use context-aware injection
if err := faultinject.InjectWithContextError(ctx, "db-insert", "database failure"); err != nil {
    return err
}
```

### 7. One-Liner Patterns

Combine multiple approaches for maximum simplicity:

```go
// Database operations
func (s *UserService) CreateUser(user User) error {
    return faultinject.PostgresInjector.WithFaultInjection("insert", func() error {
        return s.db.Create(&user).Error
    })
}

// API calls
func (s *APIClient) CallAPI() error {
    return faultinject.InjectWithError("api-call", "API call failed") || s.makeAPICall()
}

// HTTP handlers
func userHandler(w http.ResponseWriter, r *http.Request) {
    if err := faultinject.InjectWithError("user-handler", "handler failure"); err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    // Normal handler logic
}
```

### 8. Build Tag Helpers

Use the NoOp functions for automatic switching between test and production:

```go
// Same code works in both production and testing
func (s *UserService) CreateUser(user User) error {
    if err := faultinject.NoOpInjectWithError("user-create", "user creation failed"); err != nil {
        return err
    }
    return s.db.Create(&user).Error
}
```

## Comparison of Approaches

| Approach | Code Overhead | Flexibility | Use Case |
|----------|---------------|-------------|----------|
| `Inject()` | High | High | Custom logic |
| `InjectWithError()` | Medium | High | Simple error returns |
| HTTP Middleware | Low | Medium | Web applications |
| Decorators | Low | High | Function wrapping |
| Database Helpers | Low | Medium | Database operations |
| Context Overrides | Medium | High | Request-specific control |

Choose the approach that best fits your use case and coding style!

## Configuration

#### `LoadSpec(path string) error`
Loads failure configuration from a YAML file.

```yaml
# faults.yaml
failures:
  database-connect: 2      # Fail first 2 calls
  cache-get: 1             # Fail first 1 call
  external-api: 3          # Fail first 3 calls

precise-failures:
  payment-service: 5       # Fail only 5th call
  email-service: 10        # Fail only 10th call
```

```go
if err := faultinject.LoadSpec("faults.yaml"); err != nil {
    log.Fatalf("Failed to load fault spec: %v", err)
}
```

### HTTP Control Server

#### `StartControlServer(addr string, runHandler http.HandlerFunc)`
Starts an HTTP server for remote control.

```go
// Start control server on port 8081
faultinject.StartControlServer(":8081", nil)
```

**Available endpoints:**

- `POST /set?key=<key>&count=<n>` - Set failure count
- `POST /reset` - Reset all failures
- `GET /status` - Get current status
- `POST /run` - Custom handler (optional)

**Example usage:**
```bash
# Set database to fail first 3 times
curl -X POST "http://localhost:8081/set?key=database-connect&count=3"

# Check current status
curl "http://localhost:8081/status"

# Reset all failures
curl -X POST "http://localhost:8081/reset"
```

## HTTP Control Server

The control server is an **embedded HTTP server** that provides REST API endpoints to manage fault injection **while your application is running**. This allows you to dynamically change fault injection settings without restarting your application.

### What is it for?

The control server enables:

- **Runtime Configuration Changes**: Modify fault injection settings via HTTP requests
- **Chaos Engineering Experiments**: Dynamically inject failures during live testing
- **Load Testing with Dynamic Failures**: Adjust failure rates during load tests
- **Production Monitoring**: Monitor and control fault injection in production environments

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

#### 4. Production Monitoring

```go
func main() {
    // Start control server for monitoring
    faultinject.StartControlServer(":8081", nil)
    
    // Your application logic
    app := NewApp()
    app.Run()
}
```

### Benefits

- **No Restart Required**: Change fault injection settings without stopping your application
- **Remote Management**: Control fault injection from anywhere via HTTP
- **Real-time Monitoring**: Check current status and remaining failures
- **Automation Friendly**: Easy to integrate with CI/CD pipelines and testing scripts
- **Production Safe**: Can be used in production for chaos engineering experiments

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

### Strategy 2: Environment-Based Configuration

Use environment variables to control fault injection:

```go
// fault_injection.go
package main

import (
    "os"
    "github.com/talinashro/go-fi/faultinject"
)

var faultInjectionEnabled bool

func init() {
    faultInjectionEnabled = os.Getenv("ENABLE_FAULT_INJECTION") == "true"
    if faultInjectionEnabled {
        faultinject.LoadSpec("faults.yaml")
    }
}

func injectFault(key string) bool {
    if !faultInjectionEnabled {
        return false
    }
    return faultinject.Inject(key)
}
```

**Usage in application:**
```go
func (s *UserService) CreateUser(email string) (*User, error) {
    if injectFault("db-insert") {
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

**Run with fault injection:**
```bash
ENABLE_FAULT_INJECTION=true go run main.go
```

### Strategy 3: Separate Test Binary

Create a separate test binary that includes fault injection:

```go
// cmd/test/main.go
package main

import (
    "github.com/talinashro/go-fi/faultinject"
    "your-app/internal/services"
)

func main() {
    // Load fault injection configuration
    faultinject.LoadSpec("test-faults.yaml")
    
    // Start control server
    faultinject.StartControlServer(":8081", nil)
    
    // Run your application with fault injection
    app := services.NewApp()
    app.Run()
}
```

**Production binary remains clean:**
```go
// cmd/prod/main.go
package main

import "your-app/internal/services"

func main() {
    // Clean production binary - no fault injection
    app := services.NewApp()
    app.Run()
}
```

### Strategy 4: Interface-Based Approach

Use interfaces to abstract fault injection:

```go
// fault_injector.go
package main

type FaultInjector interface {
    Inject(key string) bool
}

type NoOpFaultInjector struct{}

func (n NoOpFaultInjector) Inject(key string) bool {
    return false
}

type RealFaultInjector struct{}

func (r RealFaultInjector) Inject(key string) bool {
    return faultinject.Inject(key)
}

// Service with dependency injection
type UserService struct {
    db            *gorm.DB
    faultInjector FaultInjector
}

func NewUserService(db *gorm.DB, faultInjector FaultInjector) *UserService {
    return &UserService{
        db:            db,
        faultInjector: faultInjector,
    }
}

func (s *UserService) CreateUser(email string) (*User, error) {
    if s.faultInjector.Inject("db-insert") {
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

**Production usage:**
```go
func main() {
    db := initDatabase()
    service := NewUserService(db, NoOpFaultInjector{})
    // ... rest of app
}
```

**Test usage:**
```go
func TestMain(m *testing.M) {
    faultinject.LoadSpec("test-faults.yaml")
    os.Exit(m.Run())
}

func TestUserService(t *testing.T) {
    db := initTestDatabase()
    service := NewUserService(db, RealFaultInjector{})
    // ... test logic
}
```

### Strategy 5: Compile-Time Constants

Use compile-time constants to completely eliminate fault injection code:

```go
// build.go
package main

//go:generate go run build.go

const (
    ENABLE_FAULT_INJECTION = false // Set to true for test builds
)

// fault_injection.go
package main

import "github.com/talinashro/go-fi/faultinject"

func injectFault(key string) bool {
    if !ENABLE_FAULT_INJECTION {
        return false
    }
    return faultinject.Inject(key)
}
```

**Build script:**
```bash
#!/bin/bash
# build-test.sh
sed -i 's/ENABLE_FAULT_INJECTION = false/ENABLE_FAULT_INJECTION = true/' build.go
go build -o app-test
sed -i 's/ENABLE_FAULT_INJECTION = true/ENABLE_FAULT_INJECTION = false/' build.go
```

## Recommended Approach

For most projects, we recommend **Strategy 1 (Build Tags)** because:

- **Zero runtime overhead** in production
- **Clean separation** between test and production code
- **Easy to use** with existing Go tooling
- **No conditional logic** in production binaries
- **Works well** with CI/CD pipelines

Example CI/CD pipeline:
```yaml
# .github/workflows/test.yml
- name: Run tests with fault injection
  run: go test -tags testing ./...

- name: Build production binary
  run: go build -o app
```

This ensures your production binary is completely clean while maintaining full fault injection capabilities during testing and development.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

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
