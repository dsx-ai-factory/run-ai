package cluster_config

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
)

func ResolveClusterState(icc *models.ClusterConfig, maxDuration time.Duration) (string, string) {
	// Handle the case where the cluster didn't report any status or the cluster is of an older version that doesn't report status
	// In this case, we will assume that the cluster is in the applied state
	// And the actual config in the status will always be the default one
	if len(icc.Status.Conditions) == 0 {
		return string(models.ClusterStateApplied), string(models.ClusterStateMessageApplied)
	}
	if !isDesiredConfigStateEqualsToActual(&icc.Config, &icc.Status.ActualConfig) {
		return evaluateUpdatedConfigState(icc.UpdatedAt, maxDuration)
	} else {
		return evaluateReportedClusterConfigState(icc)
	}
}

func evaluateUpdatedConfigState(lastConfigUpdateTime time.Time, maxDuration time.Duration) (string, string) {
	if time.Since(lastConfigUpdateTime) < maxDuration {
		return string(models.ClusterStateUpdating), string(models.ClusterStateMessageUpdating)
	} else {
		return string(models.ClusterStateFailed), fmt.Sprintf(string(models.ClusterStateMessageFailed), models.ClusterStatusDidNotReturnAfterMaxInterval)
	}
}

func evaluateReportedClusterConfigState(icc *models.ClusterConfig) (string, string) {
	clusterApplyingCondition := icc.Status.GetCondition(models.ClusterConfigApplying)
	clusterAppliedCondition := icc.Status.GetCondition(models.ClusterConfigApplied)
	var state, statusMessage string
	var messages []string
	if clusterApplyingCondition != nil {
		if clusterApplyingCondition.Status == models.ClusterConfigSuccessStatusIndication {
			state = string(models.ClusterStateApplying)
			statusMessage = string(models.ClusterStateMessageApplying)
		} else if clusterApplyingCondition.Message != "" {
			messages = append(messages, clusterApplyingCondition.Message)
		}
	}
	if clusterAppliedCondition != nil {
		if clusterAppliedCondition.Status == models.ClusterConfigSuccessStatusIndication {
			state = string(models.ClusterStateApplied)
			statusMessage = string(models.ClusterStateMessageApplied)
		} else if clusterAppliedCondition.Message != "" {
			messages = append(messages, clusterAppliedCondition.Message)
		}
	}
	if state == "" {
		var aggMessages string
		if len(messages) > 0 {
			aggMessages = strings.Join(messages, ", ")
		}
		return string(models.ClusterStateFailed), fmt.Sprintf(string(models.ClusterStateMessageFailed), aggMessages)
	}
	return state, statusMessage
}

func isDesiredConfigStateEqualsToActual(config *models.ClusterConfigData, actual *models.ClusterConfigData) bool {
	if config == nil || actual == nil {
		return actual == config
	}
	return reflect.DeepEqual(config.UserConfig, actual.UserConfig) && reflect.DeepEqual(config.SystemConfig, actual.SystemConfig)
}

func MergeClusterConfig(patchConfig api.ClusterUserConfigDataPartialUpdate, clusterConfig api.ClusterUserConfigData) *api.ClusterUserConfigData {
	mergeConfig := api.ClusterUserConfigData{}

	targetVal := reflect.ValueOf(clusterConfig)
	patchVal := reflect.ValueOf(patchConfig)
	mergedConfigVal := reflect.ValueOf(&mergeConfig).Elem()

	for i := 0; i < targetVal.NumField(); i++ {
		targetField := targetVal.Field(i)
		fieldName := targetVal.Type().Field(i).Name
		patchField := patchVal.FieldByName(fieldName)
		mergedFieldVal := targetField
		if !patchField.IsNil() {
			mergedFieldVal = patchField
		}
		if mergedFieldVal.Kind() == reflect.Ptr {
			mergedFieldVal = mergedFieldVal.Elem()
		}
		mergedConfigVal.FieldByName(fieldName).Set(mergedFieldVal)
	}
	return &mergeConfig
}
