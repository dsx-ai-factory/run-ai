package schema_upgrades

import (
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	"testing"
)

func TestAddiontionalTolerationsSet(t *testing.T) {
	tolerations := []coreV1.Toleration{
		{
			Key:      "key1",
			Operator: coreV1.TolerationOpEqual,
			Value:    "value1",
			Effect:   coreV1.TaintEffectNoSchedule,
		},
	}
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.Spec.Global = &global.Global{
		AdditionalTolerations: tolerations,
	}
	daemonSetsTolerationsUpgrader := newDaemonSetsTolerationsUpgrader()

	shouldUpgrade := daemonSetsTolerationsUpgrader.shouldPerformUpgrade(runaiConfig)
	assert.Equal(t, shouldUpgrade, true)

	resource, err := daemonSetsTolerationsUpgrader.updateRunaiConfig(runaiConfig)
	assert.Nil(t, err)
	assert.Equal(t, len(resource.Spec.Global.AdditionalTolerations), 0)
	assert.Equal(t, len(resource.Spec.Global.DaemonsetsTolerations), 1)
	assert.Equal(t, resource.Spec.Global.DaemonsetsTolerations, tolerations)

	shouldUpgrade = daemonSetsTolerationsUpgrader.shouldPerformUpgrade(resource)
	assert.Equal(t, shouldUpgrade, false)
}

func TestAddiontionalTolerationsNotSet(t *testing.T) {
	runaiConfig := &runaiv1.RunaiConfig{}

	daemonSetsTolerationsUpgrader := newDaemonSetsTolerationsUpgrader()

	shouldUpgrade := daemonSetsTolerationsUpgrader.shouldPerformUpgrade(runaiConfig)
	assert.Equal(t, shouldUpgrade, false)
}

func TestDaemonsetTolerationsSet(t *testing.T) {
	tolerations := []coreV1.Toleration{
		{
			Key:      "key1",
			Operator: coreV1.TolerationOpEqual,
			Value:    "value1",
			Effect:   coreV1.TaintEffectNoSchedule,
		},
	}
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.Spec.Global = &global.Global{
		DaemonsetsTolerations: tolerations,
	}
	daemonSetsTolerationsUpgrader := newDaemonSetsTolerationsUpgrader()

	shouldUpgrade := daemonSetsTolerationsUpgrader.shouldPerformUpgrade(runaiConfig)
	assert.Equal(t, shouldUpgrade, false)
}
