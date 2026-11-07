#!/usr/bin/env bash
# Create an independent app from the compiled starter. Supports macOS Bash 3.2.
set -euo pipefail
export LC_ALL=C
kit_root="$(cd "$(dirname "$0")/.." && pwd)"
usage() {
  cat <<'HELP'
Usage: bash scripts/new-app.sh DESTINATION MODULE [options]

  --name NAME           Display name (default: My App)
  --folder FOLDER       Unique data folder (default: destination name)
  --env PREFIX          Environment prefix (default: upper-case folder)
  --identifier ID       Bundle identifier (default: com.example.<folder>)
  --kit-version TAG     Published kit version (default: v0.3.1)
  --kit-dir PATH        Explicit local development replacement
  -h, --help            Show this help

The destination must not exist. After creation, run GOWORK=off go mod tidy
and make check in the new app. Requires Bash, standard Unix tools and Go.
HELP
}
fail() { printf 'new-app: %s\n' "$*" >&2; exit 1; }
if [[ "${1:-}" == --help || "${1:-}" == -h ]]; then usage; exit 0; fi
[[ $# -ge 2 ]] || { usage >&2; exit 1; }
destination=$1
module=$2
shift 2
app_name='My App'
folder=''
env_prefix=''
identifier=''
kit_version=v0.3.1
kit_dir=''
while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help) usage; exit 0 ;;
    --name|--folder|--env|--identifier|--kit-version|--kit-dir)
      [[ $# -ge 2 ]] || fail "missing value for $1"
      [[ -n "$2" ]] || fail "empty value for $1"
      case "$1" in
        --name) app_name=$2 ;;
        --folder) folder=$2 ;;
        --env) env_prefix=$2 ;;
        --identifier) identifier=$2 ;;
        --kit-version) kit_version=$2 ;;
        --kit-dir) kit_dir=$2 ;;
      esac
      shift 2 ;;
    *) fail "unknown option: $1" ;;
  esac
done
[[ -n "$destination" ]] || fail 'destination must not be empty'
case "$destination" in /*) ;; *) destination="$PWD/$destination" ;; esac
folder=${folder:-$(basename "$destination")}
env_prefix=${env_prefix:-$(printf '%s' "$folder" | tr '[:lower:]-' '[:upper:]_')}
identifier=${identifier:-com.example.$folder}
[[ "$module" =~ ^[a-z0-9][a-z0-9.-]*(/[A-Za-z0-9._-]+)+$ ]] || fail 'module must be a Go module path such as example.com/my-app'
case "/$module/" in */./*|*/../*) fail 'module must not contain . or .. path segments' ;; esac
[[ "$folder" =~ ^[a-z][a-z0-9-]*$ ]] || fail 'folder must start with a letter and contain lower-case letters, digits or hyphens'
[[ "$env_prefix" =~ ^[A-Z][A-Z0-9_]*$ ]] || fail 'env must be an upper-case environment prefix'
[[ "$app_name" =~ [^[:space:]] && ! "$app_name" =~ [[:cntrl:]] ]] || fail 'name must be nonempty single-line text without control characters'
[[ "$kit_version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z0-9.-]+)?$ ]] || fail 'kit-version must be an explicit version tag'
[[ "$identifier" =~ ^[A-Za-z][A-Za-z0-9-]*(\.[A-Za-z][A-Za-z0-9-]*)+$ ]] || fail 'identifier must be a reverse-DNS bundle identifier'
[[ ! -e "$destination" && ! -L "$destination" ]] || fail 'destination already exists; choose a new directory'
if [[ -n "$kit_dir" ]]; then
  kit_dir="$(cd "$kit_dir" && pwd)" || fail 'kit-dir must be a directory'
  [[ -f "$kit_dir/go.mod" && "$(head -n 1 "$kit_dir/go.mod")" == 'module github.com/xleotranx/wada-kit' ]] || fail "kit-dir must contain wada-kit's go.mod"
fi
# Go and JSON accept the same escaping for this validated single-line name.
name_json=${app_name//\\/\\\\}
name_json=${name_json//\"/\\\"}
name_json="\"$name_json\""
mkdir -p "$(dirname "$destination")"
mkdir "$destination" # also refuses a destination created concurrently
cp -R "$kit_root/examples/starter/." "$destination/"
while IFS= read -r source; do
  sed "s|github.com/xleotranx/wada-kit/examples/starter|$module|g" "$source" > "$source.tmp"
  mv "$source.tmp" "$source"
done < <(find "$destination" -type f -name '*.go')
# Pass user text as data, never as awk code or a replacement expression.
APP_NAME_JSON="$name_json" APP_FOLDER="$folder" APP_ENV="$env_prefix" awk '
  /Name: "Starter", Folder: "wada-kit-starter", Env: "STARTER"/ {
    printf "\t\tName: %s, Folder: \"%s\", Env: \"%s\",\n", ENVIRON["APP_NAME_JSON"], ENVIRON["APP_FOLDER"], ENVIRON["APP_ENV"]
    next
  }
  { print }
' "$destination/main.go" > "$destination/main.go.tmp"
mv "$destination/main.go.tmp" "$destination/main.go"
APP_NAME="$app_name" APP_ENV="$env_prefix" awk '
  NR == 1 { print "# " ENVIRON["APP_NAME"]; next }
  { gsub(/STARTER_/, ENVIRON["APP_ENV"] "_"); print }
' "$destination/README.md" > "$destination/README.md.tmp"
mv "$destination/README.md.tmp" "$destination/README.md"
# Keep the checkout's exact Go/UI dependency combination and tool declaration.
sed "s|^module github.com/xleotranx/wada-kit$|module $module|" "$kit_root/go.mod" > "$destination/go.mod"
printf '\nrequire github.com/xleotranx/wada-kit %s\n' "$kit_version" >> "$destination/go.mod"
if [[ -n "$kit_dir" ]]; then
  (cd "$destination" && GOWORK=off go mod edit "-replace=github.com/xleotranx/wada-kit=$kit_dir")
fi
cat > "$destination/mygo.json" <<META
{
  "name": $name_json,
  "identifier": "$identifier",
  "version": "0.1.0",
  "out": "build"
}
META
printf '/build/\n.DS_Store\n' > "$destination/.gitignore"
cat > "$destination/.golangci.yml" <<'LINT'
version: "2"
linters:
  default: none
  enable: [depguard, ineffassign, staticcheck, unused]
  settings:
    depguard:
      rules:
        plugins:
          files: ["**/internal/plugins/**/*.go", "!$test"]
          list-mode: strict
          allow:
            - "$gostd"
            - "github.com/egoist/mygo/ui"
            - "github.com/oneo-me/onevra/source"
            - "github.com/xleotranx/wada-kit/kernel"
            - "github.com/xleotranx/wada-kit/contract"
            - "github.com/xleotranx/wada-kit/native"
LINT
printf '            - "%s/internal/documents"\n            - "%s/internal/contract"\n' "$module" "$module" >> "$destination/.golangci.yml"
printf 'Created %s. Run GOWORK=off go mod tidy, then make check and make run there.\n' "$destination"
if [[ -n "$kit_dir" ]]; then
  printf '%s\n' 'Local development replacement added; remove it and verify a released tag before distribution.'
fi
