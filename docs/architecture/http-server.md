# HTTP Server Architecture

## Purpose

The Village API uses Go's `http.Server` instead of `http.ListenAndServe`.

This provides:

- Configurable timeouts
- Graceful shutdown
- Better production readiness

---

## Timeouts

### Read Timeout

Limits the amount of time allowed to read the request.

Default:

10 seconds

---

### Write Timeout

Limits the amount of time allowed to write the response.

Default:

10 seconds

---

### Idle Timeout

Maximum keep-alive duration.

Default:

60 seconds

---

### Shutdown Timeout

Maximum amount of time the application waits for in-flight requests to finish before terminating.

Default:

15 seconds

---

## Graceful Shutdown Flow

Application starts

↓

Listen for HTTP requests

↓

Receive SIGTERM

↓

Stop accepting new requests

↓

Finish active requests

↓

Exit cleanly

---

## Future Improvements

- HTTP/2
- TLS support
- ReadHeaderTimeout
- MaxHeaderBytes
- Request logging middleware
- Request ID middleware
- Prometheus metrics