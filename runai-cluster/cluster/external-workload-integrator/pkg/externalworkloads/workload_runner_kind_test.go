/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkloads

import (
	"context"
	"fmt"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/networkstatus"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/structure"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/supportedhandlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	schedulingv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var snapshotJobGVK = schema.GroupVersionKind{Group: "nvidia.com", Version: "v1alpha1", Kind: "SnapshotJob"}

const (
	captureName      = "my-wlr-snapshot-c109be00792d"
	captureUID       = types.UID("capture-uid")
	captureNamespace = "runai-proj"
)

// pytorchRunner runs the PyTorchJob that newChildObject builds.
func pytorchRunner(namespace string) *dromosv1alpha1.WorkloadRunner {
	return &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testWlrName,
			Namespace: namespace,
			UID:       testWlrUID,
		},
		Spec: dromosv1alpha1.WorkloadRunnerSpec{Template: runtime.RawExtension{Raw: []byte(fmt.Sprintf(
			`{"apiVersion":%q,"kind":%q,"metadata":{"name":%q,"namespace":%q}}`,
			pytorchGVK.GroupVersion().String(), pytorchGVK.Kind, testJobName, namespace))}},
	}
}

// capture is a SnapshotJob owned by the serving workload's runner, as snapshot-plugin creates it.
func capture(namespace string) *metav1.PartialObjectMetadata {
	object := newObjectWithOwnerRefs(wlrOwnerRefWithUID(testWlrName, testWlrUID))
	object.Name = captureName
	object.Namespace = namespace
	object.UID = captureUID
	object.GetObjectKind().SetGroupVersionKind(snapshotJobGVK)
	return object
}

// stubHandler stands in for a kind's registered handler; parsing isn't under test.
type stubHandler struct{}

func (stubHandler) Parse(context.Context, client.Object, *v1alpha1.ExternalWorkload) error {
	return nil
}

func (stubHandler) GetObject(context.Context, string, string) (client.Object, error) {
	return nil, apierrors.NewNotFound(schema.GroupResource{}, "")
}

func kindHandler(objects ...client.Object) (*ExternalWorkloadHandler, client.Client) {
	scheme := runtime.NewScheme()
	Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
	Expect(dromosv1alpha1.AddToScheme(scheme)).To(Succeed())
	Expect(corev1.AddToScheme(scheme)).To(Succeed())
	Expect(kaiorg.AddToScheme(scheme)).To(Succeed())
	c := fakeclient.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objects...).
		WithStatusSubresource(&v1alpha1.ExternalWorkload{}).
		WithIndex(&v1alpha1.ExternalWorkload{}, constants.ExternalWorkloadByOriginalUIDIndexerName,
			cachedclient.ExternalWorkloadByOriginalWorkloadUIDIndexer).
		Build()
	return &ExternalWorkloadHandler{
		cachedClient: c,
		supportedGVKsHandlers: map[schema.GroupVersionKind]supportedhandlers.TopOwnerHandler{
			snapshotJobGVK: stubHandler{},
			pytorchGVK:     stubHandler{},
		},
		onlyOwnedSupportedGVKs: map[schema.GroupVersionKind]bool{},
		networkStatusUpdater:   networkstatus.NewNetworkStatusUpdater(c),
		structureUpdater:       structure.NewStructureUpdater(c),
	}, c
}

// projectObjects are what creating an external workload reads its project from.
func projectObjects(namespace string) []client.Object {
	return []client.Object{
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
			Name:   namespace,
			Labels: map[string]string{constants.RunaiQueueLabel: "proj-1"},
		}},
		&kaiorg.Project{
			ObjectMeta: metav1.ObjectMeta{
				Name: "proj-1",
				Labels: map[string]string{
					v2.ProjectIdLabel:                    "1",
					schedulingv2alpha2.DepartmentIdLabel: "2",
				},
			},
			Spec: kaiorg.ProjectSpec{Parent: "dept-1"},
		},
	}
}

var _ = Describe("ExternalWorkloadHandler.workloadRunnerOf", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("returns the runner for its own workload", func() {
		handler, _ := kindHandler(pytorchRunner(testNamespace))
		child := newChildObject(newChildUID, wlrOwnerRefWithUID(testWlrName, testWlrUID))

		Expect(handler.workloadRunnerOf(ctx, child)).To(HaveField("UID", testWlrUID))
	})

	It("returns nothing for another kind the runner owns", func() {
		handler, _ := kindHandler(pytorchRunner(testNamespace))

		Expect(handler.workloadRunnerOf(ctx, capture(testNamespace))).To(BeNil())
	})

	It("returns nothing for the runner's kind in another group", func() {
		handler, _ := kindHandler(pytorchRunner(testNamespace))
		lookalike := newChildObject(newChildUID, wlrOwnerRefWithUID(testWlrName, testWlrUID))
		lookalike.GetObjectKind().SetGroupVersionKind(
			schema.GroupVersionKind{Group: "example.com", Version: "v1", Kind: pytorchGVK.Kind})

		Expect(handler.workloadRunnerOf(ctx, lookalike)).To(BeNil())
	})

	It("returns nothing for another kind with the workload's name", func() {
		handler, _ := kindHandler(pytorchRunner(testNamespace))
		other := newChildObject(newChildUID, wlrOwnerRefWithUID(testWlrName, testWlrUID))
		other.GetObjectKind().SetGroupVersionKind(
			schema.GroupVersionKind{Group: pytorchGVK.Group, Version: pytorchGVK.Version, Kind: "TFJob"})

		Expect(handler.workloadRunnerOf(ctx, other)).To(BeNil())
	})

	It("returns nothing for another object of the workload's kind", func() {
		handler, _ := kindHandler(pytorchRunner(testNamespace))
		sibling := newChildObject(newChildUID, wlrOwnerRefWithUID(testWlrName, testWlrUID))
		sibling.Name = testJobName + "-aux"

		Expect(handler.workloadRunnerOf(ctx, sibling)).To(BeNil())
	})

	It("keeps the runner when its template can't be read", func() {
		runner := pytorchRunner(testNamespace)
		runner.Spec.Template = runtime.RawExtension{}
		handler, _ := kindHandler(runner)

		Expect(handler.workloadRunnerOf(ctx, capture(testNamespace))).NotTo(BeNil())
	})

	It("keeps the runner when it cannot be read", func() {
		handler, _ := kindHandler()

		Expect(handler.workloadRunnerOf(ctx, capture(testNamespace))).NotTo(BeNil())
	})

	It("returns nothing without a runner owner", func() {
		handler, _ := kindHandler(pytorchRunner(testNamespace))
		object := newChildObject(newChildUID,
			metav1.OwnerReference{APIVersion: "batch/v1", Kind: "CronJob", Name: "my-cronjob"})

		Expect(handler.workloadRunnerOf(ctx, object)).To(BeNil())
	})
})

// A capture is owned by the serving workload's runner, and used to be taken for
// that runner's recreated workload, re-pointing its external workload.
var _ = Describe("a capture owned by the serving workload's runner", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("does not take over the runner's external workload", func() {
		handler, c := kindHandler(pytorchRunner(testNamespace), survivingExternalWorkload(testWlrUID, testWorkload))

		_, err := handler.getOrRebindExternalWorkloadOfObject(ctx, capture(testNamespace), captureUID)

		Expect(apierrors.IsNotFound(err)).To(BeTrue(), "the capture should get its own external workload")
		Expect(storedExternalWorkload(ctx, c).Spec.OriginalWorkloadUID).To(Equal(oldChildUID))
	})

	It("gets an external workload of its own, owned by the capture and without runner labels", func() {
		handler, c := kindHandler(append(projectObjects(captureNamespace), pytorchRunner(captureNamespace))...)

		Expect(handler.handleTopOwnedObject(ctx, capture(captureNamespace), true)).To(Succeed())

		created := onlyExternalWorkloadIn(ctx, c)
		Expect(created.Spec.OriginalWorkloadUID).To(Equal(captureUID))
		Expect(created.OwnerReferences).To(ConsistOf(HaveField("UID", captureUID)))
		Expect(created.Labels).NotTo(HaveKey(commonconstants.WorkloadRunnerUidLabel))
	})

	It("keeps its own owner and gains no runner labels on update", func() {
		existing := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      utils.GenerateExternalWorkloadName(snapshotJobGVK, captureName),
				Namespace: testNamespace,
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: snapshotJobGVK.GroupVersion().String(),
					Kind:       snapshotJobGVK.Kind,
					Name:       captureName,
					UID:        captureUID,
				}},
			},
			Spec: v1alpha1.ExternalWorkloadSpec{
				OriginalWorkloadUID:  captureUID,
				OriginalWorkloadName: captureName,
				Group:                snapshotJobGVK.Group,
				Version:              snapshotJobGVK.Version,
			},
		}
		handler, c := kindHandler(pytorchRunner(testNamespace), existing)

		Expect(handler.handleTopOwnedObject(ctx, capture(testNamespace), false)).To(Succeed())

		updated := &v1alpha1.ExternalWorkload{}
		Expect(c.Get(ctx, client.ObjectKeyFromObject(existing), updated)).To(Succeed())
		Expect(updated.OwnerReferences).To(ConsistOf(HaveField("UID", captureUID)))
		Expect(updated.Labels).NotTo(HaveKey(commonconstants.WorkloadRunnerUidLabel))
	})
})

// #873's restart flow, with a runner whose template is read and compared.
var _ = Describe("a runner's own workload", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("is created under the runner", func() {
		handler, c := kindHandler(append(projectObjects(captureNamespace), pytorchRunner(captureNamespace))...)
		child := newChildObject(newChildUID, wlrOwnerRefWithUID(testWlrName, testWlrUID))
		child.Namespace = captureNamespace

		Expect(handler.handleTopOwnedObject(ctx, child, true)).To(Succeed())

		created := onlyExternalWorkloadIn(ctx, c)
		Expect(created.OwnerReferences).To(ConsistOf(HaveField("UID", testWlrUID)))
		Expect(created.Labels).To(HaveKeyWithValue(commonconstants.WorkloadRunnerUidLabel, string(testWlrUID)))
	})

	It("moves an existing record under the runner on update", func() {
		existing := survivingExternalWorkload("", testWorkload)
		existing.Spec.OriginalWorkloadUID = newChildUID
		existing.OwnerReferences = []metav1.OwnerReference{{
			APIVersion: pytorchGVK.GroupVersion().String(), Kind: pytorchGVK.Kind, Name: testJobName, UID: newChildUID,
		}}
		handler, c := kindHandler(pytorchRunner(testNamespace), existing)
		child := newChildObject(newChildUID, wlrOwnerRefWithUID(testWlrName, testWlrUID))

		Expect(handler.handleTopOwnedObject(ctx, child, false)).To(Succeed())

		updated := storedExternalWorkload(ctx, c)
		Expect(updated.OwnerReferences).To(ConsistOf(HaveField("UID", testWlrUID)))
		Expect(updated.Labels).To(HaveKeyWithValue(commonconstants.WorkloadRunnerUidLabel, string(testWlrUID)))
	})

	It("is re-bound after a restart", func() {
		handler, c := kindHandler(pytorchRunner(testNamespace), survivingExternalWorkload(testWlrUID, testWorkload))
		recreated := newChildObject(newChildUID, wlrOwnerRefWithUID(testWlrName, testWlrUID))

		rebound, err := handler.getOrRebindExternalWorkloadOfObject(ctx, recreated, newChildUID)

		Expect(err).NotTo(HaveOccurred())
		Expect(rebound).NotTo(BeNil())
		Expect(storedExternalWorkload(ctx, c).Spec.OriginalWorkloadUID).To(Equal(newChildUID))
	})
})
