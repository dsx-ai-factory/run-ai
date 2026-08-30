package test

import (
	"context"
	dwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"time"

	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	iwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/gomega" //nolint:staticcheck
)

const (
	testNoAutoDeletion = -1
)

type TestWorkload interface {
	Build()
	Get(ctx context.Context, k8sClient client.WithWatch) error
	Delete(ctx context.Context, k8sClient client.WithWatch) error
	GetObject() *client.Object
	SetAutoDeleteTime(seconds int)
	SetCompletionTimeAndStatus(ctx context.Context, k8sClient client.WithWatch, success bool)
	SetPhase(ctx context.Context, k8sClient client.WithWatch, phase common.Phase)
}

type TestTrainingWorkload struct {
	name     string
	workload *twv2a1.TrainingWorkload
}

func (tw *TestTrainingWorkload) Build() {
	tw.name = GenerateWorkloadName()
	tw.workload = BuildTrainingWorkload(testNamespaceName, tw.name, nil)
}

func (tw *TestTrainingWorkload) GetObject() *client.Object {
	var obj client.Object = tw.workload
	return &obj
}
func (tw *TestTrainingWorkload) Get(ctx context.Context, k8sClient client.WithWatch) error {
	objKey := client.ObjectKey{Namespace: testNamespaceName, Name: tw.name}
	twGet := twv2a1.TrainingWorkload{}
	err := k8sClient.Get(ctx, objKey, &twGet)
	return err
}

func (tw *TestTrainingWorkload) Delete(ctx context.Context, k8sClient client.WithWatch) error {
	err := k8sClient.Delete(ctx, tw.workload)
	return err
}

func (tw *TestTrainingWorkload) SetAutoDeleteTime(seconds int) {
	tw.workload.Spec.AutoDeletionTimeAfterCompletionSeconds = &flds.IntField{
		Value: &seconds,
	}
}

func (tw *TestTrainingWorkload) SetCompletionTimeAndStatus(ctx context.Context, k8sClient client.WithWatch, success bool) {
	tw.workload.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	if success {
		tw.workload.Status.Phase = string(common.PhaseCompleted)
	} else {
		tw.workload.Status.Phase = string(common.PhaseFailed)
		cond := common.WorkloadCondition{Type: common.Failed, Status: corev1.ConditionTrue}
		tw.workload.Status.Conditions = append(tw.workload.Status.Conditions, cond)
	}
	err := k8sClient.Update(ctx, tw.workload)
	Expect(err).To(BeNil())
}

func (tw *TestTrainingWorkload) SetPhase(ctx context.Context, k8sClient client.WithWatch, phase common.Phase) {
	tw.workload.Status.Phase = string(phase)
	err := k8sClient.Update(ctx, tw.workload)
	Expect(err).To(BeNil())
}

type TestDistributedWorkload struct {
	name     string
	workload *dwv2a1.DistributedWorkload
}

func (dw *TestDistributedWorkload) Build() {
	dw.name = GenerateWorkloadName()
	dw.workload = BuildDistributedWorkload(testNamespaceName, dw.name, nil)
}

func (dw *TestDistributedWorkload) GetObject() *client.Object {
	var obj client.Object = dw.workload
	return &obj
}
func (dw *TestDistributedWorkload) Get(ctx context.Context, k8sClient client.WithWatch) error {
	objKey := client.ObjectKey{Namespace: testNamespaceName, Name: dw.name}
	dwGet := dwv2a1.DistributedWorkload{}
	err := k8sClient.Get(ctx, objKey, &dwGet)
	return err
}

func (dw *TestDistributedWorkload) Delete(ctx context.Context, k8sClient client.WithWatch) error {
	err := k8sClient.Delete(ctx, dw.workload)
	return err
}

func (dw *TestDistributedWorkload) SetAutoDeleteTime(seconds int) {
	dw.workload.Spec.AutoDeletionTimeAfterCompletionSeconds = &flds.IntField{
		Value: &seconds,
	}
}

func (dw *TestDistributedWorkload) SetCompletionTimeAndStatus(ctx context.Context, k8sClient client.WithWatch, success bool) {
	dw.workload.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	if success {
		dw.workload.Status.Phase = string(common.PhaseCompleted)
	} else {
		dw.workload.Status.Phase = string(common.PhaseFailed)
		cond := common.WorkloadCondition{Type: common.Failed, Status: corev1.ConditionTrue}
		dw.workload.Status.Conditions = append(dw.workload.Status.Conditions, cond)
	}
	err := k8sClient.Update(ctx, dw.workload)
	Expect(err).To(BeNil())
}

func (dw *TestDistributedWorkload) SetPhase(ctx context.Context, k8sClient client.WithWatch, phase common.Phase) {
	dw.workload.Status.Phase = string(phase)
	err := k8sClient.Update(ctx, dw.workload)
	Expect(err).To(BeNil())
}

type TestInteractiveWorkload struct {
	name     string
	workload *iwv2a1.InteractiveWorkload
}

func (tw *TestInteractiveWorkload) Build() {
	tw.name = GenerateWorkloadName()
	tw.workload = BuildInteractiveWorkload(testNamespaceName, tw.name, nil)
}

func (tw *TestInteractiveWorkload) GetObject() *client.Object {
	var obj client.Object = tw.workload
	return &obj
}
func (tw *TestInteractiveWorkload) Get(ctx context.Context, k8sClient client.WithWatch) error {
	objKey := client.ObjectKey{Namespace: testNamespaceName, Name: tw.name}
	twGet := iwv2a1.InteractiveWorkload{}
	err := k8sClient.Get(ctx, objKey, &twGet)
	return err
}

func (tw *TestInteractiveWorkload) Delete(ctx context.Context, k8sClient client.WithWatch) error {
	err := k8sClient.Delete(ctx, tw.workload)
	return err
}

func (tw *TestInteractiveWorkload) SetAutoDeleteTime(seconds int) {
	tw.workload.Spec.AutoDeletionTimeAfterCompletionSeconds = &flds.IntField{
		Value: &seconds,
	}
}

func (tw *TestInteractiveWorkload) SetCompletionTimeAndStatus(ctx context.Context, k8sClient client.WithWatch, success bool) {
	tw.workload.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	if success {
		tw.workload.Status.Phase = string(common.PhaseCompleted)
	} else {
		tw.workload.Status.Phase = string(common.PhaseFailed)
		cond := common.WorkloadCondition{Type: common.Failed, Status: corev1.ConditionTrue}
		tw.workload.Status.Conditions = append(tw.workload.Status.Conditions, cond)
	}
	err := k8sClient.Update(ctx, tw.workload)
	Expect(err).To(BeNil())
}

func (tw *TestInteractiveWorkload) SetPhase(ctx context.Context, k8sClient client.WithWatch, phase common.Phase) {
	tw.workload.Status.Phase = string(phase)
	err := k8sClient.Update(ctx, tw.workload)
	Expect(err).To(BeNil())
}

func SimulateWorkloadLifecycle(ctx context.Context, k8sClient client.WithWatch, workload TestWorkload, durationSeconds int, autoDeleteSeconds int, success bool) {
	// build the base workload
	workload.Build()

	// set AutoDeletionTimeAfterCompletionSeconds, if needed
	if autoDeleteSeconds != testNoAutoDeletion {
		workload.SetAutoDeleteTime(autoDeleteSeconds)
	}

	// create the workload
	obj := workload.GetObject()
	err := k8sClient.Create(ctx, *obj)
	Expect(err).To(BeNil())

	// simulate running time by sleeping and updating CompletionTime status.
	// before updating, verify that the workload was not deleted (to make sure auto-delete time is respected in this matter)
	time.Sleep(time.Duration(durationSeconds+1) * time.Second)

	err = workload.Get(ctx, k8sClient)
	Expect(err).To(BeNil())

	workload.SetCompletionTimeAndStatus(ctx, k8sClient, success)
}
