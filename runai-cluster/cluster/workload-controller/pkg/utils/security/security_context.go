/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package security

import (
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	"k8s.io/utils/ptr"

	corev1 "k8s.io/api/core/v1"
)

func GetSecurityContext(containerFields *field_groups.ContainerFields,
	aclFields *field_groups.AccessControlFields,
	isJupyter bool,
	envs *[]corev1.EnvVar) *corev1.SecurityContext {

	sc := corev1.SecurityContext{
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileType(containerFields.SeccompProfileType.DerefOrDefault(constants.SeccompProfileTypeRuntimeDefault)),
		},
	}

	//
	//    if "runAsUser" option was chosen in the UI, add the uid/gid to the security context
	//    of the container. for users running in ldap mode we also set the LDAP variables
	//
	if containerFields.RunAsUser.Deref() && !(isJupyter && workloadconfig.JupyterUserIdentity()) {
		if containerFields.RunAsUid.Deref32default(-1) != -1 {
			*envs = append(*envs, corev1.EnvVar{
				Name:  "LDAP_UID",
				Value: fmt.Sprintf("%v", containerFields.RunAsUid.Deref()),
			})
		}
		if containerFields.RunAsGid.Deref32default(-1) != -1 {
			*envs = append(*envs, corev1.EnvVar{
				Name:  "LDAP_GID",
				Value: fmt.Sprintf("%v", containerFields.RunAsGid.Deref()),
			})
		}
	}

	if aclFields.AllowPrivilegeEscalation.IsValueInitialized() {
		sc.AllowPrivilegeEscalation = ptr.To(aclFields.AllowPrivilegeEscalation.Deref())
	}

	if aclFields.Privileged.IsValueInitialized() {
		sc.Privileged = ptr.To(aclFields.Privileged.Deref())
	}

	if runAsNonRoot := containerFields.RunAsNonRoot.Deref(); runAsNonRoot {
		sc.RunAsNonRoot = ptr.To(true)
	}

	if readOnlyRootFilesystem := containerFields.ReadOnlyRootFilesystem.Deref(); readOnlyRootFilesystem {
		sc.ReadOnlyRootFilesystem = ptr.To(true)
	}

	//
	//   with JupyterUserIdentity, jupyter uid/gid is handled by passing environment variables NB_UID/NB_GID to the container
	//   but the pod itself has to run as root, in this case.
	//
	if containerFields.RunAsUser.Deref() && isJupyter && workloadconfig.JupyterUserIdentity() {
		sc.RunAsUser = ptr.To(int64(0))
		return &sc
	}

	if containerFields.RunAsUser.Deref() {
		if containerFields.RunAsUid != nil && containerFields.RunAsUid.Value != nil {
			sc.RunAsUser = ptr.To(int64(containerFields.RunAsUid.Deref()))
		}
		if containerFields.RunAsGid != nil && containerFields.RunAsGid.Value != nil {
			sc.RunAsGroup = ptr.To(int64(containerFields.RunAsGid.Deref()))
		}
	}

	//
	//   if spec specifies set of capabilities, need to add it to the spec
	//
	if len(containerFields.Capabilities.Deref()) != 0 || containerFields.DropAllCapabilities.Deref() {
		sc.Capabilities = &corev1.Capabilities{}

		if containerFields.DropAllCapabilities.Deref() {
			sc.Capabilities.Drop = []corev1.Capability{"all"}
		}

		if len(containerFields.Capabilities.Deref()) != 0 {
			capabilitiesMap := containerFields.Capabilities.Deref()

			for _, nick := range field_types.SortedCapabilitiesMapKeys(capabilitiesMap) {
				capability := capabilitiesMap[nick].Value
				if capability != nil {
					sc.Capabilities.Add = append(sc.Capabilities.Add, corev1.Capability(ptr.Deref(capability.Capability, "")))
				}
			}
		}
	}
	return &sc
}

func GetPodSecurityContext(containerFields *field_groups.ContainerFields, isJupyter bool, ps *corev1.PodSpec) error {

	if containerFields.RunAsUser.Deref() {
		if containerFields.SupplementalGroups.Deref() != "" {
			if ps.SecurityContext == nil {
				ps.SecurityContext = &corev1.PodSecurityContext{}
			}
			groups, err := utils.DecodeInt64List(containerFields.SupplementalGroups.Deref(), common.GroupsSeparator)
			if err != nil {
				return err
			}
			if ps.SecurityContext.SupplementalGroups == nil {
				ps.SecurityContext.SupplementalGroups = make([]int64, 0, len(groups))
			}
			ps.SecurityContext.SupplementalGroups = append(ps.SecurityContext.SupplementalGroups, groups...)
		}

		if isJupyter && workloadconfig.JupyterUserIdentity() {
			return nil
		}

		if ps.SecurityContext == nil {
			ps.SecurityContext = &corev1.PodSecurityContext{}
		}

		if containerFields.RunAsGid != nil && containerFields.RunAsGid.Value != nil {
			ps.SecurityContext.FSGroup = ptr.To(int64(containerFields.RunAsGid.Deref()))
		}

	}

	return nil
}
