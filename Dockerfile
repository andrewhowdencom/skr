# Build Stage
FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o skr .

# Runtime Stage
FROM alpine:latest

# git: required for 'skr batch publish'
# ca-certificates: required for HTTPS
# bash: required for entrypoint.sh
RUN apk add --no-cache git ca-certificates bash

WORKDIR /workspace

COPY --from=builder /app/skr /usr/local/bin/skr
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
