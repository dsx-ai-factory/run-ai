package plugins

import (
	"context"
	"fmt"

	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	templateutils "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/utils/template"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const webhookValidateNamespacePluginName = "webhook-validate-namespace-exists-plugin"

// WebhookValidateNamespaceExistsPlugin is a plugin that validates that the namespace exists in
// given resource as part of the WorkloadRunner.
// This plugin is used for testing the webhook validation.
type WebhookValidateNamespaceExistsPlugin struct {
	parser *templateutils.Parser
}

func NewWebhookValidateNamespaceExistsPlugin() *WebhookValidateNamespaceExistsPlugin {
	return &WebhookValidateNamespaceExistsPlugin{
		parser: templateutils.NewParser(),
	}
}

func (p *WebhookValidateNamespaceExistsPlugin) Name() string {
	return webhookValidateNamespacePluginName
}

// WebhookValidate validates that the namespace exists in the template of the WorkloadRunner.
func (p *WebhookValidateNamespaceExistsPlugin) WebhookValidate(
	_ context.Context,
	_ *dromosv1alpha1.WorkloadRunner,
	desiredRunner *dromosv1alpha1.WorkloadRunner,
) (admission.Warnings, field.ErrorList) {
	template, err := p.parser.Parse(desiredRunner.Spec.Template)
	if err != nil {
		return nil, field.ErrorList{
			field.Invalid(
				field.NewPath("spec", "template"),
				desiredRunner.Spec.Template,
				fmt.Sprintf("failed to parse template: %v", err),
			),
		}
	}

	var fieldErrs field.ErrorList
	if template.GetNamespace() == "" {
		fieldErrs = append(
			fieldErrs,
			field.Required(
				field.NewPath("spec", "template", "namespace"),
				"template must have namespace field set",
			),
		)
	}

	return nil, fieldErrs
}
