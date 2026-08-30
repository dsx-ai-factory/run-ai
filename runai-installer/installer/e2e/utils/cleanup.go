//go:build e2e

package utils

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/uninstall"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Uninstall performs a full RunAI uninstall using the production uninstall flow,
// then waits for the cluster to settle before returning.
// Used as test cleanup between tests to guarantee a clean cluster state.
func Uninstall(logsDir string) error {
	cfg := &config.UninstallConfig{
		DevMode:          true,
		LogsDir:          logsDir,
		StepTimeout:      10 * time.Minute,
		ExecutionTimeout: 30 * time.Minute,
	}
	deps, err := uninstall.ProductionDeps()
	if err != nil {
		return err
	}
	if err := orchestrator.RunUninstall(cfg, deps); err != nil {
		return err
	}
	return waitForClusterCalm(3 * time.Minute)
}

// waitForClusterCalm polls until all nodes are free of resource pressure conditions
// (MemoryPressure, DiskPressure, PIDPressure) and no RunAI pods are still terminating.
// This prevents the next test from starting on a cluster that is still under stress.
// API errors are treated as "not calm yet" and retried rather than returned as failures,
// since the cluster may be temporarily unable to serve requests under load.
func waitForClusterCalm(timeout time.Duration) error {
	k8sConfig, err := GetK8sConfig()
	if err != nil {
		return fmt.Errorf("failed to get k8s config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}

	const pollInterval = 10 * time.Second
	deadline := time.Now().Add(timeout)

	for {
		pressured, terminating := clusterPressureStatus(clientset)
		if len(pressured) == 0 && terminating == 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("cluster did not calm down within %v: node pressure=%v, terminating pods=%d", timeout, pressured, terminating)
		}
		if len(pressured) > 0 {
			fmt.Printf("[waitForClusterCalm] node pressure active: %v\n", pressured)
		}
		if terminating > 0 {
			fmt.Printf("[waitForClusterCalm] waiting for %d terminating RunAI pods\n", terminating)
		}
		time.Sleep(pollInterval)
	}
}

// clusterPressureStatus returns nodes with active pressure conditions and the number of
// terminating RunAI pods. API errors are silently ignored so that a struggling cluster
// is treated as "not calm yet" rather than causing the check to abort.
func clusterPressureStatus(clientset kubernetes.Interface) (pressuredNodes []string, terminatingPods int) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{}); err == nil {
		for _, node := range nodes.Items {
			for _, cond := range node.Status.Conditions {
				isPressure := cond.Type == corev1.NodeMemoryPressure ||
					cond.Type == corev1.NodeDiskPressure ||
					cond.Type == corev1.NodePIDPressure
				if isPressure && cond.Status == corev1.ConditionTrue {
					pressuredNodes = append(pressuredNodes, fmt.Sprintf("%s(%s)", node.Name, cond.Type))
				}
			}
		}
	}

	for _, ns := range []string{step.ClusterNamespace, step.CPNamespace, step.ReservationNamespace, step.ScaleAdjustNamespace} {
		pods, err := clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			continue // namespace may not exist yet, or API is temporarily unavailable
		}
		for _, pod := range pods.Items {
			if pod.DeletionTimestamp != nil {
				terminatingPods++
			}
		}
	}
	return pressuredNodes, terminatingPods
}

func GetK8sConfig() (*rest.Config, error) {
	home, _ := os.UserHomeDir()
	kubeconfigPath := home + "/.kube/config"
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		kubeconfigPath = os.Getenv("KUBECONFIG")
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}
