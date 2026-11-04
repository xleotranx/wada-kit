# Plugin installation and activation hardening

Focus on wada-kit first; Terminal and Cowork migrations follow later.

## Scope

Make plugin construction obey the same panic policy as Setup. A failed Install
batch must preserve previously installed plugins and discard all new records,
including on a strict-mode panic. Later construction failures mark only that
plugin Failed, leave its dependents Blocked, and permit an explicit retry.
Factories must remain free of resource acquisition: registry rollback cannot undo
external side effects. Keep public signatures and successful activation behavior.

## Ordered tasks

1. Reproduce installation rollback and construction panic failures with tests.
2. Implement transactional registry installation and guarded construction.
3. Document the behavior, review the diff, and run the standalone kit gate,
   build, and external Notes consumer check.

## Deferred work

Explicit restart policy and shutdown guarantees require separate changes and
compatibility review. Capability extraction, readiness and command redesign are
outside this increment. Do not modify or require sibling apps.

## Next increment: explicit restart policy

Add `Manifest.Restart *Restart`: nil inherits an explicitly configured kernel
default, a pointer to the zero value disables restart, and positive Max/Within
values opt in with a bounded budget. Snapshot the policy at installation so caller
mutation cannot alter restart safety. The shell stops setting an implicit global
policy; its plugins stay Failed unless their manifests opt in. Keep the existing
kernel Options.Restart for hosts that deliberately set a default. Nonpositive
limits disable restart, including malformed time windows.

Verify per-plugin precedence, budgets and time windows, default-off shell behavior,
failed operations not being replayed, and existing strict/cleanup/retry behavior.
Update lifecycle and shell documentation and provide a next-minor upgrade note.
Run the standalone gate, build and external Notes consumer. App migrations and
shutdown API changes remain separate work.

## Risks

Strict mode must continue propagating panics. Rollback applies to registry state,
not arbitrary factory side effects. Construction failure must not queue an
automatic restart or leave services, scopes or contributions registered.
