# Add NVIDIA Run:ai Authorization to Kubeconfig

The `runai kubeconfig set` command allows users to configure their kubeconfig file with NVIDIA Run:ai authorization token. This setup enables users to gain access to the Kubernetes (K8s) cluster seamlessly.

<Note>

**Note**

Setting kubeconfig is not required in order to use the CLI. This command is used to enable third-party workloads under NVIDIA Run:ai authorization.

</Note>

## Usage

To set the token (will be fetched automatically) inside the kubeconfig file, run the following command:

```sh
runai kubeconfig set
```

## Prerequisites

Before executing the command, ensure that

1. [Cluster authentication](../../installation-and-configuration/configuration/cluster-authentication.md) is configured and enabled.
2. The user has a kubeconfig file configured.
3. The user is logged in (use the [runai login](runai/runai_login.md) command).

## User Kubeconfig Configuration

Add the following to the Kubernetes client configuration file (.`/kube/config`). For the full command reference, see [kubeconfig set](runai/runai_kubeconfig_set.md).

* Make sure to replace values with the actual cluster information and user credentials.
* There can be multiple contexts in the kubeconfig file. The command will configure the current context.

```yaml
apiVersion: v1
kind: Config
preferences:
  colors: true
current-context: <CONTEXT_NAME>
contexts:
- context:
    cluster: <CLUSTER_NAME>
    user: <USER_NAME>
  name: <CONTEXT_NAME>
clusters:
- cluster:
    server: <CLUSTER_URL>
    certificate-authority-data: <CLUSTER_CERT>
  name: <CLUSTER_NAME>
users:
- name: <USER_NAME>
```
