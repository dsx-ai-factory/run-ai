//go:build e2e

package utils

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// WaitForRunaiClusterStable blocks until every Deployment in the runai
// namespace is Available AND the deployment set has been stable (no count or
// readiness changes) for stableDuration.
//
// orchestrator.RunInstall returns as soon as runai-agent reports liveness to
// the Control Plane. At that point the runai-operator may still be in its
// first reconcile burst, spinning up additional cluster controllers. If a
// test triggers Uninstall while the operator is still reconciling, the
// chart's pre-delete hook ends up fighting the operator instead of tearing
// down a quiescent cluster, which can hang for minutes and exceed the hook's
// BackoffLimit. Calling this helper after RunInstall lets the operator
// settle before the next BeforeEach runs.
func WaitForRunaiClusterStable(timeout time.Duration) error {
	const (
		pollInterval   = 5 * time.Second
		stableDuration = 60 * time.Second
	)

	k8sConfig, err := GetK8sConfig()
	if err != nil {
		return fmt.Errorf("failed to get k8s config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}

	deadline := time.Now().Add(timeout)
	var lastSnapshot map[string]int32
	var stableSince time.Time

	for {
		snapshot, allReady, snapshotErr := runaiDeploymentsSnapshot(clientset)
		switch {
		case snapshotErr != nil:
			fmt.Printf("[WaitForRunaiClusterStable] poll error (will retry): %v\n", snapshotErr)
			stableSince = time.Time{}
		case allReady && len(snapshot) > 0 && reflect.DeepEqual(snapshot, lastSnapshot):
			if stableSince.IsZero() {
				stableSince = time.Now()
				fmt.Printf("[WaitForRunaiClusterStable] %d deployments ready, holding for %v\n", len(snapshot), stableDuration)
			}
			if time.Since(stableSince) >= stableDuration {
				return nil
			}
		default:
			stableSince = time.Time{}
		}
		if snapshotErr == nil {
			lastSnapshot = snapshot
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("runai cluster did not stabilize within %v: last snapshot=%v", timeout, lastSnapshot)
		}
		time.Sleep(pollInterval)
	}
}

func runaiDeploymentsSnapshot(clientset kubernetes.Interface) (map[string]int32, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	deployments, err := clientset.AppsV1().Deployments(step.ClusterNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, false, err
	}

	snapshot := make(map[string]int32, len(deployments.Items))
	allReady := true
	for _, d := range deployments.Items {
		desired := int32(1)
		if d.Spec.Replicas != nil {
			desired = *d.Spec.Replicas
		}
		snapshot[d.Name] = d.Status.AvailableReplicas
		if d.Status.AvailableReplicas < desired || d.Status.ObservedGeneration < d.Generation {
			allReady = false
		}
	}
	return snapshot, allReady, nil
}
