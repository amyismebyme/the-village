# syntax=docker/dockerfile:1

FROM golang:1.26.5-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY apps/api/go.mod apps/api/go.sum ./apps/api/
WORKDIR /src/apps/api

# Optional corporate CA. Provide with: --secret id=zscaler,src=zscaler.crt
RUN --mount=type=secret,id=zscaler,dst=/tmp/zscaler.crt \
    if [ -f /tmp/zscaler.crt ]; then \
        cp /tmp/zscaler.crt /usr/local/share/ca-certificates/zscaler.crt && \
        update-ca-certificates; \
    fi

RUN go mod download

WORKDIR /src
COPY apps/api ./apps/api

WORKDIR /src/apps/api

ARG VERSION=0.1.2
ARG GIT_COMMIT=local
ARG BUILD_TIME=
ARG ENVIRONMENT=production

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w -X github.com/amyismebyme/the-village/apps/api/internal/runtime.BuildVersion=${VERSION} -X github.com/amyismebyme/the-village/apps/api/internal/runtime.GitCommit=${GIT_COMMIT} -X github.com/amyismebyme/the-village/apps/api/internal/runtime.BuildTimestamp=${BUILD_TIME} -X github.com/amyismebyme/the-village/apps/api/internal/runtime.Environment=${ENVIRONMENT}" \
    -o /out/village-api \
    ./cmd/api

FROM alpine:3.22 AS runtime

WORKDIR /app

RUN apk add --no-cache ca-certificates \
    && addgroup -S village \
    && adduser -S village -G village

COPY --from=builder /out/village-api /app/village-api

RUN chown village:village /app/village-api

USER village

EXPOSE 8080

ENTRYPOINT ["/app/village-api"]
