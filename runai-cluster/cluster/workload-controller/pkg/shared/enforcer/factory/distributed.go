package factory

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/nodepools"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	policyUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/utils/policy"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/utils/runai_rules"

	"github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	dwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	log2 "sigs.k8s.io/controller-runtime/pkg/log"
)

type DistributedWorkloadMergeCreator struct {
	k8sClient               client.Client
	useControlPlanePolicies bool
}

func NewDistributedWorkloadMergeCreator(k8sClient client.Client, useControlPlanePolicies bool) *DistributedWorkloadMergeCreator {
	return &DistributedWorkloadMergeCreator{
		k8sClient:               k8sClient,
		useControlPlanePolicies: useControlPlanePolicies,
	}
}

// CreateMergeStack  loads the workload spec and masterSpec, the base workload and the policy to merge stacks.
// once built, we can use the stack for merging and enforcing the complete set
// of values against the complete set of policy rules
// this is wrapper to the dwv2.CreateDistributedMergeStack function in the distributed workload
// mainly for testing purposes and facilitating the creation of the merge stack with interface
// TODO: refactor this function after there is generic function for creating merge stack for all workloads
func (dwmc *DistributedWorkloadMergeCreator) CreateMergeStack(ctx context.Context, workloadObject client.Object) (map[api.SubresourceComponent]*enforcer.MergeStack, error) {

	var mergeStackMap = make(map[api.SubresourceComponent]*enforcer.MergeStack, 2)
	distributed, ok := workloadObject.(*dwv2.DistributedWorkload)
	if !ok {
		return nil, fmt.Errorf("cannot cast object to DistributedWorkload")
	}

	workerMergeStack, masterMergeStack, err := dwmc.CreateDistributedMergeStack(ctx, distributed, dwmc.k8sClient)
	if err != nil {
		return nil, err
	}
	if workerMergeStack == nil {
		// this should never happen
		return nil, fmt.Errorf("worker merge stack cannot be nil")
	}
	mergeStackMap[api.WorkerSpecSubResource] = workerMergeStack
	if masterMergeStack != nil {
		mergeStackMap[api.MasterSpecSubResource] = masterMergeStack
	}

	return mergeStackMap, nil
}

func (self *DistributedWorkloadMergeCreator) CreateDistributedMergeStack(ctx context.Context, r *dwv2.DistributedWorkload, kubeClient client.Client) (*enforcer.MergeStack, *enforcer.MergeStack, error) {
	// two merge stacks are created with the following levels:
	// * worker merge stack:
	//   0. workload (spec)
	//   1. base workload
	//   2. policy
	//
	// * master merge stack: (optional - only if noMaster is nil/false)
	//   0. workload (master spec, if given, otherwise re-using workload spec)
	//   1. base workload
	//   2. master policy (if given, otherwise using workload policy)

	logDist := log2.FromContext(ctx)
	id := workloadproject.NamespacedName(r.ObjectMeta)

	// spec
	workloadSpec, err := utils.NewJsonQ(r.Spec)
	if err != nil {
		logDist.Error(err, "could not load spec as JSON", "id", id)
		return nil, nil, fmt.Errorf("failed to parse %s.Spec", id)
	}

	// master spec
	var masterSpec *gojsonq.JSONQ
	if r.MasterSpec != nil {
		masterSpec, err = utils.NewJsonQ(r.MasterSpec)
		if err != nil {
			logDist.Error(err, "could not load masterSpec as JSON", "id", id)
			return nil, nil, fmt.Errorf("failed to parse %s.MasterSpec", id)
		}
	} else if !r.Spec.NoMaster.Deref() {
		// this means we have a master, but no masterSpec is given. we should use a copy of spec for the master
		masterSpec, err = utils.NewJsonQ(r.Spec)
		if err != nil {
			logDist.Error(err, "could not load masterSpec from spec as JSON", "id", id)
			return nil, nil, fmt.Errorf("failed to parse %s.Spec", id)
		}

	}

	// runai rules
	runaiRules := v2alpha1.GetRunaiRules(ctx, nodepools.NewNameLister(kubeClient))

	runaiRulesQ, err := utils.NewJsonQ(runaiRules)
	if err != nil {
		return nil, nil, err
	}

	// policy
	policy, policyLevelName, err := self.GetPolicy(ctx, r.Namespace, r, kubeClient)
	if err != nil {
		logDist.Error(err, "could not fetch policy")
	}

	runai_rules.PatchPolicyWithRunaiDefaults(ctx, &runaiRules.ContainerFields, &policy.Spec.ContainerFields,
		&runaiRules.AccessControlFields, &policy.Spec.AccessControlFields, &policy.Spec.RunaiSchedulerFields)

	if policy.MasterSpec != nil {
		runai_rules.PatchPolicyWithRunaiDefaults(ctx, &runaiRules.ContainerFields, &policy.MasterSpec.ContainerFields,
			&runaiRules.AccessControlFields, &policy.MasterSpec.AccessControlFields, &policy.MasterSpec.RunaiSchedulerFields)
	}

	policySpec, err := utils.NewJsonQ(policy.Spec)
	if err != nil {
		logDist.Error(err, "could not load spec of policy as JSON")
		return nil, nil, fmt.Errorf("failed to parse policy")
	}

	policySpecCopy, err := utils.NewJsonQ(policy.Spec)
	if err != nil {
		logDist.Error(err, "could not load spec of policy as JSON")
		return nil, nil, fmt.Errorf("failed to parse policy")
	}

	var policyMasterSpec *gojsonq.JSONQ
	if policy.MasterSpec != nil {
		policyMasterSpec, err = utils.NewJsonQ(policy.MasterSpec)
		if err != nil {
			logDist.Error(err, "could not load masterSpec of policy as JSON")
			return nil, nil, fmt.Errorf("failed to parse policy")
		}
	}

	// base workload
	baseWorkload, err := GetBaseWorkload(ctx, r, kubeClient)
	if err != nil {
		return nil, nil, fmt.Errorf("base workload does not exist")
	}

	var baseWorkloadSpec *gojsonq.JSONQ
	var baseWorkloadMasterSpec *gojsonq.JSONQ
	if baseWorkload != nil {
		baseWorkloadSpec, err = utils.NewJsonQ(baseWorkload.Spec)
		if err != nil {
			logDist.Error(err, "could not load spec of base workload as JSON",
				"name", baseWorkload.Name, "namespace", baseWorkload.Namespace)
			return nil, nil, fmt.Errorf("failed to parse spec from base workload %s", baseWorkload.Name)
		}

		if baseWorkload.MasterSpec != nil {
			baseWorkloadMasterSpec, err = utils.NewJsonQ(baseWorkload.MasterSpec)
			if err != nil {
				logDist.Error(err, "could not load master spec of base workload as JSON",
					"name", baseWorkload.Name, "namespace", baseWorkload.Namespace)
				return nil, nil, fmt.Errorf("failed to parse master spec from base workload %s", baseWorkload.Name)
			}
		}
	}

	// finally, build the merge stacks

	// spec merge stack
	mergeStack := enforcer.NewMergeStack(workloadSpec, workloadproject.NamespacedName(r.ObjectMeta), runaiRulesQ.JumpToTop(), r.Spec.Usage == common.UsedForSubmit)
	if baseWorkload != nil {
		mergeStack.AddLevel(baseWorkloadSpec, enforcer.LevelWorkload, workloadproject.NamespacedName(baseWorkload.ObjectMeta))
	}
	mergeStack.AddLevel(policySpec, enforcer.LevelPolicy, policyLevelName)

	// master spec merge stack
	var masterMergeStack *enforcer.MergeStack
	if !r.Spec.NoMaster.Deref() {
		masterMergeStack = enforcer.NewMergeStack(masterSpec, workloadproject.NamespacedName(r.ObjectMeta), runaiRulesQ.JumpToTop(), r.Spec.Usage == common.UsedForSubmit)
		if baseWorkloadMasterSpec != nil {
			masterMergeStack.AddLevel(baseWorkloadMasterSpec, enforcer.LevelWorkload, workloadproject.NamespacedName(baseWorkload.ObjectMeta))
		}
		if policyMasterSpec != nil {
			masterMergeStack.AddLevel(policyMasterSpec, enforcer.LevelPolicy, policyLevelName)
		} else { // no master policy, re-using workload policy
			masterMergeStack.AddLevel(policySpecCopy, enforcer.LevelPolicy, policyLevelName)
		}
	}

	return mergeStack, masterMergeStack, nil
}

func (self *DistributedWorkloadMergeCreator) GetPolicy(ctx context.Context, namespace string, workload *dwv2.DistributedWorkload, k8sClient client.Client) (*v2alpha1.DistributedPolicy, string, error) {
	log := log2.FromContext(ctx)
	var levelName string
	policy := v2alpha1.NewEmptyDistributedPolicy()
	if workload.Namespace != "" {
		if utils.IsControlPlanePoliciesEnabled() && self.useControlPlanePolicies {
			log.V(4).Info("new policy sync is enabled")
			policyObject, foundLevel, err := policyUtils.FindBackendSyncedPolicy(ctx, k8sClient, workload.Namespace, &workload.ObjectMeta.Labels, policyUtils.KindDistributedPolicy)
			if err != nil {
				return nil, "", err
			}
			if policyObject != nil {
				policy = policyObject.(*v2alpha1.DistributedPolicy)
			}

			levelName = string(foundLevel) + ":" + workload.ObjectMeta.Name
		} else {
			log.V(4).Info("new policy sync is disabled")
			var err error
			policy, levelName, err = getOldDistributedPolicy(ctx, namespace, k8sClient)
			if err != nil {
				return nil, "", err
			}
		}
	}
	return policy, levelName, nil
}

func getOldDistributedPolicy(ctx context.Context, namespace string, k8sClient client.Client) (*v2alpha1.DistributedPolicy, string, error) {
	log := log2.FromContext(ctx)
	policy, policyLevelName, err := FindPolicyInNamespace(ctx, namespace, k8sClient)

	if err != nil {
		log.Error(err, "failed to find distributed policy", "ns", namespace)
		return nil, "", err
	}

	if policy == nil {
		policy, policyLevelName, err = FindPolicyInNamespace(ctx, constants.NamespaceRunAi, k8sClient)
		if err != nil {
			log.Error(err, "failed to find distributed policy", "ns", namespace)
			return nil, "", err
		}
	}

	if policy == nil {
		policy = v2alpha1.NewEmptyDistributedPolicy()
		policyLevelName = ""
	}

	return policy, policyLevelName, err
}

// FindPolicyInNamespace is a helper function for GetPolicy, that lists the policies in the given namespace and returns
// the relevant one, if found
func FindPolicyInNamespace(ctx context.Context, namespace string, k8sClient client.Client) (*v2alpha1.DistributedPolicy, string, error) {
	log := log2.FromContext(ctx)
	if namespace == "" {
		return nil, "", nil
	}

	var listResult v2alpha1.DistributedPolicyList

	err := k8sClient.List(ctx, &listResult, client.InNamespace(namespace))
	if err != nil {
		log.Error(err, "failed to list policies", "ns", namespace)
	}

	if len(listResult.Items) == 0 {
		return nil, "", nil
	}

	item := listResult.Items[0]
	levelName := workloadproject.NamespacedName(item.ObjectMeta)
	return &item, levelName, nil
}

func GetBaseWorkload(ctx context.Context, r *dwv2.DistributedWorkload, k8sClient client.Client) (*dwv2.DistributedWorkload, error) {
	log := log2.FromContext(ctx)
	//
	//    check if there is a base workload
	//
	if ptr.Deref(r.Spec.BaseWorkload, "") == "" {
		return nil, nil
	}

	//
	//    base workload must point to another workload
	//
	if ptr.Deref(r.Spec.BaseWorkload, "") == r.Name {
		return nil, fmt.Errorf("the name of the base workload must be different from the name of the workload itself (%s)", r.Name)
	}

	//
	//   find the base workload
	//
	workloadNamespacedName := client.ObjectKey{
		Namespace: r.Namespace,
		Name:      *r.Spec.BaseWorkload,
	}

	var baseWorkload dwv2.DistributedWorkload

	//   try in the project namespace
	err := k8sClient.Get(ctx, workloadNamespacedName, &baseWorkload)
	if err != nil {
		//   ...if not found, try to search it as a global template in runai namespace
		workloadNamespacedName.Namespace = constants.NamespaceRunAi
		err = k8sClient.Get(ctx, workloadNamespacedName, &baseWorkload)
	}
	if err != nil {
		log.Error(err, "could not find base workload", "name", workloadNamespacedName.Name)
		return nil, fmt.Errorf("base workload '%s' does not exist", workloadNamespacedName.Name)
	}

	//
	//   the only thing we do not want to take from the base workload is the resource name
	//   it generates. it will only cause duplication in names
	//
	baseWorkload.Spec.Name = nil

	return &baseWorkload, nil
}
