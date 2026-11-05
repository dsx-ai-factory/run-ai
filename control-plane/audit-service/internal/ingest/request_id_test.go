/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ingest

import (
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/audit-service/internal/generated/pkg/api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// The request id is stored as a UUID column, so a log whose request id is
// missing or is not a UUID has to get a generated one. Storing the zero UUID
// instead makes every such log share the same request id.
func TestParseToModelLogGeneratesARequestIdWhenMissing(t *testing.T) {
	for _, tc := range []struct {
		name      string
		requestId string
	}{
		{name: "empty request id", requestId: ""},
		{name: "request id is not a uuid", requestId: "not-a-uuid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			log := &api.AuditLog{RequestId: tc.requestId, TenantId: "7", Timestamp: time.Now()}

			parsed, err := ParseToModelLog(log)

			require.NoError(t, err)
			require.NotEqual(t, uuid.Nil, parsed.RequestId, "the stored request id is the zero UUID")
			require.Equal(t, parsed.RequestId.String(), log.RequestId, "the generated id is not the one that was stored")
		})
	}
}

func TestParseToModelLogKeepsAValidRequestId(t *testing.T) {
	requestId := uuid.New()
	log := &api.AuditLog{RequestId: requestId.String(), TenantId: "7", Timestamp: time.Now()}

	parsed, err := ParseToModelLog(log)

	require.NoError(t, err)
	require.Equal(t, requestId, parsed.RequestId)
}
