package orchestrator2

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// NodeLabeler applies a label to every worker (non-control-plane) node. The
// fake-GPU prerequisite uses it to mark nodes as simulated GPU nodes.
type NodeLabeler interface {
	LabelWorkerNodes(ctx context.Context, key, value string) (labeled int, err error)
	// LabelNodes applies key=value to each named node. Used by the system-node
	// placement feature to mark a chosen subset.
	LabelNodes(ctx context.Context, names []string, key, value string) error
}

type realNodeLabeler struct {
	cs kubernetes.Interface
}

// NewNodeLabeler builds the production NodeLabeler from a REST config.
func NewNodeLabeler(cfg *rest.Config) (NodeLabeler, error) {
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &realNodeLabeler{cs: cs}, nil
}

func (n *realNodeLabeler) LabelWorkerNodes(ctx context.Context, key, value string) (int, error) {
	nodes, err := n.cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}
	patch, err := json.Marshal(map[string]any{
		"metadata": map[string]any{"labels": map[string]string{key: value}},
	})
	if err != nil {
		return 0, err
	}
	labeled := 0
	for i := range nodes.Items {
		if isControlPlane(&nodes.Items[i]) {
			continue
		}
		if _, err := n.cs.CoreV1().Nodes().Patch(ctx, nodes.Items[i].Name, types.MergePatchType, patch, metav1.PatchOptions{}); err != nil {
			return labeled, fmt.Errorf("label node %s: %w", nodes.Items[i].Name, err)
		}
		labeled++
	}
	return labeled, nil
}

func (n *realNodeLabeler) LabelNodes(ctx context.Context, names []string, key, value string) error {
	patch, err := json.Marshal(map[string]any{
		"metadata": map[string]any{"labels": map[string]string{key: value}},
	})
	if err != nil {
		return err
	}
	for _, name := range names {
		if _, err := n.cs.CoreV1().Nodes().Patch(ctx, name, types.MergePatchType, patch, metav1.PatchOptions{}); err != nil {
			return fmt.Errorf("label node %s: %w", name, err)
		}
	}
	return nil
}

func isControlPlane(node *corev1.Node) bool {
	_, master := node.Labels["node-role.kubernetes.io/master"]
	_, cp := node.Labels["node-role.kubernetes.io/control-plane"]
	return master || cp
}
