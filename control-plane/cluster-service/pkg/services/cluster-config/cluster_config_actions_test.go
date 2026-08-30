package cluster_config

import (
	"fmt"
	"testing"
	"time"

	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"

	"github.com/stretchr/testify/assert"
)

func TestResolveStatusDesiredConfigNotEqualToActualConfigTimeoutNotReached(t *testing.T) {
	icc := &models.ClusterConfig{
		Common: models.Common{
			UpdatedAt: time.Now().Add(-3 * time.Minute),
		},
		Config: api.ClusterFullConfigData{
			SystemConfig: api.ClusterSystemConfigData{
				Deactivated: false,
			},
			UserConfig: api.ClusterUserConfigData{},
		},
		Status: models.ClusterConfigStatus{
			ActualConfig: api.ClusterFullConfigData{
				SystemConfig: api.ClusterSystemConfigData{
					Deactivated: true,
				},
				UserConfig: api.ClusterUserConfigData{},
			},
			Conditions: []api.Condition{
				{
					Type:               "Applying",
					Status:             "True",
					LastTransitionTime: time.Now(),
				},
				{
					Type:               "Applied",
					Status:             "False",
					LastTransitionTime: time.Now(),
				},
			},
		},
	}
	state, stateMessage := ResolveClusterState(icc, 5*time.Minute)
	assert.Equal(t, state, string(models.ClusterStateUpdating))
	assert.Equal(t, stateMessage, string(models.ClusterStateMessageUpdating))
}

func TestResolveStatusMaxUpdateTimePassed(t *testing.T) {
	icc := &models.ClusterConfig{
		Common: models.Common{
			UpdatedAt: time.Now().Add(-7 * time.Minute),
		},
		Config: api.ClusterFullConfigData{
			SystemConfig: api.ClusterSystemConfigData{
				Deactivated: false,
			},
			UserConfig: api.ClusterUserConfigData{},
		},
		Status: models.ClusterConfigStatus{
			ActualConfig: api.ClusterFullConfigData{
				SystemConfig: api.ClusterSystemConfigData{
					Deactivated: true,
				},
				UserConfig: api.ClusterUserConfigData{},
			},
			Conditions: []api.Condition{
				{
					Type:               "Applying",
					Status:             "True",
					LastTransitionTime: time.Now().Add(-10 * time.Minute),
				},
			},
		},
	}
	state, stateMessage := ResolveClusterState(icc, 5*time.Minute)
	assert.Equal(t, state, string(models.ClusterStateFailed))
	assert.Equal(t, stateMessage, fmt.Sprintf(string(models.ClusterStateMessageFailed), models.ClusterStatusDidNotReturnAfterMaxInterval))
}

func TestResolveStatusRemoteConfigBeingUpdatedInCluster(t *testing.T) {
	icc := &models.ClusterConfig{
		Common: models.Common{
			UpdatedAt: time.Now().Add(-3 * time.Minute),
		},
		Config: api.ClusterFullConfigData{
			SystemConfig: api.ClusterSystemConfigData{
				Deactivated: false,
			},
			UserConfig: api.ClusterUserConfigData{},
		},
		Status: models.ClusterConfigStatus{
			ActualConfig: api.ClusterFullConfigData{
				SystemConfig: api.ClusterSystemConfigData{
					Deactivated: true,
				},
				UserConfig: api.ClusterUserConfigData{},
			},
			Conditions: []api.Condition{
				{
					Type:               "Applying",
					Status:             "True",
					LastTransitionTime: time.Now().Add(-10 * time.Minute),
				},
				{
					Type:               "Applied",
					Status:             "True",
					LastTransitionTime: time.Now().Add(-10 * time.Minute),
				},
			},
		},
	}
	state, stateMessage := ResolveClusterState(icc, 5*time.Minute)
	assert.Equal(t, state, string(models.ClusterStateUpdating))
	assert.Equal(t, stateMessage, string(models.ClusterStateMessageUpdating))
}

func TestResolveStatusApplyingAndApplied(t *testing.T) {
	icc := &models.ClusterConfig{
		Common: models.Common{
			UpdatedAt: time.Now().Add(-3 * time.Minute),
		},
		Config: api.ClusterFullConfigData{
			SystemConfig: api.ClusterSystemConfigData{
				Deactivated: true,
			},
			UserConfig: api.ClusterUserConfigData{},
		},
		Status: models.ClusterConfigStatus{
			ActualConfig: api.ClusterFullConfigData{
				SystemConfig: api.ClusterSystemConfigData{
					Deactivated: true,
				},
				UserConfig: api.ClusterUserConfigData{},
			},
			Conditions: []api.Condition{
				{
					Type:               "Applying",
					Status:             "True",
					LastTransitionTime: time.Now(),
				},
				{
					Type:               "Applied",
					Status:             "True",
					LastTransitionTime: time.Now(),
				},
			},
		},
	}
	state, stateMessage := ResolveClusterState(icc, 5*time.Minute)
	assert.Equal(t, state, string(models.ClusterStateApplied))
	assert.Equal(t, stateMessage, string(models.ClusterStateMessageApplied))
}

func TestResolveStatusRemoteConfigFailed(t *testing.T) {
	icc := &models.ClusterConfig{
		Common: models.Common{
			UpdatedAt: time.Now().Add(-3 * time.Minute),
		},
		Config: api.ClusterFullConfigData{
			SystemConfig: api.ClusterSystemConfigData{
				Deactivated: true,
			},
			UserConfig: api.ClusterUserConfigData{},
		},
		Status: models.ClusterConfigStatus{
			ActualConfig: api.ClusterFullConfigData{
				SystemConfig: api.ClusterSystemConfigData{
					Deactivated: true,
				},
				UserConfig: api.ClusterUserConfigData{},
			},
			Conditions: []api.Condition{
				{
					Type:               "Applying",
					Status:             "False",
					LastTransitionTime: time.Now(),
					Message:            "message applying",
				},
				{
					Type:               "Applied",
					Status:             "False",
					LastTransitionTime: time.Now(),
					Message:            "message applies",
				},
			},
		},
	}
	state, stateMessage := ResolveClusterState(icc, 5*time.Minute)
	assert.Equal(t, state, string(models.ClusterStateFailed))
	assert.Equal(t, stateMessage, fmt.Sprintf(string(models.ClusterStateMessageFailed), "message applying, message applies"))
}

func TestResolveStatusUserUpdatedSameConfig(t *testing.T) {
	icc := &models.ClusterConfig{
		Common: models.Common{
			UpdatedAt: time.Now(),
		},
		Config: api.ClusterFullConfigData{
			SystemConfig: api.ClusterSystemConfigData{
				Deactivated: true,
			},
			UserConfig: api.ClusterUserConfigData{},
		},
		Status: models.ClusterConfigStatus{
			ActualConfig: api.ClusterFullConfigData{
				SystemConfig: api.ClusterSystemConfigData{
					Deactivated: true,
				},
				UserConfig: api.ClusterUserConfigData{},
			},
			Conditions: []api.Condition{
				{
					Type:               "Applying",
					Status:             "True",
					LastTransitionTime: time.Now().Add(-10 * time.Minute),
				},
				{
					Type:               "Applied",
					Status:             "True",
					LastTransitionTime: time.Now().Add(-10 * time.Minute),
				},
			},
		},
	}
	state, stateMessage := ResolveClusterState(icc, 5*time.Minute)
	assert.Equal(t, state, string(models.ClusterStateApplied))
	assert.Equal(t, stateMessage, string(models.ClusterStateMessageApplied))
}

func TestResolveStatusApplyingAfterApplied(t *testing.T) {
	icc := &models.ClusterConfig{
		Common: models.Common{
			UpdatedAt: time.Now().Add(-3 * time.Minute),
		},
		Config: api.ClusterFullConfigData{
			SystemConfig: api.ClusterSystemConfigData{
				Deactivated: true,
			},
			UserConfig: api.ClusterUserConfigData{},
		},
		Status: models.ClusterConfigStatus{
			ActualConfig: api.ClusterFullConfigData{
				SystemConfig: api.ClusterSystemConfigData{
					Deactivated: true,
				},
				UserConfig: api.ClusterUserConfigData{},
			},
			Conditions: []api.Condition{
				{
					Type:               "Applying",
					Status:             "True",
					LastTransitionTime: time.Now(),
				},
				{
					Type:               "Applied",
					Status:             "False",
					LastTransitionTime: time.Now(),
				},
			},
		},
	}
	state, stateMessage := ResolveClusterState(icc, 5*time.Minute)
	assert.Equal(t, state, string(models.ClusterStateApplying))
	assert.Equal(t, stateMessage, string(models.ClusterStateMessageApplying))
}
