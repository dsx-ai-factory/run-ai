/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jobs

import (
	"context"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func IndexAccessRules(mgr manager.Manager) error {
	err := mgr.GetFieldIndexer().IndexField(
		context.Background(), &v1alpha1.AccessRule{},
		common.PrivilegedWorkloadCrudPermissionsField, common.PrivilegedWorkloadCrudPermissionsIndexer)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed indexing accessrule field %s",
			common.PrivilegedWorkloadCrudPermissionsField))
		return err
	}

	return nil
}
