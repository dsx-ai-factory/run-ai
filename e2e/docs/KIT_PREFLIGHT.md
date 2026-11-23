# Kit prerequisite preflight

Before a kit runs a single test, it reads the versions of the prerequisites
deployed on the target cluster and compares them with the versions NVIDIA Run:ai
documents as supported for the cluster release. Both the workloads kit and the
certification kit run this gate.

A mismatch stops the run with an explicit expected-versus-actual line, so a kit
report never has to be read to discover that the cluster was never supported in
the first place.

The certification kit runs the same gate over the same matrix. A cluster
presented for certification is expected to meet Run:ai's published
prerequisites, so the check is part of what that kit certifies — not a
precondition for its tests. It applies regardless of which components the
kit's own tests happen to exercise.

Scope is deliberately limited to prerequisites NVIDIA Run:ai publishes a
commitment for. Validating the wider platform stack against a vendor SBOM is
cluster-build validation and belongs with whoever builds the cluster, not in a
Run:ai test kit.

Code lives in `e2e/src/modules/preflight`, in four files: `supportMatrix.ts`
holds the published versions and how to find each one, `clusterProbe.ts` reads
them off the cluster, `preflight.ts` compares and renders, and `cli.ts` is the
entry point. The gate is invoked by `run_prerequisite_preflight` in
`e2e/kit-common/kit-utils.sh`.

## What is checked

The ranges come from the NVIDIA Run:ai
[support matrix](https://run-ai-docs.nvidia.com/self-hosted/getting-started/installation/support-matrix)
and [cluster system requirements](https://run-ai-docs.nvidia.com/self-hosted/getting-started/installation/install-using-helm/system-requirements),
and are transcribed in `supportMatrix.ts`.

| Component | v2.26 supported | Required? |
| --- | --- | --- |
| Kubernetes | 1.34 - 1.36 | yes (skipped on OpenShift) |
| Red Hat OpenShift | 4.18 - 4.22 | yes (skipped on Kubernetes) |
| NVIDIA GPU Operator | 25.10 - 26.3 | yes |
| Kube-Prometheus Stack | 76.0 or above | yes (skipped on OpenShift, which ships its own monitoring) |
| Prometheus | 3.5 or above | yes |
| NVIDIA Network Operator | 25.10 - 26.7 | only with RDMA or Multi-Node NVLink nodes |
| NVIDIA DRA driver | 25.8 - 25.12 | only with Multi-Node NVLink nodes |
| Kubeflow Training Operator | 1.9.2 | only for distributed training workloads |
| MPI Operator | 0.6.0 or above | only for MPI v2 workloads |
| Knative Serving | 1.19 - 1.21 | only for inference workloads |
| Leader Worker Set (LWS) | 0.7.0 or above | only for distributed inference workloads |

Only the required components can fail the gate — when they are missing, or
deployed at a version outside the supported range. A conditional component never
fails it: the cluster may legitimately not need it, and when it is installed at
an unsupported version the only tests at risk are the ones that depend on it. It
is reported as a warning either way, naming the workload type it is needed for,
so the reader can decide whether that matters to them.

Versions are read from the Helm release metadata first, then from the
`app.kubernetes.io/version` label, then from the container image tag, so the gate
works whether or not the prerequisite was installed with Helm.

## Configuration

| Variable | Default | Meaning |
| --- | --- | --- |
| `PREFLIGHT_MODE` | `enforce` | `enforce` stops the run on any failure, `warn` reports and continues, `off` skips the gate |
| `PREFLIGHT_RUNAI_VERSION` | `2.26` | Which support matrix to check against |
| `PREFLIGHT_RESULTS_DIR` | `/app/e2e/results` | Where `preflight-report.json` and `preflight-report.md` are written |
| `PREFLIGHT_DEBUG` | unset | Also print the probe trace to the run log; `1`, `true`, `yes` or `on` |
| `KUBERNETES_PLATFORM` | auto-detected | `k8s` or `ocp`, overrides detection |

## Diagnosing a component reported as not detected

Every lookup a probe makes is recorded, whether or not it found anything, and
written to `preflight-report.json` and to the **Probe trace** section of
`preflight-report.md`. `PREFLIGHT_DEBUG=1` prints the same trace to the run log:

```
Prometheus
  prometheuses.monitoring.coreos.com across all namespaces: no version in 1 object(s)
  statefulset matching app.kubernetes.io/name=prometheus across all namespaces: no version in 0 object(s)
  statefulset matching operator.prometheus.io/mode=server across all namespaces: read v3.0.1
```

This tells apart the two outcomes that otherwise look identical: a component
that is genuinely absent, and one whose objects were found but carried no
readable version. The second case is reported as
`deployed, no version readable` rather than `not detected`, so nobody goes
looking for an install that is already there.

Every install shape the kit meets should end up in
`e2e/src/test/unit/fixtures/preflightClusterShapes.ts`, with the version each
component is expected to resolve to. Capturing a surprising cluster there turns
it into a permanent regression test instead of knowledge that lives in one
person's head.

## Running it on its own

```bash
cd e2e
npm run kit-preflight   # against the current kubeconfig context
npm run test:unit       # offline unit tests, no cluster needed
```

The unit tests live in `e2e/src/test/unit` and run on every pull request as part
of the `ESLinter` job.

## Updating for a new cluster release

Add an entry to `supportMatrices` in `supportMatrix.ts` keyed by the new minor
version, move `defaultRunaiVersion`, and update the table above. The numbers must
be transcribed from the published documentation, which is the commitment the gate
enforces; nothing else in the preflight needs to change.
