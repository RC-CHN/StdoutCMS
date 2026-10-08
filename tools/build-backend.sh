#!/bin/sh
set -eu

cd "$(dirname "$0")/.."
version=$(cat VERSION)
cd backend
exec go build -ldflags "-X stdoutcms/internal/handler.Version=$version" -o server ./cmd/server
