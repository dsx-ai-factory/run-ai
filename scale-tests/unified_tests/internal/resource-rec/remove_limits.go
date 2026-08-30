package resource_rec

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DefaultNamespaces is the default set of namespaces to strip resource limits from.
// Namespaces that don't exist in a given cluster are skipped without error, so this is a
// union across distributions: monitoring vs cattle-monitoring-system, training-operator vs
// kubeflow.
//
// Membership rule: a namespace belongs here if it holds a component the Run:ai operator
// manages, i.e. one with an operand under go-operator/internal/run/operands. That is what
// makes training-operator, mpi-operator and knative-serving in scope while gpu-operator and
// nginx-ingress are not — those are cluster prerequisites Run:ai does not manage. The
// gpu-operator namespace is doubly out on a scale-test cluster: it holds the *fake* GPU
// operator, whose consumption says nothing about the real one.
var DefaultNamespaces = []string{
	"runai",
	"runai-backend",
	"monitoring",
	"cattle-monitoring-system",
	"training-operator",
	"kubeflow",
	"mpi-operator",
	"knative-serving",
}

// highLimits is the sentinel value applied to every container: high enough that no
// pod will hit it in practice, but present as explicit keys so the Run:ai operator's
// SetDefaultsWhereNeeded skips filling in its own defaults (it only sets a key when absent).
var highLimits = corev1.ResourceList{
	corev1.ResourceCPU:    resource.MustParse("100"),
	corev1.ResourceMemory: resource.MustParse("1Ti"),
}

var runaiConfigGVK = schema.GroupVersionKind{Group: "run.ai", Version: "v1", Kind: "RunaiConfig"}

// engineConfigGVK is the GroupVersionKind of the engine operator Config CR.
// The go-operator writes to this CR but does not set engineAdmission.resources,
// so a patch to that field persists across reconciles.
var engineConfigGVK = schema.GroupVersionKind{Group: "engine.run.ai", Version: "v1alpha1", Kind: "Config"}

// runaiConfigServiceKeys lists all RunaiConfig spec keys that carry a flat resources.limits field.
// researcher-service is excluded: it uses a nested schema (resources.researcherService.limits)
// and is disabled by default in standard installs.
var runaiConfigServiceKeys = []string{
	"accessrule-controller",
	"admission-controller",
	"advice-restart-controller",
	"agent",
	"anyworkload-controller",
	"application-controller",
	"assets-sync",
	"binder",
	"cluster-api",
	"cluster-redis",
	"cluster-sync",
	"device-plugin",
	"external-workload-integrator",
	"inference-workload-controller",
	"init-ca",
	"job-controller",
	"metrics-exporter",
	"node-scale-adjuster",
	"nodepool-controller",
	"pod-group-assigner",
	"pod-group-controller",
	"pod-grouper",
	"project-controller",
	"queue-controller",
	"runai-container-toolkit",
	"runai-container-toolkit-reservation",
	"runai-node-exporter",
	"runai-scheduler",
	"shared-objects-controller",
	"status-updater",
	"workload-controller",
	"workload-exporter",
	"workload-overseer",
}

// RemoveResourceLimits sets sentinel CPU/memory limits on all containers in
// Deployments, StatefulSets, and DaemonSets across the given namespaces, patches
// the RunaiConfig and engine Config CRs so operators reconcile to no-limits, then
// waits up to waitTimeout for all pods to return to Running/Ready.
// Namespaces that don't exist in the cluster are skipped gracefully.
func RemoveResourceLimits(ctx context.Context, k8sClient client.Client, namespaces []string, waitTimeout time.Duration) error {
	runaiNamespaceFound := false

	for _, ns := range namespaces {
		nsObj := &corev1.Namespace{}
		if err := k8sClient.Get(ctx, client.ObjectKey{Name: ns}, nsObj); err != nil {
			if k8serrors.IsNotFound(err) {
				fmt.Printf("namespace %s not found, skipping\n", ns)
				continue
			}
			return fmt.Errorf("checking namespace %s: %w", ns, err)
		}

		if ns == "runai" {
			runaiNamespaceFound = true
		}

		if err := removeDeploymentLimits(ctx, k8sClient, ns); err != nil {
			return fmt.Errorf("namespace %s deployments: %w", ns, err)
		}
		if err := removeStatefulSetLimits(ctx, k8sClient, ns); err != nil {
			return fmt.Errorf("namespace %s statefulsets: %w", ns, err)
		}
		if err := removeDaemonSetLimits(ctx, k8sClient, ns); err != nil {
			return fmt.Errorf("namespace %s daemonsets: %w", ns, err)
		}
	}

	// Patch operator CRs after direct workload patches to prevent reconciliation back
	// to baked-in defaults. Done second to avoid write conflicts with the operators.
	if runaiNamespaceFound {
		if err := patchRunaiConfigLimits(ctx, k8sClient); err != nil {
			return fmt.Errorf("patching RunaiConfig: %w", err)
		}
		if err := patchEngineConfigLimits(ctx, k8sClient); err != nil {
			return fmt.Errorf("patching engine Config: %w", err)
		}
	}

	return waitForPodsReady(ctx, k8sClient, namespaces, waitTimeout)
}

// patchRunaiConfigLimits sets resources.limits to sentinel values for every service key
// in the RunaiConfig CR so the go-operator reconciles deployments to no-limits rather than
// restoring its baked-in defaults. Uses JSON merge patch so only targeted keys are touched.
func patchRunaiConfigLimits(ctx context.Context, k8sClient client.Client) error {
	sentinelLimits := map[string]interface{}{
		string(corev1.ResourceCPU):    highLimits.Cpu().String(),
		string(corev1.ResourceMemory): highLimits.Memory().String(),
	}
	spec := make(map[string]interface{}, len(runaiConfigServiceKeys))
	for _, key := range runaiConfigServiceKeys {
		spec[key] = map[string]interface{}{
			"resources": map[string]interface{}{
				"limits": sentinelLimits,
			},
		}
	}
	patchData, err := json.Marshal(map[string]interface{}{"spec": spec})
	if err != nil {
		return fmt.Errorf("marshaling RunaiConfig patch: %w", err)
	}

	rc := &unstructured.Unstructured{}
	rc.SetGroupVersionKind(runaiConfigGVK)
	rc.SetName("runai")
	rc.SetNamespace("runai")

	if err := k8sClient.Patch(ctx, rc, client.RawPatch(types.MergePatchType, patchData)); err != nil {
		if k8serrors.IsNotFound(err) {
			fmt.Println("RunaiConfig not found, skipping operator config patch")
			return nil
		}
		return err
	}

	fmt.Printf("patched RunaiConfig: set sentinel limits for %d service keys\n", len(runaiConfigServiceKeys))
	return nil
}

// patchEngineConfigLimits sets engineAdmission.resources.limits to sentinel values on the
// engine-config CR (engine.run.ai/v1alpha1). The go-operator does not write engineAdmission
// resources, so this patch persists across reconciles. The engine-operator then applies the
// updated limits to the engine-admission deployment.
func patchEngineConfigLimits(ctx context.Context, k8sClient client.Client) error {
	sentinelLimits := map[string]interface{}{
		string(corev1.ResourceCPU):    highLimits.Cpu().String(),
		string(corev1.ResourceMemory): highLimits.Memory().String(),
	}
	patchData, err := json.Marshal(map[string]interface{}{
		"spec": map[string]interface{}{
			"engineAdmission": map[string]interface{}{
				"resources": map[string]interface{}{
					"limits": sentinelLimits,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("marshaling engine Config patch: %w", err)
	}

	ec := &unstructured.Unstructured{}
	ec.SetGroupVersionKind(engineConfigGVK)
	ec.SetName("engine-config")
	ec.SetNamespace("runai")

	if err := k8sClient.Patch(ctx, ec, client.RawPatch(types.MergePatchType, patchData)); err != nil {
		if k8serrors.IsNotFound(err) {
			fmt.Println("engine Config not found, skipping engine-admission patch")
			return nil
		}
		return err
	}

	fmt.Println("patched engine Config: set sentinel limits for engineAdmission")
	return nil
}

func removeDeploymentLimits(ctx context.Context, k8sClient client.Client, namespace string) error {
	list := &appsv1.DeploymentList{}
	if err := k8sClient.List(ctx, list, client.InNamespace(namespace)); err != nil {
		return err
	}
	for i := range list.Items {
		dep := &list.Items[i]
		setHighContainerLimits(dep.Spec.Template.Spec.Containers)
		setHighContainerLimits(dep.Spec.Template.Spec.InitContainers)
		if err := k8sClient.Update(ctx, dep); err != nil {
			return fmt.Errorf("updating deployment %s/%s: %w", namespace, dep.Name, err)
		}
		fmt.Printf("set high limits: deployment %s/%s\n", namespace, dep.Name)
	}
	return nil
}

func removeStatefulSetLimits(ctx context.Context, k8sClient client.Client, namespace string) error {
	list := &appsv1.StatefulSetList{}
	if err := k8sClient.List(ctx, list, client.InNamespace(namespace)); err != nil {
		return err
	}
	for i := range list.Items {
		ss := &list.Items[i]
		setHighContainerLimits(ss.Spec.Template.Spec.Containers)
		setHighContainerLimits(ss.Spec.Template.Spec.InitContainers)
		if err := k8sClient.Update(ctx, ss); err != nil {
			return fmt.Errorf("updating statefulset %s/%s: %w", namespace, ss.Name, err)
		}
		fmt.Printf("set high limits: statefulset %s/%s\n", namespace, ss.Name)
	}
	return nil
}

func removeDaemonSetLimits(ctx context.Context, k8sClient client.Client, namespace string) error {
	list := &appsv1.DaemonSetList{}
	if err := k8sClient.List(ctx, list, client.InNamespace(namespace)); err != nil {
		return err
	}
	for i := range list.Items {
		ds := &list.Items[i]
		setHighContainerLimits(ds.Spec.Template.Spec.Containers)
		setHighContainerLimits(ds.Spec.Template.Spec.InitContainers)
		if err := k8sClient.Update(ctx, ds); err != nil {
			return fmt.Errorf("updating daemonset %s/%s: %w", namespace, ds.Name, err)
		}
		fmt.Printf("set high limits: daemonset %s/%s\n", namespace, ds.Name)
	}
	return nil
}

// minRequests is used when a container has no explicit requests: setting it prevents
// Kubernetes admission from defaulting requests=limits (which would make the pod request
// 100 CPU / 1Ti and fail to schedule on real nodes).
var minRequests = corev1.ResourceList{
	corev1.ResourceCPU:    resource.MustParse("1m"),
	corev1.ResourceMemory: resource.MustParse("1Mi"),
}

func setHighContainerLimits(containers []corev1.Container) {
	for i := range containers {
		if len(containers[i].Resources.Limits) == 0 {
			continue
		}
		if len(containers[i].Resources.Requests) == 0 {
			containers[i].Resources.Requests = minRequests
		}
		containers[i].Resources.Limits = highLimits
	}
}

func waitForPodsReady(ctx context.Context, k8sClient client.Client, namespaces []string, timeout time.Duration) error {
	const interval = 10 * time.Second

	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fmt.Printf("waiting up to %s for all pods to be Running/Ready...\n", timeout)
	for {
		select {
		case <-ctxWithTimeout.Done():
			return fmt.Errorf("pods did not become ready within %s", timeout)
		case <-ticker.C:
			ready, err := allPodsReady(ctxWithTimeout, k8sClient, namespaces)
			if err != nil {
				fmt.Printf("pod readiness check error: %v\n", err)
				continue
			}
			if ready {
				fmt.Println("all pods are Running/Ready")
				return nil
			}
		}
	}
}

func allPodsReady(ctx context.Context, k8sClient client.Client, namespaces []string) (bool, error) {
	for _, ns := range namespaces {
		nsObj := &corev1.Namespace{}
		if err := k8sClient.Get(ctx, client.ObjectKey{Name: ns}, nsObj); err != nil {
			if k8serrors.IsNotFound(err) {
				continue
			}
			return false, err
		}

		pods := &corev1.PodList{}
		if err := k8sClient.List(ctx, pods, client.InNamespace(ns)); err != nil {
			return false, fmt.Errorf("listing pods in %s: %w", ns, err)
		}

		for i := range pods.Items {
			pod := &pods.Items[i]
			if !isPodReady(pod) {
				fmt.Printf("pod %s/%s not ready\n", ns, pod.Name)
				return false, nil
			}
		}
	}
	return true, nil
}

func isPodReady(pod *corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodRunning {
		return false
	}
	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}
