# Changelog

## Unreleased

### Fixed

- Failed `Kernel.Install` batches now discard all newly added plugin records,
  preserving earlier installations and allowing a corrected retry. Rollback also
  runs before a strict-mode panic propagates.
- In non-strict mode, plugin factory and manifest panics return installation
  errors or fail reactivation instead of escaping the kernel's recovery boundary.
  Reactivation failures retain panic diagnostics, leave dependents blocked, and
  permit an explicit retry without repeatedly restarting failed construction.

Public signatures, settings formats and restart defaults are unchanged.

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
