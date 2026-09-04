/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jobs

import (
	"context"
	"reflect"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	enforcer2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"k8s.io/apimachinery/pkg/runtime"

	rjv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	mpiv2beta1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/factory"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logger.Log.WithName("jobs")

//	The job workload is a superset of distributed interactive and training fields.
//
// +k8s:deepcopy-gen=true
type JobWorkload struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec       JobSpec  `json:"spec,omitempty"`
	MasterSpec *JobSpec `json:"masterSpec,omitempty"`

	Status JobStatus `json:"status,omitempty"`
}

func (self *JobWorkload) DeepCopyObject() runtime.Object {
	return self.DeepCopy()
}

func (self *JobWorkload) IsMarkedForDeletion() bool {
	return !self.DeletionTimestamp.IsZero()
}

func (self *JobWorkload) IsCompleted() bool {
	return !self.Status.WorkloadCommonStatus.CompletionTime.IsZero()
}

func (self *JobWorkload) IsFirstTime() bool {
	return self.Status.FirstAcknowledgedTime.IsZero()
}

// IsFailedByWorkload returns true if the workload is failed due to the workload itself (not due to the failed job)
func (self *JobWorkload) IsFailedByWorkload() bool {
	for _, condition := range self.Status.Conditions {
		if condition.Type == common.Failed && (condition.Reason == common.WorkloadBuildFailedReason || condition.Reason == common.WorkloadFailedReason) {
			return true
		}
	}
	return false
}

// GetChildName returns the name of the workload's children resources (e.g. run-ai jobs, services)
func (w *JobWorkload) GetChildName() string {
	name := w.ObjectMeta.Name
	// The job name should be metadata.Name, however for backward compatibility old workload uses sepc.Name
	// Therefor need to check if spec.Name isn't empty and use it instead of metadata.Name
	if w.Spec.Name != nil && w.Spec.Name.Deref() != "" {
		name = w.Spec.Name.Deref()
	}
	return name
}

// JobSpec unifies interactive, training, distributed and inference fields
// +k8s:deepcopy-gen=true
type JobSpec struct {
	field_groups.WorkloadFields `json:""`

	field_groups.MpiFields                `json:""`
	field_groups.LifeCycleFields          `json:""`
	field_groups.ResourceAllocationFields `json:""`
	field_groups.RunaiSchedulerFields     `json:""`
	field_groups.SchedulingFields         `json:""`
	field_groups.RunaiJobSpecificRules    `json:""`
	field_groups.AccessControlFields      `json:""`
	field_groups.StateFields              `json:""`
	//   interactive specific
	field_groups.InteractiveFields `json:""`

	//   this one is only for distributed fields, it is used by CLI when building the workload
	//   it is not used by workload controller itself.
	field_groups.DistributedFields `json:""`

	// inference specific
	field_groups.InferenceFields `json:""`
}

// JobStatus unifies interactive and training status
// +k8s:deepcopy-gen=true
type JobStatus struct {
	common.WorkloadCommonStatus `json:""`
	common.WorkloadDebugStatus  `json:""`
	common.WorkloadStatus       `json:""`
	common.NetworkStatus        `json:",inline"`
	common.InferenceStatus      `json:""`
	MinRequiredReadyPods        int32 `json:"minRequiredReadyPods,omitempty"`

	// WorkloadSyncStatus defines the status of the sync, reported by cluster-sync
	SyncStatus *common.WorkloadSyncStatus `json:"syncStatus,omitempty"`
}

func (self *JobWorkload) IsStatusChanged(originalWorkload *JobWorkload) bool {
	// ignore the status last status change time
	if originalWorkload == nil {
		return true
	}
	originalTime := originalWorkload.Status.LastStatusChangeTime
	originalWorkload.Status.LastStatusChangeTime = self.Status.LastStatusChangeTime
	notChanged := reflect.DeepEqual(originalWorkload.Status, self.Status)
	originalWorkload.Status.LastStatusChangeTime = originalTime
	return !notChanged
}

// Succeeded unifies status setting to success for training and interactive
// functions to handle the workload status
func (self *JobWorkload) Succeeded(mergeStack *enforcer2.MergeStack) *JobWorkload {
	self.Status.CreationStatus = common.StatusSubmitted
	return self
}

// Builder return a builder based on the job
//
//	interactive/training -> mpi or runai builder depending on the mpi field
func (self *JobWorkload) Builder(options factory.BuilderOptions, k8sClient client.Client,
	podLabels map[string]string, podAnnotations map[string]string) factory.Builder {
	if self.Spec.Mpi.Deref() {
		builderConstructor := factory.BuildersRegistry[factory.MpiOldBuilder]
		if builderConstructor != nil {
			return builderConstructor(self, &mpiv2beta1.MPIJob{}, k8sClient, options, podLabels, podAnnotations)
		}
	} else {
		builderConstructor := factory.BuildersRegistry[factory.RunaiJobBuilder]
		if builderConstructor != nil {
			return builderConstructor(self, &rjv1.RunaiJob{}, k8sClient, options, podLabels, podAnnotations)
		}
	}

	log.Info("Failed to allocate builder, this can only happen during testing/debugging")
	return nil
}

type JobStackCreator interface {
	CreateMergeStack(ctx context.Context, k8s client.Client) (*enforcer2.MergeStack, error)
}

func IsElasticPytorchJob(job *JobWorkload) bool {
	if job == nil || job.Spec.JobType != constants.KindPyTorchJob {
		return false
	}

	pytorchFields := job.Spec.PyTorchJob.Deref()
	return pytorchFields.MinReplicas != nil || pytorchFields.MaxReplicas != nil
}

func IsUsingNodePools(spec *JobSpec) bool {
	if spec == nil {
		return false
	}
	return spec.NodePools.Deref() != "" || spec.NodePool.Deref() != ""
}
