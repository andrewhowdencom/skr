---
name: go
description: Guidelines for Go development, testing, and tooling.
---

# Go

Follow [Effective Go](https://go.dev/doc/effective_go) as the baseline.

## Project Structure
We follow the standard layout. See [Layout Reference](./references/layout.md) for details.

- **cmd/**: Main applications.
- **internal/**: Private application logic.

## Inlined Expertise: Common Patterns

### 1. Functional Options Pattern
Use for object construction with optional parameters.

```go
type Server struct {
    timeout time.Duration
}

type Option func(*Server)

func WithTimeout(t time.Duration) Option {
    return func(s *Server) { s.timeout = t }
}

func New(opts ...Option) *Server {
    s := &Server{timeout: 30 * time.Second} // Default
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

### 2. Table-Driven Tests
The standard for unit testing.

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 1, 2, 3},
        {"negative", -1, -1, -2},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := Add(tt.a, tt.b); got != tt.want {
                t.Errorf("Add() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### 3. Error Handling
Wrap upstream errors to provide context.

```go
if err := dest.Do(); err != nil {
    return fmt.Errorf("failed to do action: %w", err)
}
```

## Tooling
- **Linting**: use `golangci-lint`.
- **Testing**: always use `go test -race ./...`.
- See [Tooling Reference](./references/tooling.md) for deeper configuration.

## References
- [Project Layout](./references/layout.md)
- [Tooling Standards](./references/tooling.md)
