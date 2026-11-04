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
