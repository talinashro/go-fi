# go-fi

> A minimalist Go library for in-process fault injection (“fail N times”) with an optional HTTP control server and YAML-driven DSL.

## Features

- **Inject failures by key**: call `faultinject.Inject("my-key")` to fail the first N invocations.
- **Runtime control**: `faultinject.SetFailures`, `Reset` and `Status` APIs.
- **YAML DSL**: load failure counts from a `faults.yaml`.
- **Embedded HTTP server**: live `/set`, `/reset`, `/status` and optional `/run` endpoints.

## Installation

```bash
go get github.com/talinashro/go-fi@v0.1.0
```

Import:

```go
import "github.com/talinashro/go-fi/faultinject"
```

## Usage

### Basic in-code injection

```go
if faultinject.Inject("create-ec2") {
    return fmt.Errorf("injected EC2 failure")
}
```

Configure:

```go
faultinject.SetFailures("create-ec2", 1)
fmt.Println(faultinject.Status()) // map[string]int{"create-primary-ec2":0}
```

### YAML DSL

Create `faults.yaml`:

```yaml
failures:
  create-ec2:         1
  create-storage:     2
```

Load:

```go
if err := faultinject.LoadSpec("faults.yaml"); err != nil {
    log.Fatalf("cannot load spec: %v", err)
}
```

### HTTP Control Server

```go
faultinject.StartControlServer(":8081", nil)
```

Shell commands:

```bash
curl -X POST "http://127.0.0.1:8081/set?key=create-ec2&count=1"
curl "http://127.0.0.1:8081/status"
curl -X POST "http://127.0.0.1:8081/reset"
```

## Example

```bash
cd examples/simple
go run main.go --spec faults.yaml
```

## API Reference

```go
func Inject(key string) bool
func SetFailures(key string, n int)
func Reset()
func Status() map[string]int
func LoadSpec(path string) error
func StartControlServer(addr string, runHandler http.HandlerFunc)
```

## License

Apache License 2.0 — see the LICENSE file for details.
