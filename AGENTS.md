# AGENTS.md — wada-kit

A reusable native MyGo desktop kit. This repository owns one Go module,
`github.com/xleotranx/wada-kit`, with independent tagged releases. It never imports
an app, PiG, or an app's contracts. Apps consume it; their behavior stays in their
own repositories.

## Source map

- `kernel`: standard library only; plugin lifecycle, scopes, services, events,
  pipelines, extension points, commands, tasks and settings.
- `kernel/kerneltest`: deterministic kernel harness and fake store.
- `contract/shell` and `contract/settings`: public declarations shared with plugins.
- `shell`: native window, layout, menus, shortcuts and SQLite settings store.
- `native`: semantic palette and small shared controls.
- `plugins/settings` and `plugins/inspector`: optional built-in plugins using the
  same public contracts as app plugins.
- `kittest`: fake host, palette guard and goroutine guard for consumers.
- `examples/notes`: small reference app; must never need Wada.

## Rules

1. Register through `kernel.Context`; scoped registrations are reversible. Plugins
   never import another plugin or private shell implementation.
2. Plugin state belongs to the UI thread. Workers use `kernel.Go` or `Scope.Spawn`,
   receive immutable inputs, and return values or post closures. No raw `go`
   statements in plugins. Navigation-dependent work belongs to a child scope.
3. Rendering performs no I/O or kernel mutation. Actions belong in handlers.
   Every color comes from `native.PaletteFor`; depguard and guard tests enforce
   the boundaries.
4. Check APIs against the pinned MyGo source before UI work:
   `go list -m -f '{{.Dir}}' github.com/egoist/mygo`, especially `ui/doc.go`.
   Keep `ui.Element`, build parts and `*ui.Context` out of persistent state and
   goroutines. Use persistent handles and list states, keys before stateful
   controls, and `ItemKey` on lists.
5. Keep the public API small. App-specific contracts and feature policy belong in
   apps. Review interface additions, behavior, settings schema and UI dependency
   changes for consumer impact. Follow `docs/maintenance.md` for versioning.
6. Documentation describes the code as it is. Update `docs/kernel-api.md` or
   `docs/shell.md` in the same change as the API. Never require a sibling app to
   build or test this repository.

## Verification

`GOWORK=off make check` runs MyGo lifetime vet, pinned lint, race tests and no-cgo
tests, including the Notes app. `go build ./...` checks compilation. Native checks
run on macOS. `go run ./examples/notes` runs the reference app; use a temporary
`NOTES_DATA_DIR` for experiments and captures, never developer data.

See `README.md` for private-module access and optional local workspaces.
