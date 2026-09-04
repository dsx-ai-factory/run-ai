/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/database/models"
	metricerrors "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/errors"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/integdb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInit(t *testing.T) {
	RegisterFailHandler(Fail)

	conf, cleanup, err := integdb.SetupPostgresDB(&integdb.IntegdbConfig{
		Schema:           "metrics",
		MigrationDirPath: "file://../migrations",
	})
	if err != nil {
		t.Fatalf("failed to set up test postgres: %v", err)
	}
	t.Cleanup(cleanup)

	if err := runaidb.Init(conf); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}
	t.Cleanup(func() { _ = runaidb.Close() })

	RunSpecs(t, "Metric stores dao tests")
}

var _ = Describe("MetricStoresDao", func() {
	var (
		ctx             context.Context
		metricStoresDao *MetricStoresDao
		metricStore     models.MetricStore
	)

	BeforeEach(func() {
		ctx = context.WithValue(
			context.WithValue(context.Background(), runaictx.CfTenantId, "1"),
			runaictx.CfSubjectID,
			"user",
		)
		metricStoresDao = NewMetricStoresDao()
		endpoint := models.MetricStoreEndpoint{
			Url: "some_url",
			Auth: &models.MetricStoreEndpointAuth{
				Type: models.ATBasic,
				Basic: &models.AuthBasic{
					Username: "test",
					Password: "secret",
				},
			},
		}
		metricStore = models.MetricStore{
			ID:           uuid.New(),
			TenantId:     rand.New(rand.NewSource(time.Now().UnixNano())).Int31(),
			DatasourceId: "mimir_12345",
			MetricStoreDetails: models.MetricStoreDetails{
				Type:  models.MSTMimir,
				Read:  endpoint,
				Write: endpoint,
			},
		}
	})

	AfterEach(func() {
		_ = metricStoresDao.Delete(ctx, metricStore.TenantId)
	})

	It("returns ErrMetricStoreAlreadyExists when creating a duplicate tenant metric store", func() {
		created, err := metricStoresDao.Create(ctx, metricStore)
		Expect(err).To(BeNil())
		Expect(created).ToNot(BeNil())
		Expect(created.ID).ToNot(BeNil())

		duplicate, err := metricStoresDao.Create(ctx, metricStore)
		Expect(err).To(HaveOccurred())
		Expect(duplicate).To(BeNil())
		Expect(errors.Is(err, metricerrors.ErrMetricStoreAlreadyExists)).To(BeTrue())
		Expect(err.Error()).To(Equal("metric store already exists"))
	})
})
