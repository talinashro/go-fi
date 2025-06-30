# Environment-Based Control Example

This example demonstrates how go-fi automatically disables fault injection in production environments.

## How it Works

The library checks environment variables to determine if fault injection should be enabled:

1. `ENVIRONMENT`
2. `ENV`
3. `GO_ENV`

## Default Behavior

- **Allowed environments**: `development`, `staging`, `testing`
- **Production environments**: `production`, `prod`
- **Unknown environments**: Treated as production (fault injection disabled)

## Running the Example

### Development Environment (Fault Injection Enabled)

```bash
cd examples/environment-control
ENVIRONMENT=development go run main.go
```

**Expected Output:**
```
Current environment: development
=== Environment-Based Fault Injection Demo ===
1. Testing database connection:
   Attempt 1: database connection failed
   Attempt 2: database connection failed
   Attempt 3: Success
2. Testing API call:
   Error: API call failed
3. Current fault injection status: map[api-call:0 db-connect:0]
```

### Production Environment (Fault Injection Disabled)

```bash
cd examples/environment-control
ENVIRONMENT=production go run main.go
```

**Expected Output:**
```
Current environment: production
=== Environment-Based Fault Injection Demo ===
1. Testing database connection:
   Attempt 1: Success
   Attempt 2: Success
   Attempt 3: Success
2. Testing API call:
   Success
3. Current fault injection status: map[]
```

### Unknown Environment (Fault Injection Disabled)

```bash
cd examples/environment-control
ENVIRONMENT=unknown go run main.go
```

**Expected Output:**
```
Current environment: unknown
=== Environment-Based Fault Injection Demo ===
1. Testing database connection:
   Attempt 1: Success
   Attempt 2: Success
   Attempt 3: Success
2. Testing API call:
   Success
3. Current fault injection status: map[]
```

## Custom Environment Configuration

You can customize which environments are allowed:

```go
// Allow custom environments
faultinject.SetAllowedEnvironments([]string{"dev", "test", "qa"})

// Set custom production environments
faultinject.SetProductionEnvironments([]string{"prod", "live"})
```

## Benefits

1. **Production Safety**: Fault injection is automatically disabled in production
2. **No Configuration**: Works out of the box with common environment names
3. **Flexible**: Can customize environment detection
4. **Zero Overhead**: No runtime cost in production
5. **Safe Defaults**: Unknown environments default to production mode 