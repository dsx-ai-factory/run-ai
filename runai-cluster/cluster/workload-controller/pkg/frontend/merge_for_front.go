/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package frontend

import (
	"context"
	"fmt"

	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer/factory"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/webhook/validator/workloads"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type WebhookOptions struct {
	// this option is for both frontend and testing. It disables the builders actions and stops
	// once merge stack is ready
	SuppressBuilding bool
	// this option is only for frontend. it is used for loading a global template to a project
	TargetNamespace string
}

func CreateUnitedMergeStackFrontend(ctx context.Context, k8sClient client.Client, options WebhookOptions,
	workload *jobs.JobWorkload, mergeStackCreator factory.JobMergeCreator, originalWorkload client.Object) (*enforcer.MergeStack, error) {
	log := logf.FromContext(ctx)
	err := workloads.ValidateLabels(workload.ObjectMeta.Labels)
	if err != nil {
		return nil, err
	}

	err = workloads.ValidatePodAffinity(workload.Spec.RunaiSchedulerFields.PodAffinity, workload.Spec.RunaiSchedulerFields.PodAffinityTopology)

	if err != nil {
		return nil, err
	}

	//
	//    figure out the project the workload resides in
	//
	ctx, err = workloadproject.SetProjectInContextFromNamespace(ctx, k8sClient, workload.Namespace, workload.Name, workload.Spec.Usage)
	if err != nil {
		return nil, err
	}

	//
	//   only workloads with usage=Submit can "borrow" a set of values from another workload.
	//   the other workload can have any usage (Submit = previous job, Template = template), it does not matter,
	//
	if !options.SuppressBuilding && workload.Spec.Usage == common.UsedAsTemplate && ptr.Deref(workload.Spec.BaseWorkload, "") != "" {
		return nil, fmt.Errorf("baseWorkload is only allowed for workloads with usage=Submit")
	}

	//
	//    load the spec as json
	//
	workloadSpec, err := utils.NewJsonQ(workload.Spec)
	if err != nil {
		log.Error(err, "failed to load spec as JSON", "id", workloadproject.NamespacedName(workload.ObjectMeta))
		return nil, fmt.Errorf("failed to parse %s.Spec", workloadproject.NamespacedName(workload.ObjectMeta))
	}

	//
	//   perform syntactic validation (only things we cannot validate with open3api)
	//
	err = enforcer.ValidateSyntaxOfWorkload("", workloadSpec)
	if err != nil {
		return nil, err
	}

	//
	//   now we need to load the (optional) base workload and the policy and place
	//   everything as json in the merge stack
	//
	mergeStackMap, err := mergeStackCreator.CreateMergeStack(ctx, originalWorkload)
	if err != nil {
		return nil, err
	}

	mergeStack, ok := mergeStackMap[api.SpecSubResource]
	if !ok {
		return nil, fmt.Errorf("failed to find spec merge stack")
	}

	//
	//    now merge everything and validate against the admin policy and the runai rules
	//
	err = mergeStack.MergeAndValidate(ctx, enforcer.MergeAndValidateOptions{})
	if err != nil {
		return nil, err
	}

	//
	//   if the workload is used as template, the validation we did so far are sufficient. if it
	//   is for Submit, we have to make a "dry run" to make sure there are no issues from the
	//   admission controller
	//

	return mergeStack, nil
}
