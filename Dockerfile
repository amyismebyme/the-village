# syntax=docker/dockerfile:1

# -----------------------------------------------------------------------------
# Build stage
# -----------------------------------------------------------------------------

FROM golang:1.26.5-alpine AS builder

WORKDIR /src

# Install certificates needed for HTTPS during dependency downloads.
RUN apk add --no-cache ca-certificates

# Copy dependency manifests first so Docker can cache dependency downloads.
COPY apps/api/go.mod apps/api/go.sum ./apps/api/

WORKDIR /src/apps/api


COPY zscaler.crt /usr/local/share/ca-certificates/zscaler.crt
RUN update-ca-certificates
RUN go mod download

# Copy the API source.
WORKDIR /src
COPY apps/api ./apps/api

WORKDIR /src/apps/api

# Build a static Linux binary.
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/village-api \
    ./cmd/api


# -----------------------------------------------------------------------------
# Runtime stage
# -----------------------------------------------------------------------------

FROM alpine:3.22 AS runtime

WORKDIR /app

# Install CA certificates for outbound HTTPS requests.
RUN apk add --no-cache ca-certificates \
    && addgroup -S village \
    && adduser -S village -G village

COPY --from=builder /out/village-api /app/village-api

RUN chown village:village /app/village-api

USER village

EXPOSE 8080

ENTRYPOINT ["/app/village-api"]
