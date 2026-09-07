#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/../apps/api"
go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./...
