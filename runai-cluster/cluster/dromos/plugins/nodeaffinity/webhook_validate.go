package nodeaffinity

import (
	"context"
	"fmt"
	"strconv"

	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	affinityutils "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/affinity"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	minPreferredWeight = 1
	maxPreferredWeight = 100

	// The merge ANDs the instruction into the workload's own affinity as a
	// cartesian product, so the term count multiplies. Both sides are user
	// supplied and the CRD cannot bound an embedded corev1 type, so bound them
	// here. A real instruction is one term; these are far above any honest use.
	maxTerms                = 16
	maxValuesPerRequirement = 256
	maxMergedTerms          = maxTerms * maxTerms
)

var instructionPath = field.NewPath("spec", "instructions", "nodeAffinity", "nodeAffinity")

// WebhookValidate rejects an instruction the cluster could not act on: one that
// is structurally invalid, or one the workload's Karta has no place to write.
// Failing here means a restart patch dies before the running child is deleted,
// rather than half-way through reconciliation.
func (p *Plugin) WebhookValidate(
	ctx context.Context,
	_, desiredRunner *dromosv1alpha1.WorkloadRunner,
) (admission.Warnings, field.ErrorList) {
	desired := extractNodeAffinity(desiredRunner)
	if desired == nil {
		return nil, nil
	}

	if errs := validateNodeAffinity(instructionPath, desired); len(errs) > 0 {
		return nil, errs
	}

	return nil, p.validateKartaSupport(ctx, desiredRunner)
}

// validateKartaSupport checks that the Karta backing the template's GVK can carry
// node affinity. A missing Karta is not an admission error: the controller
// already reports it on status.
func (p *Plugin) validateKartaSupport(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner) field.ErrorList {
	if p.kartas == nil {
		return nil
	}

	template, err := p.parser.Parse(runner.Spec.Template)
	if err != nil {
		// The controller's own GVK validator already rejects unparsable templates.
		return nil
	}

	gvk := template.GroupVersionKind()
	karta, err := p.kartas.GetKartaForGVK(ctx, gvk)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return field.ErrorList{field.InternalError(instructionPath, fmt.Errorf("look up karta for %s: %w", gvk, err))}
	}

	if err := affinityutils.SupportsNodeAffinity(karta); err != nil {
		return field.ErrorList{field.Forbidden(instructionPath,
			fmt.Sprintf("karta %s for %s cannot express node affinity: %v", karta.Name, gvk.Kind, err))}
	}
	return nil
}

// The apiserver skips label-value validation only on preferred terms, so the
// webhook does the same.
func validateNodeAffinity(path *field.Path, affinity *corev1.NodeAffinity) field.ErrorList {
	var errs field.ErrorList

	if required := affinity.RequiredDuringSchedulingIgnoredDuringExecution; required != nil {
		requiredPath := path.Child("requiredDuringSchedulingIgnoredDuringExecution", "nodeSelectorTerms")

		if len(required.NodeSelectorTerms) == 0 {
			errs = append(errs, field.Required(requiredPath, "at least one node selector term is required"))
		}
		if len(required.NodeSelectorTerms) > maxTerms {
			errs = append(errs, field.TooMany(requiredPath, len(required.NodeSelectorTerms), maxTerms))
		}
		for i, term := range required.NodeSelectorTerms {
			errs = append(errs, validateTerm(requiredPath.Index(i), term, false)...)
		}
	}

	preferredPath := path.Child("preferredDuringSchedulingIgnoredDuringExecution")
	preferred := affinity.PreferredDuringSchedulingIgnoredDuringExecution
	if len(preferred) > maxTerms {
		errs = append(errs, field.TooMany(preferredPath, len(preferred), maxTerms))
	}
	for i, term := range preferred {
		termPath := preferredPath.Index(i)
		if term.Weight < minPreferredWeight || term.Weight > maxPreferredWeight {
			errs = append(errs, field.Invalid(termPath.Child("weight"), term.Weight,
				fmt.Sprintf("must be in the range %d-%d", minPreferredWeight, maxPreferredWeight)))
		}
		errs = append(errs, validateTerm(termPath.Child("preference"), term.Preference, true)...)
	}

	return errs
}

// An empty required term matches no nodes; an empty preferred term is a no-op
// the apiserver accepts.
func validateTerm(path *field.Path, term corev1.NodeSelectorTerm, preferred bool) field.ErrorList {
	if len(term.MatchExpressions) == 0 && len(term.MatchFields) == 0 {
		if preferred {
			return nil
		}
		return field.ErrorList{field.Required(path, "must have at least one matchExpressions or matchFields entry")}
	}

	var errs field.ErrorList
	for i, req := range term.MatchExpressions {
		errs = append(errs, validateMatchExpression(path.Child("matchExpressions").Index(i), req, preferred)...)
	}
	for i, req := range term.MatchFields {
		errs = append(errs, validateMatchField(path.Child("matchFields").Index(i), req)...)
	}
	return errs
}

func validateMatchExpression(path *field.Path, req corev1.NodeSelectorRequirement, allowInvalidLabelValues bool) field.ErrorList {
	var errs field.ErrorList
	for _, msg := range validation.IsQualifiedName(req.Key) {
		errs = append(errs, field.Invalid(path.Child("key"), req.Key, msg))
	}

	switch req.Operator {
	case corev1.NodeSelectorOpIn, corev1.NodeSelectorOpNotIn:
		if len(req.Values) == 0 {
			errs = append(errs, field.Required(path.Child("values"), "must have at least one value for In/NotIn"))
		}
		if len(req.Values) > maxValuesPerRequirement {
			errs = append(errs, field.TooMany(path.Child("values"), len(req.Values), maxValuesPerRequirement))
		}
		if !allowInvalidLabelValues {
			for j, value := range req.Values {
				for _, msg := range validation.IsValidLabelValue(value) {
					errs = append(errs, field.Invalid(path.Child("values").Index(j), value, msg))
				}
			}
		}
	case corev1.NodeSelectorOpExists, corev1.NodeSelectorOpDoesNotExist:
		if len(req.Values) > 0 {
			errs = append(errs, field.Forbidden(path.Child("values"), "may not be specified for Exists/DoesNotExist"))
		}
	case corev1.NodeSelectorOpGt, corev1.NodeSelectorOpLt:
		if len(req.Values) != 1 {
			errs = append(errs, field.Invalid(path.Child("values"), req.Values, "must have exactly one value for Gt/Lt"))
		} else if _, err := strconv.ParseInt(req.Values[0], 10, 64); err != nil {
			errs = append(errs, field.Invalid(path.Child("values").Index(0), req.Values[0], "must be an integer for Gt/Lt"))
		}
	default:
		errs = append(errs, field.NotSupported(path.Child("operator"), req.Operator, []string{
			string(corev1.NodeSelectorOpIn), string(corev1.NodeSelectorOpNotIn),
			string(corev1.NodeSelectorOpExists), string(corev1.NodeSelectorOpDoesNotExist),
			string(corev1.NodeSelectorOpGt), string(corev1.NodeSelectorOpLt),
		}))
	}

	return errs
}

func validateMatchField(path *field.Path, req corev1.NodeSelectorRequirement) field.ErrorList {
	if req.Key != nodeFieldSelectorKey {
		return field.ErrorList{field.NotSupported(path.Child("key"), req.Key, []string{nodeFieldSelectorKey})}
	}

	var errs field.ErrorList
	if req.Operator != corev1.NodeSelectorOpIn && req.Operator != corev1.NodeSelectorOpNotIn {
		errs = append(errs, field.NotSupported(path.Child("operator"), req.Operator,
			[]string{string(corev1.NodeSelectorOpIn), string(corev1.NodeSelectorOpNotIn)}))
	}
	if len(req.Values) != 1 {
		errs = append(errs, field.Invalid(path.Child("values"), req.Values, "must have exactly one value"))
	}
	for j, value := range req.Values {
		for _, msg := range validation.IsDNS1123Subdomain(value) {
			errs = append(errs, field.Invalid(path.Child("values").Index(j), value, msg))
		}
	}
	return errs
}

// The only field a node field selector may address.
const nodeFieldSelectorKey = "metadata.name"
