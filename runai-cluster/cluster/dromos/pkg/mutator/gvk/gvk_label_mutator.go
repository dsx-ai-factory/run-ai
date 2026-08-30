package gvk

import (
	"context"
	"fmt"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	templateutils "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/utils/template"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	LabelKeyGroup   = "run.ai/workload-group"
	LabelKeyVersion = "run.ai/workload-version"
	LabelKeyKind    = "run.ai/workload-kind"
)

// LabelMutator adds the attached resource GVK as labels to the resource during reconciliation.
type LabelMutator struct {
	parser *templateutils.Parser
}

func NewLabelMutator() *LabelMutator {
	return &LabelMutator{
		parser: templateutils.NewParser(),
	}
}

func (lm *LabelMutator) Name() string {
	return "LabelMutator"
}

// Mutate adds the runner's attached resource GVK as labels to the workload runner.
func (lm *LabelMutator) Mutate(_ context.Context, runner *dromosv1alpha1.WorkloadRunner) (admission.Warnings, error) {
	template, err := lm.parser.Parse(runner.Spec.Template)
	if err != nil {
		return admission.Warnings{fmt.Sprintf("failed to parse template: %v", err)}, nil
	}

	gvk := template.GroupVersionKind()

	labels := runner.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[LabelKeyGroup] = gvk.Group
	labels[LabelKeyVersion] = gvk.Version
	labels[LabelKeyKind] = gvk.Kind

	runner.SetLabels(labels)

	return nil, nil
}
