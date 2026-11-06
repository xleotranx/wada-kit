# Build an app with wada-kit

Start here when a user points you at this repository and asks for a new native
app. Ask for missing product requirements, then create the app in its own
repository. The kit supplies the host and lifecycle; your app owns its features,
data, backend, branding and packaging.

## Create and verify

Requirements: Bash (including macOS Bash 3.2), standard Unix tools, the Go version
in the kit's go.mod, make, and MyGo's native platform requirements. Native checks
run on macOS. No Node,
frontend bundler, sibling Wada checkout or GitHub credentials are required.

From this kit checkout, choose a new destination, module path and app identity:

```sh
bash scripts/new-app.sh ../my-app example.com/my-app \
  --name "My App" --folder my-app --env MY_APP \
  --identifier com.example.myapp
cd ../my-app
GOWORK=off go mod tidy
GOWORK=off make check
make run
```

The command refuses an existing destination. It creates an independent module
pinned to v0.3.0 and the kit checkout's Go/MyGo/Onevra combination. Use
`--kit-version vX.Y.Z` to select another published release compatible with this
starter. It never writes to developer data during generation or tests.

For coordinated development against this checkout, add
`--kit-dir /absolute/path/to/wada-kit` to the creation command. This explicitly
adds a local replacement. Before releasing, choose a published kit tag, remove
that replacement with `go mod edit -dropreplace=github.com/xleotranx/wada-kit`,
run `GOWORK=off go mod tidy`, and repeat `GOWORK=off make check`. A local checkout
must not hide an unpublished API dependency. See [maintenance](maintenance.md).

The generated README and AGENTS.md explain that app's map and commands. Keep them
current as features change. The kit's own AGENTS.md governs kit maintenance;
the generated app has consumer instructions suited to its own repository.

## Start with one feature

```text
main.go                              app identity and composition
internal/documents/repository.go      ordinary domain code and storage boundary
internal/plugins/documents/plugin.go  registration, actions and scoped work
internal/plugins/documents/view.go    renders current UI state
internal/plugins/documents/*_test.go  deterministic feature tests
internal/contract/<capability>/       add only for communication between plugins
```

Adapt the Documents plugin to the first requested feature. Keep small components
as functions and business logic as ordinary Go packages. Add a plugin when a
feature needs independent activation, disposal or optional installation. Plugins
never import each other: shared capabilities belong in app-owned contracts.
Only main imports the shell implementation and composes plugin factories.

For the first feature, an agent needs Plugin/Setup, a page contribution, handlers,
a Task, a child Scope and the palette. Services, events, pipelines and custom
extension points are available when a real interaction needs them; consult the
[kernel API](kernel-api.md) at that point.

## Choose the right example

| Need | Read |
| --- | --- |
| Smallest page, command and settings section | `examples/notes/main.go` |
| Independent app structure and checks | `examples/starter/` and `scripts/new-app.sh` |
| Document loads, navigation, errors and cancellation | [Async document recipe](recipes/async-document.md) |
| Typed communication between plugins | [Kernel concepts and contracts](kernel-api.md#concepts) |
| Window slots, menus, data directory and screenshots | [Shell API](shell.md) |

The starter's embedded files keep it runnable without writing domain data. Replace
its repository with app-owned storage or a backend. Preferences belong in
`kernel.Setting`; document bodies and domain databases belong in the app's
repository. For app files, choose a root under Host.DataDir with the plugin's ID;
handle the empty directory used in tests and perform I/O in workers.

## Build and inspect each feature

Read the pinned MyGo source before using controls:

```sh
go list -m -f '{{.Dir}}' github.com/egoist/mygo
```

Read that directory's `ui/doc.go` and the relevant control implementation. Keep
persistent handles and list state in app state, use keys before stateful controls
and ItemKey on lists. Frame elements, build parts and *ui.Context never escape
the render callback. All colors come from native.PaletteFor.

Run `GOWORK=off make check`: lifetime vet, import lint, race tests, no-cgo tests
and compilation. Extend the generated import allowlist when adding domain
packages; preserve the ban on plugin-to-plugin and plugin-to-shell imports.
Test actions with the queued kernel harness and UI with ui.Tester. Include error,
retry, stale-result, page-exit and shutdown cases for background work.

For a native capture, use disposable data and an absolute output path:

```sh
capture_data=$(mktemp -d)
MY_APP_DATA_DIR="$capture_data" MY_APP_CAPTURE=/tmp/my-app.png \
  MY_APP_APPEARANCE=light go run .
```

Inspect both light and dark appearance. The app exits after capture. Delete the
disposable directory when finished. Never substitute a developer data directory.

`make package` uses the pinned MyGo tool and generated mygo.json. The starter
provides metadata, not a distribution identity: set your actual bundle identifier,
icon, version, signing and release policy before shipping.

## Agent handoff

A useful user prompt is:

> Build a native [app purpose] for [users], with [first features]. Use the
> wada-kit checkout at [path]. Follow docs/build-an-app.md, create an independent
> app at [destination], and run its documented checks and a native capture.

Before handing back the app, record its feature map, what works, commands and
results, and remaining product decisions in its own README/AGENTS.md. The next
agent should need that app's source and the kit's public documentation to extend it.
