---
name: instrumentation
description: Guidelines for OpenTelemetry instrumentation, observability strategy, and standards.
---

# Instrumentation

## Methodology
- **Standard**: Use [OpenTelemetry](https://opentelemetry.io) (OTel).
- **Propagation**: Propagate W3C Trace Context across all boundaries.

## Naming & Standards
- **Spans**: Name after *business operation* (e.g., `checkout`), NOT transport (`POST /api/checkout`).
- **Metrics**: Use OTel naming (e.g., `foo.bar`), NOT Prometheus contextless names.
- **Attributes**: Follow [OTel Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/) (v1.37.0).
  - Use `http.path` with placeholders (e.g., `/users/{id}`). **NEVER** use `http.url` (PII risk).
- **Scopes**: Name Tracers/Meters using the fully qualified library name.

## Go Implementation

### Initialization
Inject `trace.Tracer` and `metric.Meter` as dependencies.

```go
type Service struct {
	tracer trace.Tracer
	meter  metric.Meter
}
// Initialize with namespaced tracer/meter
s.tracer = tp.Tracer("github.com/org/repo/pkg/service")
```

### Starting Spans
Always set the proper `SpanKind`.
```go
ctx, span := s.tracer.Start(ctx, "calculate-tax",
    trace.WithSpanKind(trace.SpanKindInternal), // or Server, Client, Producer, Consumer
)
defer span.End()
```

### Recording Errors
```go
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    return err
}
```

### HTTP Instrumentation
Use `httpconv` for standard attributes.

```go
// Client
attrs := httpconv.ClientRequest(req)
ctx, span := c.tracer.Start(req.Context(), "HTTP "+req.Method,
    trace.WithSpanKind(trace.SpanKindClient),
    trace.WithAttributes(attrs...),
)

// Server
attrs := httpconv.ServerRequest("server-name", r)
ctx, span := h.tracer.Start(ctx, "HTTP "+r.Method,
    trace.WithSpanKind(trace.SpanKindServer),
    trace.WithAttributes(attrs...),
)
```

