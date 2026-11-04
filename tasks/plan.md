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

## Next increment: shutdown guarantees

Use existing Scope.Spawn workers for potentially blocking resource cleanup after
Lifetime cancellation. Keep OnDispose synchronous, quick and UI-thread-only;
neither context cancellation nor the kernel can interrupt an arbitrary disposer.
No new lifecycle API is needed for this increment.

At Close entry, make the kernel terminal and cancel all active plugin lifetimes
and the host lifetime before invoking any disposer. Reject Install/Start after
Close and activation during Close. Preserve dependent-first disposal, final
settings persistence and idempotence. Return attributed disposer failures alongside
worker/settings errors. Keep host panics and Strict behavior unchanged.

Test cancellation order, rejected activation/reuse, disposer error reporting,
final settings written by disposers, and blocking worker cleanup bounded by the
close context. Verify the standalone gate, build and external Notes consumer.
Do not add hooks or move UI disposal onto workers. App migrations remain deferred.

## Next increment: scope and capability lifetimes

Exercise optional-provider disable/failure/re-enable, required dependent recovery,
and old worker results/panics. Fix only demonstrated lifetime gaps: stale contexts
must not provide services or poison a newer Setup; contributions must register
their undo before immediate PointChanged callbacks; disposed callers must not
emit events, run pipeline stages or apply pending pipeline commits.

Keep read-only Use/Items semantics and public signatures. Optional consumers must
re-query Use rather than cache a provider. Preserve required-service availability
during dependent cleanup. Cover child scopes and replacement activations, document
the resulting behavior, and run the standalone gate/build/Notes consumer check.
