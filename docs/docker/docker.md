Build image
Run image
Docker Compose
Environment variables
Health check
Troubleshooting
## Runtime healthcheck

The API image exposes a Docker `HEALTHCHECK` against `GET /health`. Compose can therefore distinguish a started container from a healthy application process.

The current Dockerfile still supports the local corporate CA workflow through the checked-in `zscaler.crt` dependency; removing that requirement is intentionally deferred.
