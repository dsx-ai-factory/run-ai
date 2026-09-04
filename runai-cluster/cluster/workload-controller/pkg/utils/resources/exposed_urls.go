/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	commonutils "github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	"k8s.io/utils/ptr"
)

func HasPrivateExposedUrl(exposedUrls map[string]flds.ExposedUrlItemAndRules) bool {
	// Verify if any authorized user is specified.
	hasPrivateExposedUrlBool := false
	for _, nick := range commonutils.SortedKeys(exposedUrls) {
		exposedUrl := exposedUrls[nick]
		// entry's value can be nil (e.g. deleted)
		if exposedUrl.Value == nil {
			continue
		}
		authorizedUsers := ptr.Deref(exposedUrl.Value.AuthorizedUsers, "")
		if authorizedUsers != "" {
			hasPrivateExposedUrlBool = true
			break // Exit the loop since we found a non-empty authorizedUsers
		}
	}

	return hasPrivateExposedUrlBool
}
