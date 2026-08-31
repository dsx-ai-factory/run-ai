/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dinf_to_lws

import (
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	corev1 "k8s.io/api/core/v1"
)

func (dinfComposer *DistributedInferenceToLWSComposer) processSecurityFields(
	src *workloadsApi.DistributedInferenceLeaderWorkerSpecSecurity,
	target *corev1.PodTemplateSpec, uidGid *composer.UidGid,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		podSpec := &target.Spec
		container := &podSpec.Containers[0]
		if podSpec.SecurityContext == nil {
			podSpec.SecurityContext = &corev1.PodSecurityContext{}
		}
		if container.SecurityContext == nil {
			container.SecurityContext = &corev1.SecurityContext{}
		}

		runner := &composer.StepRunner{}
		runner.Batch(
			dinfComposer.fc.CopyBoolean(src.HostIpc.Get(), &podSpec.HostIPC, composer.BoolToBool, nil),
			dinfComposer.fc.CopyBoolean(src.HostNetwork.Get(), &podSpec.HostNetwork, composer.BoolToBool, nil),
			dinfComposer.fc.ComposeDnsPolicy(src.HostNetwork.Get(), &podSpec.DNSPolicy, composer.StringToString, nil),
			dinfComposer.fc.CopyBoolean(src.Privileged.Get(), &container.SecurityContext.Privileged, composer.BoolToBool, nil),
			dinfComposer.fc.CopyBoolean(src.ReadOnlyRootFilesystem.Get(), &container.SecurityContext.ReadOnlyRootFilesystem, composer.BoolToBool, nil),
			dinfComposer.fc.ComposeRunAsNonRoot(src.RunAsNonRoot.Get(), src.Privileged.Get(), &container.SecurityContext.RunAsNonRoot, composer.BoolToBool, nil),
			dinfComposer.fc.ComposeCapabilities(src.Capabilities, &container.SecurityContext.Capabilities, composer.MapToK8sCapabilities, nil),
			dinfComposer.fc.ComposeSeccompProfile(src.SeccompProfileType.Get(), nil, &container.SecurityContext.SeccompProfile, composer.StringToSeccompProfile, nil),
		)

		var runAsUser bool
		runner.Run(dinfComposer.fc.ComposeRunAsUser(composer.EnumToString(src.UidGidSource.Get()), &runAsUser, composer.BoolToBool, nil))

		if runAsUser {
			runner.Batch(
				dinfComposer.fc.CopyInt(src.RunAsUid.Get(), &container.SecurityContext.RunAsUser, composer.IntToInt, nil),
				dinfComposer.fc.CopyInt(src.RunAsGid.Get(), &container.SecurityContext.RunAsGroup, composer.IntToInt, nil),
				dinfComposer.fc.ComposeSupplementalGroups(src.SupplementalGroups.Get(), &podSpec.SecurityContext.SupplementalGroups, composer.StringToSupplementalGroups, nil),
				dinfComposer.fc.CopyInt(src.RunAsGid.Get(), &podSpec.SecurityContext.FSGroup, composer.IntToInt, nil),
			)

			uidGid.Uid = src.RunAsUid.Get()
			uidGid.Gid = src.RunAsGid.Get()
		}

		if composer.IsEmpty(*podSpec.SecurityContext) {
			podSpec.SecurityContext = nil
		}
		if composer.IsEmpty(*container.SecurityContext) {
			container.SecurityContext = nil
		}
		return runner.Error()
	}
}
