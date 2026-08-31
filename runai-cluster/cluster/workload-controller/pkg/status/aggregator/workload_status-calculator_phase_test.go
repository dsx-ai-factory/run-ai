/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package aggregator

import (
	"context"
	"errors"
	fields "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/issue"
	issue2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common/issue"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common/resource"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/autogenerate/mocks"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	job "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"

	"strings"
)

type phaseMovementTest struct {
	startPhase common.Phase
	jobStates  []jobState
	endPhase   common.Phase
}

type jobState struct {
	isCreating     bool
	isRunning      bool
	isPending      bool
	isJobFailed    bool
	isSucceeded    bool
	isSuspended    bool
	isInitializing bool
	isTerminating  bool
	annotations    map[string]string
}

func (j jobState) String() string {
	var states []string
	if j.isCreating {
		states = append(states, "creating")
	}
	if j.isRunning {
		states = append(states, "running")
	}
	if j.isPending {
		states = append(states, "pending")
	}
	if j.isJobFailed {
		states = append(states, "failed")
	}
	if j.isSucceeded {
		states = append(states, "succeeded")
	}
	if j.isSuspended {
		states = append(states, "suspended")
	}
	if j.isInitializing {
		states = append(states, "initializing")
	}
	if j.isTerminating {
		states = append(states, "terminating")
	}
	return "jobState{" + strings.Join(states, ",") + "}"
}

func (j jobState) Extend(otherJobState jobState) jobState {
	// doing or between the two job states
	return jobState{
		isCreating:     j.isCreating || otherJobState.isCreating,
		isRunning:      j.isRunning || otherJobState.isRunning,
		isPending:      j.isPending || otherJobState.isPending,
		isJobFailed:    j.isJobFailed || otherJobState.isJobFailed,
		isSucceeded:    j.isSucceeded || otherJobState.isSucceeded,
		isSuspended:    j.isSuspended || otherJobState.isSuspended,
		isInitializing: j.isInitializing || otherJobState.isInitializing,
		isTerminating:  j.isTerminating || otherJobState.isTerminating,
		annotations:    j.annotations,
	}
}

var (
	jobCreating     = jobState{isCreating: true}
	jobRunning      = jobState{isRunning: true}
	jobPending      = jobState{isPending: true}
	jobFailed       = jobState{isJobFailed: true}
	jobSucceeded    = jobState{isSucceeded: true}
	jobSuspended    = jobState{isSuspended: true}
	jobInitializing = jobState{isInitializing: true}

	jobTerminating                       = jobState{isTerminating: true}
	jobPendingWithRunning                = jobRunning.Extend(jobPending)
	jobPendingWithRunningAndInitializing = jobRunning.Extend(jobPending).Extend(jobInitializing)
	jobPendingWithInitializing           = jobPending.Extend(jobInitializing)
	jobInitializingWithRunning           = jobPending.Extend(jobInitializing).Extend(jobRunning)

	jobTerminatingWithRunning = jobTerminating.Extend(jobRunning)
	jobTerminatingWithPending = jobTerminating.Extend(jobPending)

	jobTerminatingWithInitializing                                = jobTerminating.Extend(jobInitializing)
	jobTerminatingWithRunningAndInitializing                      = jobTerminating.Extend(jobRunning).Extend(jobInitializing)
	jobTerminatingWithPendingAndInitializing                      = jobTerminating.Extend(jobPending).Extend(jobInitializing)
	jobTerminatingWithPendingAndRunning                           = jobTerminating.Extend(jobPending).Extend(jobRunning)
	jobTerminatingWithPendingAndRunningAndInitializing            = jobTerminatingWithPendingAndRunning.Extend(jobInitializing)
	jobTerminatingWithPendingAndRunningAndInitializingAndCreating = jobTerminatingWithPendingAndRunningAndInitializing.Extend(jobCreating)
	jobTerminatingWithSuspended                                   = jobTerminating.Extend(jobSuspended)
	jobTerminatingWithFailed                                      = jobTerminating.Extend(jobFailed)
	jobTerminatingWithSucceeded                                   = jobTerminating.Extend(jobSucceeded)

	jobTerminatingWithCreating = jobTerminating.Extend(jobCreating)
)

var (
	creatingToRunning = phaseMovementTest{
		startPhase: common.PhaseCreating,
		jobStates: []jobState{
			jobPendingWithRunning,
			jobPendingWithRunningAndInitializing,
			jobRunning,
		},

		endPhase: common.PhaseRunning,
	}
	creatingToPending = phaseMovementTest{
		startPhase: common.PhaseCreating,
		jobStates: []jobState{
			jobPending,
		},

		endPhase: common.PhasePending,
	}
	creatingToCreating = phaseMovementTest{
		startPhase: common.PhaseCreating,
		jobStates: []jobState{
			jobCreating,
		},

		endPhase: common.PhaseCreating,
	}
	creatingToFailed = phaseMovementTest{
		startPhase: common.PhaseCreating,
		jobStates: []jobState{
			jobFailed,
		},

		endPhase: common.PhaseFailed,
	}
	creatingToSucceeded = phaseMovementTest{
		startPhase: common.PhaseCreating,
		jobStates: []jobState{
			jobSucceeded,
		},

		endPhase: common.PhaseCompleted,
	}
	creatingToSuspended = phaseMovementTest{
		startPhase: common.PhaseCreating,
		jobStates: []jobState{
			jobSuspended,
		},

		endPhase: common.PhaseSuspended,
	}
	creatingToInitializing = phaseMovementTest{
		startPhase: common.PhaseCreating,
		jobStates: []jobState{
			jobInitializing,
			jobPendingWithInitializing,
		},

		endPhase: common.PhaseInitializing,
	}
	creatingToTerminating = phaseMovementTest{
		startPhase: common.PhaseCreating,
		jobStates: []jobState{
			jobTerminating,
			jobTerminatingWithRunning,
			jobTerminatingWithPending,
			jobTerminatingWithInitializing,
			jobTerminatingWithRunningAndInitializing,
			jobTerminatingWithPendingAndInitializing,
			jobTerminatingWithPendingAndRunning,
			jobTerminatingWithPendingAndRunningAndInitializing,
			jobTerminatingWithPendingAndRunningAndInitializingAndCreating,
			jobTerminatingWithSuspended,
			jobTerminatingWithFailed,
			jobTerminatingWithSucceeded,
		},

		endPhase: common.PhaseTerminating,
	}
	pendingToPending = phaseMovementTest{
		startPhase: common.PhasePending,
		jobStates: []jobState{
			jobPending,
		},

		endPhase: common.PhasePending,
	}
	pendingToFailed = phaseMovementTest{
		startPhase: common.PhasePending,
		jobStates: []jobState{
			jobFailed,
		},

		endPhase: common.PhaseFailed,
	}
	pendingToSucceeded = phaseMovementTest{
		startPhase: common.PhasePending,
		jobStates: []jobState{
			jobSucceeded,
		},

		endPhase: common.PhaseCompleted,
	}
	pendingToSuspended = phaseMovementTest{
		startPhase: common.PhasePending,
		jobStates: []jobState{
			jobSuspended,
		},

		endPhase: common.PhaseSuspended,
	}
	pendingToRunning = phaseMovementTest{
		startPhase: common.PhasePending,
		jobStates: []jobState{
			jobRunning,
			jobPendingWithRunning,
			jobInitializingWithRunning,
			jobPendingWithRunningAndInitializing,
		},

		endPhase: common.PhaseRunning,
	}
	pendingToInitializing = phaseMovementTest{
		startPhase: common.PhasePending,
		jobStates: []jobState{
			jobInitializing,
			jobPendingWithInitializing,
		},

		endPhase: common.PhaseInitializing,
	}
	pendingToTerminating = phaseMovementTest{
		startPhase: common.PhasePending,
		jobStates: []jobState{
			jobTerminating,
			jobTerminatingWithRunning,
			jobTerminatingWithPending,
			jobTerminatingWithInitializing,
			jobTerminatingWithRunningAndInitializing,
			jobTerminatingWithPendingAndInitializing,
			jobTerminatingWithPendingAndRunning,
			jobTerminatingWithPendingAndRunningAndInitializing,
			jobTerminatingWithPendingAndRunningAndInitializingAndCreating,
			jobTerminatingWithSuspended,
			jobTerminatingWithFailed,
			jobTerminatingWithSucceeded,
		},

		endPhase: common.PhaseTerminating,
	}
	initializeToInitializing = phaseMovementTest{
		startPhase: common.PhaseInitializing,
		jobStates: []jobState{
			jobInitializing,
			jobPendingWithInitializing,
		},

		endPhase: common.PhaseInitializing,
	}
	initializeToRunning = phaseMovementTest{
		startPhase: common.PhaseInitializing,
		jobStates: []jobState{
			jobRunning,
			jobPendingWithRunning,
			jobInitializingWithRunning,
			jobPendingWithRunningAndInitializing,
		},

		endPhase: common.PhaseRunning,
	}
	initializeToPending = phaseMovementTest{
		startPhase: common.PhaseInitializing,
		jobStates: []jobState{
			jobPending,
		},

		endPhase: common.PhasePending,
	}
	initializeToFailed = phaseMovementTest{
		startPhase: common.PhaseInitializing,
		jobStates: []jobState{
			jobFailed,
		},

		endPhase: common.PhaseFailed,
	}
	initializeToSucceeded = phaseMovementTest{
		startPhase: common.PhaseInitializing,
		jobStates: []jobState{
			jobSucceeded,
		},

		endPhase: common.PhaseCompleted,
	}
	initializeToSuspended = phaseMovementTest{
		startPhase: common.PhaseInitializing,
		jobStates: []jobState{
			jobSuspended,
		},

		endPhase: common.PhaseSuspended,
	}
	initializeToTerminating = phaseMovementTest{
		startPhase: common.PhaseInitializing,
		jobStates: []jobState{
			jobTerminating,
			jobTerminatingWithRunning,
			jobTerminatingWithPending,
			jobTerminatingWithInitializing,
			jobTerminatingWithRunningAndInitializing,
			jobTerminatingWithPendingAndInitializing,
			jobTerminatingWithPendingAndRunning,
			jobTerminatingWithPendingAndRunningAndInitializing,
			jobTerminatingWithPendingAndRunningAndInitializingAndCreating,
			jobTerminatingWithSuspended,
			jobTerminatingWithFailed,
			jobTerminatingWithSucceeded,
		},

		endPhase: common.PhaseTerminating,
	}
	initializeToCreating = phaseMovementTest{
		startPhase: common.PhaseInitializing,
		jobStates: []jobState{
			jobCreating,
		},

		endPhase: common.PhaseCreating,
	}
	runningToRunning = phaseMovementTest{
		startPhase: common.PhaseRunning,
		jobStates: []jobState{
			jobRunning,
			jobPendingWithRunning,
			jobInitializingWithRunning,
			jobPendingWithRunningAndInitializing,
		},

		endPhase: common.PhaseRunning,
	}
	runningToPending = phaseMovementTest{
		startPhase: common.PhaseRunning,
		jobStates: []jobState{
			jobPending,
		},

		endPhase: common.PhasePending,
	}
	runningToFailed = phaseMovementTest{
		startPhase: common.PhaseRunning,
		jobStates: []jobState{
			jobFailed,
		},

		endPhase: common.PhaseFailed,
	}
	runningToSucceeded = phaseMovementTest{
		startPhase: common.PhaseRunning,
		jobStates: []jobState{
			jobSucceeded,
		},

		endPhase: common.PhaseCompleted,
	}
	runningToSuspended = phaseMovementTest{
		startPhase: common.PhaseRunning,
		jobStates: []jobState{
			jobSuspended,
		},

		endPhase: common.PhaseSuspended,
	}
	runningToInitializing = phaseMovementTest{
		startPhase: common.PhaseRunning,
		jobStates: []jobState{
			jobInitializing,
			jobPendingWithInitializing,
		},

		endPhase: common.PhaseInitializing,
	}
	runningToTerminating = phaseMovementTest{
		startPhase: common.PhaseRunning,
		jobStates: []jobState{
			jobTerminating,
			jobTerminatingWithRunning,
			jobTerminatingWithPending,
			jobTerminatingWithInitializing,
			jobTerminatingWithRunningAndInitializing,
			jobTerminatingWithPendingAndInitializing,
			jobTerminatingWithPendingAndRunning,
			jobTerminatingWithPendingAndRunningAndInitializing,
			jobTerminatingWithPendingAndRunningAndInitializingAndCreating,
			jobTerminatingWithSuspended,
			jobTerminatingWithFailed,
			jobTerminatingWithSucceeded,
			jobTerminatingWithCreating,
		},

		endPhase: common.PhaseTerminating,
	}
	runningToCreating = phaseMovementTest{
		startPhase: common.PhaseRunning,
		jobStates: []jobState{
			jobCreating,
		},
		endPhase: common.PhaseCreating,
	}
	suspendedToSuspended = phaseMovementTest{
		startPhase: common.PhaseSuspended,
		jobStates: []jobState{
			jobSuspended,
		},

		endPhase: common.PhaseSuspended,
	}
	suspendedToPending = phaseMovementTest{
		startPhase: common.PhaseSuspended,
		jobStates: []jobState{
			jobPending,
		},

		endPhase: common.PhasePending,
	}
	suspendedToFailed = phaseMovementTest{
		startPhase: common.PhaseSuspended,
		jobStates: []jobState{
			jobFailed,
		},

		endPhase: common.PhaseFailed,
	}
	suspendedToSucceeded = phaseMovementTest{
		startPhase: common.PhaseSuspended,
		jobStates: []jobState{
			jobSucceeded,
		},

		endPhase: common.PhaseCompleted,
	}
	suspendedToRunning = phaseMovementTest{
		startPhase: common.PhaseSuspended,
		jobStates: []jobState{
			jobRunning,
			jobPendingWithRunning,
			jobInitializingWithRunning,
			jobPendingWithRunningAndInitializing,
		},

		endPhase: common.PhaseRunning,
	}
	suspendedToInitializing = phaseMovementTest{
		startPhase: common.PhaseSuspended,
		jobStates: []jobState{
			jobInitializing,
			jobPendingWithInitializing,
		},

		endPhase: common.PhaseInitializing,
	}
	suspendedToTerminating = phaseMovementTest{
		startPhase: common.PhaseSuspended,
		jobStates: []jobState{
			jobTerminating,
			jobTerminatingWithRunning,
			jobTerminatingWithPending,
			jobTerminatingWithInitializing,
			jobTerminatingWithRunningAndInitializing,
			jobTerminatingWithPendingAndInitializing,
			jobTerminatingWithPendingAndRunning,
			jobTerminatingWithPendingAndRunningAndInitializing,
			jobTerminatingWithPendingAndRunningAndInitializingAndCreating,
			jobTerminatingWithSuspended,
			jobTerminatingWithFailed,
			jobTerminatingWithSucceeded,
			jobTerminatingWithCreating,
		},

		endPhase: common.PhaseTerminating,
	}
	terminatingToTerminating = phaseMovementTest{
		startPhase: common.PhaseTerminating,
		jobStates: []jobState{
			jobTerminating,
			jobTerminatingWithRunning,
			jobTerminatingWithPending,
			jobTerminatingWithInitializing,
			jobTerminatingWithRunningAndInitializing,
			jobTerminatingWithPendingAndInitializing,
			jobTerminatingWithPendingAndRunning,
			jobTerminatingWithPendingAndRunningAndInitializing,
			jobTerminatingWithPendingAndRunningAndInitializingAndCreating,
			jobTerminatingWithSuspended,
			jobTerminatingWithFailed,
			jobTerminatingWithSucceeded,
			jobTerminatingWithCreating,
		},

		endPhase: common.PhaseTerminating,
	}
	runningToDegraded = phaseMovementTest{
		startPhase: common.PhaseRunning,
		jobStates: []jobState{
			jobPendingWithRunning,
			jobInitializingWithRunning,
			jobPendingWithRunningAndInitializing,
		},
		endPhase: common.PhaseDegraded,
	}
	DegradedToRunning = phaseMovementTest{
		startPhase: common.PhaseDegraded,
		jobStates: []jobState{
			jobRunning,
		},
		endPhase: common.PhaseRunning,
	}
	DegradedToInitializing = phaseMovementTest{
		startPhase: common.PhaseDegraded,
		jobStates: []jobState{
			jobInitializing,
			jobPendingWithInitializing,
		},
		endPhase: common.PhaseInitializing,
	}
	DegradedToPending = phaseMovementTest{
		startPhase: common.PhaseDegraded,
		jobStates: []jobState{
			jobPending,
		},
		endPhase: common.PhasePending,
	}
	DegradedToFailed = phaseMovementTest{
		startPhase: common.PhaseDegraded,
		jobStates: []jobState{
			jobFailed,
		},
		endPhase: common.PhaseFailed,
	}
	DegradedToSucceeded = phaseMovementTest{
		startPhase: common.PhaseDegraded,
		jobStates: []jobState{
			jobSucceeded,
		},
		endPhase: common.PhaseCompleted,
	}
	DegradedToSuspended = phaseMovementTest{
		startPhase: common.PhaseDegraded,
		jobStates: []jobState{
			jobSuspended,
		},
		endPhase: common.PhaseSuspended,
	}
	DegradedToTerminating = phaseMovementTest{
		startPhase: common.PhaseDegraded,
		jobStates: []jobState{
			jobTerminating,
			jobTerminatingWithRunning,
			jobTerminatingWithPending,
			jobTerminatingWithInitializing,
			jobTerminatingWithRunningAndInitializing,
			jobTerminatingWithPendingAndInitializing,
			jobTerminatingWithPendingAndRunning,
			jobTerminatingWithPendingAndRunningAndInitializing,
			jobTerminatingWithPendingAndRunningAndInitializingAndCreating,
			jobTerminatingWithSuspended,
			jobTerminatingWithFailed,
			jobTerminatingWithSucceeded,
			jobTerminatingWithCreating,
		},
		endPhase: common.PhaseTerminating,
	}
)

var _ = Describe("WorkloadStatusCalculator", func() {
	var (
		testEntries []TableEntry
	)
	Context("Test phase movement", func() {
		testEntries = []TableEntry{
			Entry("creating to running", creatingToRunning),
			Entry("creating to pending", creatingToPending),
			Entry("creating to creating", creatingToCreating),
			Entry("creating to failed", creatingToFailed),
			Entry("creating to succeeded", creatingToSucceeded),
			Entry("creating to suspended", creatingToSuspended),
			Entry("creating to initializing", creatingToInitializing),
			Entry("creating to terminating", creatingToTerminating),
			Entry("pending to pending", pendingToPending),
			Entry("pending to failed", pendingToFailed),
			Entry("pending to succeeded", pendingToSucceeded),
			Entry("pending to suspended", pendingToSuspended),
			Entry("pending to running", pendingToRunning),
			Entry("pending to initializing", pendingToInitializing),
			Entry("pending to terminating", pendingToTerminating),
			Entry("initialize to initializing", initializeToInitializing),
			Entry("initialize to running", initializeToRunning),
			Entry("initialize to pending", initializeToPending),
			Entry("initialize to failed", initializeToFailed),
			Entry("initialize to succeeded", initializeToSucceeded),
			Entry("initialize to suspended", initializeToSuspended),
			Entry("initialize to terminating", initializeToTerminating),
			Entry("initialize to creating", initializeToCreating),
			Entry("running to running", runningToRunning),
			Entry("running to pending", runningToPending),
			Entry("running to failed", runningToFailed),
			Entry("running to succeeded", runningToSucceeded),
			Entry("running to suspended", runningToSuspended),
			Entry("running to initializing", runningToInitializing),
			Entry("running to terminating", runningToTerminating),
			Entry("running to creating", runningToCreating),
			Entry("suspended to suspended", suspendedToSuspended),
			Entry("suspended to pending", suspendedToPending),
			Entry("suspended to failed", suspendedToFailed),
			Entry("suspended to succeeded", suspendedToSucceeded),
			Entry("suspended to running", suspendedToRunning),
			Entry("suspended to initializing", suspendedToInitializing),
			Entry("suspended to terminating", suspendedToTerminating),
			Entry("pending to terminating", pendingToTerminating),
			Entry("terminating to terminating", terminatingToTerminating),
			Entry("degraded to running", DegradedToRunning),
			Entry("degraded to initializing", DegradedToInitializing),
			Entry("degraded to pending", DegradedToPending),
			Entry("degraded to failed", DegradedToFailed),
			Entry("degraded to succeeded", DegradedToSucceeded),
			Entry("degraded to suspended", DegradedToSuspended),
			Entry("degraded to terminating", DegradedToTerminating),
		}
		degradedTestEntries := []TableEntry{
			Entry("running to degraded", runningToDegraded),
		}
		DescribeTable("Test phase movement", func(test phaseMovementTest) {
			runTestsAndAssert(test)

		}, testEntries)
		Context("Test running to degraded phase movement", func() {
			var (
				issues []*issue.Issue
			)
			BeforeEach(func() {
				service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{
					Name: "service01",
				}}
				issues = []*issue.Issue{
					issue2.NewIssue(service, errors.New("danger danger"), true, ""),
				}
			})
			DescribeTable("", func(test phaseMovementTest) {
				for _, state := range test.jobStates {
					workload := calculatePhase(test, state, nil, issues)
					Expect(workload.Status.WorkloadCommonStatus.Phase).To(Equal(test.endPhase.String()),
						"start phase: %s, end phase: %s job state: %s", test.startPhase, test.endPhase, state.String())
				}

			}, degradedTestEntries)
		})

	})
})

func runTestsAndAssert(test phaseMovementTest) {
	for _, state := range test.jobStates {
		workload := calculatePhase(test, state, nil, nil)
		Expect(workload.Status.WorkloadCommonStatus.Phase).To(Equal(test.endPhase.String()), "start phase: %s, end phase: %s job state: %s", test.startPhase, test.endPhase, state.String())
	}
}

func calculatePhase(test phaseMovementTest, state jobState, currentWorkload *job.JobWorkload, issues []*issue.Issue) *job.JobWorkload {
	ctx := context.Background()
	status := &mocks.JobStatus{}
	jobAdapter := &mocks.JobAdapter{}
	statusCalculator := NewStatusCalculator(false)
	var workload *job.JobWorkload
	if currentWorkload != nil {
		workload = currentWorkload
	} else {
		workload = &job.JobWorkload{}
		workload.Status.WorkloadCommonStatus.Phase = test.startPhase.String()
	}
	workload.Spec.Active = fields.NewOptionalBoolField(ptr.To(!state.isSuspended))

	aggr := resource.NewAggregator(workload)
	if len(issues) > 0 {
		aggr.AddIssues(issues...)
	}

	setJobMock(status, jobAdapter, state.isCreating, state.isRunning, state.isPending, state.isJobFailed,
		state.isSucceeded, state.isSuspended, state.isInitializing, state.isTerminating, state.annotations)
	statusCalculator.WithJob(jobAdapter)
	statusCalculator.WithWorkload(workload)
	statusCalculator.CalculateAll(ctx, aggr)
	return workload
}
