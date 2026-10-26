/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"k8s.io/utils/ptr"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	NormalType       = "normal"
	FailureType      = "failure"
	SubmittedMsg     = "workload submitted"
	FailureMsg       = "workload failure"
	CompletedMsg     = "workload finished"
	CompletedReason  = "completed"
	FailureReason    = "failure"
	SubmissionReason = "submitted"
	EventSource      = "kubelet"
	WorkloadKind     = "Workload"
	RevisionKind     = "Workload"
	PodKind          = "Pod"

	SortAscending          = "asc"
	SortDescending         = "desc"
	EventsDefaultSortOrder = SortDescending
)

var _ = Describe("Events API tests", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	eventsService := NewEventsService()
	workloadsService := NewWorkloadsService()
	revisionsService := NewRevisionsService()
	podsService := NewPodsService(nil)

	BeforeAll(func() {
		apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
			ProjectId: DefaultProjectId,
			Workloads: []api.WorkloadSyncFields{
				SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1),
				SetupBaseWorkload(testing.TestUuid2, DefaultProjectId, testing.ClusterUuid1),
			},
		})
		Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

		id3 := uuid.NewString()
		id4 := uuid.NewString()
		id5 := uuid.NewString()

		// Sleep a second to make sure there is a difference between the first events and the workload creation time
		time.Sleep(1 * time.Second)
		apiResponse, _ = eventsService.SyncEvents(ctx, api.EventsSyncRequest{
			Events: []api.EventSyncFields{
				{
					Id:        id4,
					ClusterId: testing.ClusterUuid1,
					Type:      NormalType,
					Message:   CompletedMsg,
					Reason:    CompletedReason,
					CreatedAt: time.Now().Add(time.Hour),
					Source:    EventSource,
					InvolvedObject: api.InvolvedObject{
						Uid:       testing.TestUuid1,
						Kind:      WorkloadKind,
						Name:      DefaultWorkloadName,
						Namespace: DefaultNamespace,
					},
				},
				{
					Id:        id3,
					ClusterId: testing.ClusterUuid1,
					Type:      NormalType,
					Message:   SubmittedMsg,
					Reason:    SubmissionReason,
					CreatedAt: time.Now(),
					Source:    EventSource,
					InvolvedObject: api.InvolvedObject{
						Uid:       testing.TestUuid1,
						Kind:      WorkloadKind,
						Name:      DefaultWorkloadName,
						Namespace: DefaultNamespace,
					},
				},
				{
					Id:        id5,
					ClusterId: testing.ClusterUuid1,
					Type:      FailureType,
					Message:   FailureMsg,
					Reason:    FailureReason,
					CreatedAt: time.Now(),
					Source:    EventSource,
					InvolvedObject: api.InvolvedObject{
						Uid:       testing.TestUuid2,
						Kind:      WorkloadKind,
						Name:      DefaultWorkloadName,
						Namespace: DefaultNamespace,
					},
				},
			},
		})
		Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
	})

	AfterAll(func() {
		testing.CleanTables()
	})

	Context("Get workload events", func() {
		It("Workload with events", func() {
			apiResponse, _ := eventsService.GetWorkloadEvents(ctx, testing.TestUuid1, 0, 0, EventsDefaultSortOrder)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			eventsResponse := httpResponseBodyToEventsResponse(apiResponse)
			Expect(eventsResponse.Next).To(Equal(int32(0)))
			Expect(len(eventsResponse.Events)).To(Equal(2))
			Expect(eventsResponse.Events[1].Id).ToNot(BeNil())
			Expect(eventsResponse.Events[1].Source).To(Equal(EventSource))
			Expect(eventsResponse.Events[1].Type).To(Equal(NormalType))
			Expect(eventsResponse.Events[1].Message).To(Equal(SubmittedMsg))
			Expect(eventsResponse.Events[1].Reason).To(Equal(SubmissionReason))
			Expect(eventsResponse.Events[1].CreatedAt.Unix()).To(BeNumerically("<", eventsResponse.Events[0].CreatedAt.Unix()))
			Expect(eventsResponse.Events[1].ClusterId).To(Equal(testing.ClusterUuid1))
			Expect(eventsResponse.Events[1].InvolvedObject.Uid).To(Equal(testing.TestUuid1))
			Expect(eventsResponse.Events[0].InvolvedObject.Uid).To(Equal(testing.TestUuid1))
			Expect(eventsResponse.Events[1].InvolvedObject.Kind).To(Equal(WorkloadKind))
			Expect(eventsResponse.Events[1].InvolvedObject.Name).To(Equal(DefaultWorkloadName))
			Expect(eventsResponse.Events[1].InvolvedObject.Namespace).To(Equal(DefaultNamespace))
		})
		DescribeTable("Workload with events - sorting",
			func(sortOrder string) {
				apiResponse, _ := eventsService.GetWorkloadEvents(ctx, testing.TestUuid1, 0, 0, sortOrder)
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				eventsResponse := httpResponseBodyToEventsResponse(apiResponse)

				comparator := "<"
				if sortOrder == SortDescending {
					comparator = ">"
				}

				for i := 0; i+1 < len(eventsResponse.Events); i++ {
					Expect(eventsResponse.Events[i].CreatedAt.Unix()).To(BeNumerically(comparator, eventsResponse.Events[i+1].CreatedAt.Unix()))
				}
			},
			Entry("ascending", SortAscending),
			Entry("descending", SortDescending),
		)
		It("Workload with history", func() {
			baseWorkload := SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			baseWorkload.Phase = api.PHASE_DEGRADED
			baseWorkload.PhaseMessage = "Not enough pods"

			apiResponse, _ := workloadsService.UpsertWorkload(ctx, testing.TestUuid1, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = eventsService.GetWorkloadHistory(ctx, testing.TestUuid1, 0, 0)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			historyResponse := httpResponseBodyToHistoryResponse(apiResponse)
			Expect(historyResponse.Next).To(Equal(int32(0)))
			Expect(len(historyResponse.Records)).To(Equal(4))
			for i := 0; i < 4; i++ {
				record := historyResponse.Records[i]
				if i != 3 {
					Expect(record.Meta.CreationTimestamp.Unix()).To(BeNumerically(">=", historyResponse.Records[i+1].Meta.CreationTimestamp.Unix()))
				}

				if record.Meta.Type == api.HISTORYRECORDTYPE_PHASE_UPDATE {
					Expect(record.Spec.PhaseUpdate).ToNot(BeNil())
					Expect(record.Spec.Event).To(BeNil())
				} else {
					Expect(record.Spec.Event).ToNot(BeNil())
					Expect(record.Spec.PhaseUpdate).To(BeNil())
				}
			}

			Expect(historyResponse.Records[3].Meta.Type).To(Equal(api.HISTORYRECORDTYPE_PHASE_UPDATE))
			Expect(historyResponse.Records[2].Meta.Type).To(Equal(api.HISTORYRECORDTYPE_EVENT))
			Expect(historyResponse.Records[1].Meta.Type).To(Equal(api.HISTORYRECORDTYPE_PHASE_UPDATE))
			Expect(historyResponse.Records[0].Meta.Type).To(Equal(api.HISTORYRECORDTYPE_EVENT))

			Expect(historyResponse.Records[1].Spec.PhaseUpdate.Phase).To(Equal(baseWorkload.Phase))
			Expect(historyResponse.Records[1].Spec.PhaseUpdate.PhaseMessage).To(Equal(baseWorkload.PhaseMessage))
		})
		It("Workload phase history saved only on change", func() {
			baseWorkload := SetupBaseWorkload(uuid.NewString(), DefaultProjectId, testing.ClusterUuid1)
			baseWorkload.Type = "Inference"

			workloadsService.SyncWorkloads(ctx, []string{"type==Inference"}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					baseWorkload,
				},
			})
			workloadsService.SyncWorkloads(ctx, []string{"type==Inference"}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					baseWorkload,
				},
			})
			workloadsService.SyncWorkloads(ctx, []string{"type!=Inference"}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{},
			})
			workloadsService.SyncWorkloads(ctx, []string{"type==Inference"}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					baseWorkload,
				},
			})

			apiResponse, _ := workloadsService.UpsertWorkload(ctx, baseWorkload.Id, baseWorkload)

			apiResponse, _ = eventsService.GetWorkloadHistory(ctx, baseWorkload.Id, 0, 0)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			historyResponse := httpResponseBodyToHistoryResponse(apiResponse)
			Expect(historyResponse.Next).To(Equal(int32(0)))
			Expect(len(historyResponse.Records)).To(Equal(1))
			Expect(historyResponse.Records[0].Meta.Type).To(Equal(api.HISTORYRECORDTYPE_PHASE_UPDATE))
			Expect(historyResponse.Records[0].Spec.PhaseUpdate.Phase).To(Equal(baseWorkload.Phase))
			Expect(historyResponse.Records[0].Spec.PhaseUpdate.PhaseMessage).To(Equal(baseWorkload.PhaseMessage))
		})
	})

	Context("Get revision events", func() {
		It("Revision with events", func() {
			// syncing two revisions for a workload
			revId1 := uuid.NewString()
			revId2 := uuid.NewString()
			apiResponse, _ := revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{ // required in order to sync the pod with a revision
				ProjectId: DefaultProjectId,
				Revisions: []api.RevisionSyncFields{
					SetupBaseRevision(revId1, testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1),
					SetupBaseRevision(revId2, testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// syncing pods for the revisions, that will be added as children,
			// and will associate the following events with the revision
			pod := SetupBasePod(uuid.NewString(), testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			pod.RevisionId = ptr.To(revId1)
			pod2 := SetupBasePod(uuid.NewString(), testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			pod2.RevisionId = ptr.To(revId2)
			pod3 := SetupBasePod(uuid.NewString(), testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			apiResponse, _ = podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				Pods:      []api.PodSyncFields{pod, pod2, pod3},
			})

			// syncing events for the revision pods, as well as for a third unrelated pods
			// Sleep a second to make sure there is a difference between the first events and the revision creation time
			time.Sleep(1 * time.Second)
			eventId1 := uuid.NewString()
			eventId2 := uuid.NewString()
			apiResponse, _ = eventsService.SyncEvents(ctx, api.EventsSyncRequest{
				Events: []api.EventSyncFields{
					{
						Id:        eventId1,
						ClusterId: testing.ClusterUuid1,
						Type:      NormalType,
						Message:   CompletedMsg,
						Reason:    CompletedReason,
						CreatedAt: time.Now().Add(time.Hour),
						Source:    EventSource,
						InvolvedObject: api.InvolvedObject{
							Uid:       pod.Id,
							Kind:      PodKind,
							Name:      DefaultWorkloadName,
							Namespace: DefaultNamespace,
						},
					},
					{
						Id:        eventId2,
						ClusterId: testing.ClusterUuid1,
						Type:      FailureType,
						Message:   FailureMsg,
						Reason:    FailureReason,
						CreatedAt: time.Now(),
						Source:    EventSource,
						InvolvedObject: api.InvolvedObject{
							Uid:       pod2.Id,
							Kind:      PodKind,
							Name:      DefaultWorkloadName,
							Namespace: DefaultNamespace,
						},
					},
					{
						Id:        uuid.NewString(),
						ClusterId: testing.ClusterUuid1,
						Type:      FailureType,
						Message:   FailureMsg,
						Reason:    FailureReason,
						CreatedAt: time.Now(),
						Source:    EventSource,
						InvolvedObject: api.InvolvedObject{
							Uid:       pod3.Id,
							Kind:      PodKind,
							Name:      DefaultWorkloadName,
							Namespace: DefaultNamespace,
						},
					},
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// validate that the events are associated with the correct revision
			apiResponse, _ = eventsService.GetRevisionEvents(ctx, revId1, 0, 0, EventsDefaultSortOrder)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			eventsResponse := httpResponseBodyToEventsResponse(apiResponse)
			Expect(eventsResponse.Next).To(Equal(int32(0)))
			Expect(len(eventsResponse.Events)).To(Equal(1))
			event := eventsResponse.Events[0]
			Expect(event.Id).To(Equal(eventId1))
			Expect(event.Source).To(Equal(EventSource))
			Expect(event.Type).To(Equal(NormalType))
			Expect(event.Message).To(Equal(CompletedMsg))
			Expect(event.Reason).To(Equal(CompletedReason))
			Expect(event.ClusterId).To(Equal(testing.ClusterUuid1))
			Expect(event.InvolvedObject.Uid).To(Equal(pod.Id))
			Expect(event.InvolvedObject.Kind).To(Equal(PodKind))
			Expect(event.InvolvedObject.Name).To(Equal(DefaultWorkloadName))
			Expect(event.InvolvedObject.Namespace).To(Equal(DefaultNamespace))

			apiResponse, _ = eventsService.GetRevisionEvents(ctx, revId2, 0, 0, EventsDefaultSortOrder)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			eventsResponse = httpResponseBodyToEventsResponse(apiResponse)
			Expect(eventsResponse.Next).To(Equal(int32(0)))
			Expect(len(eventsResponse.Events)).To(Equal(1))
			event = eventsResponse.Events[0]
			Expect(event.Id).To(Equal(eventId2))
			Expect(event.Source).To(Equal(EventSource))
			Expect(event.Type).To(Equal(FailureType))
			Expect(event.Message).To(Equal(FailureMsg))
			Expect(event.Reason).To(Equal(FailureReason))
			Expect(event.ClusterId).To(Equal(testing.ClusterUuid1))
			Expect(event.InvolvedObject.Uid).To(Equal(pod2.Id))
			Expect(event.InvolvedObject.Kind).To(Equal(PodKind))
			Expect(event.InvolvedObject.Name).To(Equal(DefaultWorkloadName))
			Expect(event.InvolvedObject.Namespace).To(Equal(DefaultNamespace))
		})
		DescribeTable("Revision with events - sorting",
			func(sortOrder string) {
				// syncing a revision for a workload
				revId := uuid.NewString()
				apiResponse, _ := revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{ // required in order to sync the pod with a revision
					ProjectId: DefaultProjectId,
					Revisions: []api.RevisionSyncFields{
						SetupBaseRevision(revId, testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1),
					},
				})
				Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

				// syncing a pod for the revision
				pod := SetupBasePod(uuid.NewString(), testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
				pod.RevisionId = ptr.To(revId)
				apiResponse, _ = podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
					ProjectId: DefaultProjectId,
					Pods:      []api.PodSyncFields{pod},
				})

				// syncing multiple events for the revision pod, with different timestamps
				// Sleep a second to make sure there is a difference between the first events and the revision creation time
				time.Sleep(1 * time.Second)
				apiResponse, _ = eventsService.SyncEvents(ctx, api.EventsSyncRequest{
					Events: []api.EventSyncFields{
						{
							Id:        uuid.NewString(),
							ClusterId: testing.ClusterUuid1,
							Type:      NormalType,
							Message:   CompletedMsg,
							Reason:    CompletedReason,
							CreatedAt: time.Now().Add(time.Hour),
							Source:    EventSource,
							InvolvedObject: api.InvolvedObject{
								Uid:       pod.Id,
								Kind:      PodKind,
								Name:      DefaultWorkloadName,
								Namespace: DefaultNamespace,
							},
						},
						{
							Id:        uuid.NewString(),
							ClusterId: testing.ClusterUuid1,
							Type:      FailureType,
							Message:   FailureMsg,
							Reason:    FailureReason,
							CreatedAt: time.Now(),
							Source:    EventSource,
							InvolvedObject: api.InvolvedObject{
								Uid:       pod.Id,
								Kind:      PodKind,
								Name:      DefaultWorkloadName,
								Namespace: DefaultNamespace,
							},
						},
						{
							Id:        uuid.NewString(),
							ClusterId: testing.ClusterUuid1,
							Type:      FailureType,
							Message:   FailureMsg,
							Reason:    FailureReason,
							CreatedAt: time.Now().Add(time.Minute),
							Source:    EventSource,
							InvolvedObject: api.InvolvedObject{
								Uid:       pod.Id,
								Kind:      PodKind,
								Name:      DefaultWorkloadName,
								Namespace: DefaultNamespace,
							},
						},
					},
				})
				Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

				// validate sorting
				apiResponse, _ = eventsService.GetRevisionEvents(ctx, revId, 0, 0, sortOrder)
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				eventsResponse := httpResponseBodyToEventsResponse(apiResponse)

				comparator := "<"
				if sortOrder == SortDescending {
					comparator = ">"
				}

				for i := 0; i+1 < len(eventsResponse.Events); i++ {
					Expect(eventsResponse.Events[i].CreatedAt.Unix()).To(BeNumerically(comparator, eventsResponse.Events[i+1].CreatedAt.Unix()))
				}
			},
			Entry("ascending", SortAscending),
			Entry("descending", SortDescending),
		)
	})
	It("Revision with history", func() {
		// syncing a revision
		revisionId := uuid.NewString()
		revision := SetupBaseRevision(revisionId, testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
		apiResponse, _ := revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{
			ProjectId: DefaultProjectId,
			Revisions: []api.RevisionSyncFields{
				revision,
			},
		})
		Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

		// sync two events for the revision
		// Sleep a second to make sure there is a difference between the first events and the workload creation time
		time.Sleep(1 * time.Second)
		eventId1 := uuid.NewString()
		eventId2 := uuid.NewString()
		apiResponse, _ = eventsService.SyncEvents(ctx, api.EventsSyncRequest{
			Events: []api.EventSyncFields{
				{
					Id:        eventId1,
					ClusterId: testing.ClusterUuid1,
					Type:      NormalType,
					Message:   CompletedMsg,
					Reason:    CompletedReason,
					CreatedAt: time.Now().Add(time.Hour),
					Source:    EventSource,
					InvolvedObject: api.InvolvedObject{
						Uid:       revisionId,
						Kind:      RevisionKind,
						Name:      DefaultWorkloadName,
						Namespace: DefaultNamespace,
					},
				},
				{
					Id:        eventId2,
					ClusterId: testing.ClusterUuid1,
					Type:      NormalType,
					Message:   SubmittedMsg,
					Reason:    SubmissionReason,
					CreatedAt: time.Now(),
					Source:    EventSource,
					InvolvedObject: api.InvolvedObject{
						Uid:       revisionId,
						Kind:      RevisionKind,
						Name:      DefaultWorkloadName,
						Namespace: DefaultNamespace,
					},
				},
				{
					Id:        uuid.NewString(),
					ClusterId: testing.ClusterUuid1,
					Type:      FailureType,
					Message:   FailureMsg,
					Reason:    FailureReason,
					CreatedAt: time.Now(),
					Source:    EventSource,
					InvolvedObject: api.InvolvedObject{
						Uid:       uuid.NewString(), // a different revision id
						Kind:      RevisionKind,
						Name:      DefaultWorkloadName,
						Namespace: DefaultNamespace,
					},
				},
			},
		})
		Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

		// we should get two events and the initial phase as the revision history
		apiResponse, _ = eventsService.GetRevisionHistory(ctx, revisionId, 0, 0)
		Expect(apiResponse.Code).To(Equal(http.StatusOK))
		historyResponse := httpResponseBodyToHistoryResponse(apiResponse)
		Expect(historyResponse.Next).To(Equal(int32(0)))
		Expect(len(historyResponse.Records)).To(Equal(3))

		// perform a phase change on the revision
		revision.Phase = api.PHASE_DEGRADED
		revision.PhaseMessage = "Not enough pods"

		apiResponse, _ = revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{
			ProjectId: DefaultProjectId,
			Revisions: []api.RevisionSyncFields{
				revision,
			},
		})
		Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

		// we should get two events and two phases as the revision history
		apiResponse, _ = eventsService.GetRevisionHistory(ctx, revisionId, 0, 0)
		Expect(apiResponse.Code).To(Equal(http.StatusOK))
		historyResponse = httpResponseBodyToHistoryResponse(apiResponse)
		Expect(historyResponse.Next).To(Equal(int32(0)))
		Expect(len(historyResponse.Records)).To(Equal(4))

		for i := 0; i < 4; i++ {
			record := historyResponse.Records[i]
			if i != 3 {
				Expect(record.Meta.CreationTimestamp.Unix()).To(BeNumerically(">=", historyResponse.Records[i+1].Meta.CreationTimestamp.Unix()))
			}

			if record.Meta.Type == api.HISTORYRECORDTYPE_PHASE_UPDATE {
				Expect(record.Spec.PhaseUpdate).ToNot(BeNil())
				Expect(record.Spec.Event).To(BeNil())
			} else {
				Expect(record.Spec.Event).ToNot(BeNil())
				Expect(record.Spec.PhaseUpdate).To(BeNil())
			}
		}

		Expect(historyResponse.Records[3].Meta.Type).To(Equal(api.HISTORYRECORDTYPE_PHASE_UPDATE))
		Expect(historyResponse.Records[2].Meta.Type).To(Equal(api.HISTORYRECORDTYPE_EVENT))
		Expect(historyResponse.Records[1].Meta.Type).To(Equal(api.HISTORYRECORDTYPE_PHASE_UPDATE))
		Expect(historyResponse.Records[0].Meta.Type).To(Equal(api.HISTORYRECORDTYPE_EVENT))

		Expect(historyResponse.Records[1].Spec.PhaseUpdate.Phase).To(Equal(revision.Phase))
		Expect(historyResponse.Records[1].Spec.PhaseUpdate.PhaseMessage).To(Equal(revision.PhaseMessage))
	})
	It("Revision phase history saved only on change", func() {
		revisionId := uuid.NewString()
		revision := SetupBaseRevision(revisionId, testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
		apiResponse, _ := revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{
			ProjectId: DefaultProjectId,
			Revisions: []api.RevisionSyncFields{
				revision,
			},
		})
		Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

		revision.Command = "some new command"
		apiResponse, _ = revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{
			ProjectId: DefaultProjectId,
			Revisions: []api.RevisionSyncFields{
				revision,
			},
		})
		Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
		apiResponse, _ = revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{
			ProjectId: DefaultProjectId,
			Revisions: []api.RevisionSyncFields{
				revision,
			},
		})
		Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

		apiResponse, _ = eventsService.GetRevisionHistory(ctx, revision.Id, 0, 0)
		Expect(apiResponse.Code).To(Equal(http.StatusOK))
		historyResponse := httpResponseBodyToHistoryResponse(apiResponse)
		Expect(historyResponse.Next).To(Equal(int32(0)))
		Expect(len(historyResponse.Records)).To(Equal(1))
		Expect(historyResponse.Records[0].Meta.Type).To(Equal(api.HISTORYRECORDTYPE_PHASE_UPDATE))
		Expect(historyResponse.Records[0].Spec.PhaseUpdate.Phase).To(Equal(revision.Phase))
		Expect(historyResponse.Records[0].Spec.PhaseUpdate.PhaseMessage).To(Equal(revision.PhaseMessage))
	})
})

func httpResponseBodyToEventsResponse(httpResponse api.ImplResponse) api.GetWorkloadEvents200Response {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	response := &api.GetWorkloadEvents200Response{}
	err = json.Unmarshal(bytes, response)
	Expect(err).To(BeNil())
	return *response
}

func httpResponseBodyToHistoryResponse(httpResponse api.ImplResponse) api.GetWorkloadHistory200Response {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	response := &api.GetWorkloadHistory200Response{}
	err = json.Unmarshal(bytes, response)
	Expect(err).To(BeNil())
	return *response
}
