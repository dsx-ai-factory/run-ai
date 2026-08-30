/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package v2alpha1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/nodepools"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/utils/runai_rules"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
)

// +kubebuilder:webhook:path=/validate-run-ai-v2alpha1-inferencepolicy,mutating=false,failurePolicy=fail,sideEffects=None,groups=run.ai,resources=inferencepolicies,verbs=create;update,versions=v2alpha1,name=runai-inference-policy-validation.kb.io,admissionReviewVersions=v1

// we need a client to fetch k8s resources that we need during the validation
var k8sclient client.Client

func (r *InferencePolicy) SetupWebhookWithManager(mgr ctrl.Manager) error {

	k8sclient = mgr.GetClient()

	// controller-runtime v0.23 made the webhook builder generic over the object
	// type. We keep the validator implemented against runtime.Object (the
	// admission.CustomValidator shape), so instantiate the builder with T =
	// runtime.Object by passing the object through a runtime.Object variable.
	var obj runtime.Object = r
	return ctrl.NewWebhookManagedBy(mgr, obj).
		WithValidator(&InferencePolicy{}).
		Complete()
}

func (r *InferencePolicy) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	policy, ok := obj.(*InferencePolicy)
	if !ok {
		return nil, fmt.Errorf("expected an InferencePolicy but got a %T", obj)
	}

	log := logger.FromContext(ctx)
	log.Info("received request", "name", policy.ObjectMeta.Name,
		"namespace", policy.ObjectMeta.Namespace, "spec", policy.Spec)

	//
	//    we limit the administrator to one policy per namespace
	//
	var result InferencePolicyList
	err := k8sclient.List(ctx, &result, client.InNamespace(policy.Namespace))
	if err != nil {
		log.Error(err, "Failed to list policies")
		return nil, fmt.Errorf("failed to find inference policies")
	}

	if len(result.Items) != 0 {
		item := result.Items[0]
		if item.Name != policy.Name {
			return nil, fmt.Errorf("policy of type '%s' already exists: %s", constants.KindInferencePolicy, workloadproject.NamespacedName(item.ObjectMeta))
		}
	}

	//
	//   load the spec of the policy for validation
	//
	policySpec, err := utils.NewJsonQ(policy.Spec)
	if err != nil {
		log.Error(err, "could not load spec as JSON", "id", workloadproject.NamespacedName(policy.ObjectMeta))
		return nil, fmt.Errorf("could not parse %s.Spec", workloadproject.NamespacedName(policy.ObjectMeta))
	}

	//
	//   load the rules as jsonQ
	//
	runaiRules := GetRunaiRules(ctx, nodepools.NewNameLister(k8sclient))

	runaiRulesQ, err := utils.NewJsonQ(runaiRules)
	if err != nil {
		return nil, err
	}

	err = enforcer.ValidateSyntaxOfPolicy("", policySpec, runaiRulesQ.JumpToTop())
	if err != nil {
		return nil, err
	}

	//
	//    create a merge stack with only the inference policy and the runai rules in order
	//    to validate the policy itself against the run:ai global rules.
	//
	emptyPolicy := NewEmptyInferencePolicy()

	runai_rules.PatchPolicyWithRunaiDefaults(ctx, &runaiRules.ContainerFields, &emptyPolicy.Spec.ContainerFields,
		nil, nil, &emptyPolicy.Spec.RunaiSchedulerFields)

	emptyPolicyQ, _ := utils.NewJsonQ(emptyPolicy.Spec)

	mergeStack := enforcer.NewMergeStack(emptyPolicyQ, "", runaiRulesQ.JumpToTop(), false)
	mergeStack.AddLevel(policySpec, enforcer.LevelPolicy, policy.Name)

	err = mergeStack.MergeAndValidate(ctx, enforcer.MergeAndValidateOptions{})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *InferencePolicy) ValidateUpdate(ctx context.Context, _ runtime.Object, newObj runtime.Object) (admission.Warnings, error) {
	return r.ValidateCreate(ctx, newObj)
}

func (r *InferencePolicy) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

// InitTest is a function to allow testing with fake client
func InitTest(fakeClient client.Client) {
	k8sclient = fakeClient
}
