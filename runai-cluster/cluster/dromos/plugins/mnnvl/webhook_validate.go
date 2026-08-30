package mnnvl

import (
	"context"

	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// WebhookValidate validates MNNVL configuration in WorkloadRunner
func (p *Plugin) WebhookValidate(_ context.Context, _ *dromosv1alpha1.WorkloadRunner, desiredRunner *dromosv1alpha1.WorkloadRunner) (admission.Warnings, field.ErrorList) {
	config := extractMnnvlConfig(desiredRunner)
	if config == nil || config.Mode == dromosv1alpha1.MnnvlModeNone {
		return nil, nil
	}

	var errs field.ErrorList
	basePath := field.NewPath("spec", "instructions", "gpuOptimizations", "mnnvl")

	if config.Mode != dromosv1alpha1.MnnvlModeRequired {
		errs = append(errs, field.NotSupported(
			basePath.Child("mode"),
			config.Mode,
			[]string{string(dromosv1alpha1.MnnvlModeNone), string(dromosv1alpha1.MnnvlModeRequired)}))
	}

	return nil, errs
}
