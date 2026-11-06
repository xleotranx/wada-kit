# Building this app

Read README.md first. main.go composes the shell and plugin factories.
internal/documents owns worker-safe domain access. internal/plugins/documents
owns registration, UI state, document lifetimes, actions and rendering.

First run `GOWORK=off go mod tidy`, then `GOWORK=off make check`. Locate the kit's
public documentation with `go list -m -f '{{.Dir}}' github.com/xleotranx/wada-kit`:
read docs/kernel-api.md for lifecycle/contracts and docs/shell.md for host APIs.

- Keep domain logic in ordinary Go packages. Add a plugin when it needs its own
  lifecycle or optional activation. Add internal/contract/<capability> only when
  plugins need a shared typed service, event or extension point.
- Plugins import contracts, domain packages, kernel and native; never another
  plugin or the kit's shell implementation. Only main composes plugins and shell.
- Setup registers through kernel.Context. Acquire resources there and register
  cleanup. Factories make fresh instances; OnDispose must return promptly.
- Plugin state belongs to the UI thread. Use kernel.Go for results and Scope.Spawn
  for streams. Workers capture immutable inputs and worker-safe repositories;
  never capture mutable plugin state, ui.Element, build parts or *ui.Context.
- Use a child scope per selected document; dispose it on switch and page exit.
  Rendering performs no I/O or kernel mutation. Actions belong in handlers.
- Colors come from native.PaletteFor. Use keys before stateful controls, persistent
  handles for focus, persistent list states and ItemKey for virtual lists.
- Before using UI APIs, read the pinned source with
  `go list -m -f '{{.Dir}}' github.com/egoist/mygo`, especially ui/doc.go.
  Do not guess methods from another
  UI toolkit or update dependencies just to make an invented method compile.
- Run `GOWORK=off make check`. Tests use kerneltest's queued work/posts, kittest's
  fake host, ui.Tester and palette/goroutine guards. Test cancellation and errors.
- Use a temporary APP_DATA_DIR (with the prefix in main.go) for experiments and
  captures. Never use developer data. Follow README.md for capture commands.
- Keep the kit dependency tagged for releases. An explicit local development
  replacement must be removed and the tagged graph checked before release.

When adding a feature, update this map and README.md so the next agent can find
its domain code, plugin, tests and verification commands.
