package readiness

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// checkWorkloadsReady asserts every Deployment, StatefulSet and DaemonSet in the target
// namespaces has all of its replicas ready. DaemonSets are included because setup-vpa
// creates VPA objects for them too, so their health matters to the measurement.
func (c *Checker) checkWorkloadsReady(ctx context.Context) CheckResult {
	res := CheckResult{Check: "workloads_ready", Status: StatusPass}
	total := 0

	for _, ns := range c.namespaces {
		exists, err := c.namespaceExists(ctx, ns)
		if err != nil {
			return failf(res, "checking namespace %s: %v", ns, err)
		}
		if !exists {
			continue
		}

		deps := &appsv1.DeploymentList{}
		if err := c.clients.KubeClient.List(ctx, deps, client.InNamespace(ns)); err != nil {
			return failf(res, "listing deployments in %s: %v", ns, err)
		}
		for _, d := range deps.Items {
			total++
			want := replicaCount(d.Spec.Replicas)
			if d.Status.ReadyReplicas != want {
				res.Status = StatusFail
				res.Detail = append(res.Detail,
					fmt.Sprintf("Deployment %s/%s ready %d/%d", ns, d.Name, d.Status.ReadyReplicas, want))
			}
		}

		sets := &appsv1.StatefulSetList{}
		if err := c.clients.KubeClient.List(ctx, sets, client.InNamespace(ns)); err != nil {
			return failf(res, "listing statefulsets in %s: %v", ns, err)
		}
		for _, s := range sets.Items {
			total++
			want := replicaCount(s.Spec.Replicas)
			if s.Status.ReadyReplicas != want {
				res.Status = StatusFail
				res.Detail = append(res.Detail,
					fmt.Sprintf("StatefulSet %s/%s ready %d/%d", ns, s.Name, s.Status.ReadyReplicas, want))
			}
		}

		// A DaemonSet's desired count is scheduling-dependent rather than a fixed
		// replica number, so it is compared against DesiredNumberScheduled.
		daemons := &appsv1.DaemonSetList{}
		if err := c.clients.KubeClient.List(ctx, daemons, client.InNamespace(ns)); err != nil {
			return failf(res, "listing daemonsets in %s: %v", ns, err)
		}
		for _, d := range daemons.Items {
			total++
			if d.Status.NumberReady != d.Status.DesiredNumberScheduled {
				res.Status = StatusFail
				res.Detail = append(res.Detail,
					fmt.Sprintf("DaemonSet %s/%s ready %d/%d", ns, d.Name, d.Status.NumberReady, d.Status.DesiredNumberScheduled))
			}
		}
	}

	res.Summary = fmt.Sprintf("%d workloads checked, %d not fully ready", total, len(res.Detail))
	return res
}

// checkPodsReady asserts the Ready condition, not merely phase==Running. A pod can be
// Running with a container failing its readiness probe because it is waiting on another
// component to come up; treating that as ready is exactly the mistake this gate exists to
// prevent.
func (c *Checker) checkPodsReady(ctx context.Context) CheckResult {
	res := CheckResult{Check: "pods_ready", Status: StatusPass}
	total := 0

	for _, ns := range c.namespaces {
		pods, err := c.listPods(ctx, ns)
		if err != nil {
			return failf(res, "listing pods in %s: %v", ns, err)
		}
		for i := range pods {
			p := &pods[i]
			// Completed pods (install hooks, migration jobs) never become Ready and are
			// not a readiness problem.
			if p.Status.Phase == corev1.PodSucceeded {
				continue
			}
			total++

			if !podReady(p) {
				res.Status = StatusFail
				res.Detail = append(res.Detail,
					fmt.Sprintf("pod %s/%s not Ready (phase=%s)%s", ns, p.Name, p.Status.Phase, notReadyContainers(p)))
			}
		}
	}

	res.Summary = fmt.Sprintf("%d pods checked, %d not Ready", total, len(res.Detail))
	return res
}

func replicaCount(spec *int32) int32 {
	if spec == nil {
		return 1
	}
	return *spec
}

func notReadyContainers(p *corev1.Pod) string {
	var names []string
	for _, cs := range p.Status.ContainerStatuses {
		if !cs.Ready {
			names = append(names, cs.Name)
		}
	}
	if len(names) == 0 {
		return ""
	}
	return fmt.Sprintf(" containers=%v", names)
}
