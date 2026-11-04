# Changelog

## v0.2.0 — 2026-11-05

### Added

- `Manifest.Restart` lets a plugin declare a bounded restart policy or explicitly
  opt out of a kernel-wide policy. Policies are copied at installation; each plugin
  keeps its own restart budget and time window.

### Changed

- **Breaking behavior:** the shell no longer automatically restarts every plugin
  up to three times per minute. Plugins stay failed by default. To restore that
  behavior for a plugin whose Setup and cleanup are safe to repeat, set its
  manifest's `Restart` to `&kernel.Restart{Max: 3, Within: time.Minute}`. Review
  side effects and dependent reactivation before opting in. Direct kernel hosts
  can still explicitly set `Options.Restart`; nil manifest policies inherit it.

### Fixed

- Failed `Kernel.Install` batches now discard all newly added plugin records,
  preserving earlier installations and allowing a corrected retry. Rollback also
  runs before a strict-mode panic propagates.
- In non-strict mode, plugin factory and manifest panics return installation
  errors or fail reactivation instead of escaping the kernel's recovery boundary.
  Reactivation failures retain panic diagnostics, leave dependents blocked, and
  permit an explicit retry without repeatedly restarting failed construction.
- Restart policies with a nonpositive count or time window now disable automatic
  restart, preventing malformed windows from bypassing the restart budget.
- Close prevents new activation and cancels every active plugin lifetime before
  invoking disposers, so cleanup can observe cancellation across dependencies.
  Install and Start now reject a kernel that was closed before startup.
- In non-strict mode, Close returns attributed disposer panics alongside worker
  and settings errors while continuing cleanup. Blocking resource cleanup remains
  supported through tracked Spawn workers; synchronous OnDispose callbacks must
  return promptly and cannot be interrupted by Close's context.
- Disposed contexts can no longer register services or poison a newer activation's
  Setup. A Setup that disposes its own context is rejected rather than marked active.
- Contributions register cleanup before notifying PointChanged listeners, so
  immediate callbacks that dispose their owner cannot leak entries or reserve IDs.
- Disposed callers cannot start events or pipelines; pending pipeline commits also
  expire with the caller's scope. Disposal during a stage aborts the run, and
  disposal during a commit prevents later commits.
- Direct disposal of an active plugin root now stops that activation as failed
  and blocks its required dependents, instead of withdrawing services while
  leaving consumers active. This does not schedule automatic restart; use child
  scopes for local work and Kernel.Enable/Disable for normal plugin management.

Settings formats and dependencies are unchanged. See the
[v0.2 upgrade guide](docs/upgrading-v0.2.md) for consumer changes.

## v0.1.2 — 2026-11-04

### Added

- `Options.Capture` lets a host render the window. The shell implements
  `Capture() ([]byte, error)` as an optional method on its host, so a plugin
  finds it with a type assertion and a kit without it still compiles. Wada
  Terminal uses it for `wada-terminal shot`.

### Changed

- MyGo v0.7.1 → v0.7.4: an idle macOS window wakes far less, terminals draw
  block sextants and themed scrollbars, and `terminal.Options.ScrollMultiplier`
  sets the wheel's distance.

The Go API is otherwise unchanged; the new method is optional.

## v0.1.1 — 2026-10-31

### Added

- MIT license for the kit source and reference app.
- Public module setup instructions requiring no GitHub credentials.

### Fixed

- Pin `golang.org/x/net v0.62.0`, which addresses GO-2026-6860 and GO-2026-6866.
  Its minimum dependency also selects `golang.org/x/term v0.47.0`.

The Go API and runtime behavior are unchanged.

## v0.1.0 — 2026-10-30

### Added

- Independently versioned desktop kit with plugin kernel, native window shell,
  palette, settings, plugin diagnostics, test harnesses and Notes reference app.
- Standalone CI, consumer smoke checks, and documented maintenance and releases.

The API is unchanged by extraction. This v0 release does not yet promise v1
compatibility; breaking changes require a minor bump and consumer upgrade notes.
