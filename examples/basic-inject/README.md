# Basic Inject() Function Examples

This example demonstrates how to use the simple `Inject()` function for all fault injection scenarios without any fancy convenience functions.

## The Power of Basic Inject()

The `Inject()` function is all you need! It returns `true` when a fault should occur, allowing you to handle it however you want.

## Examples

### 1. Simple Error Injection
```go
func createUser(email string) error {
    if faultinject.Inject("user-create") {
        return fmt.Errorf("user creation failed")
    }
    log.Println("User created successfully")
    return nil
}
```

### 2. Database Operations
```go
func connectToDatabase() error {
    if faultinject.Inject("db-connect") {
        return fmt.Errorf("database connection failed")
    }
    log.Println("Database connected successfully")
    return nil
}
```

### 3. API Calls with Custom Messages
```go
func callExternalAPI() error {
    if faultinject.Inject("api-call") {
        return fmt.Errorf("API call failed: timeout")
    }
    log.Println("API call successful")
    return nil
}
```

### 4. Complex Error Handling
```go
func processPayment() error {
    if faultinject.Inject("payment-process") {
        log.Println("Simulating payment processing failure...")
        time.Sleep(100 * time.Millisecond)
        return fmt.Errorf("payment gateway timeout")
    }
    log.Println("Payment processed successfully")
    return nil
}
```

### 5. Context-Aware Injection
```go
func sendEmail(ctx context.Context, email string) error {
    // Check context override first, then use Inject
    if ctx.Value("faultinject:email-send") == true || faultinject.Inject("email-send") {
        return fmt.Errorf("email sending failed")
    }
    log.Printf("Email sent to %s successfully", email)
    return nil
}
```

### 6. HTTP Handlers
```go
func userHandler(w http.ResponseWriter, r *http.Request) {
    if faultinject.Inject("user-handler") {
        http.Error(w, "handler failure", 500)
        return
    }
    w.Write([]byte(`{"message": "user handler success"}`))
}
```

## Key Benefits

1. **Simple**: Just one function call
2. **Flexible**: Handle faults however you want
3. **Clear**: Easy to understand and debug
4. **Consistent**: Same pattern everywhere
5. **Powerful**: Can handle any scenario

## Common Patterns

### Basic Error Return
```go
if faultinject.Inject("key") {
    return fmt.Errorf("operation failed")
}
```

### Custom Error Messages
```go
if faultinject.Inject("key") {
    return fmt.Errorf("failed to %s: %v", "operation", err)
}
```

### Conditional Logic
```go
if faultinject.Inject("key") {
    if isRetryable {
        return fmt.Errorf("retryable failure")
    } else {
        return fmt.Errorf("permanent failure")
    }
}
```

### Complex Operations
```go
if faultinject.Inject("key") {
    log.Println("Simulating complex failure...")
    time.Sleep(100 * time.Millisecond)
    return fmt.Errorf("complex operation failed")
}
```

### Context Overrides
```go
if ctx.Value("faultinject:key") == true || faultinject.Inject("key") {
    return fmt.Errorf("operation failed")
}
```

## Running the Example

```bash
cd examples/basic-inject
go run main.go
```

## Expected Output

```
=== Basic Inject() Function Examples ===
1. Simple error injection:
   Error: user creation failed
2. Database operations:
   Error: database connection failed
3. API calls:
   Error: API call failed: timeout
4. Complex error handling:
   Simulating payment processing failure...
   Error: payment gateway timeout
5. Context-aware injection:
   Error: email sending failed
6. HTTP handler:
   HTTP server starting on :8080
```

## Why Basic Inject() is Enough

- **No Overhead**: Simple boolean check
- **Maximum Flexibility**: Handle faults your way
- **Easy to Understand**: Clear and straightforward
- **Consistent**: Same pattern across codebase
- **Debuggable**: Easy to trace and debug

The basic `Inject()` function is powerful enough to handle all your fault injection needs! 