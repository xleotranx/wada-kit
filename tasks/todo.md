# Plugin hardening tasks

- [x] Reproduce failed-batch rollback, installation construction panics, and
  reactivation construction panics. Cover strict mode, blocked dependents and
  successful explicit retries. Verify new tests fail against the original code.
- [x] Implement guarded construction and rollback without changing public
  signatures. Verify focused tests and the kernel suite pass.
- [x] Update kernel lifecycle documentation and changelog; review the change.
  Run `GOWORK=off make check`, `GOWORK=off go build ./...`, and
  `scripts/check-consumer.sh`. Commit the verified increment.

Verification passed: MyGo lifetime vet, pinned lint, race tests, no-cgo tests,
all-package build, and the external Notes consumer. Regression tests were observed
failing against the original implementation before the fix.

## Explicit restart policy

- [x] Reproduce shell default restarts and nonpositive-window restart loops.
  Add tests for per-plugin override, budget isolation and snapshot behavior.
- [x] Add manifest policy overrides and remove implicit shell restarts. Verify
  that a restart does not replay a failed user command and existing tests pass.
- [x] Document the minor-release behavior change and upgrade path. Review the
  diff, run the standalone gate, build and external Notes consumer, and commit.

Restart verification passed: lifetime vet, lint, race and no-cgo suites, build,
and external Notes consumer. New tests failed before behavior changes; a guard
mutation confirmed that the zero-window regression is detected.

## Shutdown guarantees

- [x] Add regression tests for all-lifetime cancellation, activation during close,
  reuse after close and attributed disposer failures; verify the gaps reproduce.
- [x] Harden Close entry and error reporting. Verify bounded worker cleanup and
  final settings persistence without a new lifecycle API.
- [x] Document the shutdown contract and blocking-cleanup pattern. Review, run
  the full standalone gate, build and external Notes consumer, then commit.

Shutdown verification passed: lifetime vet, lint, race/no-cgo suites, build and
external Notes consumer. Regression assertions failed before the fix; omitting
early cancellation also failed the cancellation-order test. No new lifecycle API.

## Scope and capability lifetimes

- [x] Reproduce stale-context registration/setup failures, immediate contribution
  disposal leaks, and disposed caller event/pipeline delivery.
- [x] Fix demonstrated gaps and verify provider removal/recovery and old-worker
  isolation, including child scopes and replacement activations.
- [x] Document lifetime semantics, review the diff, run the full standalone gate,
  build and external Notes consumer, then commit.

Lifetime verification passed: lifetime vet, lint, race/no-cgo tests, build and
external Notes consumer. Reproduction tests failed before the fixes; removing the
commit caller-lifetime guard also failed its regression assertion. Optional and
required provider recovery and old task/post isolation pass without API additions.
