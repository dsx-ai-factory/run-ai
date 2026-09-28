/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/utils"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/spf13/pflag"
)

type Security struct {
	Capabilities             []string
	AllowPrivilegeEscalation bool
	Privileged               bool
	RunAsUser                bool
	RunAsUid                 int64
	RunAsGid                 int64
	RunAsNonRoot             bool
	UidGidSource             string
	SeccompProfile           string
	SupplementalGroups       []int
	HostIpc                  bool
	HostNetwork              bool

	// Distributed Inference (Worker)
	WorkerRunAsUid                 int
	WorkerRunAsGid                 int
	WorkerRunAsNonRoot             bool
	WorkerSupplementalGroups       []int
	WorkerCapabilities             []string
	WorkerPrivileged               bool
	WorkerAllowPrivilegeEscalation bool
	WorkerHostIpc                  bool
	WorkerHostNetwork              bool
}

func (s *Security) GenerateSecuritySpec() (*openapi.WorkspaceSpecSpecAllOfSecurity, error) {
	security := openapi.NewWorkspaceSpecSpecAllOfSecurity()
	fieldsSet := false

	if len(s.Capabilities) > 0 {
		var capList []openapi.Capability
		for _, c := range s.Capabilities {
			capability, err := openapi.NewCapabilityFromValue(c)
			if err != nil {
				return nil, err
			}
			capList = append(capList, *capability)
		}
		security.SetCapabilities(capList)
		fieldsSet = true
	}
	if s.AllowPrivilegeEscalation {
		security.SetAllowPrivilegeEscalation(s.AllowPrivilegeEscalation)
		fieldsSet = true
	}
	if s.Privileged {
		security.SetPrivileged(s.Privileged)
		fieldsSet = true
	}
	if s.RunAsUid > 0 {
		security.SetRunAsUid(s.RunAsUid)
		fieldsSet = true
	}
	if s.RunAsGid > 0 {
		security.SetRunAsGid(s.RunAsGid)
		fieldsSet = true
	}
	if s.RunAsNonRoot {
		security.SetRunAsNonRoot(s.RunAsNonRoot)
		fieldsSet = true
	}

	if s.UidGidSource == specset.EmptyString && (s.RunAsUid > 0 || s.RunAsGid > 0 || s.SupplementalGroups != nil) {
		s.UidGidSource = string(openapi.UIDGIDSOURCE_CUSTOM)
	}

	if s.UidGidSource != specset.EmptyString {
		uidGidSource, err := openapi.NewUidGidSourceFromValue(s.UidGidSource)
		if err != nil {
			return nil, err
		}
		security.SetUidGidSource(*uidGidSource)
		fieldsSet = true
	}

	if s.SeccompProfile != specset.EmptyString {
		sp, err := openapi.NewSeccompProfileTypeFromValue(s.SeccompProfile)
		if err != nil {
			return nil, err
		}
		security.SetSeccompProfileType(*sp)
		fieldsSet = true
	}

	if s.SupplementalGroups != nil {
		strSupplementaryGroups := utils.IntSliceToString(s.SupplementalGroups)
		security.SetSupplementalGroups(strSupplementaryGroups)
		fieldsSet = true
	}

	if s.HostIpc {
		security.SetHostIpc(s.HostIpc)
		fieldsSet = true
	}

	if s.HostNetwork {
		security.SetHostNetwork(s.HostNetwork)
		fieldsSet = true
	}

	if !fieldsSet {
		return nil, nil
	}

	return security, nil
}

func SecurityFlagSet(flagsSecurity *pflag.FlagSet, specSecurity *Security, workloadType spectype.Workload, framework ...api.Framework) {
	if workloadType == spectype.NimService {
		flagsSecurity.Int64Var(&specSecurity.RunAsUid, consts.AttributeNameRunAsUid, 0, consts.AttributeUsageRunAsUid)
		flagsSecurity.Int64Var(&specSecurity.RunAsGid, consts.AttributeNameRunAsGid, 0, consts.AttributeUsageRunAsGid)
		return
	}

	if workloadType != spectype.Submit {
		flagsSecurity.StringArrayVar(&specSecurity.Capabilities, consts.AttributeNameCapability, []string{}, consts.AttributeUsageCapability)
	} else {
		flagsSecurity.StringArrayVar(&specSecurity.Capabilities, consts.AttributeBCNameAddCapability, []string{}, consts.AttributeUsageCapability)
	}

	if workloadType != spectype.Inference && workloadType != spectype.InferenceDistributed {
		flagsSecurity.BoolVar(&specSecurity.AllowPrivilegeEscalation, consts.AttributeNameAPE, false, consts.AttributeUsageAPE)
	}
	flagsSecurity.BoolVar(&specSecurity.HostNetwork, consts.AttributeNameHostNetwork, false, consts.AttributeUsageHostNetwork)
	flagsSecurity.BoolVar(&specSecurity.HostIpc, consts.AttributeNameHostIpc, false, consts.AttributeUsageHostIpc)

	flagsSecurity.BoolVar(&specSecurity.RunAsUser, consts.AttributeNameRunAsUser, false, consts.AttributeUsageRunAsUser)
	flagsSecurity.Int64Var(&specSecurity.RunAsUid, consts.AttributeNameRunAsUid, 0, consts.AttributeUsageRunAsUid)
	flagsSecurity.Int64Var(&specSecurity.RunAsGid, consts.AttributeNameRunAsGid, 0, consts.AttributeUsageRunAsGid)
	flagsSecurity.BoolVar(&specSecurity.RunAsNonRoot, consts.AttributeNameRunAsNonRoot, false, consts.AttributeUsageRunAsNonRoot)
	flagsSecurity.StringVar(&specSecurity.UidGidSource, consts.AttributeNameUidGidSource, "", consts.AttributeUsageUidGidSource)
	flagsSecurity.StringVar(&specSecurity.SeccompProfile, consts.AttributeNameSeccompProfile, "", consts.AttributeUsageSeccompProfile)
	flagsSecurity.IntSliceVar(&specSecurity.SupplementalGroups, consts.AttributeNameSupplementalGroups, nil, consts.AttributeUsageSupplementalGroups)
	flagsSecurity.BoolVar(&specSecurity.Privileged, consts.AttributeNamePrivileged, false, consts.AttributeUsagePrivileged)

	// Add worker-specific security flags for distributed inference
	if workloadType == spectype.InferenceDistributed {
		DistributedInferenceSecurityFlagSet(flagsSecurity, specSecurity)
	}
}

func ValidateSecurityRunAsUserFlags(specSecurity *Security) error {
	errorMsg := fmt.Errorf(consts.FlagsValidationRunAsUser, consts.AttributeNameRunAsUid, consts.AttributeNameRunAsGid,
		consts.AttributeNameUidGidSource, consts.AttributeNameSupplementalGroups)

	if specSecurity.RunAsUser && (specSecurity.RunAsUid > 0 ||
		specSecurity.RunAsGid > 0 || specSecurity.UidGidSource != "" ||
		len(specSecurity.SupplementalGroups) > 0) {
		return errorMsg
	}

	return nil
}
