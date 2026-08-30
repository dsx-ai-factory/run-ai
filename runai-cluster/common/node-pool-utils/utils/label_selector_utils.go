package utils

import (
	"github.com/rs/zerolog/log"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

// GetNodePoolLabelSelector returns a label requirement that matches the given
// node pool. When the name equals defaultNodepoolName, the requirement is
// "label key does not exist" (the default pool is represented by the absence
// of the assignment label). Otherwise it's "label key equals name".
//
// nodePoolAssignmentLabelKey is the cluster-wide label key used to record
// node pool assignment on workload-side objects.
func GetNodePoolLabelSelector(nodePoolName, nodePoolAssignmentLabelKey, defaultNodepoolName string) (requirement *labels.Requirement, err error) {
	if nodePoolName == defaultNodepoolName {
		return requirementNodePoolNameDoesNotExist(nodePoolAssignmentLabelKey)
	}
	return RequirementNodePoolNameMatching(nodePoolName, nodePoolAssignmentLabelKey)
}

// RequirementNodePoolNameMatching returns "key == nodePoolName" for the given
// assignment label key.
func RequirementNodePoolNameMatching(nodePoolName, nodePoolAssignmentLabelKey string) (requirement *labels.Requirement, err error) {
	return requirementNodePoolName(nodePoolName, nodePoolAssignmentLabelKey, true, true)
}

// GetNotNodePoolLabelSelector returns a requirement that EXCLUDES the given
// node pool. When the name equals defaultNodepoolName, the requirement is
// "label key exists" (since the default pool is represented by absence of
// the label, anything with a label is not the default). Otherwise it's
// "label key not in [nodePoolName]" (which also matches resources without
// the label).
func GetNotNodePoolLabelSelector(nodePoolName, nodePoolAssignmentLabelKey, defaultNodepoolName string) (requirement *labels.Requirement, err error) {
	if nodePoolName == defaultNodepoolName {
		return newRequirement(nodePoolAssignmentLabelKey, selection.Exists, []string{})
	}

	// "NotIn" will also match resources with no labels with the "nodePoolName" key
	return newRequirement(nodePoolAssignmentLabelKey, selection.NotIn, []string{nodePoolName})
}

func requirementNodePoolNameDoesNotExist(nodePoolAssignmentLabelKey string) (requirement *labels.Requirement, err error) {
	return requirementNodePoolName("", nodePoolAssignmentLabelKey, false, false)
}

func RequirementNodePoolLabelMatching(labelKey, labelValue string) (requirement *labels.Requirement, err error) {
	return requirementNodePoolLabel(labelKey, labelValue, true)
}

func RequirementNodePoolLabelNotMatching(labelKey, labelValue string) (requirement *labels.Requirement, err error) {
	return requirementNodePoolLabel(labelKey, labelValue, false)
}

func requirementNodePoolName(nodePoolName, nodePoolAssignmentLabelKey string, matching bool, exists bool) (requirement *labels.Requirement, err error) {
	vals := []string{nodePoolName}
	op := selection.DoubleEquals
	if !matching {
		op = selection.NotEquals
	}
	if !exists {
		op = selection.DoesNotExist
		vals = []string{}
	}

	return newRequirement(nodePoolAssignmentLabelKey, op, vals)
}

func requirementNodePoolLabel(labelKey, labelValue string, matching bool) (requirement *labels.Requirement, err error) {
	op := selection.DoubleEquals
	if !matching {
		op = selection.NotEquals
	}

	return newRequirement(labelKey, op, []string{labelValue})
}

func newRequirement(key string, op selection.Operator, vals []string) (requirement *labels.Requirement, err error) {
	requirement, err = labels.NewRequirement(key, op, vals)
	if err != nil {
		log.Error().Msgf("Failed creating requirement %v/%v/%v, err: %v",
			key, op, vals, err.Error())
	}

	return requirement, err
}
