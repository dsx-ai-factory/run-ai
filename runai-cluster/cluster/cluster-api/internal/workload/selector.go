/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	infw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/api"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	kubeflowcommon "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Selector interface {
	CreateSelector() client.ListOption
	String() string
}

type FirstPodSelector struct {
	clusterUID string
	namespace  string
}

const (
	knativePodLabelSelector = "serving.knative.dev/revision"

	lwsWorkerIndexLabelKey = "leaderworkerset.sigs.k8s.io/worker-index"
	lwsLeaderWorkerIndex   = "0"
)

func NewFirstPodSelector(workload *Workload) *FirstPodSelector {
	selector := &FirstPodSelector{
		clusterUID: getWorkloadUID(workload),
		namespace:  workload.Namespace,
	}
	log.Debug().
		Str("clusterUID", selector.clusterUID).
		Str("namespace", selector.namespace).
		Msg("Created FirstPodSelector")

	return selector
}

func (f *FirstPodSelector) CreateSelector() client.ListOption {
	selector := labels.SelectorFromValidatedSet(map[string]string{
		constants.ClusterTopOwnerUidLabel: f.clusterUID,
	})
	listOptions := &client.ListOptions{
		Limit:         1,
		LabelSelector: selector,
		Namespace:     f.namespace,
	}

	log.Debug().
		Interface("selector", selector).
		Str("namespace", f.namespace).
		Msg("FirstPodSelector: Created list options")

	return listOptions
}

func (f *FirstPodSelector) String() string {
	return "FirstPodSelector"
}

func GetSelectorCreator(workload *Workload) Selector {
	if workload == nil {
		log.Warn().Msg("GetSelectorCreator: Workload is nil")
		return nil
	}

	jobType := workload.JobType
	log.Debug().
		Str("jobType", string(jobType)).
		Msg("GetSelectorCreator: Determining selector")

	var selector Selector

	switch jobType {
	case api.InferenceJob:
		selector = buildInferenceSelector(workload)
	case api.DistributedInferenceJob:
		selector = buildDistributedInferenceSelector(workload)
	case api.MPIJob:
		selector = NewKeyValueSelector(kubeflowcommon.JobRoleLabel,
			strings.ToLower(string(kubeflowcommon.MPIJobReplicaTypeLauncher)), workload)
	case api.PyTorchJob:
		selector = NewKeyValueSelector(kubeflowcommon.JobRoleLabel,
			strings.ToLower(string(kubeflowcommon.PyTorchJobReplicaTypeMaster)), workload)
	case api.XGBoostJob:
		selector = NewKeyValueSelector(kubeflowcommon.JobRoleLabel,
			strings.ToLower(string(kubeflowcommon.XGBoostJobReplicaTypeMaster)), workload)
	case api.TFJob:
		selector = NewKeyValueSelector(kubeflowcommon.JobRoleLabel,
			strings.ToLower(string(kubeflowcommon.TFJobReplicaTypeMaster)), workload)
	case api.JAXJob:
		selector = NewKeyValueSelector(kubeflowcommon.ReplicaIndexLabel,
			strings.ToLower(api.MasterWorkerIndex), workload)
	default:
		selector = NewFirstPodSelector(workload)
	}

	log.Debug().
		Str("selectorType", fmt.Sprintf("%T", selector)).
		Msg("GetSelectorCreator: Selected selector")

	return selector
}

func buildInferenceSelector(workload *Workload) Selector {
	latestReadyRevision, err := GetLatestReadyRevision(workload.Original)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching latest ready Knative revision")
	}

	if latestReadyRevision == "" {
		log.Debug().Msg("No latest ready revision found, falling back to FirstPodSelector")
		return NewFirstPodSelector(workload)
	}

	log.Debug().
		Str("latestReadyRevision", latestReadyRevision).
		Msg("Using KeyValueSelector with Knative revision label")

	return NewKeyValueSelector(knativePodLabelSelector, latestReadyRevision, workload)
}

func buildDistributedInferenceSelector(workload *Workload) Selector {
	log.Debug().
		Str("workloadName", workload.Name).
		Msg("Building distributed inference selector for leader pod")

	return NewKeyValueSelector(lwsWorkerIndexLabelKey, lwsLeaderWorkerIndex, workload)
}

func GetLatestReadyRevision(workload client.Object) (string, error) {
	if workload == nil {
		return "", fmt.Errorf("workload is nil")
	}

	infWorkload, ok := workload.(*infw.InferenceWorkload)
	if !ok {
		return "", fmt.Errorf("workload is not of type InferenceWorkload")
	}

	latestReadyRevision := ptr.Deref(infWorkload.Status.LatestReadyRevisionName, "")
	if latestReadyRevision == "" {
		return "", fmt.Errorf("latest ready revision is empty")
	}

	return latestReadyRevision, nil
}

func FallBackSelectorCreator(workload *Workload) Selector {
	if workload == nil {
		log.Warn().Msg("Workload is nil")
		return nil
	}

	jobType := workload.JobType
	log.Debug().
		Str("jobType", string(jobType)).
		Msg("Determining fallback selector")

	var selector Selector

	switch jobType {
	case api.InferenceJob:
		selector = NewFirstPodSelector(workload)
	case api.DistributedInferenceJob:
		log.Debug().Msg("No fallback for distributed inference - leader pod must exist")
		return nil
	case api.MPIJob:
		selector = NewIndexRolePodSelector(kubeflowcommon.ReplicaIndexLabel, api.MasterWorkerIndex,
			kubeflowcommon.JobRoleLabel, strings.ToLower(string(kubeflowcommon.MPIJobReplicaTypeWorker)), workload)
	case api.PyTorchJob:
		selector = NewIndexRolePodSelector(kubeflowcommon.ReplicaIndexLabel, api.MasterWorkerIndex,
			kubeflowcommon.ReplicaTypeLabel,
			strings.ToLower(string(kubeflowcommon.PyTorchJobReplicaTypeWorker)), workload)
	case api.XGBoostJob:
		selector = NewIndexRolePodSelector(kubeflowcommon.ReplicaIndexLabel, api.MasterWorkerIndex,
			kubeflowcommon.ReplicaTypeLabel,
			strings.ToLower(string(kubeflowcommon.XGBoostJobReplicaTypeWorker)), workload)
	case api.TFJob:
		selector = NewIndexRolePodSelector(kubeflowcommon.ReplicaIndexLabel, api.MasterWorkerIndex,
			kubeflowcommon.ReplicaTypeLabel,
			strings.ToLower(string(kubeflowcommon.TFJobReplicaTypeWorker)), workload)
	case api.JAXJob:
		log.Debug().Msg("No fallback needed for JAX jobs")

	default:
		log.Debug().Msg("No fallback needed for non-distributed jobs")
		return nil
	}

	log.Debug().
		Str("selectorType", fmt.Sprintf("%T", selector)).
		Msg("Selected fallback selector")

	return selector
}

type KeyValueSelector struct {
	LabelKey   string
	LabelValue string
	namespace  string
	clusterUID string
}

func NewKeyValueSelector(LabelKey, LabelValuesValues string, workload *Workload) *KeyValueSelector {
	selector := &KeyValueSelector{
		LabelKey:   LabelKey,
		LabelValue: LabelValuesValues,
		namespace:  workload.Namespace,
		clusterUID: getWorkloadUID(workload),
	}
	log.Debug().
		Str("LabelKey", selector.LabelKey).
		Str("LabelValue", selector.LabelValue).
		Str("namespace", selector.namespace).
		Str("clusterUID", selector.clusterUID).
		Msg("Created KeyValueSelector")

	return selector
}

func (d *KeyValueSelector) String() string {
	return fmt.Sprintf("KeyValueSelector: LabelKey=%s, LabelValue=%s", d.LabelKey, d.LabelValue)
}

func (d *KeyValueSelector) CreateSelector() client.ListOption {
	selector := labels.SelectorFromValidatedSet(map[string]string{
		d.LabelKey:                        d.LabelValue,
		constants.ClusterTopOwnerUidLabel: d.clusterUID,
	})
	listOptions := &client.ListOptions{
		LabelSelector: selector,
		Limit:         1,
		Namespace:     d.namespace,
	}

	return listOptions
}

type IndexRolePodSelector struct {
	indexKey   string
	indexValue string
	roleKey    string
	roleValue  string
	namespace  string
	clusterUID string
}

func NewIndexRolePodSelector(
	indexKey, indexValue, roleKey, roleValue string, workload *Workload,
) *IndexRolePodSelector {
	selector := &IndexRolePodSelector{
		indexKey:   indexKey,
		indexValue: indexValue,
		roleKey:    roleKey,
		roleValue:  roleValue,
		namespace:  workload.Namespace,
		clusterUID: workload.ClusterUUID.String(),
	}
	log.Debug().
		Str("indexKey", selector.indexKey).
		Str("indexValue", selector.indexValue).
		Str("roleKey", selector.roleKey).
		Str("roleValue", selector.roleValue).
		Str("namespace", selector.namespace).
		Str("clusterUID", selector.clusterUID).
		Msg("Created IndexRolePodSelector")

	return selector
}

func (i *IndexRolePodSelector) String() string {
	return fmt.Sprintf("IndexRolePodSelector: indexKey=%s, indexValue=%s, roleKey=%s, roleValue=%s",
		i.indexKey, i.indexValue, i.roleKey, i.roleValue)
}

func (i *IndexRolePodSelector) CreateSelector() client.ListOption {
	selector := labels.SelectorFromValidatedSet(map[string]string{
		i.indexKey:                        i.indexValue,
		i.roleKey:                         i.roleValue,
		constants.ClusterTopOwnerUidLabel: i.clusterUID,
	})
	listOptions := &client.ListOptions{
		LabelSelector: selector,
		Limit:         1,
		Namespace:     i.namespace,
	}

	return listOptions
}

func getWorkloadUID(workload *Workload) string {
	var uid string
	if workload.Type == api.External {
		uid = getExternalWorkloadUID(workload)
	} else {
		uid = workload.ClusterUUID.String()
	}

	log.Debug().
		Str("UID", uid).
		Msg("workload UID")

	return uid
}

func getExternalWorkloadUID(workload *Workload) string {
	external, ok := workload.Original.(*v1alpha1.ExternalWorkload)
	if !ok {
		return ""
	}

	return string(external.Spec.OriginalWorkloadUID)
}
