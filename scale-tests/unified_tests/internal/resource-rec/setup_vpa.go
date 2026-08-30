package resource_rec

import (
	"context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var updateModeOff = vpav1.UpdateModeOff

// EnsureVPAObjects creates or updates a VPA object (updateMode: Off) for every
// Deployment, StatefulSet, and DaemonSet across the given namespaces.
// Namespaces that don't exist in the cluster are skipped without error.
func EnsureVPAObjects(ctx context.Context, k8sClient client.Client, namespaces []string) error {
	for _, ns := range namespaces {
		nsObj := &corev1.Namespace{}
		if err := k8sClient.Get(ctx, client.ObjectKey{Name: ns}, nsObj); err != nil {
			if k8serrors.IsNotFound(err) {
				fmt.Printf("namespace %s not found, skipping\n", ns)
				continue
			}
			return fmt.Errorf("checking namespace %s: %w", ns, err)
		}

		if err := ensureDeploymentVPAs(ctx, k8sClient, ns); err != nil {
			return fmt.Errorf("namespace %s deployments: %w", ns, err)
		}
		if err := ensureStatefulSetVPAs(ctx, k8sClient, ns); err != nil {
			return fmt.Errorf("namespace %s statefulsets: %w", ns, err)
		}
		if err := ensureDaemonSetVPAs(ctx, k8sClient, ns); err != nil {
			return fmt.Errorf("namespace %s daemonsets: %w", ns, err)
		}
	}
	return nil
}

func ensureDeploymentVPAs(ctx context.Context, k8sClient client.Client, namespace string) error {
	list := &appsv1.DeploymentList{}
	if err := k8sClient.List(ctx, list, client.InNamespace(namespace)); err != nil {
		return err
	}
	for _, dep := range list.Items {
		if err := ensureVPA(ctx, k8sClient, namespace, "Deployment", "apps/v1", dep.Name); err != nil {
			return err
		}
	}
	return nil
}

func ensureStatefulSetVPAs(ctx context.Context, k8sClient client.Client, namespace string) error {
	list := &appsv1.StatefulSetList{}
	if err := k8sClient.List(ctx, list, client.InNamespace(namespace)); err != nil {
		return err
	}
	for _, ss := range list.Items {
		if err := ensureVPA(ctx, k8sClient, namespace, "StatefulSet", "apps/v1", ss.Name); err != nil {
			return err
		}
	}
	return nil
}

func ensureDaemonSetVPAs(ctx context.Context, k8sClient client.Client, namespace string) error {
	list := &appsv1.DaemonSetList{}
	if err := k8sClient.List(ctx, list, client.InNamespace(namespace)); err != nil {
		return err
	}
	for _, ds := range list.Items {
		if err := ensureVPA(ctx, k8sClient, namespace, "DaemonSet", "apps/v1", ds.Name); err != nil {
			return err
		}
	}
	return nil
}

func ensureVPA(ctx context.Context, k8sClient client.Client, namespace, kind, apiVersion, name string) error {
	vpa := &vpav1.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vpaObjectName(kind, name),
			Namespace: namespace,
		},
	}
	result, err := controllerutil.CreateOrUpdate(ctx, k8sClient, vpa, func() error {
		vpa.Spec = vpav1.VerticalPodAutoscalerSpec{
			TargetRef: &autoscalingv1.CrossVersionObjectReference{
				APIVersion: apiVersion,
				Kind:       kind,
				Name:       name,
			},
			UpdatePolicy: &vpav1.PodUpdatePolicy{
				UpdateMode: &updateModeOff,
			},
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("vpa for %s/%s %s: %w", namespace, kind, name, err)
	}
	fmt.Printf("%s vpa: %s/%s (target: %s/%s)\n", result, namespace, vpa.Name, kind, name)
	return nil
}

func vpaObjectName(kind, name string) string {
	return fmt.Sprintf("%s-%s", strings.ToLower(kind), name)
}
