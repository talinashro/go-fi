# HTTP Middleware Example

This example demonstrates the flexible HTTP middleware that allows custom responses beyond just error messages.

## Features Demonstrated

1. **Default Error Response** - Simple 500 error
2. **Custom JSON Response** - Structured error with metadata
3. **Health Check Failure** - Custom status and content type
4. **Retry Headers** - Adding retry-after and custom headers
5. **Slow Response Simulation** - Timeout scenarios
6. **Logging Integration** - Custom error with logging

## Running the Example

```bash
cd examples/middleware
go run main.go
```

## Testing the Endpoints

### 1. Default Error Response
```bash
curl http://localhost:8080/api/users
```
**Response:** 500 Internal Server Error with "Injected failure" message

### 2. Custom JSON Response
```bash
curl http://localhost:8080/api/payments
```
**Response:** 503 Service Unavailable with JSON:
```json
{
  "error": "payment service unavailable",
  "code": "PAYMENT_DOWN",
  "retry": "true",
  "timeout": "30s"
}
```

### 3. Health Check Failure
```bash
curl http://localhost:8080/api/health
```
**Response:** 503 Service Unavailable with text: "health check failed - service degraded"

### 4. Data API with Retry Headers
```bash
curl -v http://localhost:8080/api/data
```
**Response:** 503 Service Unavailable with headers:
- `Retry-After: 30`
- `X-Failure-Reason: database_connection`

### 5. Slow Response Simulation
```bash
curl http://localhost:8080/api/slow
```
**Response:** 408 Request Timeout after 5 seconds with JSON:
```json
{
  "error": "request timeout",
  "code": "TIMEOUT"
}
```

### 6. Critical Error with Logging
```bash
curl http://localhost:8080/api/critical
```
**Response:** 500 Internal Server Error with header `X-Error-ID: CRITICAL_001`

## Middleware Usage Patterns

### Simple Error Response
```go
mux.Handle("/api/users", faultinject.HTTPMiddleware("user-api")(userHandler))
```

### Custom JSON Response
```go
mux.Handle("/api/payments", faultinject.HTTPMiddlewareWithResponse("payment-api", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(503)
    json.NewEncoder(w).Encode(map[string]string{
        "error": "payment service unavailable",
        "code": "PAYMENT_DOWN",
    })
})(paymentHandler))
```

### Custom Headers and Logging
```go
mux.Handle("/api/data", faultinject.HTTPMiddlewareWithResponse("data-api", func(w http.ResponseWriter, r *http.Request) {
    log.Println("Simulating data API failure...")
    w.Header().Set("Retry-After", "30")
    w.Header().Set("X-Failure-Reason", "database_connection")
    http.Error(w, "service temporarily unavailable", 503)
})(dataHandler))
```

### Slow Response Simulation
```go
mux.Handle("/api/slow", faultinject.HTTPMiddlewareWithResponse("slow-api", func(w http.ResponseWriter, r *http.Request) {
    time.Sleep(5 * time.Second)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(408)
    json.NewEncoder(w).Encode(map[string]string{
        "error": "request timeout",
    })
})(slowHandler))
```

## Benefits

1. **Flexible Responses**: Any type of HTTP response can be simulated
2. **Custom Headers**: Add retry headers, error codes, etc.
3. **Logging Integration**: Log failures for debugging
4. **Timeout Simulation**: Simulate slow responses
5. **Structured Errors**: Return JSON with error codes and metadata
6. **No New Methods**: Uses existing API with custom response functions

The middleware provides maximum flexibility while keeping the API simple! 