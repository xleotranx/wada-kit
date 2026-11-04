# Upgrading from v0.1.2 to v0.2

v0.2.0 hardens plugin construction, restart and shutdown behavior. Apps can upgrade
independently; existing release tags remain unchanged.
Consumers older than v0.1.2 must also review the intervening [changelog](../CHANGELOG.md),
including the MyGo upgrade in v0.1.2.

## Compatibility

This is a v0 minor release: restart defaults, cancellation order, error reporting
and disposed-owner behavior change. The module path remains
`github.com/xleotranx/wada-kit`.

The only exported API addition is `Manifest.Restart *Restart`.

No exported functions, methods or interfaces were removed or changed. Existing
implementations and fakes need no new methods. Convert any unkeyed Manifest
literal to named fields: adding a struct field can break positional literals even
when an API comparison classifies the addition as compatible. Reflection and
serialization of Manifest also see the new field; the settings database does not
persist manifests.

Go, MyGo, Onevra and all module dependencies are unchanged from v0.1.2. Settings
keys, namespaces and SQLite schema 1 are unchanged; no settings migration is
required by these kit changes. Apps still validate the versions selected by their
own module graphs.

## Behavior to review

| Area | v0.1.2 | v0.2.0 | Consumer action |
|---|---|---|---|
| Shell restart | Automatically restarted runtime failures up to three times per minute | No automatic restart without a manifest policy | Opt in only where repeating Setup is safe |
| Kernel restart | One explicitly configured default policy | Manifest can override that default; both limits must be positive | Give every enabled policy a positive count and time window |
| Installation | A rejected batch could leave earlier records from that batch installed | The whole failed batch rolls back; earlier successful batches remain | Retry a corrected batch; do not rely on partial installation |
| Construction panic | Factory/manifest panics escaped recovery | Non-strict mode returns an install error or fails activation; Strict still propagates panics | Handle errors and inspect failed plugin status |
| Close cancellation | Each lifetime was cancelled as its plugin was disposed | All active plugin lifetimes and the host lifetime are cancelled before any disposer | Review cleanup that depends on another worker staying alive |
| Close errors | Disposer panics appeared in plugin status | Non-strict Close also returns attributed disposer errors, joined with worker/settings errors | Handle the first Close result |
| Activation during/after Close | Some shutdown/reuse paths were accepted | Enable/Disable during Close and Install/Start after Close return errors | Treat a closed kernel as terminal |
| Disposed owners | Could emit events, run pipelines or retain pipeline commits | Emit does nothing; Run returns an error and no commit; pending commits expire with their caller | Send required notifications through a live owner before disposal; check Run errors |
| Stale registrations | An old context could leak a service or fail a newer Setup | Disposed contexts are rejected without poisoning a newer activation | Register through the current activation's context |
| Root disposal | Could leave a dead plugin marked active and its consumers running | Setup root disposal fails Setup; active root disposal fails the activation and blocks dependents, with no automatic restart | Dispose child scopes for local work; use Kernel.Enable/Disable for plugin management |
| Immediate point callbacks | Could dispose an owner before its contribution cleanup was registered | Cleanup is installed before notification | No callback or scheduler change is required |

An event already being emitted retains its documented listener ordering and panic
isolation. Optional capability consumers remain active when a provider disappears;
call Use each time instead of retaining its interface. Required consumers still
dispose before the provider and get fresh contexts when it returns. Use and Items
keep their read-only lookup behavior during cleanup.

## Choosing restart policy

For a shell plugin, leave Manifest.Restart nil to keep automatic restart disabled.
An eligible presentation plugin can declare:

```go
return kernel.Manifest{
	ID: "viewer",
	Restart: &kernel.Restart{Max: 3, Within: time.Minute},
}
```

For hosts using kernel.New directly, Options.Restart remains an explicit default:

| Manifest.Restart | Effective policy |
|---|---|
| nil | Inherit Options.Restart, whose zero value disables restart |
| &kernel.Restart{} | Disable restart even if the host configured a default |
| Positive Max and Within | Override the host default with this plugin's budget |
| Nonpositive Max or Within | Disable restart |

The policy is copied at Install; changing the original value or a status snapshot
does not change it. Each plugin counts its own restarts. The kernel recreates the
plugin and repeats Setup; it never replays the failed callback or user command.

Before opting in, review resource acquisition, external mutations and cleanup.
Cancelled workers from an earlier activation may still be finishing when the next
activation starts. Required dependents are also set up again when their provider
returns. Neither fresh plugin instances nor cancellation makes external side
effects idempotent. Construction and Setup failures are not automatically retried.

## Shutdown and cleanup

Keep OnDispose quick and on the UI thread. For blocking, worker-safe resources,
use the existing tracked Spawn pattern in [Blocking cleanup](kernel-api.md#blocking-cleanup).
Do not move UI handles or mutable plugin state into worker cleanup. Workers cannot
wait for a posted UI closure during Close, and posts from disposed scopes are dropped.

The shell supplies a three-second close context. It bounds the wait for workers
and is passed to the final synchronous Store.Save, which must honor it. It cannot
interrupt an OnDispose callback, terminate a worker, or enforce a deadline on a
store that ignores cancellation. Worker cleanup may overlap dependent-first UI
disposal and other cleanup workers.

Handle the first Close error. Repeated calls do nothing and return nil; they do
not retry a failed final save or wait again for unfinished cleanup. Disposer errors
can be inspected with errors.As for *kernel.PluginError and *kernel.PanicError;
errors.Is still finds joined context or store errors. Host callback panics and
Strict-mode panic propagation remain unchanged.

## Upgrade and verify

Upgrade in the app's own repository:

```sh
go get github.com/xleotranx/wada-kit@v0.2.0
GOWORK=off make check
```

Review the restart and shutdown changes above and run the app's native smoke
checks with a temporary data directory. Keep app-specific policy and contracts in
the app. Do not commit a local replacement or rely on a development workspace for
release validation. See [Maintenance and releases](maintenance.md) for the kit's
release checks.
