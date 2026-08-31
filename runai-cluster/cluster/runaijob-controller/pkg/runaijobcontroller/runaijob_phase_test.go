/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaijobcontroller

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	runaiPodPhase "github.com/run-ai/runai/runai-cluster/cluster/common/utils/pod"
	runaijobv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runaijobinformers "github.com/run-ai/runai/runai-cluster/cluster/sdk/client/informers/externalversions"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/controller"
)

func TestRunAIPodPhaseAndMessage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TestRunAIJobPhases Test Suite")
}

var _ = Describe("TestRunAIJobPhases", func() {
	Context("Pod to RunAI phase", func() {
		const (
			readyPods        int32 = 6
			runningPodsFalse int32 = 5
			schedTrue        int32 = 4
			schedFalse       int32 = 3
			initTrue         int32 = 2
			initFalse        int32 = 1
			createdPods      int32 = 11
			pendingPods      int32 = 12
			successPods      int32 = 13
			failedPods       int32 = 14
		)

		var (
			//schemes *runtime.Scheme
			actual                      *runaijobv1.RunaiJob
			manager                     *Controller
			sharedInformerFactory       informers.SharedInformerFactory
			runaiSharedInformersFactory runaijobinformers.SharedInformerFactory
		)

		BeforeEach(func() {
			manager, sharedInformerFactory, runaiSharedInformersFactory = newController()
			fakePodControl := controller.FakePodControl{}
			manager.podControl = &fakePodControl
			manager.control = &FakePodControl{}
			manager.recorder = record.NewFakeRecorder(1024)

			manager.updateHandler = func(job *runaijobv1.RunaiJob) error {
				actual = job
				return nil
			}
		})
		It("Phase and reason check", func() {
			job := newJob(99, 99, 99)
			err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job)
			if err != nil {
				klog.Errorf("Failed to add runaiJob, error: <%v>", err)
			}
			configmap := v1.ConfigMap{
				TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
				ObjectMeta: metav1.ObjectMeta{
					Name:            fmt.Sprintf("%s-pod-attempts", job.Name),
					Namespace:       job.Namespace,
					UID:             uuid.NewUUID(),
					ResourceVersion: "1",
					OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(job, controllerKind)},
				},
				Data: make(map[string]string),
			}
			index := 0
			podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()

			// created
			newPodList(podIndexer, createdPods, v1.PodPending, nil, job, &index, &configmap, nil)
			// pending
			newPodList(podIndexer, pendingPods, v1.PodRunning, nil, job, &index, &configmap, nil)
			newPodList(podIndexer, successPods, v1.PodSucceeded, nil, job, &index, &configmap, nil)
			newPodList(podIndexer, failedPods, v1.PodFailed, nil, job, &index, &configmap, nil)

			newPodList(podIndexer, initFalse, v1.PodRunning, createCondition(CreateConditionInitializingFalse), job, &index, &configmap, nil)
			newPodList(podIndexer, initTrue, v1.PodRunning, createCondition(CreateConditionInitializingTrue), job, &index, &configmap, nil)
			newPodList(podIndexer, schedFalse, v1.PodRunning, createCondition(CreateConditionScheduledFalse), job, &index, &configmap, nil)
			newPodList(podIndexer, schedTrue, v1.PodRunning, createCondition(CreateConditionScheduledTrue), job, &index, &configmap, nil)
			newPodList(podIndexer, runningPodsFalse, v1.PodRunning, createCondition(CreateConditionRunningFalse), job, &index, &configmap, nil)
			newPodList(podIndexer, readyPods, v1.PodRunning, createCondition(CreateConditionRunningTrue), job, &index, &configmap, nil)
			err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(&configmap)
			if err != nil {
				klog.Errorf("failed to add configmap")
			}

			_, err = manager.syncJob(GetKey(job))

			Expect(fmt.Sprintf("%s/%s", actual.Namespace, actual.Name)).To(Equal(fmt.Sprintf("%s/%s", job.Namespace, job.Name)))
			Expect(actual.Status.Ready).To(Equal(readyPods))
			Expect(actual.Status.Initialized).To(Equal(runningPodsFalse + initTrue + initFalse + schedFalse + schedTrue))
			Expect(actual.Status.Pending).To(Equal(pendingPods + createdPods))
			Expect(actual.Status.Failed).To(Equal(failedPods))
			Expect(actual.Status.Succeeded).To(Equal(successPods))

			conditions := actual.Status.Conditions
			Expect(len(conditions)).To(Equal(3))

			var foundPending, foundRunning, foundInit bool
			for _, condition := range conditions {
				switch condition.Type {
				case runaijobv1.JobPending:
					foundPending = true
					Expect(condition.Status).To(Equal(v1.ConditionTrue))
					Expect(condition.Message).To(Equal(runaiPodPhase.EmptyStr))
				case runaijobv1.JobRunning:
					foundRunning = true
					Expect(condition.Status).To(Equal(v1.ConditionTrue))
					Expect(condition.Message).To(Equal(runaiPodPhase.EmptyStr))
				case runaijobv1.JobInitializing:
					foundInit = true
					Expect(condition.Status).To(Equal(v1.ConditionTrue))
					Expect(countNumberOfAppearancesInString(condition.Message, "InitFalse")).To(Equal(initFalse))
					Expect(countNumberOfAppearancesInString(condition.Message, "SchedFalse")).To(Equal(schedFalse))
					Expect(countNumberOfAppearancesInString(condition.Message, "RunFalse")).To(Equal(runningPodsFalse))
				}
			}
			Expect(foundPending).To(Equal(true))
			Expect(foundRunning).To(Equal(true))
			Expect(foundInit).To(Equal(true))
		})

		It("Phase and reason check condition empty", func() {
			job := newJob(99, 99, 99)
			job.Name = "conditions"
			err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job)
			if err != nil {
				klog.Errorf("Failed to add runaiJob, error: <%v>", err)
			}
			configmap := v1.ConfigMap{
				TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
				ObjectMeta: metav1.ObjectMeta{
					Name:            fmt.Sprintf("%s-pod-attempts", job.Name),
					Namespace:       job.Namespace,
					UID:             uuid.NewUUID(),
					ResourceVersion: "1",
					OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(job, controllerKind)},
				},
				Data: make(map[string]string),
			}
			index := 0
			podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()

			newPodList(podIndexer, createdPods, v1.PodPending, nil, job, &index, &configmap, nil)
			newPodList(podIndexer, initTrue, v1.PodRunning, createCondition(CreateConditionInitializingTrue), job, &index, &configmap, nil)
			newPodList(podIndexer, pendingPods, v1.PodRunning, nil, job, &index, &configmap, nil)

			err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(&configmap)
			if err != nil {
				klog.Errorf("failed to add configmap")
			}

			_, err = manager.syncJob(GetKey(job))

			Expect(actual.Status.Initialized).To(Equal(initTrue))
			Expect(actual.Status.Pending).To(Equal(pendingPods + createdPods))

			conditions := actual.Status.Conditions
			Expect(len(conditions)).To(Equal(2))

			var foundPending, foundRunning, foundInit bool
			for _, condition := range conditions {
				switch condition.Type {
				case runaijobv1.JobPending:
					foundPending = true
					Expect(condition.Status).To(Equal(v1.ConditionTrue))
					Expect(condition.Message).To(Equal(runaiPodPhase.EmptyStr))
				case runaijobv1.JobInitializing:
					foundInit = true
					Expect(condition.Status).To(Equal(v1.ConditionTrue))
					Expect(condition.Message).To(Equal(runaiPodPhase.EmptyStr))
				}
			}

			Expect(foundPending).To(Equal(true))
			Expect(foundRunning).To(Equal(false))
			Expect(foundInit).To(Equal(true))

			// TODO add pod deletion and show condition exist with status false
		})
	})
})

func countNumberOfAppearancesInString(sentence string, lookup string) int32 {
	var counter int32
	words := strings.Split(sentence, ",")
	for _, word := range words {
		if strings.Contains(word, lookup) {
			counter++
		}
	}
	return counter
}
