/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	runaictx_http "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"

	. "github.com/onsi/gomega"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
)

const PermittedScopesKey = "permitted-scopes"

func CleanupCrudTable() {
	gormDb := runaidb.DB()
	gormDb.Unscoped().Delete(&models_crud.WorkloadSyncUpdate{}, "1 = 1")
	gormDb.Unscoped().Delete(&models_crud.Workload{}, "1 = 1")
}

func RestartCrudUpdateCounter() {
	runaidb.DB().Exec(fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH 1", "crud_workloads_sequence"))
}

func GetProjectResponse(randoms testutils.RandomValues) openapi.Project {
	return openapi.Project{
		Id:   randoms.ProjectIdStr,
		Name: "project-1",
		Parent: *openapi.NewNullableDepartmentProjectParent(&openapi.DepartmentProjectParent{
			Id:   randoms.DepartmentId,
			Name: randoms.DepartmentId,
		}),
		ParentId:  randoms.DepartmentId,
		ClusterId: randoms.ClusterId,
	}
}

func headerForAdmin(randoms *testutils.RandomValues) *http.Header {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, randoms.TenantIdStr)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "admin")
	return runaictx_http.GetHeaders(ctx)
}

// makePayload prepares a go object to be sent as a payload to REST call
func makePayload(payloadObject interface{}) *strings.Reader {
	payloadString, err := json.Marshal(payloadObject)
	Expect(err).To(BeNil())
	return strings.NewReader(string(payloadString))
}
