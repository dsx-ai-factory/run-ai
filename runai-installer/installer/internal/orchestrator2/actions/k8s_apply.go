package actions

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sApply struct {
	Object client.Object
}

func (a K8sApply) Kind() string { return "k8s.apply" }

func (a K8sApply) Args() map[string]any {
	return map[string]any{
		"name":      a.Object.GetName(),
		"namespace": a.Object.GetNamespace(),
		"kind":      a.Object.GetObjectKind().GroupVersionKind().Kind,
	}
}

func (a K8sApply) Describe() string {
	gvk := a.Object.GetObjectKind().GroupVersionKind()
	return fmt.Sprintf("kubectl apply %s %s", gvk.Kind, a.Object.GetName())
}

func (a K8sApply) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	return orchestrator2.ActionResult{}, deps.K8sExecutor.Apply(ctx, a.Object)
}
