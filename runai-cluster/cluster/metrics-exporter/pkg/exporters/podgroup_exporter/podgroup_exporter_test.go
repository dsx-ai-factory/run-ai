/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package podgroup_exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/metrics_exporter/pkg/exporters/common"
	runaiv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"

	"github.com/stretchr/testify/assert"
	k8sschedulingv1 "k8s.io/api/scheduling/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	nodepoolName = "my-nodepool-2"

	testDepartmentName      = "my-department"
	testDepartmentId        = "7"
	testDepartmentQueueName = "q-7"
)

func TestGetName(t *testing.T) {
	podGroup := createPodGroup()
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	assert.Equal(t, "pg-name", exporter.GetName())
}

func TestGetNamespace(t *testing.T) {
	podGroup := createPodGroup()
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	assert.Equal(t, "some-namespace", exporter.GetNamespace())
}

func TestGetQueue(t *testing.T) {
	podGroup := createPodGroup()
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	assert.Equal(t, "some-queue", exporter.GetQueueName())
}

func TestGetProjectName(t *testing.T) {
	podGroup := createPodGroup()
	queue := createQueue()
	fakeClient := createFakeClient([]client.Object{queue})
	exporter := NewPodGroupExporter(context.TODO(), fakeClient, podGroup)

	project, err := exporter.GetProjectName()
	assert.Nil(t, err)
	assert.Equal(t, "some-project", project)
}

func TestGetProjectNameNotFound(t *testing.T) {
	podGroup := createPodGroup()
	queue := createQueue()
	queue.Labels = map[string]string{}
	fakeClient := createFakeClient([]client.Object{queue})
	exporter := NewPodGroupExporter(context.TODO(), fakeClient, podGroup)

	_, err := exporter.GetProjectName()
	assert.NotNil(t, err)
}

func TestGetNodePoolName(t *testing.T) {
	podGroup := createPodGroup()
	queue := createQueue()
	fakeClient := createFakeClient([]client.Object{queue})
	exporter := NewPodGroupExporter(context.TODO(), fakeClient, podGroup)

	assert.Equal(t, nodepoolName, exporter.GetNodePoolName())
}

func TestGetNodePoolNameLabelNotFound(t *testing.T) {
	podGroup := createPodGroup()
	delete(podGroup.Labels, runaiv1alpha1.NodePoolNameLabel)
	queue := createQueue()
	fakeClient := createFakeClient([]client.Object{queue})
	exporter := NewPodGroupExporter(context.TODO(), fakeClient, podGroup)

	assert.Equal(t, runaiv1alpha1.DefaultNodePoolName, exporter.GetNodePoolName())
}

func TestGetProjectNameNoQueue(t *testing.T) {
	podGroup := createPodGroup()
	fakeClient := createFakeClient(nil)
	exporter := NewPodGroupExporter(context.TODO(), fakeClient, podGroup)

	_, err := exporter.GetProjectName()
	assert.NotNil(t, err)
}

func TestGetDepartmentName(t *testing.T) {
	podGroup := createPodGroup()
	queue := createQueue()
	parentQueue := createQueueInner(testDepartmentQueueName, "", "", testDepartmentName, testDepartmentId)
	fakeClient := createFakeClient([]client.Object{queue, parentQueue})
	exporter := NewPodGroupExporter(context.TODO(), fakeClient, podGroup)

	departmentName, err := exporter.GetDepartmentName()
	assert.Nil(t, err)
	assert.Equal(t, testDepartmentName, departmentName)
}

func TestGetDepartmentNameNoQueue(t *testing.T) {
	podGroup := createPodGroup()
	fakeClient := createFakeClient([]client.Object{})
	exporter := NewPodGroupExporter(context.TODO(), fakeClient, podGroup)

	departmentName, err := exporter.GetDepartmentName()
	assert.NotNil(t, err)
	assert.Equal(t, "", departmentName)
}

func TestGetDepartmentNameNoParentQueue(t *testing.T) {
	podGroup := createPodGroup()
	queue := createQueue()
	fakeClient := createFakeClient([]client.Object{queue})
	exporter := NewPodGroupExporter(context.TODO(), fakeClient, podGroup)

	departmentName, err := exporter.GetDepartmentName()
	assert.NotNil(t, err)
	assert.Equal(t, "", departmentName)
}

func TestGetDepartmentNameParentQueueSpecEmpty(t *testing.T) {
	podGroup := createPodGroup()
	queue := createQueue()
	queue.Spec.ParentQueue = ""
	parentQueue := createQueueInner(testDepartmentQueueName, "", "", testDepartmentName, testDepartmentId)
	fakeClient := createFakeClient([]client.Object{queue, parentQueue})
	exporter := NewPodGroupExporter(context.TODO(), fakeClient, podGroup)

	departmentName, err := exporter.GetDepartmentName()
	assert.NotNil(t, err)
	assert.Equal(t, "", departmentName)
}

func TestGetDepartmentNameNoLabelOnParentQueue(t *testing.T) {
	podGroup := createPodGroup()
	queue := createQueue()
	parentQueue := createQueueInner(testDepartmentQueueName, "", "", "", testDepartmentId)
	fakeClient := createFakeClient([]client.Object{queue, parentQueue})
	exporter := NewPodGroupExporter(context.TODO(), fakeClient, podGroup)

	departmentName, err := exporter.GetDepartmentName()
	assert.NotNil(t, err)
	assert.Equal(t, "", departmentName)
}

func TestGetUserFromAnnotation(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Annotations[userAnnotationOrLabel] = "some-user"
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	assert.Equal(t, "some-user", exporter.GetUser())
}

func TestGetUserFromLabel(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Labels[userAnnotationOrLabel] = "some-user"
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	assert.Equal(t, "some-user", exporter.GetUser())
}

func TestGetUserNotFound(t *testing.T) {
	podGroup := createPodGroup()
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	assert.Empty(t, exporter.GetUser())
}

func TestGetNodes_OneNode(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Annotations[usedNodesAnnotation] = "some-node1"
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	nodes, err := exporter.GetNodes()
	assert.Nil(t, err)
	assert.Equal(t, "some-node1", nodes)
}

func TestGetNodes_MultipleNodes(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Annotations[usedNodesAnnotation] = "some-node1, some-node2"
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	nodes, err := exporter.GetNodes()
	assert.Nil(t, err)
	assert.Equal(t, "-", nodes)
}

func TestGetNodesNotFound(t *testing.T) {
	podGroup := createPodGroup()
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	_, err := exporter.GetNodes()
	assert.NotNil(t, err)
}

func TestGetStatus(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Annotations[calculatedStatusAnnotation] = "Running"
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	status, err := exporter.GetStatus()
	assert.Nil(t, err)
	assert.Equal(t, "Running", status)
}

func TestGetStatusNotFound(t *testing.T) {
	podGroup := createPodGroup()
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	err, _ := exporter.GetStatus()
	assert.NotNil(t, err)
}

func TestGetNumRunningPods(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Annotations[runningPods] = "5"
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	numRunningPods, err := exporter.GetNumRunningPods()
	assert.Nil(t, err)
	assert.Equal(t, int64(5), numRunningPods)
}

func TestGetNumRunningPodsNotFound(t *testing.T) {
	podGroup := createPodGroup()
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	_, err := exporter.GetNumRunningPods()
	assert.NotNil(t, err)
}

func TestGetNumRunningPodsNotValid(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Annotations[runningPods] = "five"
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	_, err := exporter.GetNumRunningPods()
	assert.NotNil(t, err)
}

func TestGetAllocatedGPUs(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Annotations[allocatedGPUsAnnotation] = "1.234"
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	allocatedGPUs, err := exporter.GetAllocatedGPUs()
	assert.Nil(t, err)
	assert.Equal(t, 1.234, allocatedGPUs)
}

func TestGetAllocatedGPUsNotFound(t *testing.T) {
	podGroup := createPodGroup()
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	_, err := exporter.GetAllocatedGPUs()
	assert.NotNil(t, err)
}

func TestGetAllocatedGPUsNotValid(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Annotations[allocatedGPUsAnnotation] = "half-gpu"
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	_, err := exporter.GetAllocatedGPUs()
	assert.NotNil(t, err)
}

func TestGetRequestedGPUs(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Annotations[requestedGPUsAnnotation] = "2.456"
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	allocatedGPUs, err := exporter.GetRequestedGPUs()
	assert.Nil(t, err)
	assert.Equal(t, 2.456, allocatedGPUs)
}

func TestGetRequestedGPUsNotFound(t *testing.T) {
	podGroup := createPodGroup()
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	_, err := exporter.GetRequestedGPUs()
	assert.NotNil(t, err)
}

func TestGetRequestedGPUsNotValid(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Annotations[allocatedGPUsAnnotation] = "quarter-gpu"
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	_, err := exporter.GetRequestedGPUs()
	assert.NotNil(t, err)
}

func TestGetRequestedGPUsMemory(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Annotations[requestedGPUsMemoryAnnotation] = "5.678"
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	allocatedGPUs, err := exporter.GetRequestedGPUsMemory()
	assert.Nil(t, err)
	assert.Equal(t, 5.678, allocatedGPUs)
}

func TestGetRequestedGPUsMemoryNotFound(t *testing.T) {
	podGroup := createPodGroup()
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	_, err := exporter.GetRequestedGPUsMemory()
	assert.NotNil(t, err)
}

func TestGetRequestedGPUsMemoryNotValid(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Annotations[requestedGPUsMemoryAnnotation] = "one-third-gpu"
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	_, err := exporter.GetRequestedGPUsMemory()
	assert.NotNil(t, err)
}

func TestGetWorkloadTypeDefaultPriority(t *testing.T) {
	podGroup := createPodGroup()
	fakeClient := createFakeClient(nil)
	exporter := NewPodGroupExporter(context.TODO(), fakeClient, podGroup)

	workloadType := exporter.GetWorkloadType()
	assert.Equal(t, jobTypeTrain, workloadType)
}

func TestGetWorkloadTypeDefaultPriorityCouldNotList(t *testing.T) {
	podGroup := createPodGroup()
	fakeClient := createFakeClient(nil)
	fakeErroringClient := &possiblyErroringFakeCtrlRuntimeClient{
		fakeClient,
		true,
	}
	exporter := NewPodGroupExporter(context.TODO(), fakeErroringClient, podGroup)

	workloadType := exporter.GetWorkloadType()
	assert.Equal(t, jobTypeTrain, workloadType)
}

func TestGetWorkloadTypeDefaultClusterPriority(t *testing.T) {
	podGroup := createPodGroup()
	priorityClasses := []client.Object{
		&k8sschedulingv1.PriorityClass{
			ObjectMeta: v1.ObjectMeta{
				Name: "train",
			},
			Value: priorityTrainNumber,
		},
		&k8sschedulingv1.PriorityClass{
			ObjectMeta: v1.ObjectMeta{
				Name: "inference",
			},
			Value:         priorityInferenceNumber,
			GlobalDefault: true,
		},
	}
	fakeClient := createFakeClient(priorityClasses)
	exporter := NewPodGroupExporter(context.TODO(), fakeClient, podGroup)

	workloadType := exporter.GetWorkloadType()
	assert.Equal(t, jobTypeInference, workloadType)
}

func TestGetWorkloadTypeFromPodGroupPriority(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Spec.PriorityClassName = "interactive-preemptible"
	priorityClasses := []client.Object{
		&k8sschedulingv1.PriorityClass{
			ObjectMeta: v1.ObjectMeta{
				Name: "interactive-preemptible",
			},
			Value: priorityInteractivePreemptibleNumber,
		},
	}
	fakeClient := createFakeClient(priorityClasses)
	exporter := NewPodGroupExporter(context.TODO(), fakeClient, podGroup)

	workloadType := exporter.GetWorkloadType()
	assert.Equal(t, jobTypeInteractivePreemptible, workloadType)
}

func TestGetWorkloadTypeInteractive(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Spec.PriorityClassName = "build"
	priorityClasses := []client.Object{
		&k8sschedulingv1.PriorityClass{
			ObjectMeta: v1.ObjectMeta{
				Name: "build",
			},
			Value: priorityBuildNumber,
		},
	}
	fakeClient := createFakeClient(priorityClasses)
	exporter := NewPodGroupExporter(context.TODO(), fakeClient, podGroup)

	workloadType := exporter.GetWorkloadType()
	assert.Equal(t, jobTypeInteractive, workloadType)
}

func TestGetWorkloadTypeInference(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.OwnerReferences = []v1.OwnerReference{
		{
			Name: "some-deployment",
			Kind: "Deployment",
		},
	}
	podGroup.Spec.PriorityClassName = "build"
	priorityClasses := []client.Object{
		&k8sschedulingv1.PriorityClass{
			ObjectMeta: v1.ObjectMeta{
				Name: "build",
			},
			Value: priorityBuildNumber,
		},
	}
	fakeClient := createFakeClient(priorityClasses)
	exporter := NewPodGroupExporter(context.TODO(), fakeClient, podGroup)

	workloadType := exporter.GetWorkloadType()
	assert.Equal(t, jobTypeInference, workloadType)
}

func TestGetWorkloadUnexpectedType(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Spec.PriorityClassName = "some-priorityclass-type"
	priorityClasses := []client.Object{
		&k8sschedulingv1.PriorityClass{
			ObjectMeta: v1.ObjectMeta{
				Name: "some-priorityclass-type",
			},
			Value: 123,
		},
	}
	fakeClient := createFakeClient(priorityClasses)
	exporter := NewPodGroupExporter(context.TODO(), fakeClient, podGroup)

	workloadType := exporter.GetWorkloadType()
	assert.Equal(t, jobTypeTrain, workloadType)
}

func TestGetWorkloadTopOwner(t *testing.T) {
	podGroup := createPodGroup()
	topOwner := TopOwnerMetadata{
		Name:    "gpu-deployment",
		UID:     "f260cda6-163a-4c05-8c42-220df64c7fd4",
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	}
	topOwnerJson, err := json.Marshal(topOwner)
	assert.Nil(t, err)
	podGroup.Annotations[topOwnerAnnotation] = string(topOwnerJson)
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	actualTopOwner, err := exporter.GetTopOwnerMetadata()
	assert.Nil(t, err)
	assert.Equal(t, topOwner, *actualTopOwner)
}

func TestGetWorkloadTopOwnerBC(t *testing.T) {
	podGroup := createPodGroup()
	topOwner := TopOwnerMetadata{
		Name:    "gpu-deployment-bc",
		UID:     "f260cda6-163a-4c05-8c42-220df64c7fd4",
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	}
	topOwnerJson, err := json.Marshal(topOwner)
	assert.Nil(t, err)
	podGroup.Annotations[topOwnerAnnotationBC] = string(topOwnerJson)
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	actualTopOwner, err := exporter.GetTopOwnerMetadata()
	assert.Nil(t, err)
	assert.Equal(t, topOwner, *actualTopOwner)
}

func TestGetWorkloadTopOwnerNotFound(t *testing.T) {
	podGroup := createPodGroup()
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	_, err := exporter.GetTopOwnerMetadata()
	assert.NotNil(t, err)
}

func TestGetWorkloadTopOwnerNotValid(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Annotations[topOwnerAnnotation] = "malformed-top-owner"
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	_, err := exporter.GetTopOwnerMetadata()
	assert.NotNil(t, err)
}

func TestGetRequestedMigDevices(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Annotations[requestedMigDevicesAnnotation] = "{\"1g.5gb\":1, \"2g.10gb\":2}"
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	migDevices, found, err := exporter.GetRequestedMigDevices()
	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, 2, len(migDevices))
	assert.Equal(t, map[string]int64{"1g.5gb": 1, "2g.10gb": 2}, migDevices)
}

func TestGetRequestedMigDevicesEmptyRequest(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Annotations[requestedMigDevicesAnnotation] = "{}"
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	migDevices, found, err := exporter.GetRequestedMigDevices()
	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, 0, len(migDevices))
	assert.Equal(t, map[string]int64{}, migDevices)
}

func TestGetRequestedMigDevicesInvalid(t *testing.T) {
	podGroup := createPodGroup()
	podGroup.Annotations[requestedMigDevicesAnnotation] = ""
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	migDevices, found, err := exporter.GetRequestedMigDevices()
	assert.Nil(t, migDevices)
	assert.True(t, found)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "could not unmarshal value")
}

func TestGetRequestedMigDevicesAnnotationNotFound(t *testing.T) {
	podGroup := createPodGroup()
	exporter := NewPodGroupExporter(context.TODO(), nil, podGroup)

	migDevices, found, err := exporter.GetRequestedMigDevices()
	assert.Nil(t, migDevices)
	assert.False(t, found)
	assert.Nil(t, err)
}

func createPodGroup() *kaiv2alpha2.PodGroup {
	podGroup := kaiv2alpha2.PodGroup{
		ObjectMeta: v1.ObjectMeta{
			Name:        "pg-name",
			Namespace:   "some-namespace",
			Annotations: map[string]string{},
			Labels: map[string]string{
				runaiv1alpha1.NodePoolNameLabel: nodepoolName,
			},
		},
		Spec: kaiv2alpha2.PodGroupSpec{
			Queue: "some-queue",
		},
	}

	return &podGroup
}

func createQueue() *kaiv2.Queue {
	return createQueueInner("some-queue", testDepartmentQueueName, "some-project", "", "")
}

func createQueueInner(queueName, parentQueueName, projectName, departmentName, departmentId string) *kaiv2.Queue {
	queue := kaiv2.Queue{
		ObjectMeta: v1.ObjectMeta{
			Name:   queueName,
			Labels: map[string]string{},
		},
		Spec: kaiv2.QueueSpec{
			ParentQueue: parentQueueName,
		},
	}
	if projectName != "" {
		queue.Labels[common.ProjectLabel] = projectName
	}
	if departmentName != "" {
		queue.Labels[v2alpha2.DepartmentNameLabel] = departmentName
	}
	if departmentId != "" {
		queue.Labels[v2alpha2.DepartmentIdLabel] = departmentId
	}
	return &queue
}

func createFakeClient(objects []client.Object) client.Client {
	scheme := runtime.NewScheme()
	utilruntime.Must(v2alpha2.AddToScheme(scheme))
	utilruntime.Must(kaiv2alpha2.AddToScheme(scheme))
	utilruntime.Must(kaiv2.AddToScheme(scheme))
	utilruntime.Must(k8sschedulingv1.AddToScheme(scheme))
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

type possiblyErroringFakeCtrlRuntimeClient struct {
	client.Client
	shouldError bool
}

func (p *possiblyErroringFakeCtrlRuntimeClient) List(ctx context.Context, list client.ObjectList,
	opts ...client.ListOption) error {
	if p.shouldError {
		return fmt.Errorf("some error to list")
	}
	return p.Client.List(ctx, list, opts...)
}
