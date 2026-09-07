#!/usr/bin/env sh
set -eu

: "${REDDIT_CLIENT_ID:?REDDIT_CLIENT_ID must be set}"
: "${REDDIT_CLIENT_SECRET:?REDDIT_CLIENT_SECRET must be set}"
: "${REDDIT_USER_AGENT:?REDDIT_USER_AGENT must be set}"

export REDDIT_INGEST_SUBREDDIT="${1:-toronto}"

cd "$(dirname "$0")/../apps/api"
go test -tags=smoke ./internal/external/reddit \
    -run '^TestRedditLiveSmoke$' \
    -count=1 \
    -v
