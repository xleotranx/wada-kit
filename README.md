# wada-kit

The core and window of a native MyGo desktop app, without the app: a plugin kernel,
a window shell (title strip, rounded panel, clear or frosted backdrop, light and dark
appearance, menus, shortcuts, a SQLite settings store), the palette every pane
paints with, a settings page and plugin diagnostics. An app adds its own plugins and
contracts on top. Apps consume tagged releases from this independent repository.

| Package | What it is |
| --- | --- |
| `kernel` | Plugin lifecycle, services, events, pipelines, extension points, commands, tasks, settings; standard library only |
| `kernel/kerneltest` | A kernel for tests, with background work and posts the test runs |
| `contract/shell` | The window's slots (`Pages`, `Strip`, `TitleBar`, `Overlays`) and `shell.Host` |
| `contract/settings` | Settings sections |
| `shell` | The window host: `shell.Run(app, plugins...)`, and `shell.New` for tests |
| `native` | The palette (`native.PaletteFor`) and small shared controls |
| `plugins/settings` | The settings page, with Appearance and Plugins |
| `plugins/inspector` | Plugin diagnostics |
| `kittest` | A fake `shell.Host` and the palette guard for an app's tests |
| `examples/notes` | The smallest app on the kit |

Notes: [kernel API](docs/kernel-api.md), [shell and the kit's plugins](docs/shell.md).

## Use

```go
shell.Run(shell.App{Name: "Notes", Folder: "notes", Env: "NOTES", Width: 900, Height: 640},
	newNotes, inspector.New, settings.New)
```

Requires the Go version in `go.mod`. The desktop packages use MyGo and Onevra;
`kernel` uses only the standard library. Native checks run on macOS.

The repository is public. Normal Go module downloads need no GitHub credentials,
deploy keys, or `GOPRIVATE` configuration.

Add the kit to your app:

```sh
go get github.com/xleotranx/wada-kit@v0.1.0
```

An app owns its `main.go`, plugins, contracts, packaging configuration, bundle
identifier and assets. Give each app its own `shell.App.Folder` and `Env` so data
and environment variables are separate. Use `examples/notes` as a small reference;
`go run ./examples/notes` runs it from a kit checkout.

The supported dependency combination is pinned in `go.mod`. Apps that also depend
on MyGo or Onevra should test the versions Go selects in their own module graph.
The shell opens one main window; the palette's clear-window setting is process-wide.

## Develop with an app

Keep normal app dependencies on tagged versions. For coordinated local changes,
create an uncommitted workspace in a directory containing both repositories:

```sh
go work init ./wada-kit ./my-app
```

Run app tests normally to use the local kit. Run `GOWORK=off make check` in each
repository to validate its declared dependencies. Do not commit a local `replace`
or workspace as an app's release dependency.

[Maintenance and releases](docs/maintenance.md) describes ownership, compatibility,
private CI access, and the release checks. [Changelog](CHANGELOG.md) records releases.

## Check

`make check`: MyGo lifetime vet, pinned lint (the import boundaries; nothing here
imports an app), race tests and no-cgo tests.
