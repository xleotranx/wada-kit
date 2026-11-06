# Starter

A native desktop app with two sample documents, background loading, cancellation,
retry, saved document selection, Settings and Inspector.

After generation, run:

```sh
GOWORK=off go mod tidy
GOWORK=off make check
make run
```

The kit's app creation command uses Bash and
standard Unix tools; the generated app needs Go, make and the native MyGo platform
requirements. Native development and checks are supported on macOS.

Public API references are in the kit module's `docs/kernel-api.md` and
`docs/shell.md`; locate its source with
`go list -m -f '{{.Dir}}' github.com/xleotranx/wada-kit`. Consult those files
before inventing a service or adding a dependency. This app owns domain code,
storage, contracts, branding and packaging.

- `main.go`: app identity and composition.
- `internal/documents/repository.go`: worker-safe document access, with embedded
  samples. Replace Samples with your app's repository; keep storage out of views.
- `internal/plugins/documents/plugin.go`: setup, navigation and child scopes.
- `internal/plugins/documents/view.go`: loading, error/retry, empty and ready UI.
- `internal/plugins/documents/plugin_test.go`: stale results, cancellation, retry,
  persistence across reopen and disable cleanup.
- `guards_test.go`: palette and goroutine boundaries.

The samples never write document files. Only the last successful selection is a
setting. Domain content belongs in an app-owned repository; it should not become
one large settings value. A file-backed implementation can use documents.Files
with os.DirFS rooted at the app's chosen document directory. Keep any directory
creation or file writes in workers and report errors back to the UI.

For screenshots, set `STARTER_DATA_DIR` to a temporary directory and run with
`STARTER_CAPTURE=/absolute/path.png` and `STARTER_APPEARANCE=light` or `dark`.
`STARTER_CAPTURE_WAIT=2s` controls the delay; the app quits after capture. The
creation command replaces STARTER with your app's environment prefix.

`make package` runs the pinned MyGo CLI using mygo.json. Before distribution,
choose your real identifier, icon, version and platform signing configuration.
The creation command writes mygo.json; in the kit checkout run the reference with
`go run ./examples/starter` and check it with the root Makefile.
