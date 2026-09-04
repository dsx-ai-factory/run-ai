/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	mpi "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v1alpha2"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testProjectName   = "test-project"
	testNamespaceName = "runai-test-namespace"
	RunaiQueueLabel   = "runai/queue"
)

var workerSleepSeconds = 5
var eventList = []int{50}

// MockEnv Notice that after test is done env variables are cleaned up
func MockEnv(t *testing.T) {
	t.Setenv("NUMBER_OF_WORKERS_PER_QUEUE", "2")
	t.Setenv("WORKER_SLEEP_SECONDS", strconv.Itoa(workerSleepSeconds))
	t.Setenv("EVENT_LIST", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(eventList)), ","), "[]"))
}

func GenerateWorkloadName() string {
	return fmt.Sprintf("test-workload-%s", rand.String(5))
}

func BuildNamespace(name string, labels map[string]string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

var (
	testSchemeOnce sync.Once
	testSchemeInst *apiruntime.Scheme
)

func initSchema() *apiruntime.Scheme {
	testSchemeOnce.Do(func() {
		scheme := scheme_init.Scheme()
		if err := mpi.AddToScheme(scheme); err != nil {
			zap.S().Error("failed to add mpijob to scheme", err)
		}
		if err := dromosv1alpha1.AddToScheme(scheme); err != nil {
			zap.S().Error("failed to add workloadrunner to scheme", err)
		}
		testSchemeInst = scheme
	})
	return testSchemeInst
}

// BuildProjectFixtures returns the project objects to seed a fake client with.
func BuildProjectFixtures(projectName, testNamespaceName string, trainingJobTimeLimitSecs, InteractiveJobTimeLimitSecs int) []client.Object {
	return []client.Object{BuildKaiProject(projectName, testNamespaceName, trainingJobTimeLimitSecs, InteractiveJobTimeLimitSecs)}
}

// BuildKaiProject builds a kai.resources project. The CRD has no home for the job time limits,
// so they live on annotations - which is where the runtime-rules read them from.
func BuildKaiProject(projectName, testNamespaceName string, trainingJobTimeLimitSecs, InteractiveJobTimeLimitSecs int) *kaiorg.Project {
	return &kaiorg.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
			Annotations: map[string]string{
				constants.TrainingJobTimeLimitSecondsAnnotation:    strconv.Itoa(trainingJobTimeLimitSecs),
				constants.InteractiveJobTimeLimitSecondsAnnotation: strconv.Itoa(InteractiveJobTimeLimitSecs),
			},
		},
		Spec: kaiorg.ProjectSpec{
			Queues: []kaiorg.QueueConfig{
				{
					Name: projectName,
				},
			},
			Namespace: testNamespaceName,
		},
		Status: kaiorg.ProjectStatus{
			Namespace: testNamespaceName,
		},
	}
}
