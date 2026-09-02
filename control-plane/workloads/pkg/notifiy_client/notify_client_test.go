/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package notifiy_client

import (
	"context"
	"encoding/json"
	"fmt"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"net/http"
	"strconv"
	ginkgoTesting "testing"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	notifyclient "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
)

const (
	TrainingType          = "Training"
	WorkspaceType         = "Workspace"
	InferenceType         = "Inference"
	DefaultSubmittedBy    = "user@run.ai"
	DefaultProjectName    = "test"
	DefaultProjectId      = "1"
	AnotherProjectId      = "2"
	AnotherProjectName    = "test2"
	ProjectIdCluster2     = "3"
	ProjectNameCluster2   = "test3"
	DefaultDepartmentName = "dep-1"
	DefaultDepartmentId   = "1"
	DefaultPhase          = api.PHASE_RUNNING
	DefaultK8sPhase       = "Running"
	DefaultSource         = api.SOURCE_CLI

	NotifyClientIntervalMillis = 50
)

var DefaultCreationTime = time.Now()

func TestNotifyClient(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)

	dao_utils.SetPostgresEnv("", "file://../database/migrations")
	config.Init()
	conf := config.Database()
	conf.SkipDefaultTransaction = true
	mockConfigForNotify()
	dao_utils.OpenPostgresForTesting(t, conf)

	RunSpecs(t, "Notify client tests")
	t.Cleanup(func() {
		testing.CleanTables()
	})
}

var _ = Describe("Workloads API tests", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	workloadsService := services.NewWorkloadsService()
	eventsService := services.NewEventsService()

	checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
		WithCluster(testing.ClusterUuid1, testing.BuildDisplayedCluster(testing.ClusterUuid1, testing.TestDefaultClusterVersion, nil)))
	checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

	id1 := uuid.NewString()
	id2 := uuid.NewString()
	id3 := uuid.NewString()
	id4 := uuid.NewString()
	id5 := uuid.NewString()
	var workload1, workload2, workload3, workload4, workload5 api.WorkloadSyncFields
	var stopper chan struct{}
	var cancelCtx context.Context
	var cancel context.CancelFunc
	var notifyClient *NotifyClienterMock

	BeforeAll(func() {
		workload1 = setupBaseWorkload(id1, "workload1", TrainingType, DefaultProjectId, DefaultProjectName,
			api.PHASE_CREATING, "creating", testing.ClusterUuid1)
		workload2 = setupBaseWorkload(id2, "workload2", WorkspaceType, DefaultProjectId, DefaultProjectName,
			api.PHASE_CREATING, "creating", testing.ClusterUuid1)
		workload3 = setupBaseWorkload(id3, "workload3", TrainingType, AnotherProjectId, AnotherProjectName,
			api.PHASE_CREATING, "creating", testing.ClusterUuid1)
		workload4 = setupBaseWorkload(id4, "workload4", TrainingType, ProjectIdCluster2, ProjectNameCluster2,
			api.PHASE_CREATING, "creating", testing.ClusterUuid2)
		workload5 = setupBaseWorkload(id5, "workload5", InferenceType, ProjectIdCluster2, ProjectNameCluster2,
			api.PHASE_CREATING, "creating", testing.ClusterUuid2)
		syncTestWorkloads(ctx, workloadsService, []api.WorkloadSyncFields{workload1}, DefaultProjectId)
		syncTestWorkloads(ctx, workloadsService, []api.WorkloadSyncFields{workload3}, AnotherProjectId)
		syncTestWorkloads(ctx, workloadsService, []api.WorkloadSyncFields{workload4, workload5}, ProjectIdCluster2)
	})

	AfterEach(func() {
		close(stopper)
		cancel()
		notifyClient.ResetNotifiedNotifications()
	})

	AfterAll(func() {
		testing.CleanTables()
	})

	Context("Phase Notifier", func() {
		BeforeEach(func() {
			stopper = make(chan struct{})
			cancelCtx, cancel = context.WithCancel(context.Background())

			notifyClient = NewNotifyClienterMock()

			conf := runai_config.Get[*config.Notify]()
			initNotifyClienter(nil, conf, notifyClient)
			InitWorkloadPhaseNotifier(cancelCtx, stopper, conf, notifyClient)

			time.Sleep(NotifyClientIntervalMillis * time.Millisecond)
			notifyClient.ResetNotifiedNotifications()
			time.Sleep(NotifyClientIntervalMillis * time.Millisecond)
		})

		It("Sanity", func() {
			syncTestWorkloads(ctx, workloadsService, []api.WorkloadSyncFields{workload1, workload2}, DefaultProjectId)

			workload3.Phase = api.PHASE_DELETING
			workload3.PhaseMessage = ""
			upsertWorkload(ctx, workloadsService, workload3)
			workload4.Phase = api.PHASE_UPDATING //  a phase that is not reported in notifications
			upsertWorkload(ctx, workloadsService, workload4)

			time.Sleep(NotifyClientIntervalMillis * 3 * time.Millisecond)

			Expect(notifyClient.errors).To(BeEmpty())

			expectedNotifications := map[string]map[string][]*notifyclient.Notification{
				workloadStatusChangeCategory: {
					string(WorkloadStatusChangeCreating): {
						constructExpectedNotificationForWorkloadWithDescription(ctx, workloadsService, workload2,
							workloadStatusChangeCategory, string(WorkloadStatusChangeCreating),
							"Workload workload2 status changed to Creating",
							fmt.Sprintf("Workload workload2 status changed to Creating%sAdditional details: creating",
								lineSeparator), nil),
					},
					string(WorkloadStatusChangeDeleting): {
						constructExpectedNotificationForWorkloadWithDescription(ctx, workloadsService, workload3,
							workloadStatusChangeCategory, string(WorkloadStatusChangeDeleting),
							"Workload workload3 status changed from Creating to Deleting",
							"Workload workload3 status changed from Creating to Deleting", nil),
					},
				},
			}
			validateExpectedNotifications(expectedNotifications, notifyClient.notifiedNotifications)
		})

		It("Notify on workload status change", func() {
			workload1.Phase = api.PHASE_PENDING
			workload1.PhaseMessage = "blabla"
			workload2.Phase = api.PHASE_INITIALIZING
			workload2.PhaseMessage = "blabla 2"
			workload3.Phase = api.PHASE_RUNNING
			workload3.PhaseMessage = "blabla 3"
			workload4.Phase = api.PHASE_PENDING
			workload4.PhaseMessage = "blabla 4"
			workload5.Phase = api.PHASE_INITIALIZING
			workload5.PhaseMessage = "blabla 5"

			upsertWorkload(ctx, workloadsService, workload1)
			upsertWorkload(ctx, workloadsService, workload2)
			upsertWorkload(ctx, workloadsService, workload3)
			syncTestWorkloads(ctx, workloadsService, []api.WorkloadSyncFields{workload4, workload5}, ProjectIdCluster2)
			time.Sleep(NotifyClientIntervalMillis * 3 * time.Millisecond)

			Expect(notifyClient.errors).To(BeEmpty())

			expectedNotifications := map[string]map[string][]*notifyclient.Notification{
				workloadStatusChangeCategory: {
					string(WorkloadStatusChangePending): {
						constructExpectedNotificationForWorkload(ctx, workloadsService, workload1,
							workloadStatusChangeCategory, string(WorkloadStatusChangePending)),
						constructExpectedNotificationForWorkload(ctx, workloadsService, workload4,
							workloadStatusChangeCategory, string(WorkloadStatusChangePending)),
					},
					string(WorkloadStatusChangeInitializing): {
						constructExpectedNotificationForWorkload(ctx, workloadsService, workload2,
							workloadStatusChangeCategory, string(WorkloadStatusChangeInitializing)),
						constructExpectedNotificationForWorkload(ctx, workloadsService, workload5,
							workloadStatusChangeCategory, string(WorkloadStatusChangeInitializing)),
					},
					string(WorkloadStatusChangeRunning): {
						constructExpectedNotificationForWorkload(ctx, workloadsService, workload3,
							workloadStatusChangeCategory, string(WorkloadStatusChangeRunning)),
					},
				},
			}
			validateExpectedNotifications(expectedNotifications, notifyClient.notifiedNotifications)
		})

		It("Notify on workload status change 2", func() {
			workload1.Phase = api.PHASE_FAILED
			workload1.PhaseMessage = "blabla"
			workload2.Phase = api.PHASE_STOPPED
			workload2.PhaseMessage = "blabla 2"
			workload3.Phase = api.PHASE_STOPPED
			workload3.PhaseMessage = "blabla 3"
			workload4.Phase = api.PHASE_DEGRADED
			workload4.PhaseMessage = "blabla 4"
			workload5.Phase = api.PHASE_COMPLETED
			workload5.PhaseMessage = "blabla 5"

			upsertWorkload(ctx, workloadsService, workload1)
			upsertWorkload(ctx, workloadsService, workload2)
			upsertWorkload(ctx, workloadsService, workload3)
			syncTestWorkloads(ctx, workloadsService, []api.WorkloadSyncFields{workload4, workload5}, ProjectIdCluster2)
			time.Sleep(NotifyClientIntervalMillis * 3 * time.Millisecond)

			Expect(notifyClient.errors).To(BeEmpty())

			expectedNotifications := map[string]map[string][]*notifyclient.Notification{
				workloadStatusChangeCategory: {
					string(WorkloadStatusChangeFailed): {
						constructExpectedNotificationForWorkload(ctx, workloadsService, workload1,
							workloadStatusChangeCategory, string(WorkloadStatusChangeFailed)),
					},
					string(WorkloadStatusChangeStopped): {
						constructExpectedNotificationForWorkload(ctx, workloadsService, workload2,
							workloadStatusChangeCategory, string(WorkloadStatusChangeStopped)),
						constructExpectedNotificationForWorkload(ctx, workloadsService, workload3,
							workloadStatusChangeCategory, string(WorkloadStatusChangeStopped)),
					},
					string(WorkloadStatusChangeDegraded): {
						constructExpectedNotificationForWorkload(ctx, workloadsService, workload4,
							workloadStatusChangeCategory, string(WorkloadStatusChangeDegraded)),
					},
					string(WorkloadStatusChangeCompleted): {
						constructExpectedNotificationForWorkload(ctx, workloadsService, workload5,
							workloadStatusChangeCategory, string(WorkloadStatusChangeCompleted)),
					},
				},
			}
			validateExpectedNotifications(expectedNotifications, notifyClient.notifiedNotifications)
		})

		It("Notify on multiple status changes for single workload", func() {
			// update to initial status and then clear notifications
			workload2.Phase = api.PHASE_CREATING
			upsertWorkload(ctx, workloadsService, workload2)
			workload3.Phase = api.PHASE_PENDING
			upsertWorkload(ctx, workloadsService, workload3)
			time.Sleep(NotifyClientIntervalMillis * 3 * time.Millisecond)
			notifyClient.ResetNotifiedNotifications()

			workload2.Phase = api.PHASE_INITIALIZING
			workload2.PhaseMessage = "blabla 2"
			workload22 := workload2
			workload22.Phase = api.PHASE_RUNNING
			workload22.PhaseMessage = ""

			workload3.Phase = api.PHASE_RUNNING
			workload3.PhaseMessage = "blabla 3"
			workload33 := workload3
			workload33.Phase = api.PHASE_DEGRADED
			workload33.PhaseMessage = ""

			// test-only workaround: syncing the same workload with 2 different statuses will result in updating the
			// workload twice with each status change.
			// (the interval for notifications in the tests is very short, so we need the updates to happen very close to each other).
			now := time.Now().UTC().Format(time.DateTime)
			syncTestWorkloads(ctx, workloadsService, []api.WorkloadSyncFields{workload2, workload22}, DefaultProjectId)
			syncTestWorkloads(ctx, workloadsService, []api.WorkloadSyncFields{workload3, workload33}, AnotherProjectId)
			time.Sleep(NotifyClientIntervalMillis * 3 * time.Millisecond)

			Expect(notifyClient.errors).To(BeEmpty())

			expectedNotifications := map[string]map[string][]*notifyclient.Notification{
				workloadStatusChangeCategory: {
					string(WorkloadStatusChangeRunning): {
						constructExpectedNotificationForWorkloadWithDescription(ctx, workloadsService, workload2,
							workloadStatusChangeCategory, string(WorkloadStatusChangeRunning),
							"Workload workload2 status changed from Creating to Running",
							fmt.Sprintf("Workload workload2 status changed from Creating to Running"+
								"%sDetails:%sStatus changed from Creating to Initializing. Additional details: blabla 2, at %s UTC"+
								"%sStatus changed from Initializing to Running, at %s UTC",
								lineSeparator, lineSeparator, now, lineSeparator, now), nil),
					},
					string(WorkloadStatusChangeDegraded): {
						constructExpectedNotificationForWorkloadWithDescription(ctx, workloadsService, workload3,
							workloadStatusChangeCategory, string(WorkloadStatusChangeDegraded),
							"Workload workload3 status changed from Pending to Degraded",
							fmt.Sprintf("Workload workload3 status changed from Pending to Degraded"+
								"%sDetails:%sStatus changed from Pending to Running. Additional details: blabla 3, at %s UTC"+
								"%sStatus changed from Running to Degraded, at %s UTC",
								lineSeparator, lineSeparator, now, lineSeparator, now), nil),
					},
				},
			}
			validateExpectedNotifications(expectedNotifications, notifyClient.notifiedNotifications)
		})

		It("Notify and check message with old status as well", func() {
			// update to initial status and then clear notifications
			workload2.Phase = api.PHASE_CREATING
			upsertWorkload(ctx, workloadsService, workload2)
			time.Sleep(NotifyClientIntervalMillis * 3 * time.Millisecond)
			notifyClient.ResetNotifiedNotifications()

			workload2.Phase = api.PHASE_INITIALIZING
			workload2.PhaseMessage = "blabla 2"

			syncTestWorkloads(ctx, workloadsService, []api.WorkloadSyncFields{workload2}, DefaultProjectId)
			time.Sleep(NotifyClientIntervalMillis * 3 * time.Millisecond)

			Expect(notifyClient.errors).To(BeEmpty())

			expectedNotifications := map[string]map[string][]*notifyclient.Notification{
				workloadStatusChangeCategory: {
					string(WorkloadStatusChangeInitializing): {
						constructExpectedNotificationForWorkloadWithDescription(ctx, workloadsService, workload2,
							workloadStatusChangeCategory, string(WorkloadStatusChangeInitializing),
							"Workload workload2 status changed from Creating to Initializing",
							fmt.Sprintf("Workload workload2 status changed from Creating to Initializing"+
								"%sAdditional details: blabla 2",
								lineSeparator), nil),
					},
				},
			}
			validateExpectedNotifications(expectedNotifications, notifyClient.notifiedNotifications)
		})
	})

	Context("Timeout Notifier", func() {
		var workload6, workload7, workload8 api.WorkloadSyncFields

		BeforeAll(func() {
			workload6 = setupBaseWorkload(uuid.NewString(), "workload6", TrainingType, DefaultProjectId, DefaultProjectName,
				api.PHASE_RUNNING, "running", testing.ClusterUuid1)
			workload7 = setupBaseWorkload(uuid.NewString(), "workload7", TrainingType, DefaultProjectId, DefaultProjectName,
				api.PHASE_RUNNING, "running", testing.ClusterUuid1)
			workload8 = setupBaseWorkload(uuid.NewString(), "workload8", TrainingType, DefaultProjectId, DefaultProjectName,
				api.PHASE_RUNNING, "running", testing.ClusterUuid1)
			syncTestWorkloads(ctx, workloadsService, []api.WorkloadSyncFields{workload1, workload2, workload6, workload7, workload8}, DefaultProjectId)
			syncTestWorkloads(ctx, workloadsService, []api.WorkloadSyncFields{workload3}, AnotherProjectId)
			syncTestWorkloads(ctx, workloadsService, []api.WorkloadSyncFields{workload4, workload5}, ProjectIdCluster2)
		})

		BeforeEach(func() {
			stopper = make(chan struct{})
			cancelCtx, cancel = context.WithCancel(context.Background())

			notifyClient = NewNotifyClienterMock()

			conf := runai_config.Get[*config.Notify]()
			initNotifyClienter(nil, conf, notifyClient)
			InitWorkloadTimeoutNotifier(cancelCtx, stopper, conf, notifyClient)

			time.Sleep(NotifyClientIntervalMillis * time.Millisecond)
			notifyClient.ResetNotifiedNotifications()
			time.Sleep(NotifyClientIntervalMillis * time.Millisecond)
		})

		It("Sanity", func() {
			now := time.Now()

			workload6Event := getWorkloadTimeoutEvent(workload6, false, now, "")
			workload6Event.Message = "blab blaaaaa"

			workload1Time := now.Add(9 * time.Minute)
			workload2Time := now.Add(2*time.Hour + 45*time.Minute)
			workload3Time := now.Add(24*time.Hour + 2*time.Minute)

			events := []api.EventSyncFields{
				getWorkloadTimeoutEvent(workload1, false, now, "9 minutes"),
				getWorkloadTimeoutEvent(workload2, false, now, "2 hours - 45 minutes"),
				getWorkloadTimeoutEvent(workload3, true, now, "1 days - 0 hours - 2 minutes"),
				getWorkloadEvent(workload4, true, now, "0 hours - 0 minutes", false), // wrong event reason
				getWorkloadTimeoutEvent(workload5, true, now, "blabla"),              // wrong workload type
				workload6Event,
				getWorkloadTimeoutEvent(workload7, true, now, "0 hours - 0 minutes"),
				getWorkloadTimeoutEvent(workload8, true, now, "blabla"),
			}
			syncEvents(ctx, eventsService, events)

			time.Sleep(NotifyClientIntervalMillis * 3 * time.Millisecond)

			Expect(notifyClient.errors).To(BeEmpty())

			expectedNotifications := map[string]map[string][]*notifyclient.Notification{
				workloadTimeoutCategory: {
					string(WorkloadTimeoutTypeTraining): {
						constructExpectedNotificationForWorkloadWithDescription(ctx, workloadsService, workload1,
							workloadTimeoutCategory, string(WorkloadTimeoutTypeTraining),
							"Workload workload1 is about to stop due to timeout",
							fmt.Sprintf("Based on the time limit configured for your project, the workload %s will be stopped at %s UTC.",
								workload1.Name, workload1Time.UTC().Format(time.DateTime)), &now),
						constructExpectedNotificationForWorkloadWithDescription(ctx, workloadsService, workload3,
							workloadTimeoutCategory, string(WorkloadTimeoutTypeTraining),
							"Workload workload3 is about to stop due to timeout",
							fmt.Sprintf("Based on the idle GPU timeout configured for your project, as long as the GPU(s) used by %s remain idle, the workload will be terminated at %s UTC.",
								workload3.Name, workload3Time.UTC().Format(time.DateTime)), &now),
						constructExpectedNotificationForWorkloadWithDescription(ctx, workloadsService, workload6,
							workloadTimeoutCategory, string(WorkloadTimeoutTypeTraining),
							"Workload workload6 is about to stop due to timeout",
							"blab blaaaaa", &now),
						constructExpectedNotificationForWorkloadWithDescription(ctx, workloadsService, workload7,
							workloadTimeoutCategory, string(WorkloadTimeoutTypeTraining),
							"Workload workload7 is about to stop due to timeout",
							fmt.Sprintf("Based on the idle GPU timeout configured for your project, as long as the GPU(s) used by %s remain idle, the workload will be terminated at %s UTC.",
								workload7.Name, now.UTC().Format(time.DateTime)), &now),
						constructExpectedNotificationForWorkloadWithDescription(ctx, workloadsService, workload8,
							workloadTimeoutCategory, string(WorkloadTimeoutTypeTraining),
							"Workload workload8 is about to stop due to timeout",
							fmt.Sprintf("Based on the idle GPU timeout configured for your project, as long as the GPU(s) used by %s remain idle, the workload will be terminated at %s UTC.",
								workload8.Name, now.UTC().Format(time.DateTime)), &now),
					},
					string(WorkloadTimeoutTypeWorkspace): {
						constructExpectedNotificationForWorkloadWithDescription(ctx, workloadsService, workload2,
							workloadTimeoutCategory, string(WorkloadTimeoutTypeWorkspace),
							"Workload workload2 is about to stop due to timeout",
							fmt.Sprintf("Based on the time limit configured for your project, the workload %s will be stopped at %s UTC.",
								workload2.Name, workload2Time.UTC().Format(time.DateTime)), &now),
					},
				},
			}
			validateExpectedNotifications(expectedNotifications, notifyClient.notifiedNotifications)

			notifyClient.ResetNotifiedNotifications()
			time.Sleep(NotifyClientIntervalMillis * time.Millisecond * 2)

			syncEvents(ctx, eventsService, events)

			// no new notifications - the events were already fetched
			expectedNotifications = map[string]map[string][]*notifyclient.Notification{}
			validateExpectedNotifications(expectedNotifications, notifyClient.notifiedNotifications)
		})
	})
})

func validateExpectedNotifications(expected, actual map[string]map[string][]*notifyclient.Notification) {
	ExpectWithOffset(1, actual).To(HaveLen(len(expected)), "Expected notifications: %v, Actual notifications: %v", expected, actual)

	for category, expectedNotifications := range expected {
		actualCategoryNotifications, ok := actual[category]
		ExpectWithOffset(1, ok).To(BeTrue(), "Expected category <%s> not found in actual notifications", category)

		ExpectWithOffset(1, actualCategoryNotifications).To(HaveLen(len(expectedNotifications)),
			"Expected category <%s> notifications: %v, Actual notifications: %v", category, expectedNotifications, actualCategoryNotifications)

		for notificationType, expectedTypeNotifications := range expectedNotifications {
			actualTypeNotifications, ok := actualCategoryNotifications[notificationType]
			ExpectWithOffset(1, ok).To(BeTrue(), "Expected type <%s> not found in actual notifications of category <%s>", notificationType, category)

			ExpectWithOffset(1, actualTypeNotifications).To(HaveLen(len(expectedTypeNotifications)),
				"Expected category/type <%s/%s> notifications: %v, Actual notifications: %v", category, notificationType, expectedTypeNotifications, actualTypeNotifications)

			for _, expectedTypeNotification := range expectedTypeNotifications {
				validated := false
				for _, actualTypeNotification := range actualTypeNotifications {
					if actualTypeNotification.AdditionalProperties[nameProperty] == expectedTypeNotification.AdditionalProperties[nameProperty] &&
						actualTypeNotification.AdditionalProperties[projectProperty] == expectedTypeNotification.AdditionalProperties[projectProperty] {
						validateNotification(expectedTypeNotification, actualTypeNotification)
						validated = true
						break
					}
				}
				ExpectWithOffset(1, validated).To(BeTrue(), "Expected notification not found in actual notifications, expectedTypeNotifications: %v, actualTypeNotifications: %v", expectedTypeNotifications, actualTypeNotifications)
			}
		}
	}
}

func validateNotification(expected, actual *notifyclient.Notification) {
	ExpectWithOffset(2, actual.Category).To(Equal(expected.Category), "category")
	ExpectWithOffset(2, actual.Type).To(Equal(expected.Type), "type")
	ExpectWithOffset(2, actual.Tenant).To(Equal(expected.Tenant), "tenant")
	ExpectWithOffset(2, actual.User).To(Equal(expected.User), "user")

	if expected.Title != "" {
		ExpectWithOffset(2, actual.Title).To(Equal(expected.Title),
			"title expected: <%s>, actual: <%s>", expected.Title, actual.Title)
	}
	if expected.Description != "" {
		// check only the length of the description, as the actual value may differ in terms of time
		ExpectWithOffset(2, actual.Description).To(HaveLen(len(expected.Description)),
			"description expected: <%s>, actual: <%s>", expected.Description, actual.Description)
	}

	ExpectWithOffset(2, actual.AdditionalProperties).To(HaveLen(len(expected.AdditionalProperties)), "additional properties len")
	ExpectWithOffset(2, actual.AdditionalProperties[nameProperty]).To(Equal(expected.AdditionalProperties[nameProperty]), "name")
	ExpectWithOffset(2, actual.AdditionalProperties[typeProperty]).To(Equal(expected.AdditionalProperties[typeProperty]), "type")
	ExpectWithOffset(2, actual.AdditionalProperties[projectProperty]).To(Equal(expected.AdditionalProperties[projectProperty]), "project")
	// ExpectWithOffset(2, actual.AdditionalProperties[clusterProperty]).To(Equal(expected.AdditionalProperties[clusterProperty]), "cluster")
	ExpectWithOffset(2, actual.AdditionalProperties[timeProperty]).To(Equal(expected.AdditionalProperties[timeProperty]), "time")
	ExpectWithOffset(2, actual.AdditionalProperties[urlProperty]).To(Equal(expected.AdditionalProperties[urlProperty]), "url")
}

func constructExpectedNotificationForWorkload(ctx context.Context, workloadsService api.WorkloadsAPIServicer,
	w api.WorkloadSyncFields,
	category, typ string,
) *notifyclient.Notification {
	return constructExpectedNotificationForWorkloadWithDescription(ctx, workloadsService, w, category, typ, "", "", nil)
}

func constructExpectedNotificationForWorkloadWithDescription(ctx context.Context, workloadsService api.WorkloadsAPIServicer,
	w api.WorkloadSyncFields,
	category, typ, title, description string,
	tim *time.Time,
) *notifyclient.Notification {
	By(fmt.Sprintf("getting workload %s", w.Name))
	apiW := getWorkloadApi(ctx, workloadsService, w.Id)

	timeValue := tim
	if tim == nil {
		timeValue = &apiW.PhaseUpdatedAt
	}

	return &notifyclient.Notification{
		Category:    category,
		Type:        typ,
		Title:       title,
		Description: description,
		Tenant:      strconv.Itoa(int(apiW.TenantId)),
		User:        apiW.SubmittedBy,
		AdditionalProperties: map[string]interface{}{
			nameProperty:    apiW.Name,
			typeProperty:    apiW.Type,
			projectProperty: apiW.ProjectName,
			// clusterProperty: apiW.ClusterId,
			timeProperty: getTimeStr(*timeValue),
			urlProperty:  fmt.Sprintf("https://mock.com/workloads?sortBy=name&advancedFilters=[{\"term\":\"%s\",\"value\":\"==\",\"label\":\"Project\",\"name\":\"projectName\",\"type\":\"FreeText\"},{\"term\":\"%s\",\"value\":\"==\",\"label\":\"Workload\",\"name\":\"name\",\"type\":\"FreeText\"},{\"term\":\"%s\",\"value\":\"=@\",\"label\":\"Type\",\"name\":\"type\",\"type\":\"FreeText\"}]", apiW.ProjectName, apiW.Name, apiW.Type),
		},
	}
}

func syncTestWorkloads(ctx context.Context, workloadsService api.WorkloadsAPIServicer,
	workloadsToSync []api.WorkloadSyncFields, projectId string,
) {
	apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
		ProjectId: projectId, Workloads: workloadsToSync,
	})
	Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
}

func upsertWorkload(ctx context.Context, workloadsService api.WorkloadsAPIServicer, workload api.WorkloadSyncFields) {
	apiResponse, _ := workloadsService.UpsertWorkload(ctx, workload.Id, workload)
	Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
}

func getWorkloadApi(ctx context.Context, workloadsService api.WorkloadsAPIServicer, id string) api.Workload {
	apiResponse, _ := workloadsService.GetWorkload(ctx, id)
	Expect(apiResponse.Code).To(Equal(http.StatusOK), "Failed to get workload %s", id)
	bytes, err := json.Marshal(apiResponse.Body)
	Expect(err).To(BeNil())

	response := &api.Workload{}
	err = json.Unmarshal(bytes, response)
	Expect(err).To(BeNil())
	return *response
}

func setupBaseWorkload(id, name, typ, projectId, projectName string, phase api.Phase, phaseMessage, clusterId string) api.WorkloadSyncFields {
	priority := int32(50)
	return api.WorkloadSyncFields{
		Type:             typ,
		Name:             name,
		Source:           DefaultSource,
		Id:               id,
		SubmittedBy:      DefaultSubmittedBy,
		ClusterId:        clusterId,
		ProjectName:      projectName,
		ProjectId:        projectId,
		DepartmentId:     DefaultDepartmentId,
		DepartmentName:   DefaultDepartmentName,
		Namespace:        "runai-" + projectName,
		CreatedAt:        DefaultCreationTime,
		Phase:            phase,
		PhaseMessage:     phaseMessage,
		K8sPhase:         DefaultK8sPhase,
		AdditionalFields: "",
		Priority:         &priority,
	}
}

func syncEvents(ctx context.Context, eventsService api.EventsAPIServicer, events []api.EventSyncFields) {
	apiResponse, _ := eventsService.SyncEvents(ctx, api.EventsSyncRequest{
		Events: events,
	})
	Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
}

func getWorkloadTimeoutEvent(workload api.WorkloadSyncFields, isIdle bool, createdAt time.Time, timeStr string) api.EventSyncFields {
	return getWorkloadEvent(workload, isIdle, createdAt, timeStr, true)
}

func getWorkloadEvent(workload api.WorkloadSyncFields, isIdle bool, createdAt time.Time, timeStr string, isTimeoutApproaching bool) api.EventSyncFields {
	message := "Based on the time limit configured for your project, the job job-int will be terminated in %s."
	if isIdle {
		message = "Based on the idle GPU timeout configured for your project, as long as the GPU(s) used by runai-idle-gpu-workspace-preem remain idle, the job will be terminated in %s."
	}
	message = fmt.Sprintf(message, timeStr)

	reason := workloadTimeoutApproachingEventReason
	if !isTimeoutApproaching {
		reason = "Sync"
	}

	return api.EventSyncFields{
		Id:        uuid.NewString(),
		ClusterId: workload.ClusterId,
		Type:      "Normal",
		Message:   message,
		Reason:    reason,
		CreatedAt: createdAt,
		Source:    "runtime-rules",
		InvolvedObject: api.InvolvedObject{
			Uid:       workload.Id,
			Kind:      workload.Type + "Workload",
			Name:      workload.Name,
			Namespace: workload.Namespace,
		},
	}
}

func mockConfigForNotify() {
	runai_config.Set(&config.Notify{
		PhaseChangeIntervalMillis: NotifyClientIntervalMillis,
		TimeoutIntervalMillis:     NotifyClientIntervalMillis,
	})
}
