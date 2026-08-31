/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"testing"

	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dao"
	runaictx_http "github.com/run-ai/runai/runai-common-packages/runaictx/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authClient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/config"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)

	SupportTenantMode = true

	config.MockConfig()
	config.Init()

	authClient.SetMockAuthorizationClient(auth.SimpleMockAuthService{})

	//
	//   we run the service test with "real" database, so need to start it prior
	//   to running any services related test
	//
	dao_utils.OpenPostgresForTesting(t, config.Database())

	dbConf := config.Database()
	dbConf.MigrationCommand = "upmost"
	dao_utils.Migrate(dbConf)

	dao.HandleSystemPolicy()

	// Seed the system formatters / field groups / mappings ONCE for the whole suite.
	// `database.DeleteAllTablesForTesting()` preserves system-scoped rows between specs,
	// so we don't need to re-seed in BeforeEach (which previously did 17 GET+CREATE+PATCH
	// round-trips per spec in the GVK Policy Service tests).
	HandleSystemMapping()

	RunSpecs(t, "Services Suite")

	// No `downmost` migration teardown here: the testcontainers Postgres started by
	// `OpenPostgresForTesting` is destroyed at suite end, so re-running every migration
	// down would just be wasted seconds against a container that's about to be killed.
}

func tenantIdStr(tenantId int32) string {
	return fmt.Sprintf("%d", tenantId)
}

func requester(role string, tenantId int32) string {
	return fmt.Sprintf("%s-of-tenant-%d", role, tenantId)
}

func headerWithMissingInfo() *http.Header {
	return &http.Header{}
}

func headerForAdmin(tenantId int32) *http.Header {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "admin")
	return runaictx_http.GetHeaders(ctx)
}

// makePayload prepares a go object to be sent as a payload to REST call
func makePayload(payloadObject interface{}) *strings.Reader {
	payloadString, err := json.Marshal(payloadObject)
	Expect(err).To(BeNil())
	return strings.NewReader(string(payloadString))
}

// letters used for creating a random string of given length
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func tenantContextForTest(tenantId int32) context.Context {
	return context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
}
