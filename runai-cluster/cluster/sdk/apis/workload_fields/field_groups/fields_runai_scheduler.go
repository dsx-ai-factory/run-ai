/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_groups

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	log "github.com/sirupsen/logrus"
	"k8s.io/utils/ptr"
)

// +k8s:deepcopy-gen=true
type RunaiSchedulerFields struct {

	// Specifies nodes (machines) or a group of nodes on which the workload will run. To use this
	// feature, your Administrator will need to label nodes as explained in the Group Nodes guide at
	// https://docs.run.ai/admin/researcher-setup/limit-to-node-group.
	// This flag can be used in conjunction with Project-based affinity. In this case, the flag is
	// used to refine the list of allowable node groups set in the Project. For more information see
	// the Projects setup guide at https://docs.run.ai/admin/admin-ui-setup/project-setup.
	NodeType *flds.StringField `json:"nodeType,omitempty"`

	// Specifies the memory profile to be used for workload running on NVIDIA Multi-Instance GPU (MIG) technology. - DEPRECATED
	MigProfile *flds.StringField `json:"migProfile,omitempty"`

	// Specifies the NodePool name to be used to schedule this job on - DEPRECATED use NodePools instead
	NodePool *flds.StringField `json:"nodePool,omitempty"`

	// Specifies the list of node pools to use for scheduling the job, ordered by preference.
	NodePools *flds.StringField `json:"nodePools,omitempty"`

	// Specifies the node affinity requirements for the job's pods.
	// https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#node-affinity
	flds.NodeAffinityFields `json:""`

	// Indicates whether pod affinity scheduling rules apply.
	PodAffinity *flds.BoolField `json:"podAffinity,omitempty"`

	// indicates the number of times the job can be preempted
	PreemptionLimit *flds.IntField `json:"preemptionLimit,omitempty"`

	// Specifies the Pod Affinity Topology to be used for scheduling the job
	// This field can be specified only if PodAffinity is set to true
	PodAffinityTopology *flds.StringField `json:"podAffinityTopology,omitempty"`

	// Indicates if we want to use the Pod affinity rule as :
	// the "hard" (required) or the "soft" (preferred)
	// This field can be specified only if PodAffinity is set to true
	PodAffinitySchedulingRule *flds.StringField `json:"podAffinitySchedulingRule,omitempty"`

	// Specifies the duration after which it is possible for a finished workload to be
	// automatically deleted. When the workload is being deleted, its lifecycle
	// guarantees (e.g. finalizers) will be honored. If this field is unset,
	// the workload won't be automatically deleted. If this field is set to zero,
	// the workload becomes eligible to be deleted immediately after it finishes.
	AutoDeletionTimeAfterCompletionSeconds *flds.IntField `json:"autoDeletionTimeAfterCompletionSeconds,omitempty"`

	// Optional duration in seconds the workload needs to terminate gracefully.
	// Value must be non-negative integer. The value zero indicates stop immediately
	// via the kill signal (no opportunity to shut down). Set this value longer than
	// the expected cleanup time for your process. Defaults to 30 seconds.
	TerminationGracePeriodSeconds *flds.IntField `json:"terminationGracePeriodSeconds,omitempty"`
}

func RunaiSchedulerRunaiRules(ctx context.Context, nodePoolLister NodePoolNameLister) RunaiSchedulerFields {
	var result RunaiSchedulerFields

	result.MigProfile = getMigProfileRules()
	result.NodePool = getNodePoolRules(ctx, nodePoolLister)
	// the rules will be calculated from NodePool field;
	// but we need it here just so the existence of this field is known when asked.
	result.NodePools = &flds.StringField{}

	result.PodAffinityTopology = getPodAffinityTopologyRules()
	result.PodAffinitySchedulingRule = getPodAffinitySchedulingRules()

	result.AutoDeletionTimeAfterCompletionSeconds = getAutoDeletionTimeAfterCompletionSeconds()

	return result
}

func getPodAffinityTopologyRules() *flds.StringField {
	return &flds.StringField{
		Rules: &flds.StringRules{
			Format: ptr.To(workload_fields.FormatTopologyKey),
		},
	}
}

func getPodAffinitySchedulingRules() *flds.StringField {
	return &flds.StringField{
		Value: ptr.To(constants.RequiredDuringSchedulingIgnoredDuringExecutionValue),
		Rules: &flds.StringRules{
			Options: &[]flds.OptionRule{
				{
					Name:  ptr.To(constants.RequiredDuringSchedulingIgnoredDuringExecutionUiName),
					Value: ptr.To(constants.RequiredDuringSchedulingIgnoredDuringExecutionValue),
				},
				{
					Name:  ptr.To(constants.PreferredDuringSchedulingIgnoredDuringExecutionUiName),
					Value: ptr.To(constants.PreferredDuringSchedulingIgnoredDuringExecutionValue),
				},
			},
		},
	}
}

func getMigProfileRules() *flds.StringField {
	return &flds.StringField{
		Rules: &flds.StringRules{
			Options: &[]flds.OptionRule{
				{
					Value: ptr.To(fmt.Sprintf("%v%v", constants.AnnotationMigProfilePrefix, "1g.5gb")),
					Name:  ptr.To("1g.5gb"),
				},
				{
					Value: ptr.To(fmt.Sprintf("%v%v", constants.AnnotationMigProfilePrefix, "1g.10gb")),
					Name:  ptr.To("1g.10gb"),
				},
				{
					Value: ptr.To(fmt.Sprintf("%v%v", constants.AnnotationMigProfilePrefix, "2g.10gb")),
					Name:  ptr.To("2g.10gb"),
				},
				{
					Value: ptr.To(fmt.Sprintf("%v%v", constants.AnnotationMigProfilePrefix, "2g.20gb")),
					Name:  ptr.To("2g.20gb"),
				},
				{
					Value: ptr.To(fmt.Sprintf("%v%v", constants.AnnotationMigProfilePrefix, "3g.20gb")),
					Name:  ptr.To("3g.20gb"),
				},
				{
					Value: ptr.To(fmt.Sprintf("%v%v", constants.AnnotationMigProfilePrefix, "3g.40gb")),
					Name:  ptr.To("3g.40gb"),
				},
				{
					Value: ptr.To(fmt.Sprintf("%v%v", constants.AnnotationMigProfilePrefix, "4g.20gb")),
					Name:  ptr.To("4g.20gb"),
				},
				{
					Value: ptr.To(fmt.Sprintf("%v%v", constants.AnnotationMigProfilePrefix, "4g.40gb")),
					Name:  ptr.To("4g.40gb"),
				},
				{
					Value: ptr.To(fmt.Sprintf("%v%v", constants.AnnotationMigProfilePrefix, "7g.40gb")),
					Name:  ptr.To("7g.40gb"),
				},
				{
					Value: ptr.To(fmt.Sprintf("%v%v", constants.AnnotationMigProfilePrefix, "7g.80gb")),
					Name:  ptr.To("7g.80gb"),
				},
			},
			Format: ptr.To(workload_fields.FormatMigProfile),
		},
	}
}

// only for ui and maybe for submit ?
func getNodePoolRules(ctx context.Context, nodePoolLister NodePoolNameLister) *flds.StringField {
	names, err := nodePoolLister.ListSchedulableNodePoolNames(ctx)
	if err != nil || len(names) == 0 {
		if err != nil {
			log.Error(err, "failed to list nodepools in cluster")
		}
		return &flds.StringField{
			//RUNAI DEFAULT is None
			Rules: &flds.StringRules{
				Options: &[]flds.OptionRule{
					{
						Value: ptr.To(v1alpha1.DefaultNodePoolName),
						Name:  ptr.To(v1alpha1.DefaultNodePoolName),
					},
				},
			},
		}
	}

	options := make([]flds.OptionRule, 0, len(names))
	for _, name := range names {
		options = append(options, flds.OptionRule{
			Name:  ptr.To(name),
			Value: ptr.To(name),
		})
	}

	return &flds.StringField{
		//RUNAI DEFAULT is None
		Rules: &flds.StringRules{
			Options: &options,
		},
	}
}

func getAutoDeletionTimeAfterCompletionSeconds() *flds.IntField {
	return &flds.IntField{
		Rules: &flds.IntRules{
			Min: ptr.To(0),
		},
	}
}

// GetNodepoolsValue returns the final value of the used nodepools by the workload
// format: `nodepool-1 nodepool2`
func GetNodepoolsValue(schedulerFields RunaiSchedulerFields) string {
	if schedulerFields.NodePools.IsValueInitialized() {
		return schedulerFields.NodePools.Deref()
	} else if schedulerFields.NodePool.IsValueInitialized() { // deprecated
		return schedulerFields.NodePool.Deref()
	}

	return ""
}
