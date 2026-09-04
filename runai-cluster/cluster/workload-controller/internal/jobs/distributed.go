/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jobs

import (
	"fmt"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	clusterCommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	podPhase "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils/pod"
	mpiv2beta1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/jobs"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DistributedJobAdapter struct {
	client.Object
	jobKind string
	spec    *DistributedSpec
	status  *DistributedStatus
}

func (p *DistributedJobAdapter) GetOriginalObject() client.Object {
	return p.Object
}

func (p *DistributedJobAdapter) GetJobKind() string {
	return p.jobKind
}

func (p *DistributedJobAdapter) GetSpec() Spec {
	return p.spec
}

func (p *DistributedJobAdapter) GetStatus() JobStatus {
	return p.status
}

func (p *DistributedJobAdapter) GetConcurrentJob() client.Object {
	return p.Object
}

type DistributedSpec struct {
	replicaTypeSpec map[constants.ReplicaType]ReplicaTypeSpec
}

func (d *DistributedSpec) GetReplicaTypeSpecs() map[constants.ReplicaType]ReplicaTypeSpec {
	return d.replicaTypeSpec
}

func (d *DistributedSpec) GetReplicaSpec(replicaType constants.ReplicaType) ReplicaTypeSpec {
	return d.replicaTypeSpec[replicaType]
}

func NewDistributedSpec(replicaSpecs map[constants.ReplicaType]ReplicaTypeSpec) *DistributedSpec {
	return &DistributedSpec{replicaTypeSpec: replicaSpecs}
}

type DistributedReplicaSpec struct {
	internalReplicaTypeSpec *kubeflowv1.ReplicaSpec
}

func (d *DistributedReplicaSpec) GetPodTemplateSpec() *corev1.PodTemplateSpec {
	return &d.internalReplicaTypeSpec.Template
}

func (d *DistributedReplicaSpec) GetReplicas() *int32 {
	return d.internalReplicaTypeSpec.Replicas
}

func NewDistributedReplicaTypeSpec() ReplicaTypeSpec {
	return &DistributedReplicaSpec{}
}

type DistributedJobAdapterFactory struct {
	k8sClient client.Client
}

func (d *DistributedJobAdapterFactory) CreateJobAdapter(fromObject client.Object) (JobAdapter, error) {
	switch castObject := fromObject.(type) {
	case *kubeflowv1.PyTorchJob:
		return NewPytorchJobAdapter(castObject, d.k8sClient)
	case *kubeflowv1.TFJob:
		return NewTFJobAdapter(castObject, d.k8sClient)
	case *kubeflowv1.XGBoostJob:
		return NewXGBoostJobAdapter(castObject, d.k8sClient)
	case *mpiv2beta1.MPIJob:
		return NewMPIJobAdapter(castObject, d.k8sClient)
	case *kubeflowv1.JAXJob:
		return NewJAXJobAdapter(castObject, d.k8sClient)
	default:
		// should never happen
		return nil, fmt.Errorf("unknown job type %v", fromObject)

	}
}

func NewPytorchJobAdapter(pytorchJob *kubeflowv1.PyTorchJob, k8sClient client.Client) (*DistributedJobAdapter, error) {
	replicaTypeSpecMap, err := createDistributedReplicaSpec(pytorchJob.Spec.PyTorchReplicaSpecs)
	if err != nil {
		return nil, err
	}
	spec := NewDistributedSpec(replicaTypeSpecMap)
	status := NewDistributedStatusFromJobStatus(pytorchJob.Status, pytorchJob, k8sClient, &podPhase.RunaiPhaseCalculator{}, getTrainingJobManualPodLabelSelector)
	return &DistributedJobAdapter{Object: pytorchJob, jobKind: clusterCommon.KindPyTorchJob, spec: spec, status: status}, nil
}

func NewJAXJobAdapter(jaxJob *kubeflowv1.JAXJob, k8sClient client.Client) (*DistributedJobAdapter, error) {
	replicaTypeSpecMap, err := createDistributedReplicaSpec(jaxJob.Spec.JAXReplicaSpecs)
	if err != nil {
		return nil, err
	}
	spec := NewDistributedSpec(replicaTypeSpecMap)
	status := NewDistributedStatusFromJobStatus(jaxJob.Status, jaxJob, k8sClient, &podPhase.RunaiPhaseCalculator{}, getTrainingJobManualPodLabelSelector)
	return &DistributedJobAdapter{Object: jaxJob, jobKind: clusterCommon.KindJAXJob, spec: spec, status: status}, nil
}

func NewTFJobAdapter(tfJob *kubeflowv1.TFJob, k8sClient client.Client) (*DistributedJobAdapter, error) {
	replicaTypeSpecMap, err := createDistributedReplicaSpec(tfJob.Spec.TFReplicaSpecs)
	if err != nil {
		return nil, err
	}
	spec := NewDistributedSpec(replicaTypeSpecMap)
	status := NewDistributedStatusFromJobStatus(tfJob.Status, tfJob, k8sClient, &podPhase.RunaiPhaseCalculator{}, getTrainingJobManualPodLabelSelector)
	return &DistributedJobAdapter{Object: tfJob, jobKind: clusterCommon.KindTFJob, spec: spec, status: status}, nil
}

func NewXGBoostJobAdapter(xgboostJob *kubeflowv1.XGBoostJob, k8sClient client.Client) (*DistributedJobAdapter, error) {
	replicaTypeSpecMap, err := createDistributedReplicaSpec(xgboostJob.Spec.XGBReplicaSpecs)
	if err != nil {
		return nil, err
	}
	spec := NewDistributedSpec(replicaTypeSpecMap)
	status := NewDistributedStatusFromJobStatus(xgboostJob.Status, xgboostJob, k8sClient, &podPhase.RunaiPhaseCalculator{}, getTrainingJobManualPodLabelSelector)
	return &DistributedJobAdapter{Object: xgboostJob, jobKind: clusterCommon.KindXGBoostJob, spec: spec, status: status}, nil
}

func NewMPIJobAdapter(mpiJob *mpiv2beta1.MPIJob, k8sClient client.Client) (*DistributedJobAdapter, error) {
	replicaTypeSpecMap := make(map[constants.ReplicaType]ReplicaTypeSpec)
	for kubeReplicaType, replicaSpec := range mpiJob.Spec.MPIReplicaSpecs {
		replicaType, err := constants.ReplicaTypeFromString(string(kubeReplicaType))
		if err != nil {
			return nil, err
		}
		replicaTypeSpecMap[replicaType] = &DistributedReplicaSpec{internalReplicaTypeSpec: replicaSpec}
	}

	status := NewDistributedStatusFromJobStatus(mpiJob.Status, mpiJob, k8sClient, &podPhase.RunaiPhaseCalculator{}, getMpiManualPodLabelSelector)
	spec := NewDistributedSpec(replicaTypeSpecMap)
	return &DistributedJobAdapter{Object: mpiJob, jobKind: clusterCommon.KindMpiJob, spec: spec, status: status}, nil
}

func createDistributedReplicaSpec(kubeReplicaSpecs map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec) (map[constants.ReplicaType]ReplicaTypeSpec, error) {
	replicaTypeSpecMap := make(map[constants.ReplicaType]ReplicaTypeSpec)
	for kubeReplicaType, replicaSpec := range kubeReplicaSpecs {
		replicaType, err := constants.ReplicaTypeFromString(string(kubeReplicaType))
		if err != nil {
			return nil, err
		}
		replicaTypeSpecMap[replicaType] = &DistributedReplicaSpec{internalReplicaTypeSpec: replicaSpec}
	}
	return replicaTypeSpecMap, nil
}
