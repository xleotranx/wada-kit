#!/usr/bin/env bash
# Test public packages from a separate module, or download a published tag.
set -euo pipefail
kit_root="$(cd "$(dirname "$0")/.." && pwd)"
consumer_dir="$(mktemp -d)"
trap 'rm -rf "$consumer_dir"' EXIT
cp "$kit_root"/examples/notes/*.go "$consumer_dir/"
cd "$consumer_dir"
export GOWORK=off
go mod init example.com/wada-kit-consumer
if [[ $# -eq 0 ]]; then
  go mod edit -require=github.com/xleotranx/wada-kit@v0.0.0
  go mod edit "-replace=github.com/xleotranx/wada-kit=$kit_root"
else
  go mod edit "-require=github.com/xleotranx/wada-kit@$1"
fi
go mod tidy
go test ./...
go build ./...
