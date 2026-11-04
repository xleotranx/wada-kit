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
