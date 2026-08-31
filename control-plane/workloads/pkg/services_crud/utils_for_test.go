/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/run-ai/runai/runai-common-packages/runaictx"
	runaictx_http "github.com/run-ai/runai/runai-common-packages/runaictx/http"

	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-application-packages/testutils"
	openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"

	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models_crud"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
)

const PermittedScopesKey = "permitted-scopes"

func CleanupCrudTable() {
	gormDb := dao_utils.DB()
	gormDb.Unscoped().Delete(&models_crud.WorkloadSyncUpdate{}, "1 = 1")
	gormDb.Unscoped().Delete(&models_crud.Workload{}, "1 = 1")
}

func RestartCrudUpdateCounter() {
	dao_utils.DB().Exec(fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH 1", "crud_workloads_sequence"))
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
