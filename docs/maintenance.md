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
CI also runs `scripts/check-consumer.sh`, which copies the public Notes example
into a temporary external module and tests it against the candidate kit through
a temporary replacement. This checks external imports and exported APIs.

For a published release, `scripts/check-consumer.sh v0.1.0` instead downloads and
tests that tag with no replacement. Private repository authentication and
`GOPRIVATE` must be configured as described in the README.

## Private CI consumers

Each consuming repository needs read access to this private repository in its own
CI; its default `GITHUB_TOKEN` cannot read another private repository. Wada uses a
read-only deploy key registered here, with the private half in Wada's
`WADA_KIT_SSH_KEY` Actions secret. It configures `GOPRIVATE` and an SSH URL rewrite
for this repository only. Host keys come from GitHub's HTTPS metadata endpoint.
Never use a personal account's broad token as the consumer build credential.

A new app needs its own read-only deploy key and its own secret. To rotate one,
register a new public key here, replace that app's secret, verify its CI, then
revoke the old deploy key. Do not share one deploy key between consuming apps.

## Release procedure

1. Update `CHANGELOG.md`, the API notes, and consumer upgrade instructions.
2. Run `GOWORK=off make check`, `go build ./...`, and
   `scripts/check-consumer.sh` on the release commit. Review the public API diff
   against the preceding tag, including behavioral and persistence changes.
3. Test known consumers against the candidate kit in a temporary workspace or
   checkout. Wada's `make check` covers its plugin integration and native views.
4. Commit the release changes, push `main`, and wait for its CI to pass.
5. Create an annotated version tag and push that tag. Tags use `v0.1.0` format at
   the repository root. Create a GitHub release with the matching changelog notes.
6. Run `scripts/check-consumer.sh <tag>` to verify private module download and
   external use without a local replacement.
7. Upgrade each app's `go.mod` and `go.sum` in its own change; run its full gate
   with `GOWORK=off` before merging. CI must use the tagged module too.
