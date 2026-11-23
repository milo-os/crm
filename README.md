# Milo CRM

Manage customer relationships and help customers get started with Milo.

## Product enhancements

- [From sales opportunity to organization onboarding](docs/enhancements/opportunity-led-onboarding.md)

## Technical design

- [Native Milo API and onboarding system design](docs/design/opportunity-led-onboarding.md)

## Development

CRM is a native Milo service built from [controller-template](https://github.com/milo-os/controller-template). It serves the `crm.miloapis.com/v1alpha1` API group as CRDs, reconciled by a single controller manager.

| Tool | Version |
| --- | --- |
| Go | 1.27+ |
| [Task](https://taskfile.dev/installation) | v3+ |
| Docker and kind | recent (only for `task dev:*`) |

The Taskfile includes a remote Taskfile from `datum-cloud/test-infra`, so set `TASK_X_REMOTE_TASKFILES=1`.

```bash
task generate    # deepcopy and defaults code generation
task manifests   # CRDs, RBAC, and the CRD kustomization
task build       # build bin/crm
task test        # run Go tests
task lint        # run golangci-lint
task dev:setup   # kind cluster plus a dev deployment
```

### End-to-end tests

Chainsaw tests live in `test/e2e/<suite>/` and run against the test-infra kind cluster.

```bash
task dev:setup                          # create the cluster and deploy CRM (once)
task dev:deploy                         # reapply manifests after config changes
task dev:redeploy                       # rebuild and restart after code changes
task test:e2e                           # run every suite
task test:e2e -- exampleresource        # run only test/e2e/exampleresource
task test-infra:cluster-down            # delete the cluster
```

The E2E workflow runs the same steps in CI and can be started manually with a chosen suite.

### Activity policies

`ActivityPolicy` resources live in `config/components/activity-policies` and are applied to the Milo control plane on the deploy side. Their unit tests live in `test/activity-policies` and run offline in the activitytool image (Docker required):

```bash
task test:activity-policies
```

The Activity Policy Tests workflow runs the same check when policies or tests change.

### Adding a resource

1. Add `api/crm/v1alpha1/<kind>_types.go` with the type, its list type, kubebuilder markers, and an `init()` that registers both with `SchemeBuilder`.
2. Mark it `+kubebuilder:resource:scope=Cluster`; CRM is a platform-level, staff-only tool, so every CRM resource is cluster-scoped (see [Resource scoping](docs/design/opportunity-led-onboarding.md#resource-scoping)).
3. Run `task generate && task manifests`.

The CRD is written to `config/base/crd/bases`, listed in the CRD kustomization, and registered with the manager's scheme automatically.

The template's example `Resource` kind lives in its own `example.miloapis.com` group under `api/example/v1alpha1`, with its controller, webhooks, and sample. It is kept as a reference for new resources. It is not part of the CRM product.
