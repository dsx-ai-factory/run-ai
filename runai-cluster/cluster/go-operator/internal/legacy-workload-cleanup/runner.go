/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacycleanup

import (
	"context"
	"fmt"

	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"k8s.io/client-go/dynamic"
	authorizationclient "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const CleanupJobName = "runai-legacy-workload-cleanup"

type Runner struct {
	ManagementClient   client.Client
	WorkloadClient     client.Client
	WorkloadDynamic    dynamic.Interface
	ManagementReviewer permissionReviewer
	WorkloadReviewer   permissionReviewer
}

func NewRunner(managementClient, workloadClient client.Client, workloadDynamic dynamic.Interface,
	managementAuthorization, workloadAuthorization authorizationclient.AuthorizationV1Interface,
) *Runner {
	return &Runner{
		ManagementClient:   managementClient,
		WorkloadClient:     workloadClient,
		WorkloadDynamic:    workloadDynamic,
		ManagementReviewer: newSelfSubjectAccessReviewer(managementAuthorization),
		WorkloadReviewer:   newSelfSubjectAccessReviewer(workloadAuthorization),
	}
}

func (r *Runner) Run(ctx context.Context) error {
	config, err := loadLiveRunaiConfig(ctx, r.ManagementClient)
	if err != nil {
		return err
	}
	if err := validateCleanupRequest(config); err != nil {
		return err
	}
	state, err := classifyControllerState(config)
	if err != nil {
		return err
	}
	if state == controllersDisabled {
		return nil
	}

	servedWorkloads, err := discoverServedWorkloads(ctx, r.WorkloadDynamic)
	if err != nil {
		return err
	}
	if err := r.checkAllPermissions(ctx, servedWorkloads); err != nil {
		return err
	}
	if err := validateLegacyWebhookState(
		ctx, r.WorkloadClient, servedWorkloads, ownershipProtectionEnabled(config),
	); err != nil {
		return err
	}
	if err := r.prepareWorkloads(ctx, servedWorkloads); err != nil {
		return err
	}
	return commitControllerDisable(ctx, r.ManagementClient)
}

func validateCleanupRequest(config *runaiv1.RunaiConfig) error {
	sourceVersion := config.GetAnnotations()[SourceVersionAnnotation]
	if sourceVersion == "" {
		return fmt.Errorf("cleanup source version annotation is missing")
	}
	shouldRun, err := SourceRequiresCleanup(sourceVersion)
	if err != nil {
		return err
	}
	if !shouldRun {
		return fmt.Errorf("cleanup request source version %s does not require legacy cleanup", sourceVersion)
	}
	return nil
}

func loadLiveRunaiConfig(ctx context.Context, k8sClient client.Client) (*runaiv1.RunaiConfig, error) {
	config := &runaiv1.RunaiConfig{}
	key := client.ObjectKey{Namespace: runaiv1.DefaultNamespace, Name: runaiv1.DefaultName}
	if err := k8sClient.Get(ctx, key, config); err != nil {
		return nil, fmt.Errorf("get RunaiConfig %s: %w", key, err)
	}
	return config, nil
}

func (r *Runner) prepareWorkloads(ctx context.Context, servedWorkloads []resourceDescriptor) error {
	for _, descriptor := range servedWorkloads {
		if !descriptor.RemoveWorkloadFinalizer {
			continue
		}
		objects, err := listObjects(ctx, r.WorkloadDynamic, descriptor)
		if err != nil {
			return err
		}
		for i := range objects {
			if err := prepareWorkload(ctx, r.WorkloadDynamic, descriptor, &objects[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Runner) checkAllPermissions(ctx context.Context, servedWorkloads []resourceDescriptor) error {
	management := []requiredPermission{
		{Verb: "create", Group: "authorization.k8s.io", Resource: "selfsubjectaccessreviews"},
		{Verb: "get", Group: "run.ai", Resource: "runaiconfigs", Namespace: runaiv1.DefaultNamespace, Name: runaiv1.DefaultName},
		{Verb: "update", Group: "run.ai", Resource: "runaiconfigs", Namespace: runaiv1.DefaultNamespace, Name: runaiv1.DefaultName},
	}
	if err := checkPermissions(ctx, r.ManagementReviewer, management); err != nil {
		return err
	}

	workload := []requiredPermission{
		{Verb: "create", Group: "authorization.k8s.io", Resource: "selfsubjectaccessreviews"},
		{Verb: "list", Group: "admissionregistration.k8s.io", Resource: "validatingwebhookconfigurations"},
		{Verb: "list", Group: "admissionregistration.k8s.io", Resource: "mutatingwebhookconfigurations"},
	}
	for _, descriptor := range legacyWorkloadResources {
		workload = append(workload, requiredPermission{
			Verb: "get", Group: "apiextensions.k8s.io", Resource: "customresourcedefinitions", Name: descriptor.CRDName,
		})
	}
	for _, descriptor := range servedWorkloads {
		if !descriptor.RemoveWorkloadFinalizer {
			continue
		}
		for _, verb := range []string{"list", "get", "patch"} {
			workload = append(workload, requiredPermission{
				Verb: verb, Group: descriptor.GVR.Group, Resource: descriptor.GVR.Resource,
			})
		}
	}
	return checkPermissions(ctx, r.WorkloadReviewer, workload)
}
