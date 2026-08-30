# E2E Tests

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `KUBECONFIG` | Yes | Path to kubeconfig file |
| `RUNAI_SECRETS_DIR` | Yes | Directory containing secret files (see below) |
| `E2E_TEST_DOMAIN` | Yes | Test domain (default: `test.runai.local`, change to your env-in-a-click domain) |
| `E2E_TEST_WILDCARD_DOMAIN` | Yes | Wildcard domain (default: `*.inference.test.runai.local`, change to `*.inference.<Your env-in-a-click domain>`) |
| `E2E_K8S_DISTRIBUTION` | No | Set to `openshift` for OpenShift clusters |
| `E2E_COMPOSE_PROVISIONER_IMAGE` | No | Custom provisioner image for docker-compose tests (e.g. `nvcr.io/nvidia/runai/runai-provisioner:v1.2.3`) |

## Required Secret Files

The `RUNAI_SECRETS_DIR` must be a real dir in your machine, which must contain:
- `tls.crt`
- `tls.key`
- `wildcard.crt`
- `wildcard.key`
- `runai_jfrog_token`

## Running locally

```bash
export KUBECONFIG=~/.kube/config
export RUNAI_SECRETS_DIR=/path/to/secrets
export E2E_TEST_DOMAIN=myenv.runailabs.com
export E2E_TEST_WILDCARD_DOMAIN='*.inference.myenv.runailabs.com'

cd provisioner-service
go test -v -timeout 3h ./e2e
```

## Docker Compose Tests

The e2e suite includes tests that run the provisioner via docker-compose (simulating the poller workflow).

By default, these tests use the production image (`nvcr.io/nvidia/runai/runai-provisioner:latest`).
To test with a specific version (e.g. a branch build), set:

```bash
export E2E_COMPOSE_PROVISIONER_IMAGE=nvcr.io/nvidia/runai/runai-provisioner:your-version
```

In CI, this is automatically set to the branch-specific image built by the pipeline.
