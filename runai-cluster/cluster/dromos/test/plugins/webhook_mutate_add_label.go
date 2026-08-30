package plugins

import (
	"context"

	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	templateutils "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/utils/template"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const webhookMutateAddLabelPluginName = "webhook-mutate-add-label-plugin"

// WebhookMutateAddLabelPlugin is a plugin that mutates the WorkloadRunner
// by adding a label "example" with value "true".
// This plugin is used for testing the webhook mutation.
type WebhookMutateAddLabelPlugin struct {
	parser *templateutils.Parser
}

func NewWebhookMutateAddLabelPlugin() *WebhookMutateAddLabelPlugin {
	return &WebhookMutateAddLabelPlugin{
		parser: templateutils.NewParser(),
	}
}

func (p *WebhookMutateAddLabelPlugin) Name() string {
	return webhookMutateAddLabelPluginName
}

// WebhookMutate adds the label "example" with value "true" to the WorkloadRunner.
func (p *WebhookMutateAddLabelPlugin) WebhookMutate(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
) (admission.Warnings, error) {
	if runner.Labels == nil {
		runner.Labels = make(map[string]string)
	}

	if _, exists := runner.Labels["example"]; !exists {
		runner.Labels["example"] = "true"
	}

	return nil, nil
}

var _ framework.MutatorPlugin = &WebhookMutateAddLabelPlugin{}
