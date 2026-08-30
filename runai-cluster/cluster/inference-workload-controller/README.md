# inference-workload-controller

# Pod Spec Conventions

## hostNetwork and dnsPolicy

When `hostNetwork: true` is set on a workload, the pod spec builder must also set
`dnsPolicy: ClusterFirstWithHostNet`. Without it, Kubernetes falls back to using the
node's DNS resolver instead of kube-dns, and Kubernetes service names won't resolve.

This applies to all builders that set `hostNetwork` on a pod spec:
- `internal/builders/utils.go` (Knative inference)
- `internal/builders/lws_build.go` (LeaderWorkerSet distributed inference)

Reference: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy
