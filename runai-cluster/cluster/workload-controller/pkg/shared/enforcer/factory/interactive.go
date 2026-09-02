/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package factory

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/nodepools"
	workloadproject "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	ipv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	iwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	policyUtils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/shared/utils/policy"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/shared/utils/runai_rules"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type InteractiveWorkloadMergeCreator struct {
	k8sClient               client.Client
	useControlPlanePolicies bool
}

func NewInteractiveWorkloadMergeCreator(k8sClient client.Client, useControlPlanePolicies bool) *InteractiveWorkloadMergeCreator {
	return &InteractiveWorkloadMergeCreator{
		k8sClient:               k8sClient,
		useControlPlanePolicies: useControlPlanePolicies,
	}
}

func (self *InteractiveWorkloadMergeCreator) CreateMergeStack(ctx context.Context, workloadObject client.Object) (map[api.SubresourceComponent]*enforcer.MergeStack, error) {
	var mergeStackMap = make(map[api.SubresourceComponent]*enforcer.MergeStack)
	log := logf.FromContext(ctx)
	// case base object to interactive
	interactive, ok := workloadObject.(*iwv2.InteractiveWorkload)
	if !ok {
		return nil, fmt.Errorf("cannot cast object to InteractiveWorkload")
	}

	id := workloadproject.NamespacedName(interactive.ObjectMeta)

	//
	//    load the spec as json
	//
	workloadSpec, err := utils.NewJsonQ(interactive.Spec)
	if err != nil {
		log.Error(err, "failed to load spec as JSON", "id", id)
		return nil, fmt.Errorf("failed to parse %s.Spec", id)
	}

	runaiRules := ipv1.GetRunaiRules(ctx, nodepools.NewNameLister(self.k8sClient))

	runaiRulesQ, err := utils.NewJsonQ(runaiRules)
	if err != nil {
		return nil, err
	}

	err = enforcer.ValidateSyntaxOfWorkload("", workloadSpec)
	if err != nil {
		return nil, err
	}

	mergeStack := enforcer.NewMergeStack(workloadSpec, workloadproject.NamespacedName(interactive.ObjectMeta), runaiRulesQ.JumpToTop(), interactive.Spec.Usage == common.UsedForSubmit)

	//
	//    if the workload inherits from another workload...
	//
	if ptr.Deref(interactive.Spec.BaseWorkload, "") != "" {
		err = self.handleBaseWorkload(ctx, interactive, log, mergeStack)
		if err != nil {
			return nil, err
		}
	}
	//
	// get the policy
	//
	policy, levelName, err := self.GetPolicy(ctx, interactive)
	if err != nil {
		return mergeStackMap, err
	}
	runai_rules.PatchPolicyWithRunaiDefaults(ctx, &runaiRules.ContainerFields, &policy.Spec.ContainerFields,
		&runaiRules.AccessControlFields, &policy.Spec.AccessControlFields, &policy.Spec.RunaiSchedulerFields)

	//
	//    load the policy as json and add it to the stack
	//
	policySpec, err := utils.NewJsonQ(policy.Spec)
	if err != nil {
		log.Error(err, "could not load the policy spec as JSON")
		return nil, fmt.Errorf("failed to parse the interactive policy")
	}

	mergeStack.AddLevel(policySpec, enforcer.LevelPolicy, levelName)
	mergeStackMap[api.SpecSubResource] = mergeStack
	return mergeStackMap, nil
}

func (self *InteractiveWorkloadMergeCreator) GetPolicy(ctx context.Context, interactive *iwv2.InteractiveWorkload) (*ipv1.InteractivePolicy, string, error) {
	log := logf.FromContext(ctx)
	var levelName string
	policy := ipv1.NewEmptyInteractivePolicy()
	if interactive.Namespace != "" {
		if utils.IsControlPlanePoliciesEnabled() && self.useControlPlanePolicies {
			log.V(4).Info("new policy sync is enabled")
			policyObject, foundLevel, err := policyUtils.FindBackendSyncedPolicy(ctx, self.k8sClient, interactive.Namespace, &interactive.ObjectMeta.Labels, policyUtils.KindInteractivePolicy)
			if err != nil {
				return nil, "", err
			}
			if policyObject != nil {
				policy = policyObject.(*ipv1.InteractivePolicy)
			}

			levelName = string(foundLevel) + ":" + interactive.ObjectMeta.Name
		} else {
			log.V(4).Info("new policy sync is disabled")
			var err error
			policy, levelName, err = self.getOldInteractivePolicy(ctx, interactive, log)
			if err != nil {
				return nil, "", err
			}
		}
	}
	return policy, levelName, nil
}
func (self *InteractiveWorkloadMergeCreator) findPolicyInNamespace(ctx context.Context, namespace string) (*ipv1.InteractivePolicy, string, error) {
	log := logf.FromContext(ctx)
	if namespace == "" {
		return nil, "", nil
	}

	var listResult ipv1.InteractivePolicyList

	err := self.k8sClient.List(ctx, &listResult, client.InNamespace(namespace))
	if err != nil {
		log.Error(err, "failed to list policies", "ns", namespace)
		return nil, "", err
	}

	for _, item := range listResult.Items {
		levelName := workloadproject.NamespacedName(item.ObjectMeta)
		if policyUtils.IsBackendSyncedPolicy(item.ObjectMeta.Labels) {
			return &item, levelName, nil
		}
	}
	return nil, "", nil
}
func (self *InteractiveWorkloadMergeCreator) getOldInteractivePolicy(ctx context.Context, interactive *iwv2.InteractiveWorkload, log logr.Logger) (*ipv1.InteractivePolicy, string, error) {
	//
	//   the topmost level in the stack is an (optional) policy. find the one relevant policy
	//   (it is premised to have only one of each kind). it is ok if there is no policy for this type.
	//   in this case only run:ai rules will be validated
	//
	policy, levelName, err := self.findPolicyInNamespace(ctx, interactive.Namespace)
	if err != nil {
		log.Error(err, "failed to find interactive policy", "ns", interactive.Namespace)
		return nil, "", err
	}
	// if there is no policy in the namespace, try to find one in the run:ai namespace
	if policy == nil {
		policy, levelName, err = self.findPolicyInNamespace(ctx, constants.NamespaceRunAi)
		if err != nil {
			log.Error(err, "failed to find for interactive", "ns", interactive.Namespace)
			return nil, "", err
		}
	}
	if policy == nil {
		// if there is no policy, create an empty one
		policy = ipv1.NewEmptyInteractivePolicy()
		levelName = ""
	}
	return policy, levelName, nil
}

func (self *InteractiveWorkloadMergeCreator) handleBaseWorkload(ctx context.Context, interactive *iwv2.InteractiveWorkload, log logr.Logger, mergeStack *enforcer.MergeStack) error {
	//
	//    base workload cannot be our own workload...
	//
	if ptr.Deref(interactive.Spec.BaseWorkload, "") == interactive.Name {
		return fmt.Errorf("base workload cannot be the same workload")
	}

	//
	//   find the base workload
	//
	workloadNamespacedName := client.ObjectKey{
		Namespace: interactive.Namespace,
		Name:      *interactive.Spec.BaseWorkload,
	}

	var baseWorkload iwv2.InteractiveWorkload
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

	baseWorkloadSpec, createJsonQError := utils.NewJsonQ(baseWorkload.Spec)
	if createJsonQError != nil {
		log.Error(createJsonQError, "could not load spec of source workload workload as JSON",
			"name", baseWorkload.Name, "namespace", baseWorkload.Namespace)
		return fmt.Errorf("failed to parse base workload %s", baseWorkload.Name)
	}
	mergeStack.AddLevel(baseWorkloadSpec, enforcer.LevelWorkload, workloadproject.NamespacedName(baseWorkload.ObjectMeta))
	return nil
}
