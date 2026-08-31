/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config_maps

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var mu sync.Mutex

const (
	AgentConfigName = "agent-config"

	TrainingPoliciesSyncCounter    = "training-policies-sync-counter"
	InteractivePoliciesSyncCounter = "interactive-policies-sync-counter"
	DistributedPoliciesSyncCounter = "distributed-policies-sync-counter"
	InferencePoliciesSyncCounter   = "inference-policies-sync-counter"

	TrainingWorkloadsSyncCounter             = "training-workloads-sync-counter"
	InteractiveWorkloadsSyncCounter          = "interactive-workloads-sync-counter"
	DistributedWorkloadsSyncCounter          = "distributed-workloads-sync-counter"
	InferenceWorkloadsSyncCounter            = "inference-workloads-sync-counter"
	DistributedInferenceWorkloadsSyncCounter = "distributed-inference-workloads-sync-counter"

	WorkloadsV2SyncCounter = "workloads-v2-sync-counter"

	DatasourcesSyncCounter  = "datasources-sync-counter"
	CredentialsSyncCounter  = "credentials-sync-counter"
	ApplicationsSyncCounter = "applications-sync-counter"
)

type AgentConfigData struct {
	TrainingPoliciesSyncCounter    int64
	InteractivePoliciesSyncCounter int64
	DistributedPoliciesSyncCounter int64
	InferencePoliciesSyncCounter   int64

	TrainingWorkloadsSyncCounter             int64
	InteractiveWorkloadsSyncCounter          int64
	DistributedWorkloadsSyncCounter          int64
	InferenceWorkloadsSyncCounter            int64
	DistributedInferenceWorkloadsSyncCounter int64

	DatasourcesSyncCounter int64
	CredentialsSyncCounter int64

	WorkloadsV2SyncCounter int64

	ApplicationsSyncCounter int64
}

func NewAgentConfig() *AgentConfigData {
	return &AgentConfigData{
		TrainingPoliciesSyncCounter:    0,
		InteractivePoliciesSyncCounter: 0,
		DistributedPoliciesSyncCounter: 0,
		InferencePoliciesSyncCounter:   0,

		TrainingWorkloadsSyncCounter:    0,
		InteractiveWorkloadsSyncCounter: 0,
		DistributedWorkloadsSyncCounter: 0,
		InferenceWorkloadsSyncCounter:   0,

		DatasourcesSyncCounter: 0,
		CredentialsSyncCounter: 0,

		WorkloadsV2SyncCounter: 0,

		ApplicationsSyncCounter: 0,
	}
}

func (ac *AgentConfigData) updateCluster(runaiClient client.Client, cachedClient client.Reader, counterToUpdate string) error {
	mu.Lock()
	defer mu.Unlock()

	clusterAgentData, err := GetAgentConfigData(cachedClient)
	if err != nil {
		return err
	}

	err = ac.updateAgentConfigDataCounter(counterToUpdate, clusterAgentData)
	if err != nil {
		log.Error().Msgf("Failed to update agent config data counter for %s: %s", counterToUpdate, err.Error())
		return err
	}

	err = runaiClient.Update(context.Background(), clusterAgentData)
	if err != nil {
		log.Error().Msgf("Failed to update agent configmap, error: %s", err.Error())
		return err
	}

	return nil
}

func (ac *AgentConfigData) updateAgentConfigDataCounter(counterToUpdate string, clusterAgentData *corev1.ConfigMap) error {
	var counterKey string
	var counter int64

	switch counterToUpdate {
	case common.PolicyTypeInteractive:
		counterKey = InteractivePoliciesSyncCounter
		counter = ac.InteractivePoliciesSyncCounter
	case common.PolicyTypeTraining:
		counterKey = TrainingPoliciesSyncCounter
		counter = ac.TrainingPoliciesSyncCounter
	case common.PolicyTypeDistributed:
		counterKey = DistributedPoliciesSyncCounter
		counter = ac.DistributedPoliciesSyncCounter
	case common.PolicyTypeInference:
		counterKey = InferencePoliciesSyncCounter
		counter = ac.InferencePoliciesSyncCounter

	case common.WorkloadTypeInteractive:
		counterKey = InteractiveWorkloadsSyncCounter
		counter = ac.InteractiveWorkloadsSyncCounter
	case common.WorkloadTypeTraining:
		counterKey = TrainingWorkloadsSyncCounter
		counter = ac.TrainingWorkloadsSyncCounter
	case common.WorkloadTypeDistributed:
		counterKey = DistributedWorkloadsSyncCounter
		counter = ac.DistributedWorkloadsSyncCounter
	case common.WorkloadTypeInference:
		counterKey = InferenceWorkloadsSyncCounter
		counter = ac.InferenceWorkloadsSyncCounter
	case common.WorkloadTypeDistributedInference:
		counterKey = DistributedInferenceWorkloadsSyncCounter
		counter = ac.DistributedInferenceWorkloadsSyncCounter
	case common.DatasourceAssets:
		counterKey = DatasourcesSyncCounter
		counter = ac.DatasourcesSyncCounter
	case common.CredentialAssets:
		counterKey = CredentialsSyncCounter
		counter = ac.CredentialsSyncCounter
	case common.WorkloadsV2Type:
		counterKey = WorkloadsV2SyncCounter
		counter = ac.WorkloadsV2SyncCounter
	case common.ApplicationsType:
		counterKey = ApplicationsSyncCounter
		counter = ac.ApplicationsSyncCounter
	default:
		log.Error().Msgf("Unknown counter type: %s", counterToUpdate)
		return errors.New("unknown counter type")
	}

	clusterAgentData.Data[counterKey] = strconv.FormatInt(counter, 10)
	return nil
}

func (ac *AgentConfigData) SetData(configMap corev1.ConfigMap) (*AgentConfigData, error) {
	agentConfigData := AgentConfigData{}

	var err error

	var trainingPoliciesSyncCounter int64
	var interactivePoliciesSyncCounter int64
	var distributedPoliciesSyncCounter int64
	var inferencePoliciesSyncCounter int64

	var trainingWorkloadsSyncCounter int64
	var interactiveWorkloadsSyncCounter int64
	var distributedWorkloadsSyncCounter int64
	var inferenceWorkloadsSyncCounter int64

	var datasourcesSyncCounter int64
	var credentialsSyncCouunter int64

	var workloadsV2SyncCounter int64

	if configMap.Data[TrainingPoliciesSyncCounter] != "" {
		trainingPoliciesSyncCounter, err = strconv.ParseInt(configMap.Data[TrainingPoliciesSyncCounter], 10, 64)
		if err != nil {
			log.Error().Msgf("Failed to parse training policies sync counter, TrainingPoliciesSyncCounter: %s, error: %s", configMap.Data[TrainingPoliciesSyncCounter], err.Error())
			return &agentConfigData, err
		}
	}
	agentConfigData.SetTrainingPoliciesCounter(trainingPoliciesSyncCounter)

	if configMap.Data[InteractivePoliciesSyncCounter] != "" {
		interactivePoliciesSyncCounter, err = strconv.ParseInt(configMap.Data[InteractivePoliciesSyncCounter], 10, 64)
		if err != nil {
			log.Error().Msgf("Failed to parse interactive policies sync counter, InteractivePoliciesSyncCounter: %s, error: %s", configMap.Data[InteractivePoliciesSyncCounter], err.Error())
			return &agentConfigData, err
		}
	}
	agentConfigData.SetInteractivePoliciesCounter(interactivePoliciesSyncCounter)

	if configMap.Data[DistributedPoliciesSyncCounter] != "" {
		distributedPoliciesSyncCounter, err = strconv.ParseInt(configMap.Data[DistributedPoliciesSyncCounter], 10, 64)
		if err != nil {
			log.Error().Msgf("Failed to parse distributed policies sync counter, DistributedPoliciesSyncCounter: %s, error: %s", configMap.Data[DistributedPoliciesSyncCounter], err.Error())
			return &agentConfigData, err
		}
	}
	agentConfigData.SetDistributedPoliciesCounter(distributedPoliciesSyncCounter)

	if configMap.Data[InferencePoliciesSyncCounter] != "" {
		inferencePoliciesSyncCounter, err = strconv.ParseInt(configMap.Data[InferencePoliciesSyncCounter], 10, 64)
		if err != nil {
			log.Error().Msgf("Failed to parse inference policies sync counter, InferencePoliciesSyncCounter: %s, error: %s", configMap.Data[InferencePoliciesSyncCounter], err.Error())
			return &agentConfigData, err
		}
	}
	agentConfigData.SetInferencePoliciesCounter(inferencePoliciesSyncCounter)

	if configMap.Data[InteractiveWorkloadsSyncCounter] != "" {
		interactiveWorkloadsSyncCounter, err = strconv.ParseInt(configMap.Data[InteractiveWorkloadsSyncCounter], 10, 64)
		if err != nil {
			log.Error().Msgf("Failed to parse interactive workloads sync counter, InteractiveWorkloadsSyncCounter: %s, error: %s", configMap.Data[InteractiveWorkloadsSyncCounter], err.Error())
			return &agentConfigData, err
		}
	}
	agentConfigData.SetInteractiveWorkloadsCounter(interactiveWorkloadsSyncCounter)

	if configMap.Data[TrainingWorkloadsSyncCounter] != "" {
		trainingWorkloadsSyncCounter, err = strconv.ParseInt(configMap.Data[TrainingWorkloadsSyncCounter], 10, 64)
		if err != nil {
			log.Error().Msgf("Failed to parse training workloads sync counter, TrainingWorkloadsSyncCounter: %s, error: %s", configMap.Data[TrainingWorkloadsSyncCounter], err.Error())
			return &agentConfigData, err
		}
	}
	agentConfigData.SetTrainingWorkloadsCounter(trainingWorkloadsSyncCounter)

	if configMap.Data[DistributedWorkloadsSyncCounter] != "" {
		distributedWorkloadsSyncCounter, err = strconv.ParseInt(configMap.Data[DistributedWorkloadsSyncCounter], 10, 64)
		if err != nil {
			log.Error().Msgf("Failed to parse training workloads sync counter, DistributedWorkloadsSyncCounter: %s, error: %s", configMap.Data[DistributedWorkloadsSyncCounter], err.Error())
			return &agentConfigData, err
		}
	}
	agentConfigData.SetDistributedWorkloadsCounter(distributedWorkloadsSyncCounter)

	if configMap.Data[InferenceWorkloadsSyncCounter] != "" {
		inferenceWorkloadsSyncCounter, err = strconv.ParseInt(configMap.Data[InferenceWorkloadsSyncCounter], 10, 64)
		if err != nil {
			log.Error().Msgf("Failed to parse training workloads sync counter, InferenceWorkloadsSyncCounter: %s, error: %s", configMap.Data[InferenceWorkloadsSyncCounter], err.Error())
			return &agentConfigData, err
		}
	}
	agentConfigData.SetInferenceWorkloadsCounter(inferenceWorkloadsSyncCounter)

	if configMap.Data[DatasourcesSyncCounter] != "" {
		datasourcesSyncCounter, err = strconv.ParseInt(configMap.Data[DatasourcesSyncCounter], 10, 64)
		if err != nil {
			log.Error().Msgf("Failed to parse datasources sync counter: %s, error: %s", configMap.Data[DatasourcesSyncCounter], err.Error())
			return &agentConfigData, err
		}
	}
	agentConfigData.SetDatasourcesSyncCounter(datasourcesSyncCounter)

	if configMap.Data[CredentialsSyncCounter] != "" {
		credentialsSyncCouunter, err = strconv.ParseInt(configMap.Data[CredentialsSyncCounter], 10, 64)
		if err != nil {
			log.Error().Msgf("Failed to parse credentials sync counter: %s, error: %s", configMap.Data[CredentialsSyncCounter], err.Error())
			return &agentConfigData, err
		}
	}
	agentConfigData.SetCredentialsSyncCounter(credentialsSyncCouunter)

	if configMap.Data[WorkloadsV2SyncCounter] != "" {
		workloadsV2SyncCounter, err = strconv.ParseInt(configMap.Data[WorkloadsV2SyncCounter], 10, 64)
		if err != nil {
			log.Error().Msgf("Failed to parse workloads V2 sync counter: %s, error: %s", configMap.Data[WorkloadsV2SyncCounter], err.Error())
			return &agentConfigData, err
		}
	}
	agentConfigData.SetWorkloadsV2Counter(workloadsV2SyncCounter)

	var applicationsSyncCounter int64
	if configMap.Data[ApplicationsSyncCounter] != "" {
		applicationsSyncCounter, err = strconv.ParseInt(configMap.Data[ApplicationsSyncCounter], 10, 64)
		if err != nil {
			log.Error().Msgf("Failed to parse applications sync counter: %s, error: %s", configMap.Data[ApplicationsSyncCounter], err.Error())
			return &agentConfigData, err
		}
	}
	agentConfigData.SetApplicationsSyncCounter(applicationsSyncCounter)

	return &agentConfigData, nil
}

func (ac *AgentConfigData) SetTrainingPoliciesCounter(updateCounter int64) *AgentConfigData {
	ac.TrainingPoliciesSyncCounter = updateCounter
	return ac
}

func (ac *AgentConfigData) SetDistributedPoliciesCounter(updateCounter int64) *AgentConfigData {
	ac.DistributedPoliciesSyncCounter = updateCounter
	return ac
}

func (ac *AgentConfigData) SetInteractivePoliciesCounter(updateCounter int64) *AgentConfigData {
	ac.InteractivePoliciesSyncCounter = updateCounter
	return ac
}

func (ac *AgentConfigData) SetInferencePoliciesCounter(updateCounter int64) *AgentConfigData {
	ac.InferencePoliciesSyncCounter = updateCounter
	return ac
}

func (ac *AgentConfigData) SetTrainingWorkloadsCounter(updateCounter int64) *AgentConfigData {
	ac.TrainingWorkloadsSyncCounter = updateCounter
	return ac
}

func (ac *AgentConfigData) SetInteractiveWorkloadsCounter(updateCounter int64) *AgentConfigData {
	ac.InteractiveWorkloadsSyncCounter = updateCounter
	return ac
}

func (ac *AgentConfigData) SetDistributedWorkloadsCounter(updateCounter int64) *AgentConfigData {
	ac.DistributedWorkloadsSyncCounter = updateCounter
	return ac
}

func (ac *AgentConfigData) SetInferenceWorkloadsCounter(updateCounter int64) *AgentConfigData {
	ac.InferenceWorkloadsSyncCounter = updateCounter
	return ac
}

func (ac *AgentConfigData) SetDistributedInferenceWorkloadsCounter(updateCounter int64) *AgentConfigData {
	ac.DistributedInferenceWorkloadsSyncCounter = updateCounter
	return ac
}

func (ac *AgentConfigData) SetDatasourcesSyncCounter(updateCounter int64) *AgentConfigData {
	ac.DatasourcesSyncCounter = updateCounter
	return ac
}

func (ac *AgentConfigData) SetCredentialsSyncCounter(updateCounter int64) *AgentConfigData {
	ac.CredentialsSyncCounter = updateCounter
	return ac
}

func (ac *AgentConfigData) SetWorkloadsV2Counter(updateCounter int64) *AgentConfigData {
	ac.WorkloadsV2SyncCounter = updateCounter
	return ac
}

func (ac *AgentConfigData) SetApplicationsSyncCounter(updateCounter int64) *AgentConfigData {
	ac.ApplicationsSyncCounter = updateCounter
	return ac
}

func agentConfigMapToAgentConfigData(agentConfigMap *corev1.ConfigMap) (*AgentConfigData, error) {
	agentConfig, err := NewAgentConfig().SetData(*agentConfigMap)
	if err != nil {
		return nil, err
	}
	return agentConfig, err
}

func GetAgentConfigData(cachedClient client.Reader) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}
	err := cachedClient.Get(context.Background(), types.NamespacedName{
		Name:      AgentConfigName,
		Namespace: constants.NamespaceRunAi,
	}, configMap)

	if err != nil {
		return nil, err
	}

	return configMap, nil
}

func GetAgentConfig(cachedClient client.Reader) (*AgentConfigData, error) {
	configMap := &corev1.ConfigMap{}
	err := cachedClient.Get(context.Background(), types.NamespacedName{
		Name:      AgentConfigName,
		Namespace: constants.NamespaceRunAi,
	}, configMap)
	if err != nil {
		return nil, err
	}
	return agentConfigMapToAgentConfigData(configMap)
}

func UpdateInteractivePoliciesSyncCounters(runaiClient client.Client, cachedClient client.Reader, counter int64) error {
	return NewAgentConfig().SetInteractivePoliciesCounter(counter).updateCluster(runaiClient, cachedClient, common.PolicyTypeInteractive)
}

func UpdateTrainingPoliciesSyncCounters(runaiClient client.Client, cachedClient client.Reader, counter int64) error {
	return NewAgentConfig().SetTrainingPoliciesCounter(counter).updateCluster(runaiClient, cachedClient, common.PolicyTypeTraining)
}

func UpdateDistributedPoliciesSyncCounters(runaiClient client.Client, cachedClient client.Reader, counter int64) error {
	return NewAgentConfig().SetDistributedPoliciesCounter(counter).updateCluster(runaiClient, cachedClient, common.PolicyTypeDistributed)
}

func UpdateInferencePoliciesSyncCounters(runaiClient client.Client, cachedClient client.Reader, counter int64) error {
	return NewAgentConfig().SetInferencePoliciesCounter(counter).updateCluster(runaiClient, cachedClient, common.PolicyTypeInference)
}

func UpdateTrainingWorkloadsSyncCounters(runaiClient client.Client, cachedClient client.Reader, counter int64) error {
	return NewAgentConfig().SetTrainingWorkloadsCounter(counter).updateCluster(runaiClient, cachedClient, common.WorkloadTypeTraining)
}

func UpdateInteractiveWorkloadsSyncCounters(runaiClient client.Client, cachedClient client.Reader, counter int64) error {
	return NewAgentConfig().SetInteractiveWorkloadsCounter(counter).updateCluster(runaiClient, cachedClient, common.WorkloadTypeInteractive)
}

func UpdateDistributedWorkloadsSyncCounters(runaiClient client.Client, cachedClient client.Reader, counter int64) error {
	return NewAgentConfig().SetDistributedWorkloadsCounter(counter).updateCluster(runaiClient, cachedClient, common.WorkloadTypeDistributed)
}

func UpdateInferenceWorkloadsSyncCounters(runaiClient client.Client, cachedClient client.Reader, counter int64) error {
	return NewAgentConfig().SetInferenceWorkloadsCounter(counter).updateCluster(runaiClient, cachedClient, common.WorkloadTypeInference)
}

func UpdateDistributedInferenceWorkloadsSyncCounters(runaiClient client.Client, cachedClient client.Reader, counter int64) error {
	return NewAgentConfig().SetDistributedInferenceWorkloadsCounter(counter).updateCluster(runaiClient, cachedClient, common.WorkloadTypeDistributedInference)
}

func UpdateDatasourceAssetsSyncCounter(runaiClient client.Client, cachedClient client.Reader, counter int64) error {
	return NewAgentConfig().SetDatasourcesSyncCounter(counter).updateCluster(runaiClient, cachedClient, common.DatasourceAssets)
}

func UpdateCredentialsAssetsSyncCounter(runaiClient client.Client, cachedClient client.Reader, counter int64) error {
	return NewAgentConfig().SetCredentialsSyncCounter(counter).updateCluster(runaiClient, cachedClient, common.CredentialAssets)
}

func UpdateWorkloadsV2SyncCounter(runaiClient client.Client, cachedClient client.Reader, counter int64) error {
	return NewAgentConfig().SetWorkloadsV2Counter(counter).updateCluster(runaiClient, cachedClient, common.WorkloadsV2Type)
}

func UpdateApplicationsSyncCounter(runaiClient client.Client, cachedClient client.Reader, counter int64) error {
	return NewAgentConfig().SetApplicationsSyncCounter(counter).updateCluster(runaiClient, cachedClient, common.ApplicationsType)
}

func CreateAgentConfigMap(runaiClient client.Client) (*AgentConfigData, error) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.NamespaceRunAi,
			Name:      AgentConfigName,
		},
		Data: map[string]string{
			TrainingPoliciesSyncCounter:    "0",
			InteractivePoliciesSyncCounter: "0",
			DistributedPoliciesSyncCounter: "0",
			InferencePoliciesSyncCounter:   "0",

			TrainingWorkloadsSyncCounter:    "0",
			InteractiveWorkloadsSyncCounter: "0",
			DistributedWorkloadsSyncCounter: "0",
			InferenceWorkloadsSyncCounter:   "0",

			DatasourcesSyncCounter: "0",
			CredentialsSyncCounter: "0",

			WorkloadsV2SyncCounter: "0",

			ApplicationsSyncCounter: "0",
		},
	}
	err := runaiClient.Create(context.Background(), configMap)
	if err != nil {
		return nil, err
	}
	return agentConfigMapToAgentConfigData(configMap)
}

func CreateAgentConfigIfNotExists(runaiClient client.Client, cachedClient client.Reader) (*AgentConfigData, error) {
	agentConfig, err := GetAgentConfig(cachedClient)
	if err == nil {
		return agentConfig, nil
	}

	if client.IgnoreNotFound(err) == nil {
		agentConfig, err = CreateAgentConfigMap(runaiClient)
		if err != nil {
			log.Error().Msgf("Failed to create agent configmap, error: %s", err.Error())
			return nil, err
		}
		return agentConfig, nil
	}

	log.Error().Msgf("Failed to get agent configmap, error: %s", err.Error())
	return nil, err
}
