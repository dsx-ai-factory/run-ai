/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runai_rules

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

// PatchPolicyWithRunaiDefaults patch the policy with default values from runai rules.
// typically the builder does it, but for some cases it is easier to simply copy them from runai rules to the policy
func PatchPolicyWithRunaiDefaults(ctx context.Context, runaiContainerFields *field_groups.ContainerFields, policyContainerFields *field_groups.ContainerFields,
	runaiAclFields *field_groups.AccessControlFields, policyAclFields *field_groups.AccessControlFields, policyRunaiSchedulerFields *field_groups.RunaiSchedulerFields) {

	if policyContainerFields != nil {
		if runaiContainerFields != nil && runaiContainerFields.RunAsUser != nil {
			if policyContainerFields.RunAsUser == nil {
				policyContainerFields.RunAsUser = runaiContainerFields.RunAsUser
			} else if policyContainerFields.RunAsUser.Value == nil {
				policyContainerFields.RunAsUser.Value = runaiContainerFields.RunAsUser.Value
			}
		}
	}

	if policyAclFields != nil {
		if runaiAclFields != nil && runaiAclFields.AllowPrivilegeEscalation != nil {
			if policyAclFields.AllowPrivilegeEscalation == nil {
				policyAclFields.AllowPrivilegeEscalation = runaiAclFields.AllowPrivilegeEscalation
			} else if policyAclFields.AllowPrivilegeEscalation.Value == nil {
				policyAclFields.AllowPrivilegeEscalation.Value = runaiAclFields.AllowPrivilegeEscalation.Value
			}
		}
	}

	// node affinity is not supported in cluster policies - ignore it
	if policyRunaiSchedulerFields != nil && policyRunaiSchedulerFields.NodeAffinityRequired != nil {
		log := logger.FromContext(ctx)
		log.V(4).Info("Ignoring node affinity value provided via policy")

		policyRunaiSchedulerFields.NodeAffinityRequired = nil
	}
}
