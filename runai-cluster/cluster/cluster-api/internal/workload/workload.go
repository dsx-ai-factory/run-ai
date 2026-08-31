/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"errors"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	distinfwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"

	infwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"
	clustercommonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	workloadIsNilErrorMsg   = "workload is nil"
	invalidWorkloadErrorMsg = "invalid workload, expected %s workload but got different type"
	undefinedPhase          = "Undefined"
)

type UID string

func (u UID) String() string {
	return string(u)
}

func (u UID) IsEmpty() bool {
	isEmpty := u == ""
	log.Debug().Bool("isEmpty", isEmpty).Str("uid", u.String()).Msg("Checked if UID is empty")

	return isEmpty
}

func (u UID) ToK8sUID() types.UID {
	k8sUID := types.UID(u)
	log.Debug().Str("uid", u.String()).Str("k8sUID", string(k8sUID)).Msg("Converted UID to K8s UID")

	return k8sUID
}

func FromK8sUID(uid types.UID) UID {
	newUID := UID(uid)
	log.Debug().Str("k8sUID", string(uid)).Str("newUID", newUID.String()).Msg("Converted K8s UID to UID")

	return newUID
}

type Workload struct {
	Name        string
	Namespace   string
	Project     string
	BackendUUID UID
	ClusterUUID UID
	JobType     api.JobType
	Type        api.WorkloadType
	Phase       string
	Original    client.Object
}

func FromK8sResource(resource client.Object, workloadType api.WorkloadType, project string) (*Workload, error) {
	log.Debug().
		Str("resourceName", resource.GetName()).
		Str("resourceNamespace", resource.GetNamespace()).
		Str("workloadType", string(workloadType)).
		Str("project", project).
		Msg("Starting FromK8sResource")

	annotations := resource.GetAnnotations()
	uuid := annotations[clustercommonconstants.ControlPlaneWorkloadIdAnnotationForWorkload]
	log.Debug().Str("uuid", uuid).Msg("Retrieved UUID from annotations")

	jobType, err := getJobType(resource, workloadType)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get job type")
		return nil, err
	}

	phase, err := getWorkloadPhase(resource, workloadType)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get workload phase")
		return nil, err
	}

	workload := &Workload{
		Name:        resource.GetName(),
		Namespace:   resource.GetNamespace(),
		Project:     project,
		Type:        workloadType,
		JobType:     jobType,
		BackendUUID: UID(uuid),
		ClusterUUID: FromK8sUID(resource.GetUID()),
		Phase:       phase,
		Original:    resource,
	}

	log.Info().
		Str("name", workload.Name).
		Str("namespace", workload.Namespace).
		Str("project", workload.Project).
		Str("type", string(workload.Type)).
		Str("jobType", string(workload.JobType)).
		Str("phase", workload.Phase).
		Str("backendUUID", workload.BackendUUID.String()).
		Str("clusterUUID", workload.ClusterUUID.String()).
		Msg("Created Workload from K8s Resource")

	return workload, nil
}

func getJobType(workload client.Object, workloadType api.WorkloadType) (api.JobType, error) {
	if workload == nil {
		log.Error().Msg(workloadIsNilErrorMsg)
		return api.Unknown, errors.New(workloadIsNilErrorMsg)
	}

	log.Debug().
		Str("workloadType", string(workloadType)).
		Msg("Starting job type determination")

	if workloadType == api.Inference {
		log.Debug().Msg(fmt.Sprintf("%s workload, returning InferenceJob", workloadType))
		return api.InferenceJob, nil
	}

	if workloadType == api.DistributedInference {
		log.Debug().Msg(fmt.Sprintf("%s workload, returning DistributedInferenceJob", workloadType))
		return api.DistributedInferenceJob, nil
	}

	if workloadType == api.Distributed {
		distributedWorkload, ok := workload.(*dwv2.DistributedWorkload)
		if !ok {
			errorMsg := fmt.Sprintf(invalidWorkloadErrorMsg, api.Distributed)
			log.Error().Msg(errorMsg)

			return api.Unknown, errors.New(errorMsg)
		}

		log.Debug().Str("jobType", distributedWorkload.Spec.JobType).Msg("Determining job type for distributed workload")

		var jobType api.JobType

		switch distributedWorkload.Spec.JobType {
		case constants.KindMpiJob:
			jobType = api.MPIJob
		case constants.KindPyTorchJob:
			jobType = api.PyTorchJob
		case constants.KindXGBoostJob:
			jobType = api.XGBoostJob
		case constants.KindTFJob:
			jobType = api.TFJob
		case constants.KindJAXJob:
			jobType = api.JAXJob
		default:
			err := fmt.Errorf("invalid workload, unknown job type %s", distributedWorkload.Spec.JobType)
			log.Error().Err(err).Str("jobType", distributedWorkload.Spec.JobType).Msg("Unknown job type")

			return api.Unknown, err
		}

		log.Info().Str("jobType", string(jobType)).Msg("Job type determined")

		return jobType, nil
	}

	if workloadType == api.External {
		log.Debug().Msg(fmt.Sprintf("%s workload, returning ExternalJob", workloadType))

		return api.ExternalJob, nil
	}

	log.Debug().Msg(fmt.Sprintf("%s workload, returning RunaiJob", workloadType))

	return api.RunaiJob, nil
}

func getWorkloadPhase(workload client.Object, workloadType api.WorkloadType) (string, error) {
	if workload == nil {
		log.Error().Msg(workloadIsNilErrorMsg)
		return "", errors.New(workloadIsNilErrorMsg)
	}

	log.Debug().
		Str("workloadName", workload.GetName()).
		Str("resourceNamespace", workload.GetNamespace()).
		Str("workloadType", string(workloadType)).
		Msg("Starting workload phase determination")

	var phase string

	switch workloadType {
	case api.Inference:
		inference, ok := workload.(*infwv2.InferenceWorkload)
		if !ok {
			msg := fmt.Sprintf(invalidWorkloadErrorMsg, api.Inference)
			log.Error().Msg(msg)

			return "", errors.New(msg)
		}

		phase = inference.Status.Phase
	case api.DistributedInference:
		distInference, ok := workload.(*distinfwv2.DistributedInferenceWorkload)
		if !ok {
			msg := fmt.Sprintf(invalidWorkloadErrorMsg, api.DistributedInference)
			log.Error().Msg(msg)

			return "", errors.New(msg)
		}

		phase = distInference.Status.Phase
	case api.Workspace:
		interactive, ok := workload.(*iwv2.InteractiveWorkload)
		if !ok {
			msg := fmt.Sprintf(invalidWorkloadErrorMsg, api.Workspace)
			log.Error().Msg(msg)

			return "", errors.New(msg)
		}

		phase = interactive.Status.Phase
	case api.Training:
		training, ok := workload.(*twv2.TrainingWorkload)
		if !ok {
			msg := fmt.Sprintf(invalidWorkloadErrorMsg, api.Training)
			log.Error().Msg(msg)

			return "", errors.New(msg)
		}

		phase = training.Status.Phase
	case api.Distributed:
		dist, ok := workload.(*dwv2.DistributedWorkload)
		if !ok {
			msg := fmt.Sprintf(invalidWorkloadErrorMsg, api.Distributed)
			log.Error().Msg(msg)

			return "", errors.New(msg)
		}

		phase = dist.Status.Phase
	case api.External:
		ew, ok := workload.(*v1alpha1.ExternalWorkload)
		if !ok {
			msg := fmt.Sprintf(invalidWorkloadErrorMsg, api.External)
			log.Error().Msg(msg)

			return "", errors.New(msg)
		}

		phase = string(ew.Status.ObservedStatus.Phase)
		if phase == "" {
			phase = undefinedPhase
		}
	default:
		msg := fmt.Sprintf(unknownTypeErrorMsg, workloadType)
		log.Error().Msg(msg)

		return "", errors.New(msg)
	}

	log.Info().Str("workloadPhase", phase).Msg("Workload phase determined")

	return phase, nil
}
