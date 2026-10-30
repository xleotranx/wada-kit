# Changelog

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
