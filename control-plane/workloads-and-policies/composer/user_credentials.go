/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"fmt"
	"net/http"
	"strings"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	workloadMgrApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
)

// ParseUserCredentialRef strips the "USERCREDENTIAL:" prefix and splits the
// remaining "<sourceSecret>[,<key>]" into its parts.
func ParseUserCredentialRef(value string) (sourceSecret, key string, err *workloadsApi.Error) {
	prefix := UserCredentialsPrefix + ":"
	if !strings.HasPrefix(value, prefix) {
		return "", "", workloadsApi.NewError(http.StatusBadRequest,
			fmt.Sprintf("invalid user credential reference: missing %q prefix", UserCredentialsPrefix))
	}

	rest := strings.TrimPrefix(value, prefix)
	if rest == "" {
		return "", "", workloadsApi.NewError(http.StatusBadRequest,
			"invalid user credential reference: empty source secret name")
	}

	if idx := strings.LastIndex(rest, ","); idx >= 0 {
		sourceSecret = rest[:idx]
		key = rest[idx+1:]
	} else {
		sourceSecret = rest
	}

	if sourceSecret == "" {
		return "", "", workloadsApi.NewError(http.StatusBadRequest,
			"invalid user credential reference: empty source secret name")
	}

	return sourceSecret, key, nil
}

// RecordUserCredSource appends sourceSecret to cfg.UserCredentials.Sources,
// deduping so each source appears once in the agent's annotation. Initializes
// the nullable wrapper on first call.
func RecordUserCredSource(cfg *workloadMgrApi.WorkloadV2Configuration, sourceSecret string) {
	if cfg == nil || sourceSecret == "" {
		return
	}
	if !convutils.HasValue(cfg.UserCredentials) {
		cfg.UserCredentials = *workloadMgrApi.NewNullableWorkloadV2UpdateConfigurationUserCredentials(
			&workloadMgrApi.WorkloadV2UpdateConfigurationUserCredentials{},
		)
	}
	pc := cfg.UserCredentials.Get()
	for _, s := range pc.Sources {
		if s.Ref == sourceSecret {
			return
		}
	}
	pc.Sources = append(pc.Sources, workloadMgrApi.WorkloadV2UpdateConfigurationUserCredentialSource{
		Ref: sourceSecret,
	})
}
