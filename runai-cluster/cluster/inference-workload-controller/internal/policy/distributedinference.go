/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/nodepools"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	policyUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/utils/policy"

	"github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dinfp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferencepolicy/v1alpha1"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	log2 "sigs.k8s.io/controller-runtime/pkg/log"
)

type DistributedInferenceWorkloadMergeCreator struct {
	k8sClient               client.Client
	useControlPlanePolicies bool
}

func NewDistributedInferenceWorkloadMergeCreator(k8sClient client.Client, useControlPlanePolicies bool) *DistributedInferenceWorkloadMergeCreator {
	return &DistributedInferenceWorkloadMergeCreator{
		k8sClient:               k8sClient,
		useControlPlanePolicies: useControlPlanePolicies,
	}
}

// CreateMergeStack loads the workload spec, leaderSpec, and workerSpec, the base workload and the policy to merge stacks.
// once built, we can use the stack for merging and enforcing the complete set
// of values against the complete set of policy rules
func (diwmc *DistributedInferenceWorkloadMergeCreator) CreateMergeStack(ctx context.Context, workloadObject client.Object) (map[api.SubresourceComponent]*enforcer.MergeStack, error) {
	var mergeStackMap = make(map[api.SubresourceComponent]*enforcer.MergeStack, 3)
	distributed, ok := workloadObject.(*dinfw.DistributedInferenceWorkload)
	if !ok {
		return nil, fmt.Errorf("cannot cast object to DistributedInferenceWorkload")
	}

	specMergeStack, leaderMergeStack, workerMergeStack, err := diwmc.CreateDistributedInferenceMergeStack(ctx, distributed, diwmc.k8sClient)
	if err != nil {
		return nil, err
	}
	if specMergeStack == nil {
		// this should never happen
		return nil, fmt.Errorf("spec merge stack cannot be nil")
	}
	mergeStackMap[api.SpecSubResource] = specMergeStack
	if leaderMergeStack != nil {
		mergeStackMap[api.LeaderSpecSubResource] = leaderMergeStack
	}
	if workerMergeStack != nil {
		mergeStackMap[api.WorkerSpecSubResource] = workerMergeStack
	}

	return mergeStackMap, nil
}

func (diwmc *DistributedInferenceWorkloadMergeCreator) CreateDistributedInferenceMergeStack(
	ctx context.Context, r *dinfw.DistributedInferenceWorkload, kubeClient client.Client,
) (*enforcer.MergeStack, *enforcer.MergeStack, *enforcer.MergeStack, error) {
	// three merge stacks are created with the following levels:
	// * spec merge stack:
	//   0. workload (spec)
	//   1. base workload
	//   2. policy
	//
	// * leader merge stack:
	//   0. workload (leader spec)
	//   1. base workload
	//   2. leader policy
	//
	// * worker merge stack: (optional)
	//   0. workload (worker spec, if given)
	//   1. base workload
	//   2. worker policy

	logDist := log2.FromContext(ctx)
	id := workloadproject.NamespacedName(r.ObjectMeta)

	// spec
	workloadSpec, err := utils.NewJsonQ(r.Spec)
	if err != nil {
		logDist.Error(err, "could not load spec as JSON", "id", id)
		return nil, nil, nil, fmt.Errorf("failed to parse %s.Spec", id)
	}

	// leader spec
	leaderSpec, err := utils.NewJsonQ(r.LeaderSpec)
	if err != nil {
		logDist.Error(err, "could not load leaderSpec as JSON", "id", id)
		return nil, nil, nil, fmt.Errorf("failed to parse %s.LeaderSpec", id)
	}

	// worker spec (optional)
	var workerSpec *gojsonq.JSONQ
	if r.WorkerSpec != nil {
		workerSpec, err = utils.NewJsonQ(r.WorkerSpec)
		if err != nil {
			logDist.Error(err, "could not load workerSpec as JSON", "id", id)
			return nil, nil, nil, fmt.Errorf("failed to parse %s.WorkerSpec", id)
		}
	}

	// policy
	policy, policyLevelName, err := diwmc.GetPolicy(ctx, r.Namespace, r, kubeClient)
	if err != nil {
		logDist.Error(err, "could not fetch policy")
	}

	policySpec, err := utils.NewJsonQ(policy.Spec)
	if err != nil {
		logDist.Error(err, "could not load spec of policy as JSON")
		return nil, nil, nil, fmt.Errorf("failed to parse policy")
	}

	var policyLeaderSpec *gojsonq.JSONQ
	policyLeaderSpec, err = utils.NewJsonQ(policy.LeaderSpec)
	if err != nil {
		logDist.Error(err, "could not load leaderSpec of policy as JSON")
		return nil, nil, nil, fmt.Errorf("failed to parse policy")
	}

	var policyWorkerSpec *gojsonq.JSONQ
	if policy.WorkerSpec != nil {
		policyWorkerSpec, err = utils.NewJsonQ(policy.WorkerSpec)
		if err != nil {
			logDist.Error(err, "could not load workerSpec of policy as JSON")
			return nil, nil, nil, fmt.Errorf("failed to parse policy")
		}
	}

	// runai rules
	runaiRules := dinfp.GetRunaiRules(ctx, nodepools.NewNameLister(kubeClient))

	runaiRulesQ, err := utils.NewJsonQ(runaiRules)
	if err != nil {
		return nil, nil, nil, err
	}

	// base workload
	baseWorkload, err := GetBaseWorkload(ctx, r, kubeClient)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("base workload does not exist")
	}

	var baseWorkloadSpec *gojsonq.JSONQ
	var baseWorkloadLeaderSpec *gojsonq.JSONQ
	var baseWorkloadWorkerSpec *gojsonq.JSONQ
	if baseWorkload != nil {
		baseWorkloadSpec, err = utils.NewJsonQ(baseWorkload.Spec)
		if err != nil {
			logDist.Error(err, "could not load spec of base workload as JSON",
				"name", baseWorkload.Name, "namespace", baseWorkload.Namespace)
			return nil, nil, nil, fmt.Errorf("failed to parse spec from base workload %s", baseWorkload.Name)
		}

		baseWorkloadLeaderSpec, err = utils.NewJsonQ(baseWorkload.LeaderSpec)
		if err != nil {
			logDist.Error(err, "could not load leader spec of base workload as JSON",
				"name", baseWorkload.Name, "namespace", baseWorkload.Namespace)
			return nil, nil, nil, fmt.Errorf("failed to parse leader spec from base workload %s", baseWorkload.Name)
		}

		if baseWorkload.WorkerSpec != nil {
			baseWorkloadWorkerSpec, err = utils.NewJsonQ(baseWorkload.WorkerSpec)
			if err != nil {
				logDist.Error(err, "could not load worker spec of base workload as JSON",
					"name", baseWorkload.Name, "namespace", baseWorkload.Namespace)
				return nil, nil, nil, fmt.Errorf("failed to parse worker spec from base workload %s", baseWorkload.Name)
			}
		}
	}

	// finally, build the merge stacks

	// spec merge stack
	specMergeStack := enforcer.NewMergeStack(workloadSpec, workloadproject.NamespacedName(r.ObjectMeta), runaiRulesQ.JumpToTop(), false)
	if baseWorkload != nil {
		specMergeStack.AddLevel(baseWorkloadSpec, enforcer.LevelWorkload, workloadproject.NamespacedName(baseWorkload.ObjectMeta))
	}
	specMergeStack.AddLevel(policySpec, enforcer.LevelPolicy, policyLevelName)

	// leader spec merge stack
	leaderMergeStack := enforcer.NewMergeStack(leaderSpec, workloadproject.NamespacedName(r.ObjectMeta), runaiRulesQ.JumpToTop(), false)
	if baseWorkloadLeaderSpec != nil {
		leaderMergeStack.AddLevel(baseWorkloadLeaderSpec, enforcer.LevelWorkload, workloadproject.NamespacedName(baseWorkload.ObjectMeta))
	}
	leaderMergeStack.AddLevel(policyLeaderSpec, enforcer.LevelPolicy, policyLevelName)

	// worker spec merge stack (optional)
	var workerMergeStack *enforcer.MergeStack
	if r.WorkerSpec != nil {
		workerMergeStack = enforcer.NewMergeStack(workerSpec, workloadproject.NamespacedName(r.ObjectMeta), runaiRulesQ.JumpToTop(), false)
		if baseWorkloadWorkerSpec != nil {
			workerMergeStack.AddLevel(baseWorkloadWorkerSpec, enforcer.LevelWorkload, workloadproject.NamespacedName(baseWorkload.ObjectMeta))
		}
		if policyWorkerSpec != nil {
			workerMergeStack.AddLevel(policyWorkerSpec, enforcer.LevelPolicy, policyLevelName)
		}
	}

	return specMergeStack, leaderMergeStack, workerMergeStack, nil
}

func (diwmc *DistributedInferenceWorkloadMergeCreator) GetPolicy(
	ctx context.Context, namespace string, workload *dinfw.DistributedInferenceWorkload, k8sClient client.Client,
) (*dinfp.DistributedInferencePolicy, string, error) {
	log := log2.FromContext(ctx)
	var levelName string
	policy := dinfp.NewEmptyDistributedInferencePolicy()
	if workload.Namespace != "" {
		if utils.IsControlPlanePoliciesEnabled() && diwmc.useControlPlanePolicies {
			log.V(4).Info("new policy sync is enabled")
			policyObject, foundLevel, err := policyUtils.FindBackendSyncedPolicy(ctx, k8sClient, workload.Namespace, &workload.ObjectMeta.Labels, policyUtils.KindDistributedInferencePolicy)
			if err != nil {
				return nil, "", err
			}
			if policyObject != nil {
				policy = policyObject.(*dinfp.DistributedInferencePolicy)
			}

			levelName = string(foundLevel) + ":" + workload.ObjectMeta.Name
		} else {
			log.V(4).Info("new policy sync is disabled")
			var err error
			policy, levelName, err = getOldDistributedInferencePolicy(ctx, namespace, k8sClient)
			if err != nil {
				return nil, "", err
			}
		}
	}
	return policy, levelName, nil
}

func getOldDistributedInferencePolicy(ctx context.Context, namespace string, k8sClient client.Client) (*dinfp.DistributedInferencePolicy, string, error) {
	log := log2.FromContext(ctx)
	policy, policyLevelName, err := FindPolicyInNamespace(ctx, namespace, k8sClient)

	if err != nil {
		log.Error(err, "failed to find distributed inference policy", "ns", namespace)
		return nil, "", err
	}

	if policy == nil {
		policy, policyLevelName, err = FindPolicyInNamespace(ctx, constants.NamespaceRunAi, k8sClient)
		if err != nil {
			log.Error(err, "failed to find distributed inference policy", "ns", namespace)
			return nil, "", err
		}
	}

	if policy == nil {
		policy = dinfp.NewEmptyDistributedInferencePolicy()
		policyLevelName = ""
	}

	return policy, policyLevelName, err
}

// FindPolicyInNamespace is a helper function for GetPolicy, that lists the policies in the given namespace and returns
// the relevant one, if found
func FindPolicyInNamespace(ctx context.Context, namespace string, k8sClient client.Client) (*dinfp.DistributedInferencePolicy, string, error) {
	log := log2.FromContext(ctx)
	if namespace == "" {
		return nil, "", nil
	}

	var listResult dinfp.DistributedInferencePolicyList

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

func GetBaseWorkload(ctx context.Context, r *dinfw.DistributedInferenceWorkload, k8sClient client.Client) (*dinfw.DistributedInferenceWorkload, error) {
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

	var baseWorkload dinfw.DistributedInferenceWorkload

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

	return &baseWorkload, nil
}
