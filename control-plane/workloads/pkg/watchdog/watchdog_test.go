/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package watchdog

import (
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	ginkgoTesting "testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
)

const (
	TrainingType    = "Training"
	DefaultK8sPhase = "Running"
	TenantId        = 1
)

func TestWatchDog(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)

	dao_utils.SetPostgresEnv("", "file://../database/migrations")
	config.Init()
	conf := config.Database()
	conf.SkipDefaultTransaction = true
	dao_utils.OpenPostgresForTesting(t, conf)

	RunSpecs(t, "Watchdog tests")
	t.Cleanup(func() {
		testing.CleanTables()
	})
}

var _ = Describe("Delete Stuck Workloads tests", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})

	workloadDao := dao.NewWorkloadsDao(runai_config.Get[*config.ServerConfig]())

	AfterEach(func() {
		testing.CleanTables()
	})

	It("Delete stuck workloads", func() {
		// var workloads []*models.Workload
		workloads := make(map[api.Phase]models.Workload)
		workload := composeWorkload(api.PHASE_CREATING, services.ControlPlaneCreatingMessage)
		workloads[api.PHASE_CREATING] = workload
		workload = composeWorkload(api.PHASE_RUNNING, "message")
		workloads[api.PHASE_RUNNING] = workload

		for _, w := range workloads {
			upsertErr := workloadDao.DbUpsertWorkload(ctx, w)
			Expect(upsertErr).To(BeNil())
		}
		// Ensure upset is ok
		results, apiErr := workloadDao.DbGetWorkloads(ctx, TenantId, nil, false, 0, 0, "asc", "phase", nil, "")
		Expect(apiErr).To(BeNil())
		Expect(len(*results)).To(Equal(2))
		expectWorkload(&(*results)[0], api.PHASE_CREATING, services.ControlPlaneCreatingMessage, true)
		expectWorkload(&(*results)[1], api.PHASE_RUNNING, "message", true)

		// Ensure that nothing changed if no time passed
		runWatchDog(dao_utils.DB(), "30 seconds", int64(5*1000))
		expectWorkload(&(*results)[0], api.PHASE_CREATING, services.ControlPlaneCreatingMessage, true)
		expectWorkload(&(*results)[1], api.PHASE_RUNNING, "message", true)

		// Wait for 5 sec, reducing the max "stuck duration" to 3 seconds, to allow faster test time.
		// however this may cause flakiness in the test. in such case suggest to inc the sleep duration.
		// Ran the test 100 times locally and didn't encounter any flakiness.
		time.Sleep(5 * time.Second)
		runWatchDog(dao_utils.DB(), "3 seconds", int64(5*1000))

		// Ensure that the stuck in creating workload was set to failed
		results, apiErr = workloadDao.DbGetWorkloads(ctx, TenantId, nil, false, 0, 0, "asc", "phase", nil, "")
		Expect(apiErr).To(BeNil())
		Expect(len(*results)).To(Equal(2))
		expectWorkload(&(*results)[0], api.PHASE_FAILED, services_utils.WorkloadFailedPhaseMsg_StuckInCreation, false)
		expectWorkload(&(*results)[1], api.PHASE_RUNNING, "message", true)
	})
})

func composeWorkload(phase api.Phase, phaseMessage string) models.Workload {
	id, err := uuid.NewRandom()
	Expect(err).To(BeNil())
	now := time.Now()
	return models.Workload{
		ID:                id,
		TenantID:          TenantId,
		UpdatedAt:         now,
		Phase:             phase,
		PhaseMessage:      phaseMessage,
		PhaseUpdatedAt:    now,
		Urls:              []string{"http://test.com"},
		CreatedAt:         time.Now().Add(-3 * time.Second),
		SubmittedBy:       testing.TestUser,
		Name:              "name",
		Source:            "source",
		Type:              TrainingType,
		ClusterID:         id,
		ProjectName:       "project",
		ProjectID:         "project-id",
		Namespace:         "namespace",
		SupportDelete:     true,
		SupportSuspend:    true,
		K8sPhase:          DefaultK8sPhase,
		K8sPhaseUpdatedAt: now,
		AdditionalFields:  models.AdditionalFields{},
	}
}

func expectWorkload(actual *models.WorkloadView, expectedPhase api.Phase, expectPhaseMessage string, expectSupportSuspend bool) {
	Expect(actual.Phase).To(Equal(expectedPhase))
	Expect(actual.PhaseMessage).To(Equal(expectPhaseMessage))
	Expect(actual.SupportSuspend).To(Equal(expectSupportSuspend))
}
