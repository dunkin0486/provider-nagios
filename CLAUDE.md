# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## What this is

`provider-nagios` is a [Crossplane](https://crossplane.io/) provider for managing
[Nagios XI](https://www.nagios.com/products/nagios-xi/) monitoring configuration as
Kubernetes custom resources, built on `crossplane-runtime/v2`. It talks to the same
underlying Nagios XI REST API (`/api/v1/config/*`, `/api/v1/system/*`) as
[`dunkin0486/terraform-provider-nagios`](https://github.com/dunkin0486/terraform-provider-nagios);
`internal/client` is a direct port of that provider's client package — same API, same
quirks, different consumer (a Crossplane `managed.ExternalClient` instead of
`terraform-plugin-framework`). Only `Host` is implemented so far (early development).

## Commands

```bash
make submodules   # one-time: init the "build" submodule (CI/CD scripts), required first
make reviewable   # generate + lint + test — run before every commit/PR
make build        # build the provider binary and package
go test ./...     # unit tests only, no live Nagios XI needed
go test ./internal/client/... -run TestName -v   # single test
```

To add a new managed-resource type:

```bash
export group=monitoring   # lower case, e.g. monitoring
export type=Service       # Camel case, e.g. Service, Contact, Command
make provider.addtype provider=Nagios group=${group} kind=${type}
```

Then register the scheme in `apis/nagios.go` and the controller in
`internal/controller/nagios.go`; port the matching object file from the Terraform
provider's `internal/client` (see below); wire it into the new type's controller
following `internal/controller/host` as the reference implementation.

CI (`.github/workflows/ci.yml`) runs `golangci-lint`, a codegen-diff check, and
`make test` on every push/PR; images build/publish only on tagged releases. There is
no live-Nagios-XI integration suite in CI — `make test-integration` is a manual target
that spins up a `kind` cluster, but nothing here talks to a real Nagios XI instance.

## Architecture

Three packages, deliberately separated:

- **`apis/`** — CRD Go types (e.g. `apis/monitoring/v1alpha1/host_types.go`) plus
  generated deepcopy/managed-resource boilerplate (`zz_generated.*.go`). Never
  hand-edit generated files; edit the source type and run `make reviewable` (or
  `make generate`) to regenerate.
- **`internal/client`** — a plain Go HTTP client for Nagios XI's REST API, with zero
  dependency on crossplane-runtime or Kubernetes, so it's unit-testable standalone.
  One file per object type (`host.go`, and future `service.go`, `contact.go`, ...),
  each exporting `NewX`/`GetX`/`UpdateX`/`DeleteX` on `*Client`.
- **`internal/controller/<kind>`** — one package per managed-resource type: a
  `connector`/`external` split (`connector.Connect` builds an `*nagiosclient.Client`
  from the `ProviderConfig`'s credentials; `external` implements
  `Observe`/`Create`/`Update`/`Delete`), plus `convert.go` with
  `<kind>FromParameters`/`observationFrom<Kind>` mapping functions.
  `internal/controller/host/{host,convert}.go` is the reference implementation.

Patterns worth reusing from `host`: `isUpToDate` compares the client struct built from
spec against the one read back from Nagios via `cmp.Equal`, with
`cmpopts.IgnoreFields` for provider-managed fields (e.g. `Register`) and
`cmpopts.SortSlices` for list fields Nagios doesn't order consistently — don't compare
API structs with `==`/`reflect.DeepEqual`. `boolToNagios`/`nagiosToBool` convert
`*bool` spec fields to/from Nagios's `"0"`/`"1"` convention, treating `nil` as "unset"
rather than `false` — never dereference an optional bool spec field directly. `Create`
wraps the post-write `GetX` in `nagiosclient.RetryUntilFound`; `Observe`/`Update`/
`Delete` never retry, so a genuinely deleted resource surfaces as not-found promptly.

## Real Nagios XI API quirks `internal/client` encodes

Load-bearing behaviors of the live API (see the referenced doc comments), not
stylistic choices — don't refactor them away without confirming the wire behavior:

1. **Every response is HTTP 200, even on failure.** Success/failure is only knowable
   by parsing the JSON body for a `success`/`error` key — `response.go`'s
   `parseCommandResponse` is the single choke point every mutating call routes through.
2. **Every write needs a follow-up `applyconfig` call.** `client.applyConfig`
   (`client.go`) posts to `system/applyconfig` after every `NewX`/`UpdateX`/`DeleteX` —
   without it the change lands in Nagios's DB but never takes effect.
3. **PUT (rename) addresses the *old* name** as a URL path segment, not the new one —
   `url.go`'s `buildURL` builds this per verb: GET/DELETE filter by a query param, PUT
   takes `oldVal` as a path segment plus `force=1` (and appends `/<description>` for
   `service`, addressed by `(name, description)`), POST needs `force=1` for every
   object type except `applyconfig` itself.
4. **`existsErrorFor` fallback**: if a PUT fails with `"Does the <type> exist?"`,
   `UpdateHost` (and future `UpdateX`) falls back to calling `NewX` fresh — handles a
   rename targeting a name renamed/deleted outside Crossplane. **`RetryUntilFound`**
   (`retry.go`) tolerates the eventual-consistency window right after a write; use it
   only from `Create`, never from `Observe`.
5. **`free_variables`** (custom `_`-prefixed macros) come back as dynamic top-level
   keys on the object itself, not nested under a `free_variables` key —
   `params.go`'s `extractFreeVariables` picks these out of the raw JSON by prefix,
   separately from the typed `json.Unmarshal`.
6. **`NotificationOptions` is a comma-joined string on `Host` but will be `[]string`
   on `Service`** — a real, intentionally preserved API asymmetry inherited from the
   Terraform provider, not a rewrite oversight. Don't "fix" one to match the other.

## Porting from the Terraform provider

New object types are ported, not written from scratch: take the matching file from
`dunkin0486/terraform-provider-nagios`'s `internal/client`, adapt it to this repo's
package (it should end up with zero `terraform-plugin-framework` dependency), and
preserve its quirks and asymmetries verbatim rather than normalizing them — that
provider's doc comments describe the same live-API behavior this repo depends on.
`GetX` must return `(nil, nil)` for zero results, never a non-nil empty struct —
`Observe` depends on that to report `ResourceExists: false`.
