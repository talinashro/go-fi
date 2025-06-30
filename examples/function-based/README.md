# Function-Based Fault Injection Example

This example demonstrates the new function-based fault injection approach, which allows you to execute custom functions when faults occur.

## What is Function-Based Injection?

Instead of manually checking if a fault should occur and then handling it, you can pass a function that will be executed automatically when the fault injection triggers.

## Before vs After

### Before (Manual Approach)
```go
if faultinject.Inject("db-insert") {
    return fmt.Errorf("database connection failed")
}
// Continue with normal logic...
```

### After (Function-Based Approach)
```go
if err := faultinject.InjectWithFn("db-insert", func() error {
    return fmt.Errorf("database connection failed")
}); err != nil {
    return err
}
// Continue with normal logic...
```

## Key Benefits

1. **Cleaner Code**: No need to manually check return values
2. **Custom Logic**: Execute any custom logic when faults occur
3. **Context Support**: Context-aware function execution
4. **Build Tag Compatible**: Works with NoOp functions for production

## Available Functions

### `InjectWithFn(key string, fn func() error) error`
Executes the provided function if fault injection should occur.

```go
if err := faultinject.InjectWithFn("db-insert", func() error {
    return fmt.Errorf("database connection failed")
}); err != nil {
    return err
}
```

### `InjectWithFnContext(ctx context.Context, key string, fn func() error) error`
Context-aware function execution.

```go
if err := faultinject.InjectWithFnContext(ctx, "api-call", func() error {
    return fmt.Errorf("API call failed")
}); err != nil {
    return err
}
```

### `NoOpInjectWithFn(key string, fn func() error) error`
Build tag compatible version that becomes no-op in production.

```go
if err := faultinject.NoOpInjectWithFn("user-create", func() error {
    return fmt.Errorf("user creation failed")
}); err != nil {
    return err
}
```

## Use Cases

### 1. Complex Error Handling
```go
if err := faultinject.InjectWithFn("payment-process", func() error {
    log.Println("Simulating payment processing failure...")
    time.Sleep(100 * time.Millisecond)
    return fmt.Errorf("payment gateway timeout")
}); err != nil {
    return err
}
```

### 2. Database Operations
```go
if err := faultinject.InjectWithFn("db-query", func() error {
    log.Println("Simulating database query failure...")
    return fmt.Errorf("database connection pool exhausted")
}); err != nil {
    return err
}
```

### 3. External Service Calls
```go
if err := faultinject.InjectWithFn("email-service", func() error {
    log.Println("Simulating email service failure...")
    return fmt.Errorf("SMTP server unreachable")
}); err != nil {
    return err
}
```

## Running the Example

```bash
cd examples/function-based
go run main.go
```

## Expected Output

```
=== Function-Based Fault Injection Examples ===
1. Simple function-based injection:
   Error: injected failure: database connection failed
2. Context-aware function injection:
   Error: injected failure: API call failed
3. Complex error handling:
   Simulating payment processing failure...
   Error: injected failure: payment gateway timeout
4. Database operations:
   Simulating database query failure...
   Error: injected failure: database connection pool exhausted
5. External service calls:
   Simulating email service failure...
   Error: injected failure: SMTP server unreachable
6. Build tag helpers with functions:
   Error: injected failure: user creation failed
=== Examples completed ===
```

## Advantages Over Traditional Approach

1. **Less Code**: Single line instead of multiple lines
2. **More Expressive**: Clear intent with function parameters
3. **Flexible**: Can execute any custom logic
4. **Consistent**: Same pattern across the codebase
5. **Testable**: Easy to mock and test 