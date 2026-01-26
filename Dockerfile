# Build Stage
FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o skr .

# Runtime Stage
FROM alpine:latest

# Install runtime dependencies
# git: required for 'skr batch publish' (change detection) and 'build' (annotations)
# ca-certificates: required for HTTPS (registry interactions)
RUN apk add --no-cache git ca-certificates

WORKDIR /workspace

COPY --from=builder /app/skr /usr/local/bin/skr

ENTRYPOINT ["skr"]
