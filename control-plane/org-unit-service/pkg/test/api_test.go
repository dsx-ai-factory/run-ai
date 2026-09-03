/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	org_unit_config "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/settings"
	tm_client_mock "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
	mocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/mocks"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var mockTenantsClient *tm_client_mock.ITenantsManager

func TestDao(t *testing.T) {
	RegisterFailHandler(Fail)

	// Register this test to prevent premature cleanup
	RegisterTest()

	MockEnv(t)
	SharedOpenPostgresForTesting(t)
	// Initialize the settings package against the test DB so the migrated LimitQuotaOverSubscription
	// setting can be read/written by tests (defaults to disabled when no row exists).
	setting_pkg.Init(&setting_pkg.SettingConfig{}, nil)
	_ = setting_pkg.Register[*settings.LimitQuotaOverSubscription]()
	metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{})
	SetAuthMock()
	mockTenantsClient = tm_client_mock.NewITenantsManager(t)
	pubSuber := mocks.PubSuber{}
	pubSuber.On("Publish", mock.Anything, mock.Anything).Return(nil)
	pubSuber.On("SubscribeAsync", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	runaipubsub.SetMock(&pubSuber)
	RunSpecs(t, "Api tests")

	t.Cleanup(func() {
		UnregisterTest()
	})
}

func MockEnv(t *testing.T) {
	config.Set(&org_unit_config.OvertimeMetrics{
		Query24hRangeIntervalSeconds:     1,
		Query7dRangeIntervalSeconds:      2,
		Query30dRangeIntervalSeconds:     3,
		Query24hRangeInitialSleepSeconds: 1,
		Query7dRangeInitialSleepSeconds:  1,
		Query30dRangeInitialSleepSeconds: 1,
		RateLimiterLimit:                 10,
		RateLimiterBurst:                 10,
	})
}
