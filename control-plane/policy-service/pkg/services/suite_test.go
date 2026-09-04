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

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	runaictx_http "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/integdb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
)

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)

	SupportTenantMode = true

	config.MockConfig()
	config.Init()

	authClient.SetMockAuthorizationClient(auth.SimpleMockAuthService{})

	conf, cleanup, err := integdb.SetupPostgresDB(&integdb.IntegdbConfig{
		Schema:           "policies",
		MigrationDirPath: "file://../database/migrations",
	})
	if err != nil {
		t.Fatalf("failed to set up test postgres: %v", err)
	}
	t.Cleanup(cleanup)

	conf.UseServiceSession = new(bool)
	conf.LogLevel = "Debug"
	runaidb.SetLogger(dbgLogger{})
	if err := runaidb.Init(conf); err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	t.Cleanup(func() { _ = runaidb.Close() })

	dao.HandleSystemPolicy()

	HandleSystemMapping()

	RunSpecs(t, "Services Suite")
}

type dbgLogger struct{}

func (dbgLogger) Infof(f string, a ...any)  { fmt.Printf("[db] "+f+"\n", a...) }
func (dbgLogger) Errorf(f string, a ...any) { fmt.Printf("[db] "+f+"\n", a...) }
func (dbgLogger) Warnf(f string, a ...any)  { fmt.Printf("[db] "+f+"\n", a...) }
func (dbgLogger) Debugf(f string, a ...any) { fmt.Printf("[db] "+f+"\n", a...) }

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
