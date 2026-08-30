package utils

import (
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/config"
	nodepoolutils "github.com/run-ai/runai/runai-cluster/common/node-pool-utils/utils"

	"k8s.io/apimachinery/pkg/labels"
)

func GetNodePoolNameFromLabels(labels map[string]string) string {
	return nodepoolutils.GetNodePoolNameFromLabels(labels, config.Get().NodePoolNameLabel, config.Get().DefaultNodepoolName)
}

func GetNodePoolLabelSelector(nodePoolName string) (requirement *labels.Requirement, err error) {
	return nodepoolutils.GetNodePoolLabelSelector(nodePoolName, config.Get().NodePoolNameLabel, config.Get().DefaultNodepoolName)
}

func RequirementNodePoolNameMatching(nodePoolName string) (requirement *labels.Requirement, err error) {
	return nodepoolutils.RequirementNodePoolNameMatching(nodePoolName, config.Get().NodePoolNameLabel)
}

func GetNotNodePoolLabelSelector(nodePoolName string) (requirement *labels.Requirement, err error) {
	return nodepoolutils.GetNotNodePoolLabelSelector(nodePoolName, config.Get().NodePoolNameLabel, config.Get().DefaultNodepoolName)
}

func RequirementNodePoolLabelMatching(labelKey, labelValue string) (requirement *labels.Requirement, err error) {
	return nodepoolutils.RequirementNodePoolLabelMatching(labelKey, labelValue)
}

func RequirementNodePoolLabelNotMatching(labelKey, labelValue string) (requirement *labels.Requirement, err error) {
	return nodepoolutils.RequirementNodePoolLabelNotMatching(labelKey, labelValue)
}
