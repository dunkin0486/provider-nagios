# provider-nagios

`provider-nagios` is a [Crossplane](https://crossplane.io/) Provider for
managing [Nagios XI](https://www.nagios.com/products/nagios-xi/) monitoring
configuration as Kubernetes resources. It is based on
`crossplane/provider-template` and currently ships with:

- A `ProviderConfig`/`ClusterProviderConfig` type that points to a
  credentials `Secret`.
- A `monitoring.nagios.crossplane.io` API group with a `Host` managed
  resource type, backed by `internal/client` — a plain Go HTTP client for
  Nagios XI's REST API (`/api/v1/config/*`, `/api/v1/system/*`), ported from
  [`dunkin0486/terraform-provider-nagios`](https://github.com/dunkin0486/terraform-provider-nagios).

## Credentials

The `Secret` referenced by a `ProviderConfig`/`ClusterProviderConfig` must
contain JSON of the form:

```json
{"url": "http://nagios.example.com/nagiosxi", "token": "<api key>"}
```

The API token is found in the Nagios XI web UI under Admin > API Key. See
`examples/provider/config.yaml`.

## Status

Early development. Only `Host` is implemented so far. `internal/client`
carries real, load-bearing quirks of the Nagios XI API (every response is
HTTP 200 even on failure, every write needs a follow-up `applyconfig` call,
PUT addresses the *old* name, etc.) — see that package's doc comments before
touching it, and port additional object types (`service.go`, `contact.go`,
etc.) from the Terraform provider's `internal/client` the same way `host.go`
was ported, rather than reimplementing them from scratch.

## Developing

1. Run `make submodules` to initialize the "build" Make submodule used for CI/CD.
2. Add a new type:
   ```shell
   export group=monitoring # lower case, e.g. monitoring
   export type=Service     # Camel case, e.g. Service, Contact, Command
   make provider.addtype provider=Nagios group=${group} kind=${type}
   ```
3. Register the new type's scheme in `apis/nagios.go` and its controller in
   `internal/controller/nagios.go`.
4. Port the matching object file from `internal/client` in the Terraform
   provider (or write a new one following its pattern) and wire it into the
   new type's controller, following `internal/controller/host` as the
   reference implementation.
5. Run `make reviewable` to run code generation, linters, and tests.
6. Run `make build` to build the provider.

Refer to Crossplane's [CONTRIBUTING.md] file for more information on how the
Crossplane community prefers to work. The [Provider Development][provider-dev]
guide may also be of use.

[CONTRIBUTING.md]: https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md
[provider-dev]: https://github.com/crossplane/crossplane/blob/master/contributing/guide-provider-development.md
