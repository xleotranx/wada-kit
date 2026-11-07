# Two-plugin extension example

A headless notebook demonstrates services, topics, pipelines and points without
window setup. It uses only the standard library and the kit's kernel. Titles live
in the notebook plugin's memory; a new activation starts empty. It writes no files
and uses no settings or background workers.

From the kit checkout:

```sh
GOWORK=off go run ./examples/extensions
GOWORK=off go test ./examples/extensions/...
```

Output:

```text
Add welcome note: Hello plugins (committed=1, observed=1)
List notes: Hello plugins
```

| File | Role |
| --- | --- |
| [contract/notebook](internal/contract/notebook/notebook.go) | App-owned interfaces, payloads and declarations; no implementation |
| [plugins/notebook](internal/plugins/notebook/plugin.go) | Provides the notebook service, prepares titles and contributes List notes |
| [plugins/workspace](internal/plugins/workspace/plugin.go) | Requires the service, adds a length policy, observes additions and contributes Add welcome note |
| [main.go](main.go) | Composes both plugins and invokes ordered actions through `kernel.Invoke` |
| [main_test.go](main_test.go) | Exercises both plugins together, including rejection and disposal/recovery |

The workspace is installed first, but its required service makes it activate
after the notebook. Its action calls `Notebook.Add`, which runs both plugins'
preparation stages, appends the title, commits the workspace counter and emits
`Added`. The workspace's listener increments its observation counter before the
action returns. The host then invokes List notes.

The immediate `Post` callback in main is specific to this synchronous example.
Desktop apps use `shell.Run` for UI scheduling and settings persistence; use
`kerneltest`'s queued scheduler for deterministic integration tests.

Read [the extension guide](../../docs/extensions.md) for how to choose each
mechanism and define its contract. Notes remains the smallest native app, and
`examples/starter` remains the source for generated independent apps.
