/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package webhook

import (
	"context"

	"github.com/golang/glog"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	v1 "k8s.io/api/authentication/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetWorkloadControllerUID(ctx context.Context, k8sClient client.Client) (string, error) {
	workloadControllerServiceAccount, err := GetServiceAccount(
		ctx, k8sClient,
		constants.WorkloadControllerOperandName, constants.NamespaceRunAi)
	if err != nil {
		glog.Error(err, "Failed getting workload-controller service account")
		return "", err
	}

	return string(workloadControllerServiceAccount.UID), nil
}

func GetRequesterUserId(admissionRequestUserInfo v1.UserInfo) string {
	if admissionRequestUserInfo.UID != "" {
		return admissionRequestUserInfo.UID
	}
	return admissionRequestUserInfo.Username
}

func GetRuntimeRulesUID(ctx context.Context, k8sClient client.Client) (string, error) {
	runtimeRulesServiceAccount, err := GetServiceAccount(
		ctx, k8sClient,
		constants.RuntimeRulesOperandName, constants.NamespaceRunAi)
	if err != nil {
		glog.Error(err, "Failed getting runtime-rules service account")
		return "", err
	}

	return string(runtimeRulesServiceAccount.UID), nil
}
