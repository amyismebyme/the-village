# ---------- Build Stage ----------
FROM golang:1.26 AS builder

WORKDIR /app

COPY apps/api/go.mod .
COPY apps/api/go.sum .

#only needed to bypass the prpxy on laptop
COPY zscaler.crt /usr/local/share/ca-certificates/
RUN update-ca-certificates
RUN go mod download

COPY apps/api .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o village-api ./cmd/api

# ---------- Runtime Stage ----------
FROM alpine:3.18.6

RUN addgroup -S village && \
    adduser -S village -G village

WORKDIR /app

COPY --from=builder /app/village-api .

USER village

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s \
CMD wget --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["./village-api"]