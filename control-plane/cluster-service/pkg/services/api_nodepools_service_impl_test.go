/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"

	"github.com/google/uuid"
	"net/http"
	"testing"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	authorizationUtils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"

	authorizationapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

const (
	tenantId    = 1
	tenantIdStr = "1"
	testUser    = "user"
)

func addAuditToCtx(ctx context.Context) context.Context {
	req, _ := http.NewRequest(http.MethodPost, "http://whatever.com", nil)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "1.1.1.1"

	al := audit.NewAuditLog().WithRequest(req).WithCtx(ctx)
	return audit.CtxWithAudit(ctx, al)
}

func createContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantIdStr)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, testUser)
	ctx = context.WithValue(ctx, authorizationUtils.PermittedScopesKey, &authorizationapi.PermittedScopes{
		Tenants: []string{tenantIdStr},
	})
	ctx = addAuditToCtx(ctx)
	return ctx
}

func TestNodepoolSetDefaults(t *testing.T) {
	RegisterFailHandler(Fail)

	nodepool := models.Nodepool{}
	nodepoolSetDefaults(&nodepool)

	Expect(nodepool.GpuPlacementStrategy).To(Equal(api.PLACEMENTSTRATEGY_BINPACK))
	Expect(nodepool.CpuPlacementStrategy).To(Equal(api.PLACEMENTSTRATEGY_BINPACK))
	Expect(*nodepool.GPUNetworkAccelerationDetection).To(Equal(api.NODEPOOLGPUNETWORKACCELERATIONDETECTION_AUTO))
}

func TestValidateLabels(t *testing.T) {
	RegisterFailHandler(Fail)
	ptrToString := ptr.To("abc")

	err := validateLabels(ptrToString, ptrToString, false)
	Expect(err).To(BeNil())

	err = validateLabels(nil, nil, false)
	Expect(err).To(Not(BeNil()))

	err = validateLabels(ptrToString, nil, false)
	Expect(err).To(Not(BeNil()))

	err = validateLabels(nil, ptrToString, false)
	Expect(err).To(Not(BeNil()))

	err = validateLabels(nil, nil, true)
	Expect(err).To(BeNil())

	err = validateLabels(ptrToString, ptrToString, true)
	Expect(err).To(Not(BeNil()))

	err = validateLabels(ptrToString, nil, true)
	Expect(err).To(Not(BeNil()))

	err = validateLabels(nil, ptrToString, true)
	Expect(err).To(Not(BeNil()))
}

func TestClusterFieldsUpdatedDetectsNetworkTopologyChange(t *testing.T) {
	RegisterFailHandler(Fail)

	topologyA := uuid.New()
	topologyB := uuid.New()

	left := models.Nodepool{NetworkTopologyId: &topologyA}
	right := models.Nodepool{NetworkTopologyId: &topologyB}
	Expect(clusterFieldsUpdated(left, right)).To(BeTrue())

	Expect(clusterFieldsUpdated(left, models.Nodepool{NetworkTopologyId: &topologyA})).To(BeFalse())

	// a detached topology differs from an attached one
	Expect(clusterFieldsUpdated(left, models.Nodepool{})).To(BeTrue())
}

func TestNodepoolDeleteValidations(t *testing.T) {
	RegisterFailHandler(Fail)

	err := nodepoolDeleteValidations(models.Nodepool{IsDefault: true})
	Expect(err).To(Not(BeNil()))

	err = nodepoolDeleteValidations(models.Nodepool{Phase: api.NODEPOOLPHASE_DELETING})
	Expect(err).To(Not(BeNil()))

	err = nodepoolDeleteValidations(models.Nodepool{Phase: api.NODEPOOLPHASE_DELETED})
	Expect(err).To(Not(BeNil()))
}
