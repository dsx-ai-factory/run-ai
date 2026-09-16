# External Access to Containers

Researchers may need to access containers remotely during workload execution. Common use cases include:

* Running a Jupyter Notebook inside the container
* Connecting PyCharm for remote Python development
* Viewing machine learning visualizations using TensorBoard

To enable this access, you must expose the relevant container ports.

## Exposing Container Ports

Accessing the containers remotely requires exposing container ports. In Docker, ports are exposed by [declaring](https://docs.docker.com/reference/cli/docker/container/run/) them when launching the container. NVIDIA Run:ai provides similar functionality within a Kubernetes environment.

Since Kubernetes abstracts the container's physical location, exposing ports is more complex. Kubernetes supports multiple methods for exposing container ports. For more details, refer to the [Kubernetes services and networking documentation](https://kubernetes.io/docs/concepts/services-networking/service).

| Method | Description | NVIDIA Run:ai Support |
| --- | --- | --- |
| Port Forwarding | Simple port forwarding allows access to the container via local and/or remote port. | Supported natively via Kubernetes |
| NodePort | Exposes the service on each Node’s IP at a static port (the NodePort). You’ll be able to contact the NodePort service from outside the cluster by requesting `<NODE-IP>:<NODE-PORT>` regardless of which node the container actually resides in. | Supported |
| LoadBalancer | Exposes the service externally using a cloud provider’s load balancer. | Supported |

## Access to the Running Workload's Container

Many tools used by researchers, such as Jupyter, TensorBoard, or VSCode, require remote access to the running workload's container. In NVIDIA Run:ai, this access is provided through dynamically generated URLs. NVIDIA Run:ai supports two routing models for exposing workloads:

* Host-based routing (default)
* Path-based routing&#x20;

### Host-Based Routing

By default, NVIDIA Run:ai uses host-based routing to expose workload URLs using subdomains. This allows all workloads to run at the root path, avoiding file path issues and ensuring proper application behavior.

```bash
https://project-name-workload-name.<CLUSTER_URL>/
```

Host-based routing is enabled by default via the `subdomainSupport: true` cluster configuration. For setup instructions, see [Host-Based Routing (Ingress Resource)](../../installation/install-using-helm/preparations.md#host-based-routing-ingress-resource).

### Path-Based Routing

Path-based routing uses the [Cluster URL](../../system-requirements/networking-setup.md#fqdn-and-dns) provided to dynamically create SSL-secured URLs in the following format:

```bash
https://<CLUSTER_URL>/project-name/workload-name
```

To use path-based routing instead, disable host-based routing by setting `subdomainSupport: false` in [Advanced cluster configurations](../cluster-config.md).&#x20;

While path-based routing works with applications such as Jupyter Notebooks, it may not be compatible with other applications. Some applications assume they are running at the root file system, so hardcoded file paths and settings within the container may become invalid when running at a path other than the root. For example, if an application expects to access `/etc/config.json` but is served at `/project-name/workspace-name`, the file will not be found. This can cause the container to fail or not function as intended.

<Note>

**Note**

For existing clusters, changing the routing model affects how workload URLs are generated. Administrators should plan this transition accordingly.

</Note>

### OpenShift Route TLS Termination

On OpenShift, NVIDIA Run:ai exposes a workload's URL through a `Route` object. By default, the `Route` uses edge termination: TLS terminates at the OpenShift Router, and traffic to the pod inside the cluster is plain HTTP.

To terminate TLS inside the pod itself and preserve end-to-end encryption from the client to the workload, set the Route termination policy to `passthrough`:

<Tabs>

<Tab title="Helm">

```yaml
clusterConfig:
  anyworkload-controller:
    openshiftRouteConfiguration:
      terminationPolicy: passthrough # edge (default) | passthrough
```

</Tab>

<Tab title="runaiconfig">

```yaml
spec:
  anyworkload-controller:
    openshiftRouteConfiguration:
      terminationPolicy: passthrough # edge (default) | passthrough
```

</Tab>

</Tabs>

This setting applies cluster-wide to every `Route` object NVIDIA Run:ai creates, across all workload types.

<Warning>

**Prerequisite**

With `passthrough` termination, NVIDIA Run:ai does not provision or manage a TLS certificate for the pod. You must terminate TLS inside the container yourself, for example, using your own certificate mounted as a Kubernetes secret and a web server or sidecar that you own and maintain.

</Warning>

<Warning>

**Impact on existing workloads**

This setting applies to every `Route` NVIDIA Run:ai creates, including Routes for workloads that are already running. Switching to `passthrough` on a cluster where existing workloads serve plain HTTP makes those workloads unreachable until each one terminates TLS inside the pod. Plan this as a cluster-wide change rather than a change scoped to the single workload that requires end-to-end encryption.

</Warning>
