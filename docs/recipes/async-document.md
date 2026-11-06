# Load a document safely

The executable recipe is `examples/starter/internal/plugins/documents/`, with
worker-safe storage in `examples/starter/internal/documents/`. Generate an app
with [the consumer guide](../build-an-app.md) to use it outside this repository.

## Ownership and flow

1. Setup contributes the page and command, creates a typed last-document setting,
   and subscribes to PageChanged through its plugin context.
2. A posted startup action opens the saved selection when the Documents page is
   active. The starter also supports v0.2.0's empty default-page ID; the updated
   shell reports the effective page directly.
3. Selecting a document disposes the previous document scope and cancels its Task.
   It creates a new child scope, clears the body/error and marks the view loading.
4. `kernel.Go` captures only a repository and document ID. The worker receives
   context.Context, performs I/O, and returns an owned string and error.
5. The apply callback updates UI state and saves the successful selection. The
   kernel calls it on the UI thread only while the scope is live and the task
   generation is current. A late result cannot replace a newer document.
6. Leaving the page disposes its child scope. Returning reloads the last successful
   selection. Disabling the plugin or closing the app cancels all child lifetimes.

Render only reads these fields and builds controls. OnClick handlers initiate
loads and retries. Rendering shows loading, error with Retry, empty, or body.
A repository can ignore cancellation and still have its stale result discarded;
it should honor cancellation so shutdown does not wait for unnecessary work.
Cancellation cannot forcibly interrupt arbitrary blocking I/O.

## Adapt the repository

`documents.Repository.Load(context.Context, string) (string, error)` is owned by
the app. The sample Files implementation accepts fs.FS; Samples supplies embedded
files. An app can supply os.DirFS for local documents or implement the same
interface with a context-aware backend. Validate document IDs at the domain
boundary, and copy mutable inputs before workers receive them. Do not pass a
plugin, frame context or mutable UI state into a repository.

The starter deliberately offers two known document IDs. For a real document list,
replace that selection policy, add a repository list operation and stable keys,
and keep errors and permissions explicit. Settings retain preferences; domain
files, credentials and database migrations stay in app-owned code.

For subscriptions or streams, use Scope.Spawn and post immutable updates through
the same child scope. Do not start raw goroutines or wait for a posted UI callback
during shutdown. OnDispose is synchronous UI cleanup and must return promptly.

## Verify the behavior

The plugin tests use kerneltest.RunWork to complete the newer load before an older
one, including an older repository result that ignores cancellation. They also
leave the page while loading, return to it, retry a failure, close/reopen with the
same settings store and disable with work pending. The harness queues posts and
workers and checks cleanup; an immediate `Post: func(fn func()) { fn() }` is
unsuitable when worker results must reach the UI thread.

Run `GOWORK=off go test -race ./examples/starter/...` in the kit or `make check`
in the generated app. Extend these tests when adding saves, streams or navigation.
