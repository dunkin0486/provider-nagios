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

## Acceptance Tests (Docker Compose)

This repository includes a Docker Compose acceptance harness that:

- Starts a local `k3s` API server.
- Starts Nagios XI using your private GHCR image.
- Runs the provider out-of-cluster against that API server.
- Applies provider CRDs plus a `ProviderConfig` and a sample `Host`.
- Waits for the `Host` to become `Ready` and validates observed state.

### Prerequisites

- Docker with Compose plugin (`docker compose`).
- Access to your private GHCR image (`docker login ghcr.io`).
- Optional: a Nagios XI API token. If omitted, the acceptance runner derives
   and enables the `nagiosadmin` API token from inside the running Nagios XI
   container.

### Run

Option A: export env vars directly.

```shell
export NAGIOS_XI_IMAGE=ghcr.io/<org>/<nagios-xi-image>:<tag>
# Optional. If not set, it is auto-derived from the running nagiosxi container.
export NAGIOS_API_TOKEN=<nagios-xi-api-token>
# Optional. Defaults to http://nagiosxi/nagiosxi inside the compose network.
export NAGIOS_URL=http://nagiosxi/nagiosxi
# Optional fast local mode: keep compose stack running between runs.
export ACCEPTANCE_PRESERVE_STACK=true

make acceptance.run
```

Option B: use an env file.

```shell
cp cluster/acceptance/.env.example cluster/acceptance/.env
# edit cluster/acceptance/.env with your image + token
make acceptance.run
```

By default, `cluster/acceptance/.env.example` uses the same private GHCR image
published by `terraform-provider-nagios` (`ghcr.io/dunkin0486/terraform-provider-nagios-test-nagiosxi:latest`).

You can tune the readiness timeout with `ACCEPTANCE_WAIT_TIMEOUT` (default: `900s`).

For faster local iteration after the first bootstrap:

- Set `ACCEPTANCE_PRESERVE_STACK=true` to leave the stack running after tests (fastest).
- Set `ACCEPTANCE_KEEP_VOLUMES=true` to tear down containers/network but keep volumes.

When using `ACCEPTANCE_PRESERVE_STACK=true`, stop and clean up manually when needed:

```shell
make acceptance.clean
```

Refer to Crossplane's [CONTRIBUTING.md] file for more information on how the
Crossplane community prefers to work. The [Provider Development][provider-dev]
guide may also be of use.

[CONTRIBUTING.md]: https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md
[provider-dev]: https://github.com/crossplane/crossplane/blob/master/contributing/guide-provider-development.md
