# Build Tags Example

This example demonstrates how to use Go's modern build tags (`//go:build`) to conditionally include fault injection in your application.

## Files

- `main.go` - Main application code (works in both production and testing)
- `test_setup.go` - Test-specific setup (only included with `testing` build tag)
- `test-faults.yaml` - Fault injection configuration for testing

## How it works

### 1. Main Application Code
The `main.go` file uses `faultinject.Inject()` which:
- **Works normally** in test builds (with fault injection)
- **Returns false** in production builds (no fault injection)

### 2. Test Setup
The `test_setup.go` file uses the `//go:build testing` tag, so it's only included when building with the `testing` tag.

### 3. Build Commands

```bash
# Production build (no fault injection)
go build -o app

# Test build (with fault injection)
go build -tags testing -o app-test

# Run tests
go test -tags testing ./...
```

## Expected Output

### Production Build
```bash
$ go build -o app
$ ./app
Starting application...
Operation 1:
  Creating user: user1@example.com
  Succeeded
Operation 2:
  Creating user: user2@example.com
  Succeeded
...
```

### Test Build
```bash
$ go build -tags testing -o app-test
$ ./app-test
Loading fault injection configuration...
Fault injection configured: map[user-create:2]
Starting application...
Operation 1:
  Failed: injected user creation failure
Operation 2:
  Failed: injected user creation failure
Operation 3:
  Creating user: user3@example.com
  Succeeded
...
```

## Benefits

- **Single codebase** for both production and testing
- **Zero runtime overhead** in production
- **Simple approach** using basic Inject() function
- **Clean separation** at compile time
- **Modern Go syntax** using `//go:build` tags 