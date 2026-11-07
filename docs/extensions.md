# Choose an extension contract

Start with ordinary Go functions and domain packages. Add a plugin when a feature
needs its own activation, optional installation or disposal. Add a shared contract
when plugins need to communicate across that boundary.

The [two-plugin notebook example](../examples/extensions/README.md) exercises all
four mechanisms. These are app-owned contracts under the example's
`internal/contract/notebook`; they add no public kit API.

## Choose by intent

| Intent | Mechanism | Example | Responsibility |
| --- | --- | --- | --- |
| Call one capability and get its result | `Key[T]` service | `Notebook.Add(title)` | The provider owns the operation and its state |
| Tell interested plugins something happened | `Topic[T]` | `Added` after an addition | Listeners observe; their return values cannot veto the operation |
| Prepare a value through ordered extensions | `Pipeline[T]` | Normalize a title, then apply workspace policy | Stages prepare; the caller applies deferred commits after using the result |
| Collect contributions for a host to consume | `Point[T]` | Add welcome note and List notes actions | Contributors publish entries; the host chooses when to invoke them |

A method call expresses a request with an owner and a result. An event announces
a fact. A pipeline makes preparation extensible. A point makes a collection
extensible. Use the smallest mechanism that expresses the interaction.

## Keep the contract with the app

Put interfaces, payload types and declarations in `internal/contract/<capability>`.
Keep implementation and mutable state in the owning plugin. Both plugins import
the contract; neither imports the other. Only main composes their factories. The
example's import boundaries are enforced by the kit's depguard configuration.

For each declaration, document:

- **Owner and intent:** who provides the service, emits the topic, runs the
  pipeline or consumes the point; what an observer or extension may change.
- **Payload:** its meaning and ownership. Prefer immutable values; document any
  copying required for slices, maps or pointers.
- **Thread:** callbacks and plugin state belong to the UI thread. Workers use
  immutable inputs and return results or post closures through a live scope.
- **Ordering:** whether it is subscription order, transform order or contribution
  order. Assign explicit order numbers when one extension must precede another.
- **Errors:** ordinary rejection versus panic, who handles the result, and which
  state changes or notifications have already happened.
- **Lifetime:** required versus optional service use, the scope owning each
  registration, and what remains usable after disposal or reactivation.

The example's [contract comments](../examples/extensions/internal/contract/notebook/notebook.go)
record these details next to its declarations.

## Follow the example

Run from the kit checkout:

```sh
GOWORK=off go run ./examples/extensions
GOWORK=off go test ./examples/extensions/...
```

The two plugins have different responsibilities:

| Plugin | Service | Pipeline | Topic | Point |
| --- | --- | --- | --- | --- |
| `notebook` | Provides `NotebookKey` | Trims titles and rejects blanks at order 10; runs preparation in `Add` | Emits `Added` after append and commits | Contributes `notebook.list` at order 20 |
| `workspace` | Requires `NotebookKey` | Limits prepared titles to 32 runes at order 20; defers its counter increment | Observes `Added` and counts additions | Contributes `workspace.add` at order 10 |

1. Main installs workspace before notebook. `Requires` delays workspace's Setup
   until `NotebookKey` exists, so installation order is not manual boot sequencing.
2. Main reads `Items(Actions)`. The add action sorts before the list action because
   of its explicit `Order`, even though the notebook registered its action first.
3. Main uses `Invoke` to run the add action. The action calls the required notebook
   service with `"  Hello plugins  "`.
4. `Add` calls `Run(Preparing)`. Notebook trims the title; workspace validates its
   length and returns a commit closure without changing its counter.
5. After preparation succeeds, `Add` appends the title, calls the returned commit
   and emits `Added`. Workspace's listener runs synchronously before `Add` returns.
6. The add action returns `Hello plugins (committed=1, observed=1)`. Main then
   invokes List notes, which returns `Hello plugins`.

The example performs no I/O in either plugin. Main prints action results. A native
app can invoke the same contracts from handlers and render stored state.

## Use each mechanism deliberately

### Services: a capability with one provider

Declare provided and required keys in manifests; use `Provide` and `Require` in
Setup. A required consumer can retain the service for its activation: provider
removal disposes the consumer first and leaves it Blocked. When the provider
returns, both receive fresh activations as needed.

An optional consumer uses `Use` on each operation and handles absence. It remains
active when the provider disappears, so retaining the old interface is unsafe.
Declaring a topic, pipeline or point does not require a service dependency; add
`Requires` only when the plugin actually needs that service to activate.

### Topics: synchronous facts

`Emit` calls listeners synchronously in subscription order. It does not wait for
background work or collect results. A listener panic fails its owning plugin in
non-strict mode while other live listeners continue. Register with `On` on the
scope that owns the observation; disposal removes the subscription.

Emit `Added` after the operation has happened. Validation belongs before the
append; a listener cannot undo it. In this example, the service returns a snapshot
for initial/current state and the topic announces subsequent changes.

### Pipelines: preparation and deferred effects

`Transform` stages run by numeric order, then registration order. They return a
prepared value, an optional commit and an error. Stages leave plugin state alone
while preparing. An error stops the chain; `Run` returns no commit, and the caller
must not use the rejected result. An ordinary error is attributed to its stage's
plugin without failing that plugin; a panic can fail it.

The caller may drop a successful result without committing. Each returned commit
runs at most once and only while the caller and relevant stage scopes remain
alive. This is deferred application, not a database transaction: commits do not
provide rollback if a later commit panics, and external writes need their own
consistency policy. This pipeline is also not around middleware with a `next`
continuation or a first-handler-wins decision hook.

### Points: contributions with a host

`Contribute` installs entries under plugin-prefixed IDs. `Items` orders them by
`Order`, then host before plugins in install order, then registration order.
The returned slice is cached and must not be modified. A point does nothing until
a host reads and consumes it.

Use `Invoke` for contributed callbacks so panics are attributed to the contributor
and cached entries from disposed scopes cannot run. Reading plain entry data is
fine; calling `entry.Value.Run()` directly skips these protections. Collections
that must react outside rendering can listen for `PointChanged` and re-read
`Items`. A queued host coalesces changes until that notification runs; an
immediate `Post` callback announces each change as it happens.

## Test boundaries and lifetimes

The example's tests exercise reversed installation order, point ordering, title
normalization, rejection, dropped preparations and downstream failure without
premature commits or events. They also check that disabling workspace removes its
policy and action while notebook remains active, and that reactivation starts new
workspace counters. Disabling notebook blocks workspace and removes both plugins'
actions; enabling it restores fresh activations and an empty notebook.

Use `kerneltest` for deterministic integration tests. For navigation-dependent
subscriptions or transforms, register on a child scope and dispose it on exit;
the same ownership rules apply. See [kernel API](kernel-api.md) for exact lifecycle
semantics and the [async document recipe](recipes/async-document.md) for workers
and navigation lifetimes.
