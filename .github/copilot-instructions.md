# provider-nagios

A [Crossplane](https://crossplane.io/) provider for managing [Nagios XI](https://www.nagios.com/products/nagios-xi/) monitoring configuration as Kubernetes custom resources, built on `crossplane-runtime/v2`. It talks to the same underlying Nagios XI REST API (`/api/v1/config/*`, `/api/v1/system/*`) as [`dunkin0486/terraform-provider-nagios`](https://github.com/dunkin0486/terraform-provider-nagios); `internal/client` is a direct port of that provider's client package — same API, same quirks, different consumer (a Crossplane `managed.ExternalClient` instead of `terraform-plugin-framework`). `Host` and `Service` are implemented; other object types follow the same pattern.

## Commands

```bash
make submodules   # one-time: init the "build" submodule (CI/CD scripts)
make reviewable   # generate + lint + test — run before every commit/PR
make build        # build the provider binary and package
go test ./...     # unit tests only, no live Nagios XI needed
```

To add a new managed-resource type:

```bash
make provider.addtype provider=Nagios group=monitoring kind=<Kind>
# then register scheme in apis/nagios.go and controller in internal/controller/nagios.go
```

## Architecture

Three packages, deliberately separated:

- **`apis/`** — CRD Go types (e.g. `apis/monitoring/v1alpha1/host_types.go`) plus generated deepcopy/managed-resource boilerplate (`zz_generated.*.go`). Never hand-edit generated files; edit the source type and run `make reviewable` to regenerate.
- **`internal/client`** — a plain Go HTTP client for Nagios XI's REST API, with zero dependency on crossplane-runtime or Kubernetes, so it is unit-testable standalone. One file per object type (`host.go`, `service.go`, ...), each exporting `NewX`/`GetX`/`UpdateX`/`DeleteX` on `*Client`.
- **`internal/controller/<kind>`** — one package per managed-resource type: a `connector`/`external` split (`connector.Connect` builds a `*nagiosclient.Client` from the `ProviderConfig`'s credentials; `external` implements `Observe`/`Create`/`Update`/`Delete`), plus `convert.go` with `<kind>FromParameters`/`observationFrom<Kind>` mapping functions. `internal/controller/host` is the reference implementation.

### Patterns to reuse from `host`

- `isUpToDate` compares the client struct built from spec against the one read back from Nagios via `cmp.Equal`, with `cmpopts.IgnoreFields` for provider-managed fields (e.g. `Register`) and `cmpopts.SortSlices` for list fields Nagios doesn't order consistently — never use `==`/`reflect.DeepEqual` on API structs.
- `boolToNagios`/`nagiosToBool` convert `*bool` spec fields to/from Nagios's `"0"`/`"1"` convention, treating `nil` as "unset" rather than `false` — never dereference an optional bool spec field directly.
- `Create` wraps the post-write `GetX` in `nagiosclient.RetryUntilFound`; `Observe`/`Update`/`Delete` never retry.

## Nagios XI API quirks encoded in `internal/client`

These are load-bearing behaviors of the live API, not stylistic choices — do not refactor them away without confirming wire behavior against a live instance:

1. **Every response is HTTP 200, even on failure.** Success/failure is determined by parsing the JSON body for a `success`/`error` key — `response.go`'s `parseCommandResponse` is the single choke point every mutating call routes through.
2. **Every write needs a follow-up `applyconfig` call.** `client.applyConfig` posts to `system/applyconfig` after every `NewX`/`UpdateX`/`DeleteX` — without it the change lands in Nagios's DB but never takes effect.
3. **PUT (rename/update) addresses the *old* name** as a URL path segment, not the new one. `url.go`'s `buildURL` builds this per verb: GET/DELETE filter by query param, PUT takes `oldVal` as a path segment plus `force=1`. For `service`, PUT is compound-keyed by `(config_name, description)`, appending `/<description>` to the path.
4. **`existsErrorFor` fallback:** if a PUT fails with `"Does the <type> exist?"`, `UpdateX` falls back to `NewX` — handles a rename targeting a name deleted outside Crossplane.
5. **`free_variables`** (custom `_`-prefixed macros) come back as dynamic top-level keys on the object itself, not nested under a `free_variables` key — `params.go`'s `extractFreeVariables` picks these out of the raw JSON by prefix, separately from the typed `json.Unmarshal`.
6. **`service` is keyed differently per verb:** `GetService` keys off `config_name` alone; `UpdateService` is compound-keyed off `(config_name, description)`; `DeleteService` keys off `(host_name, description)` where `host_name` is the full host set comma-joined into one value. These are not normalized — it is what the live API requires.
7. **`NotificationOptions` is `[]string` on `Service` but a comma-joined `string` on `Host`** — a real, intentionally preserved API asymmetry. Do not "fix" one to match the other.

## Porting new object types

New object types are ported from `dunkin0486/terraform-provider-nagios`'s `internal/client`, not written from scratch. Adapt the matching file to this repo's package (zero `terraform-plugin-framework` dependency), and preserve its quirks verbatim rather than normalizing them. `GetX` must return `(nil, nil)` for zero results, never a non-nil empty struct — `Observe` depends on that to report `ResourceExists: false`.
