/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

import (
	"context"
	"errors"
	"fmt"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	workloadbuilders "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	ComputeDomainCpuOnlyErr    = errors.New("ComputeDomain is not supported for CPU-only workloads")
	FailedToSetReferenceErrMsg = fmt.Sprintf("failed to set controller reference for %s", nvapi.ComputeDomainKind)
)

type ComputeDomainWorkloadParams struct {
	WorkloadKind      string
	JobType           string
	WorkloadInitials  string
	WorkloadName      string
	WorkloadNamespace string
	NumNodes          int
	ElasticEnabled    bool
}

func BuildComputeDomain(ctx context.Context, params ComputeDomainWorkloadParams, scheme *runtime.Scheme, owner client.Object) (*nvapi.ComputeDomain, error) {
	objectMeta, err := buildObjectMeta(ctx, params)
	if err != nil {
		return nil, err
	}

	spec, err := buildSpec(params)
	if err != nil {
		return nil, err
	}

	computeDomain := &nvapi.ComputeDomain{
		TypeMeta: metav1.TypeMeta{
			Kind:       nvapi.ComputeDomainKind,
			APIVersion: nvapi.Version,
		},
		ObjectMeta: *objectMeta,
		Spec:       *spec,
	}

	if owner != nil {
		if scheme == nil {
			return nil, fmt.Errorf("%s, scheme is nil", FailedToSetReferenceErrMsg)
		}

		err = controllerutil.SetControllerReference(owner, computeDomain, scheme)
		if err != nil {
			return nil, fmt.Errorf("%s, %v", FailedToSetReferenceErrMsg, err)
		}
	}

	return computeDomain, nil
}

func buildObjectMeta(ctx context.Context, params ComputeDomainWorkloadParams) (*metav1.ObjectMeta, error) {
	cdName := ComposeComputeDomainName(params.WorkloadInitials, params.WorkloadName)

	objectMeta := &metav1.ObjectMeta{
		Name:      cdName,
		Namespace: params.WorkloadNamespace,
	}

	err := workloadbuilders.SetObjectMetaLabels(ctx, params.WorkloadKind, params.JobType, params.WorkloadName, objectMeta)
	if err != nil {
		return nil, err
	}

	return objectMeta, nil
}

func buildSpec(params ComputeDomainWorkloadParams) (*nvapi.ComputeDomainSpec, error) {
	if params.NumNodes == 0 {
		return nil, ComputeDomainCpuOnlyErr
	}

	rct := CreateResourceClaimTemplate(params)

	// Setting the compute domain node number to 0, enable elastic compute domain
	cdNodes := 0
	if !params.ElasticEnabled {
		cdNodes = params.NumNodes
	}

	return &nvapi.ComputeDomainSpec{
		NumNodes: cdNodes, // elastic compute domain feature
		Channel: &nvapi.ComputeDomainChannelSpec{
			ResourceClaimTemplate: *rct,
		},
	}, nil
}

func PreviewComputeDomain(workloadInitials, workloadName, namespace string) *nvapi.ComputeDomain {
	computeDomainName := ComposeComputeDomainName(workloadInitials, workloadName)

	return &nvapi.ComputeDomain{
		TypeMeta: metav1.TypeMeta{
			Kind:       nvapi.ComputeDomainKind,
			APIVersion: nvapi.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      computeDomainName,
			Namespace: namespace,
		},
	}
}

func ComposeComputeDomainName(workloadInitials, workloadName string) string {
	return fmt.Sprintf("%s-%s", workloadInitials, workloadName)
}
