# Maintenance and releases

## Ownership

The kit is maintained in `xleotranx/wada-kit`. Apps depend on tagged releases and
upgrade independently. There is one source of truth; app repositories contain no
vendored kit source or permanent local replacement.

The kernel stays independent of UI and app features. The desktop layer owns MyGo
hosting, generic controls, settings and diagnostics. Apps own domain contracts,
plugins, storage, backends, credentials, branding and packaging. A change needed by
one app enters the kit only when it is a general host capability.

One module versions the packages together. The desktop API exposes MyGo types;
compatibility includes the tested MyGo and Onevra versions in `go.mod`. Each app
must validate the versions selected by its own module graph. The current shell
supports one main window. Its clear-window palette state is process-wide.

## Compatibility

During v0, patch releases preserve the public API; breaking changes use a minor
version bump, explain consumer changes in the changelog, and are tested against
known consumers. Release tags are immutable.

v0.2 changes lifecycle behavior; [Upgrading to v0.2](upgrading-v0.2.md) describes
the API addition and consumer changes. Test known consumers against each release
candidate independently of the kit's standalone gate.

A v1 release requires a reviewed public API and evidence from Wada and another
independent app. After v1, fixes use patch versions, compatible additions use
minor versions, and breaking changes require a new major version and the Go module
path suffix. Adding methods to public interfaces can break consumer fakes and
implementations; it is not automatically a compatible addition.

Compatibility covers lifecycle/disposal, event and contribution ordering, task
cancellation and stale-result handling, settings keys/schema, and source API.
The SQLite settings schema is 1; unsupported schemas are refused. Extraction
changes no persisted format. Schema changes need an explicit compatibility policy
and tests before release.

Maintain one current release line. Apps may remain on older tags; backport fixes
only when a consumer requires it. Do not move an existing tag to repair a release.

## Checks

`GOWORK=off make check` runs the kit gate independently of any sibling app.
The gate includes `scripts/check-consumer.sh`: it copies Notes into a temporary
external module and generates an independent starter, then checks both against
the candidate kit through temporary replacements. The starter runs its own
lifetime vet, import lint, race/no-cgo tests and build. Generation and its
safety/identity tests use Bash and standard Unix tools. This checks exported APIs
and the documented new-app workflow.

For a published release, `scripts/check-consumer.sh v0.2.0` instead downloads and
tests that tag with no replacement. The public module needs no GitHub credentials
or `GOPRIVATE` configuration.

## Consumers

The kit is public. Apps fetch tagged versions through normal Go module
downloads; no cross-repository token, deploy key, SSH URL rewrite, or `GOPRIVATE`
setting is required. Each app owns its checks and pinned kit version. Keep
`GOWORK=off` in release checks so a local checkout cannot hide an unpublished
change.

## Release procedure

1. Update `CHANGELOG.md`, the API notes, and consumer upgrade instructions.
2. Run `GOWORK=off make check`, `go build ./...`, and
   `scripts/check-consumer.sh` on the release commit. Review the public API diff
   against the preceding tag, including behavioral and persistence changes.
3. Test known consumers against the candidate kit in a temporary workspace or
   checkout. Wada's `make check` covers its plugin integration and native views.
4. Commit the release changes and push `main`. Verification runs locally; this
   repository has no automated CI workflow.
5. Create an annotated version tag and push that tag. Tags use `v0.1.0` format at
   the repository root. Create a GitHub release with the matching changelog notes.
6. Run `scripts/check-consumer.sh <tag>` to verify public module download and
   external use without a local replacement.
7. Upgrade each app's `go.mod` and `go.sum` in its own change; run its full gate
   with `GOWORK=off` before merging. Release checks must use the tagged module too.
