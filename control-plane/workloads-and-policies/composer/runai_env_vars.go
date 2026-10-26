/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"fmt"
	"strconv"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
)

const (
	LdapUidEnvVar = "LDAP_UID"
	LdapGidEnvVar = "LDAP_GID"

	DefaultJupyterPort = 8888
	JupyterToolType    = "jupyter-notebook"
	JupyterPortEnvVar  = "JUPYTER_PORT"
)

func RunaiEnvVars(projectName string, workloadName string, jupyterContainerPort int, uidGid *UidGid) []workloadsApi.EnvironmentVariable {
	result := make([]workloadsApi.EnvironmentVariable, 0)

	// add run:ai internal environment variables
	result = append(result, workloadsApi.EnvironmentVariable{
		Name:  *workloadsApi.NewNullableString(convutils.Ref("RUNAI_PROJECT")),
		Value: *workloadsApi.NewNullableString(convutils.Ref(projectName)),
	})
	result = append(result, workloadsApi.EnvironmentVariable{
		Name:  *workloadsApi.NewNullableString(convutils.Ref("RUNAI_JOB_NAME")),
		Value: *workloadsApi.NewNullableString(convutils.Ref(workloadName)),
	})

	if jupyterContainerPort != 0 {
		result = append(result, workloadsApi.EnvironmentVariable{
			Name:  *workloadsApi.NewNullableString(convutils.Ref(JupyterPortEnvVar)),
			Value: *workloadsApi.NewNullableString(convutils.Ref(strconv.Itoa(jupyterContainerPort))),
		})
	}

	if jupyterContainerPort == 0 || !workloadconfig.JupyterUserIdentity() {
		if uidGid.Uid != nil {
			result = append(result, workloadsApi.EnvironmentVariable{
				Name:  *workloadsApi.NewNullableString(convutils.Ref(LdapUidEnvVar)),
				Value: *workloadsApi.NewNullableString(convutils.Ref(fmt.Sprintf("%d", *uidGid.Uid))),
			})
		}
		if uidGid.Gid != nil {
			result = append(result, workloadsApi.EnvironmentVariable{
				Name:  *workloadsApi.NewNullableString(convutils.Ref(LdapGidEnvVar)),
				Value: *workloadsApi.NewNullableString(convutils.Ref(fmt.Sprintf("%d", *uidGid.Gid))),
			})
		}
	}

	return result
}
