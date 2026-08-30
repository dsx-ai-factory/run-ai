package factory

// TODO remove this file after copying the content to inference-workload-controller
import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/nodepools"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	tpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	policyUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/utils/policy"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/utils/runai_rules"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type InferenceWorkloadMergeCreator struct {
	k8sClient client.Client
}

func NewInferenceWorkloadMergeCreator(k8sClient client.Client) *InferenceWorkloadMergeCreator {
	return &InferenceWorkloadMergeCreator{k8sClient: k8sClient}
}

func (self *InferenceWorkloadMergeCreator) CreateMergeStack(ctx context.Context, workloadObject client.Object) (map[api.SubresourceComponent]*enforcer.MergeStack, error) {
	var mergeStackMap = make(map[api.SubresourceComponent]*enforcer.MergeStack, 1)
	log := logf.FromContext(ctx)
	inference, ok := workloadObject.(*twv2.InferenceWorkload)
	if !ok {
		return nil, fmt.Errorf("cannot cast object to InferenceWorkload")
	}

	id := workloadproject.NamespacedName(inference.ObjectMeta)

	//
	//    load the spec as json
	//
	workloadSpec, err := utils.NewJsonQ(inference.Spec)
	if err != nil {
		log.Error(err, "failed to load spec as JSON", "id", id)
		return nil, fmt.Errorf("failed to parse %s.Spec", id)
	}

	runaiRules := tpv1.GetRunaiRules(ctx, nodepools.NewNameLister(self.k8sClient))

	runaiRulesQ, err := utils.NewJsonQ(runaiRules)
	if err != nil {
		return nil, err
	}

	mergeStack := enforcer.NewMergeStack(workloadSpec, workloadproject.NamespacedName(inference.ObjectMeta), runaiRulesQ.JumpToTop(), inference.Spec.Usage == common.UsedForSubmit)

	//
	//    if the workload inherits from another workload...
	//
	if ptr.Deref(inference.Spec.BaseWorkload, "") != "" {
		err = self.handleBaseInferenceWorkload(ctx, inference, mergeStack)
		if err != nil {
			return nil, err
		}
	}

	//
	// get the policy
	//
	policy, levelName, err := self.GetPolicy(ctx, inference)
	if err != nil {
		return nil, err
	}

	runai_rules.PatchPolicyWithRunaiDefaults(ctx, &runaiRules.ContainerFields, &policy.Spec.ContainerFields,
		nil, nil, &policy.Spec.RunaiSchedulerFields)

	//
	//    load the policy as json and add it to the stack
	//
	policySpec, err := utils.NewJsonQ(policy.Spec)
	if err != nil {
		log.Error(err, "could not load the policy spec as JSON")
		return nil, fmt.Errorf("failed to parse the inference policy")
	}

	mergeStack.AddLevel(policySpec, enforcer.LevelPolicy, levelName)
	mergeStackMap[api.SpecSubResource] = mergeStack
	return mergeStackMap, nil

}

func (self *InferenceWorkloadMergeCreator) GetPolicy(ctx context.Context, workload *twv2.InferenceWorkload) (*tpv1.InferencePolicy, string, error) {
	log := logf.FromContext(ctx)
	var levelName string
	policy := tpv1.NewEmptyInferencePolicy()
	if workload.Namespace != "" {
		if utils.IsControlPlanePoliciesEnabled() {
			log.V(4).Info("new policy sync is enabled")
			policyObject, foundLevel, err := policyUtils.FindBackendSyncedPolicy(ctx, self.k8sClient, workload.Namespace, &workload.ObjectMeta.Labels, policyUtils.KindInferencePolicy)
			if err != nil {
				return nil, "", err
			}
			if policyObject != nil {
				policy = policyObject.(*tpv1.InferencePolicy)
			}

			levelName = string(foundLevel) + ":" + workload.ObjectMeta.Name
		} else {
			log.V(4).Info("new policy sync is disabled")
			var err error
			policy, levelName, err = self.getOldInferencePolicy(ctx, workload)
			if err != nil {
				return nil, "", err
			}
		}
	}
	return policy, levelName, nil
}

func (self *InferenceWorkloadMergeCreator) getOldInferencePolicy(ctx context.Context, inference *twv2.InferenceWorkload) (*tpv1.InferencePolicy, string, error) {
	//
	//   the topmost level in the stack is an (optional) policy. find the one relevant policy
	//   (it is premised to have only one of each kind). it is ok if there is no policy for this type.
	//   in this case only run:ai rules will be validated
	//
	policy, levelName, err := self.findPolicyInNamespace(ctx, inference.Namespace)
	if err != nil {
		return nil, "", fmt.Errorf("failed to find for inference policy in namespace %s", inference.Namespace)

	}

	if policy == nil {
		//   if the policy is not found in the namespace, look for it in the run:ai namespace
		policy, levelName, err = self.findPolicyInNamespace(ctx, constants.NamespaceRunAi)
		if err != nil {
			return nil, "", fmt.Errorf("failed to find for inference policy in namespace %s", constants.NamespaceRunAi)

		}
	}
	if policy == nil {
		policy = tpv1.NewEmptyInferencePolicy()
		levelName = ""
	}
	//  if the policy is not found in the run:ai namespace, create an empty policy
	return policy, levelName, nil
}

// findPolicyInNamespace scan for inference policy in a given namespace
// return:
//
//	the policy itself
//	the "level name" which goes to the stack, composed of the resource type (policy/workload) and its name
//	error in case something happened
func (self *InferenceWorkloadMergeCreator) findPolicyInNamespace(ctx context.Context, namespace string) (*tpv1.InferencePolicy, string, error) {
	log := logf.FromContext(ctx)
	if namespace == "" {
		return nil, "", nil
	}

	var listResult tpv1.InferencePolicyList

	err := self.k8sClient.List(ctx, &listResult, client.InNamespace(namespace))
	if err != nil {
		log.Error(err, "failed to list policies", "ns", namespace)
	}

	for _, item := range listResult.Items {
		levelName := workloadproject.NamespacedName(item.ObjectMeta)
		if policyUtils.IsBackendSyncedPolicy(item.ObjectMeta.Labels) {
			return &item, levelName, nil
		}
	}

	return nil, "", nil
}

func (self *InferenceWorkloadMergeCreator) handleBaseInferenceWorkload(ctx context.Context, inference *twv2.InferenceWorkload, mergeStack *enforcer.MergeStack) error {
	//
	//    base workload cannot be our own workload...
	//
	if ptr.Deref(inference.Spec.BaseWorkload, "") == inference.Name {
		return fmt.Errorf("base workload cannot be the same workload")
	}

	//
	//   find the base workload
	//
	workloadNamespacedName := client.ObjectKey{
		Namespace: inference.Namespace,
		Name:      *inference.Spec.BaseWorkload,
	}

	var baseWorkload twv2.InferenceWorkload

	//   try user workload
	err := self.k8sClient.Get(ctx, workloadNamespacedName, &baseWorkload)
	if err != nil {
		//   ...if not found, try to search it as a global workload
		workloadNamespacedName.Namespace = constants.NamespaceRunAi
		err = self.k8sClient.Get(ctx, workloadNamespacedName, &baseWorkload)
	}
	if err != nil {
		return fmt.Errorf("failed to find base workload %s", workloadNamespacedName.Name)
	}

	//
	//   the only thing we do not want to take from the base workload is its name
	//   it will only cause duplication in names
	//
	baseWorkload.Spec.Name = nil

	baseWorkloadSpec, baseWorkloadError := utils.NewJsonQ(baseWorkload.Spec)
	if baseWorkloadError != nil {
		return fmt.Errorf("failed to parse  %s.Spec", workloadNamespacedName.Name)

	}
	mergeStack.AddLevel(baseWorkloadSpec, enforcer.LevelWorkload, workloadproject.NamespacedName(baseWorkload.ObjectMeta))
	return nil
}
