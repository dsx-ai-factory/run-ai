/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dinf_to_lws

import (
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

func (dinfComposer *DistributedInferenceToLWSComposer) composeTopology(
	workers *int32,
	worker *workloadsApi.DistributedInferenceLeaderWorkerSpec,
	target **int32,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		workerCount := convutils.Deref(workers, int32(0))
		if workerCount == 0 && worker != nil && !composer.IsEmpty(*worker) {
			return workloadsApi.NewError(http.StatusBadRequest, "worker spec requires workers field to be set to a positive value")
		}
		size := workerCount + 1
		return dinfComposer.fc.CopyInt(&size, target, composer.IntToInt[int64], nil)()
	}
}

func (dinfComposer *DistributedInferenceToLWSComposer) composeStartupPolicy(
	src *workloadsApi.DistributedInferenceStartupPolicy,
	target *lwsv1.StartupPolicyType,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		policy := convutils.Deref(src, workloadsApi.DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_CREATED)
		if policy != workloadsApi.DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_CREATED &&
			policy != workloadsApi.DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_READY {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("invalid startup policy: %s", policy))
		}
		return startupPolicyToLWS(convutils.Ref(string(policy)), target, nil)
	}
}

func (dinfComposer *DistributedInferenceToLWSComposer) composeRestartPolicy(
	src *workloadsApi.DistributedInferenceRestartPolicy,
	target *lwsv1.RestartPolicyType,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		policy := convutils.Deref(src, workloadsApi.DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART)
		if policy != workloadsApi.DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART &&
			policy != workloadsApi.DISTRIBUTEDINFERENCERESTARTPOLICY_NONE {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("invalid restart policy: %s", policy))
		}
		return restartPolicyToLWS(convutils.Ref(string(policy)), target, nil)
	}
}
