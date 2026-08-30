package gvk

import (
	"context"
	"fmt"

	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	templateutils "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/utils/template"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// GVKValidator validates that attached resource name, namespace, and GVK have values.
// It also validates that immutable fields (attached resource name, namespace, and GVK) are not changed on updates.
type GVKValidator struct {
	parser *templateutils.Parser
}

func NewGVKValidator() *GVKValidator {
	return &GVKValidator{parser: templateutils.NewParser()}
}

func (v *GVKValidator) Name() string {
	return "GVKValidator"
}

func (v *GVKValidator) Validate(_ context.Context, existingRunner, desiredRunner *dromosv1alpha1.WorkloadRunner) (admission.Warnings, field.ErrorList) {
	var fieldErrs field.ErrorList

	desiredTemplate, err := v.parser.Parse(desiredRunner.Spec.Template)
	if err != nil {
		return nil, field.ErrorList{
			field.Invalid(
				field.NewPath("spec", "template"),
				desiredRunner.Spec.Template,
				fmt.Sprintf("failed to parse template: %v", err),
			),
		}
	}

	basePath := field.NewPath("spec", "template")

	if desiredTemplate.GetName() == "" {
		fieldErrs = append(
			fieldErrs,
			field.Required(
				basePath.Child("metadata", "name"),
				"template name must be set",
			),
		)
	}

	if desiredTemplate.GetNamespace() == "" {
		fieldErrs = append(
			fieldErrs,
			field.Required(
				basePath.Child("metadata", "namespace"),
				"template namespace must be set",
			),
		)
	}

	newGVK := desiredTemplate.GroupVersionKind()

	if newGVK.Kind == "" {
		fieldErrs = append(
			fieldErrs,
			field.Required(
				basePath.Child("kind"),
				"kind must be set",
			),
		)
	}

	if newGVK.Version == "" {
		fieldErrs = append(
			fieldErrs,
			field.Required(
				basePath.Child("apiVersion"),
				"apiVersion must be set",
			),
		)
	}

	// Validate immutable fields on updates
	if existingRunner != nil {
		existingTemplate, err := v.parser.Parse(existingRunner.Spec.Template)
		if err != nil {
			return nil, field.ErrorList{
				field.Invalid(
					field.NewPath("spec", "template"),
					existingRunner.Spec.Template,
					fmt.Sprintf("failed to parse template: %v", err),
				),
			}
		}
		fieldErrs = append(fieldErrs, v.validateImmutableFields(desiredTemplate, existingTemplate)...)
	}

	return nil, fieldErrs
}

// validateImmutableFields validates that attached resource name, namespace, and GVK are not changed on updates.
func (v *GVKValidator) validateImmutableFields(newTemplate, oldTemplate *unstructured.Unstructured) field.ErrorList {
	var fieldErrs field.ErrorList
	basePath := field.NewPath("spec", "template")

	oldTemplateName := oldTemplate.GetName()
	newTemplateName := newTemplate.GetName()
	if oldTemplateName != newTemplateName {
		fieldErrs = append(
			fieldErrs,
			field.Forbidden(
				basePath.Child("metadata", "name"),
				fmt.Sprintf("template name is immutable and cannot be changed from %q to %q", oldTemplateName, newTemplateName),
			),
		)
	}

	oldTemplateNamespace := oldTemplate.GetNamespace()
	newTemplateNamespace := newTemplate.GetNamespace()
	if oldTemplateNamespace != newTemplateNamespace {
		fieldErrs = append(
			fieldErrs,
			field.Forbidden(
				basePath.Child("metadata", "namespace"),
				fmt.Sprintf("template namespace is immutable and cannot be changed from %q to %q", oldTemplateNamespace, newTemplateNamespace),
			),
		)
	}

	oldGVK := oldTemplate.GroupVersionKind()
	newGVK := newTemplate.GroupVersionKind()
	if oldGVK != newGVK {
		fieldErrs = append(
			fieldErrs,
			field.Forbidden(
				basePath,
				fmt.Sprintf("template GVK is immutable and cannot be changed from %q to %q", oldGVK.String(), newGVK.String()),
			),
		)
	}

	return fieldErrs
}
