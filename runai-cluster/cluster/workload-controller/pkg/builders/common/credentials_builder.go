/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"

	common "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"

	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	job "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
)

type CredentialsBuilder struct {
	builder *workloadbuilders.CredentialsBuilder
}

func NewCredentialsBuilder(k8sClient client.Client) ResourceBuilder {
	return &CredentialsBuilder{
		builder: workloadbuilders.NewCredentialsBuilder(
			k8sClient,
			constants.NamespaceRunAi),
	}
}

func (b *CredentialsBuilder) GetPreview(ctx context.Context, workload *job.JobWorkload) ([]common.Resource, error) {
	info, err := mkWorkloadInfo(workload, nil)
	if err != nil {
		return nil, err
	}
	var envs []*flds.DictionaryField
	if workload.Spec.Environment != nil {
		envs = append(envs, workload.Spec.Environment)
	}
	if workload.MasterSpec != nil && workload.MasterSpec.Environment != nil {
		envs = append(envs, workload.MasterSpec.Environment)
	}

	var imagePullSecrets []*flds.StringField
	if workload.Spec.ImagePullSecrets != nil {
		imagePullSecrets = append(imagePullSecrets, workload.Spec.ImagePullSecrets)
	}
	if workload.MasterSpec != nil && workload.MasterSpec.ImagePullSecrets != nil {
		imagePullSecrets = append(imagePullSecrets, workload.MasterSpec.ImagePullSecrets)
	}

	resources, err := b.builder.Preview(ctx, info, envs, imagePullSecrets)
	if err != nil {
		return nil, err
	}

	result := common.CommonResourceToResource(resources)

	return result, nil
}

func (b *CredentialsBuilder) Build(ctx context.Context, workload *job.JobWorkload, options ResourceBuilderOptions) ([]client.Object, error) {
	info, err := mkWorkloadInfo(workload, &options.Owner)
	if err != nil {
		return nil, err
	}

	var envs []*flds.DictionaryField
	if workload.Spec.Environment != nil {
		envs = append(envs, workload.Spec.Environment)
	}
	if workload.MasterSpec != nil && workload.MasterSpec.Environment != nil {
		envs = append(envs, workload.MasterSpec.Environment)
	}

	var imagePullSecrets []*flds.StringField
	if workload.Spec.ImagePullSecrets != nil {
		imagePullSecrets = append(imagePullSecrets, workload.Spec.ImagePullSecrets)
	}
	if workload.MasterSpec != nil && workload.MasterSpec.ImagePullSecrets != nil {
		imagePullSecrets = append(imagePullSecrets, workload.MasterSpec.ImagePullSecrets)
	}

	secrets, err := b.builder.Build(ctx, info, envs, imagePullSecrets)
	if err != nil {
		return nil, err
	}
	results := make([]client.Object, 0, len(secrets))
	for _, secret := range secrets {
		results = append(results, &secret)
	}

	return results, nil
}

func mkWorkloadInfo(workload *job.JobWorkload, owner *client.Object) (*workloadbuilders.WorkloadInfo, error) {
	initials, err := GetWorkloadInitials(workload.Kind)
	if err != nil {
		return nil, err
	}

	info := &workloadbuilders.WorkloadInfo{
		Name:         workload.GetChildName(),
		Namespace:    workload.ObjectMeta.Namespace,
		Initials:     string(initials),
		WorkloadKind: workload.Kind,
		JobType:      GetJobType(workload),
	}

	if owner != nil {
		info.Owner = *owner
	}
	return info, nil
}
