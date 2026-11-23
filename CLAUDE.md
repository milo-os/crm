# crm

Native Milo service for the CRM: Companies, People, Opportunities, and opportunity-led organization onboarding. Built from [controller-template](https://github.com/milo-os/controller-template), modeled after the `billing` service conventions.

- Product scope: `docs/enhancements/opportunity-led-onboarding.md`
- Technical design: `docs/design/opportunity-led-onboarding.md`
- Delivery tracking: milo-os/crm#2 and its sub-issues

## Architecture

- **CRD-based**: kubebuilder v4 CRDs under `crm.miloapis.com/v1alpha1`, served by Milo's API server.
- **One controller manager**: `cmd/crm` runs the `operator` subcommand (controller-runtime). Reconcilers and admission webhooks are registered in `cmd/crm/cmd/operator.go`.
- **Milo control plane aware**: the `kubeconfigPath` operator config field points the manager at Milo's API server; empty means in-cluster config.
- **Scoping**: CRM is a platform-level, staff-only tool used from the staff portal. Every CRM resource is cluster-scoped, outside any organization, and its IAM protected resource has no parent organization. Never add a namespaced CRM resource.

## Repo layout

```
cmd/crm/                # binary entrypoint and cobra commands
api/crm/v1alpha1/       # CRM CRD type definitions
api/example/v1alpha1/   # template example resource (reference only)
internal/webhook/       # admission webhooks, per API group
internal/validation/    # admission validation rules, per API group
test/e2e/               # chainsaw e2e tests, one folder per suite
internal/config/        # operator configuration (CRMOperator)
config/base/            # CRDs and manager
config/components/      # namespace, RBAC, leader election, IAM, activity policies
config/overlays/dev/    # kind development overlay
hack/                   # boilerplate and generation helpers
docs/                   # enhancement and design
```

## Adding a resource

1. Add `api/crm/v1alpha1/<kind>_types.go` with the type and list type, `+kubebuilder:object:root=true`, `+kubebuilder:resource:scope=Cluster`, and an `init()` that registers both with `SchemeBuilder.Register(func(scheme *runtime.Scheme) error {...})` (see `api/crm/v1alpha1/company_types.go`).
2. Put field rules in the type with kubebuilder markers and CEL (`+kubebuilder:validation:XValidation` with a clear `message`). Webhooks are only for checks the schema can't express, such as references to other records or who is making the request.
3. Run `task generate && task manifests`. `hack/sync-crd-kustomization.sh` adds the CRD to `config/base/crd/kustomization.yaml`.
4. Test the schema rules with a chainsaw suite under `test/e2e/<kind>/`.
5. Add a sample to `config/samples/`, a `ProtectedResource` to `config/components/iam/protected-resources/`, its permissions to the `crm.miloapis.com-viewer` and `-admin` roles, and an activity policy (below).

## Admission webhooks

Follow the service-catalog layout. Webhooks live in `internal/webhook/<group>/v1alpha1` (package `v1alpha1`), expose `Setup<Kind>WebhookWithManager`, handle defaulting and logging, and delegate every rule to `internal/validation/<group>`. Validation functions are named `Validate<Kind>Create/Update/Delete`, return a `field.ErrorList`, and are covered by table-driven unit tests; the webhook turns the list into `apierrors.NewInvalid`. Every `+kubebuilder:webhook` marker must set `serviceName=crm-webhook`, the manager's webhook Service; controller-gen otherwise writes `webhook-service`, which does not exist.

## Activity policies

Each user-facing resource gets an `ActivityPolicy` in `config/components/activity-policies/<kind>-policy.yaml` (listed in that component's kustomization) and a `test/activity-policies/<kind>.test.yaml`. Follow the `test-policy` skill: test every rule, assert exact summaries and links, and cover no-match cases (system actors, reads, status, failed and dry-run requests). Never copy actual output into an expectation. `task test:activity-policies` runs with `--fail-on-untested --fail-on-untested-rules`. Keep `ACTIVITYTOOL_VERSION` in the Taskfile and `activitytool-version` in `.github/workflows/activity-policies.yaml` in sync.

## Example resource

The template's example `Resource` kind is kept as a working reference: in its own `example.miloapis.com` API group: `api/example/v1alpha1/resource_types.go`, its reconciler in `internal/controller/example/`, its defaulting and validating webhooks in `internal/webhook/example/v1alpha1/` with their rules in `internal/validation/example/`, and a sample in `config/samples/`. Follow it when adding CRM resources, controllers, and webhooks. It is not part of the CRM product.

## Reconcile errors

Return an error from `Reconcile` only when a retry can fix it.

- **A failure no retry can fix**: set a condition naming the case and return `reconcile.TerminalError(err)`.
- **A wait for something else**: set a condition, watch what you wait for, and return no error; if you cannot watch it, return `ctrl.Result{RequeueAfter: <interval>}`. Do not use `Result.Requeue`.
- **Anything else** (timeouts, conflicts, server errors): return the error.

## Verification

```bash
export TASK_X_REMOTE_TASKFILES=1
task generate && task manifests
task build && task test && task lint
kustomize build config/overlays/dev

# End to end, against the test-infra kind cluster (the user creates it with task dev:setup)
task dev:deploy
task test:e2e -- <suite>    # a folder under test/e2e; omit to run all
```

Add a chainsaw suite under `test/e2e/<suite>/` for each resource's controller and webhook behavior.

## Metrics and logs

Create every metric and logger through `go.miloapis.com/instrumentation`, naming labels with its labels constants. The instrumentation adoption test enforces this; see https://github.com/milo-os/instrumentation/blob/main/ADOPTING.md.
