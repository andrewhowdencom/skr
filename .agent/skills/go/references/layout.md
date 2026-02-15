# Standard Go Project Layout

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout), adapted for our organizational needs.

## Directory Structure

### `/cmd`
Main applications for this project.
- Each subdirectory should match the name of the executable.
- `main.go` content should be minimal (initialization and dependency injection).
- **Example**: `cmd/server/main.go`

### `/internal`
Private application and library code. This is the code you don't want others importing in their applications or libraries.
- The Go compiler enforces this layout pattern.
- **Example**: `internal/app/server.go`

### `/api`
OpenAPI/Swagger specs, JSON schema files, protocol definition files.

### `/web`
Web application specific components: static assets, server side templates and SPAs.
