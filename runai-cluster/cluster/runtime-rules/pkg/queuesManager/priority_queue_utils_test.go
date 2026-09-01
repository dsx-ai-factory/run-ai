/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package queuesManager

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/logger"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/validation"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var eventList = []int{50, 75, 90}

func SetEnvironment(t *testing.T) {
	t.Setenv("EVENT_LIST", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(eventList)), ","), "[]"))
}

func TestRuntimeRules(t *testing.T) {
	logger.InitLogger(int(zap.DebugLevel))
	zap.S().Infof("Started service tests")
	RegisterFailHandler(Fail)
	SetEnvironment(t)
	RunSpecs(t, "runtime-rules service Test Suite")
}

var _ = Describe("getMessage Test Suite", Ordered, func() {
	Context("getMessage function test", func() {
		timeNow := time.Now()
		It("Interactive Workload: check idle flow of events (naive)", func() {

			eventClock1 := EventClock{NextEventType: TaskTimerIdle}
			eventClock2 := EventClock{NextEventType: TaskTimerTimeout}

			eventClock1.DeactivateEvent = timeNow.Add(time.Minute * 5)
			eventClock2.DeactivateEvent = timeNow.Add(time.Minute * 5)
			msg := getMessage("test-message", eventClock1)
			Expect(msg).To(Equal("Based on the idle GPU timeout configured for your project, as long as the GPU(s) used by test-message remain idle, the job will be terminated in 4 minutes."))
			msg = getMessage("test-message", eventClock2)
			Expect(msg).To(Equal("Based on the time limit configured for your project, the job test-message will be terminated in 4 minutes."))

			eventClock1.DeactivateEvent = eventClock1.DeactivateEvent.Add(time.Hour * 3)
			eventClock2.DeactivateEvent = eventClock2.DeactivateEvent.Add(time.Hour * 3)
			msg = getMessage("test-message", eventClock1)
			Expect(msg).To(Equal("Based on the idle GPU timeout configured for your project, as long as the GPU(s) used by test-message remain idle, the job will be terminated in 3 hours - 4 minutes."))
			msg = getMessage("test-message", eventClock2)
			Expect(msg).To(Equal("Based on the time limit configured for your project, the job test-message will be terminated in 3 hours - 4 minutes."))

			eventClock1.DeactivateEvent = timeNow.Add(time.Hour * 30)
			eventClock2.DeactivateEvent = timeNow.Add(time.Hour * 30)
			msg = getMessage("test-message", eventClock1)
			Expect(msg).To(Equal("Based on the idle GPU timeout configured for your project, as long as the GPU(s) used by test-message remain idle, the job will be terminated in 1 days - 5 hours - 59 minutes."))
			msg = getMessage("test-message", eventClock2)
			Expect(msg).To(Equal("Based on the time limit configured for your project, the job test-message will be terminated in 1 days - 5 hours - 59 minutes."))
		})
	})
})

var _ = Describe("getLimitsByWorkloadType — WorkloadRunner category routing", func() {
	const (
		iwTime    = int64(100)
		iwIdle    = int64(200)
		iwPreempt = int64(250)
		trainTime = int64(300)
		trainIdle = int64(400)
	)
	pq := &ProjectQueue{
		InteractiveJobTimeLimitSecs:                  iwTime,
		InteractiveJobMaxIdleDurationSecs:            iwIdle,
		InteractivePreemptibleJobMaxIdleDurationSecs: iwPreempt,
		TrainingJobTimeLimitSecs:                     trainTime,
		TrainingJobMaxIdleDurationSecs:               trainIdle,
	}

	It("routes an Interactive-category runner to interactive limits", func() {
		task := &WorkloadTask{WorkloadType: constants.KindWorkloadRunner, Category: constants.KindInteractiveWorkload}
		timeLimit, idleLimit, err := getLimitsByWorkloadType(pq, task)
		Expect(err).To(BeNil())
		Expect(timeLimit).To(Equal(iwTime))
		Expect(idleLimit).To(Equal(iwIdle))
	})

	It("uses the preemptible idle limit for a preemptible Interactive runner", func() {
		task := &WorkloadTask{WorkloadType: constants.KindWorkloadRunner, Category: constants.KindInteractiveWorkload, IsPreemptible: true}
		_, idleLimit, err := getLimitsByWorkloadType(pq, task)
		Expect(err).To(BeNil())
		Expect(idleLimit).To(Equal(iwPreempt))
	})

	It("routes Training and Distributed runners to training limits", func() {
		for _, category := range []string{constants.KindTrainingWorkload, constants.KindDistributedWorkload} {
			task := &WorkloadTask{WorkloadType: constants.KindWorkloadRunner, Category: category}
			timeLimit, idleLimit, err := getLimitsByWorkloadType(pq, task)
			Expect(err).To(BeNil())
			Expect(timeLimit).To(Equal(trainTime), "category %s", category)
			Expect(idleLimit).To(Equal(trainIdle), "category %s", category)
		}
	})

	It("yields zero limits (observed but not enforced) when Category is empty", func() {
		task := &WorkloadTask{WorkloadType: constants.KindWorkloadRunner}
		timeLimit, idleLimit, err := getLimitsByWorkloadType(pq, task)
		Expect(err).To(BeNil())
		Expect(timeLimit).To(BeZero())
		Expect(idleLimit).To(BeZero())
	})

})

var _ = Describe("HandleWorkloadTask — WorkloadRunner timeout enforcement", func() {
	const namespace = "runner-timeout-tests"

	newRunnerClient := func(name, apiVersion, kind string) client.Client {
		template, err := json.Marshal(map[string]any{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata":   map[string]any{"name": name},
		})
		Expect(err).NotTo(HaveOccurred())

		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				Namespace:         namespace,
				CreationTimestamp: metav1.NewTime(time.Now().Add(-2 * time.Hour)),
				Annotations: map[string]string{
					constants.StartTimeAnnotation: time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339),
				},
			},
			Spec: dromosv1alpha1.WorkloadRunnerSpec{
				Template: apiruntime.RawExtension{Raw: template},
			},
			Status: dromosv1alpha1.WorkloadRunnerStatus{
				Phase: dromosv1alpha1.WorkloadRunnerPhaseRunning,
			},
		}

		scheme := apiruntime.NewScheme()
		Expect(dromosv1alpha1.AddToScheme(scheme)).To(Succeed())
		return fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(runner).Build()
	}

	templateSuspend := func(ctx context.Context, k8sClient client.Client, name string, path ...string) (bool, bool) {
		var runner dromosv1alpha1.WorkloadRunner
		Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &runner)).To(Succeed())
		obj := &unstructured.Unstructured{}
		Expect(json.Unmarshal(runner.Spec.Template.Raw, &obj.Object)).To(Succeed())
		value, found, err := unstructured.NestedBool(obj.Object, path...)
		Expect(err).NotTo(HaveOccurred())
		return value, found
	}

	newProjectQueue := func(category string) *ProjectQueue {
		queue := &ProjectQueue{
			ProjectName:       "timeout-project",
			queue:             NewPriorityQueue(),
			Recorder:          record.NewFakeRecorder(10),
			EventList:         []int{DeactivateEvent},
			WorkloadValidator: validation.NewWorkloadValidator(),
		}
		if category == constants.KindInteractiveWorkload {
			queue.InteractiveJobTimeLimitSecs = 60
		} else {
			queue.TrainingJobTimeLimitSecs = 60
		}
		return queue
	}

	newTask := func(name, category string, startTime time.Time) *WorkloadTask {
		return &WorkloadTask{
			StartTime:    startTime,
			WorkloadName: name,
			ProjectName:  "timeout-project",
			Namespace:    namespace,
			JobName:      name,
			WorkloadType: constants.KindWorkloadRunner,
			Category:     category,
			TaskAction:   *NewActionList([]int{DeactivateEvent}, 0, 0, nil),
		}
	}

	DescribeTable("suspends an expired runner through the queue-worker path",
		func(category, apiVersion, wrappedKind string, suspendPath []string) {
			ctx := context.Background()
			name := "expired-" + wrappedKind
			k8sClient := newRunnerClient(name, apiVersion, wrappedKind)
			projectQueue := newProjectQueue(category)

			err := HandleWorkloadTask(ctx, nil, k8sClient, nil,
				newTask(name, category, time.Now().Add(-2*time.Hour)), projectQueue, true)
			Expect(err).NotTo(HaveOccurred())

			suspended, found := templateSuspend(ctx, k8sClient, name, suspendPath...)
			Expect(found).To(BeTrue())
			Expect(suspended).To(BeTrue())
			Expect(projectQueue.IsEmpty()).To(BeTrue())
		},
		Entry("Interactive", constants.KindInteractiveWorkload, "batch/v1", "Job", []string{"spec", "suspend"}),
		Entry("Training", constants.KindTrainingWorkload, "batch/v1", "Job", []string{"spec", "suspend"}),
		Entry("Distributed", constants.KindDistributedWorkload, "kubeflow.org/v1", constants.KindMpiJob, []string{"spec", "runPolicy", "suspend"}),
	)

	It("keeps a runner whose deadline is in the future queued and unsuspended", func() {
		ctx := context.Background()
		const name = "future-Job"
		k8sClient := newRunnerClient(name, "batch/v1", "Job")
		projectQueue := newProjectQueue(constants.KindInteractiveWorkload)

		err := HandleWorkloadTask(ctx, nil, k8sClient, nil,
			newTask(name, constants.KindInteractiveWorkload, time.Now()), projectQueue, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(projectQueue.queue.Len()).To(Equal(1))

		_, found := templateSuspend(ctx, k8sClient, name, "spec", "suspend")
		Expect(found).To(BeFalse())
	})
})
