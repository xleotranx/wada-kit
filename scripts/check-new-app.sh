#!/usr/bin/env bash
# Generator safety and identity checks, without downloading dependencies.
set -euo pipefail
kit_root="$(cd "$(dirname "$0")/.." && pwd)"
test_dir="$(mktemp -d)"
trap 'rm -rf "$test_dir"' EXIT
fail() { printf '%s\n' "$*" >&2; exit 1; }
app_dir="$test_dir/path with spaces/new-app"
bash "$kit_root/scripts/new-app.sh" "$app_dir" example.com/new-app --name 'My "App" & \\ tools $HOME' --env NEW_APP
grep -F 'Name: "My \"App\" & \\\\ tools $HOME", Folder: "new-app", Env: "NEW_APP"' "$app_dir/main.go" >/dev/null || fail 'app name was not escaped'
grep -F '"example.com/new-app/internal/documents"' "$app_dir/main.go" >/dev/null || fail 'imports were not rewritten'
grep -F '"name": "My \"App\" & \\\\ tools $HOME"' "$app_dir/mygo.json" >/dev/null || fail 'metadata name was not escaped'
grep -F 'NEW_APP_CAPTURE' "$app_dir/README.md" >/dev/null || fail 'environment prefix was not rewritten'
if grep -q 'replace' "$app_dir/go.mod"; then fail 'release module has a replacement'; fi
printf '%s\n' 'developer work' > "$app_dir/main.go"
if bash "$kit_root/scripts/new-app.sh" "$app_dir" example.com/new-app 2>/dev/null; then fail 'existing destination accepted'; fi
[[ "$(cat "$app_dir/main.go")" == 'developer work' ]] || fail 'existing file changed'
for option in --folder --env --kit-version --identifier; do
  if bash "$kit_root/scripts/new-app.sh" "$test_dir/rejected" example.com/new-app "$option" '../invalid value' 2>/dev/null; then fail "invalid $option accepted"; fi
  [[ ! -e "$test_dir/rejected" ]] || fail 'invalid input created a directory'
done
ln -s "$test_dir/missing" "$test_dir/symlink"
if bash "$kit_root/scripts/new-app.sh" "$test_dir/symlink" example.com/new-app 2>/dev/null; then fail 'dangling symlink accepted'; fi
if bash "$kit_root/scripts/new-app.sh" "$test_dir/rejected" example.com/new-app --name 2>/dev/null; then fail 'missing option value accepted'; fi
for module in 'invalid' 'example.com/../app' 'example.com/a b'; do
  if bash "$kit_root/scripts/new-app.sh" "$test_dir/rejected" "$module" 2>/dev/null; then fail 'invalid module accepted'; fi
  [[ ! -e "$test_dir/rejected" ]] || fail 'invalid module created a directory'
done
if bash "$kit_root/scripts/new-app.sh" "$test_dir/rejected" example.com/new-app --kit-dir "$test_dir" 2>/dev/null; then fail 'invalid kit directory accepted'; fi
[[ ! -e "$test_dir/rejected" ]] || fail 'invalid kit directory created a destination'
printf '%s\n' 'App generator checks passed.'
