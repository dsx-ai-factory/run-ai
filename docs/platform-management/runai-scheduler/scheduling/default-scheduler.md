# Setting the Default Scheduler

By default, Kubernetes uses its own native scheduler to determine pod placement. The NVIDIA Run:ai platform provides a custom scheduler, `runai-scheduler`, which is used by default for workloads submitted using the [NVIDIA Run:ai](../../../workloads-in-nvidia-run-ai/introduction-to-workloads.md) platform.

This guide outlines how to configure workloads submitted directly to Kubernetes or through external frameworks to run with the [NVIDIA Run:ai Scheduler](concepts-and-principles.md), instead of the default Kubernetes scheduler.

## Enforce the Scheduler at the Namespace Level

When submitting workloads in a given namespace (i.e., NVIDIA Run:ai [project](../../../platform-management/aiinitiatives/organization/projects.md)), the parameter `enforceRunaiScheduler` is enabled (true) by default. This ensures that any workload associated with a NVIDIA Run:ai project automatically uses the `runai-scheduler`, including workloads submitted directly to Kubernetes or through external frameworks.

If this parameter is disabled, `enforceRunaiScheduler=false`, workloads will no longer default to the NVIDIA Run:ai Scheduler. In this case, you can still use the NVIDIA Run:ai Scheduler by specifying it manually in the workload YAML.

## Specify the Scheduler in the Workload YAML

To use the NVIDIA Run:ai Scheduler, specify it in the workload’s YAML file. This instructs Kubernetes to schedule the workload using the NVIDIA Run:ai Scheduler instead of the default one.

```yaml
spec:
  schedulerName: runai-scheduler
```

**For example:**

```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    user: test
    gpu-fraction: "0.5"
    gpu-fraction-num-devices: "2"
  labels:
    runai/queue: test
  name: multi-fractional-pod-job
  namespace: test
spec:
  containers:
  - image: gcr.io/run-ai-demo/quickstart-cuda
    imagePullPolicy: Always
    name: job
    env:
    - name: RUNAI_VERBOSE
      value: "1"
    resources:
      limits:
        cpu: 200m
        memory: 200Mi
      requests:
        cpu: 100m
        memory: 100Mi
    securityContext:
      capabilities:
        drop: ["ALL"]
  schedulerName: runai-scheduler
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 5
```
