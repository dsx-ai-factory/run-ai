/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"fmt"
	"testing"

	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
	openapi "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
	org_unit_config "github.com/run-ai/runai/control-plane/org-unit-service/pkg/config"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/settings"
	tm_client_mock "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client/mocks"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
	mocks "github.com/run-ai/runai/runai-common-packages/runaipubsub/mocks"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting"
	"github.com/stretchr/testify/mock"
)

var mockTenantsClient *tm_client_mock.ITenantsManager

func TestDao(t *testing.T) {
	RegisterFailHandler(Fail)

	// Register this test to prevent premature cleanup
	RegisterTest()

	dao_utils.SetPostgresEnv("", "file://../database/migrations")
	MockEnv(t)
	dbConf := dao_utils.ConstructConfig(org_unit_config.Schema)
	dbConf.FullSaveAssociations = true
	dbConf.SkipDefaultTransaction = true
	logging.Infof("DB PORT <%s>", dbConf.Port)
	SharedOpenPostgresForTesting(t, dbConf)
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
	fmt.Printf("DB Port %s \n", dbConf.Port)
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
