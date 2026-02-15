# Go Tooling Standards

## Linting (`golangci-lint`)
We use `golangci-lint` as our primary linter.

### Standard Configuration
Ensure your `.golangci.yaml` enables:

- `errcheck`: Checking for unchecked errors.
- `govet`: Reports suspicious constructs.
- `staticcheck`: Static analysis.
- `gosec`: Inspects source code for security problems.

## Testing
Always run tests with the race detector enabled.

```bash
go test -race ./...
```

## Formatting
We strictly follow `gofmt`. No custom configurations.
