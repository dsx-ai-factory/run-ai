/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package k8sobject

import (
	"context"
	"errors"
	"net/http"
	"time"

	workloadsapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/auth"
	openapi "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/generated/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	fakeBearer       = "fake-bearer-token" // #nosec G101 -- test fixture, not a real credential
	testPodA         = "pod-a"
	keepMeAnnotation = "keep-me"
	keepMeValue      = "yes"
	kubectlManager   = "kubectl"
	junkAnnotation   = "{junk}"
)

var _ = Describe("K8sObject service", func() {
	var (
		ctx          context.Context
		workloadInfo *AuthorizedWorkloadInfo
	)

	BeforeEach(func() {
		workloadInfo = &AuthorizedWorkloadInfo{
			Namespace: "runai-proj-1",
			Name:      "my-workload",
			Kind:      "Training",
		}
		ctx = WithWorkloadInfo(context.Background(), workloadInfo)
		ctx = auth.ComposeAuthContext(ctx, &auth.Info{RawToken: fakeBearer})
	})

	Context("GetWorkloadObject", func() {
		It("returns the native CR for a Training workload", func() {
			scheme := runtime.NewScheme()
			Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
			seed := newUnstructured("run.ai/v2alpha1", "TrainingWorkload", workloadInfo.Namespace, workloadInfo.Name, map[string]string{
				lastAppliedAnnotation: junkAnnotation,
				keepMeAnnotation:      keepMeValue,
			})
			// The newer controller-runtime fake client validates managedFields on Build and
			// rejects entries without a valid operation, so set one. The test still asserts
			// the managedFields are stripped from the returned object below.
			seed.SetManagedFields([]metav1.ManagedFieldsEntry{{
				Manager:    kubectlManager,
				Operation:  metav1.ManagedFieldsOperationUpdate,
				APIVersion: "run.ai/v2alpha1",
				FieldsType: "FieldsV1",
				FieldsV1:   &metav1.FieldsV1{Raw: []byte("{}")},
			}})

			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(seed).Build()
			svc := NewService(c, nil, nil)

			before := time.Now().UTC().Add(-time.Second)
			resp, err := svc.GetWorkloadObject(ctx, "ignored")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			obj := resp.Body.(openapi.Object)
			Expect(obj.RetrievedAt).To(BeTemporally(">=", before))
			Expect(obj.RetrievedAt).To(BeTemporally("<=", time.Now().UTC().Add(time.Second)))

			Expect(obj.Object["metadata"].(map[string]interface{})["managedFields"]).To(BeNil())
			annotations, _ := obj.Object["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
			Expect(annotations).ToNot(HaveKey(lastAppliedAnnotation))
			Expect(annotations).To(HaveKeyWithValue(keepMeAnnotation, keepMeValue))
		})

		It("returns the External (Deployment) CR via kubernetesRef", func() {
			workloadInfo.Kind = "Deployment"
			workloadInfo.Group = "apps"
			workloadInfo.Version = "v1"
			ctx = WithWorkloadInfo(context.Background(), workloadInfo)
			ctx = auth.ComposeAuthContext(ctx, &auth.Info{RawToken: fakeBearer})

			scheme := runtime.NewScheme()
			Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
			seed := newUnstructured("apps/v1", "Deployment", workloadInfo.Namespace, workloadInfo.Name, nil)
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(seed).Build()
			svc := NewService(c, nil, nil)

			resp, err := svc.GetWorkloadObject(ctx, "ignored")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
		})

		DescribeTable("resolves the real k8s GVK for transitioned (WorkloadsV2) workloads",
			func(kind, framework, apiVersion, k8sKind string) {
				workloadInfo.Kind = kind
				workloadInfo.SourceApi = string(workloadsapi.SOURCEAPI_WORKLOADS_V2)
				workloadInfo.DistributedFramework = framework
				ctx = WithWorkloadInfo(context.Background(), workloadInfo)
				ctx = auth.ComposeAuthContext(ctx, &auth.Info{RawToken: fakeBearer})

				scheme := runtime.NewScheme()
				Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
				seed := newUnstructured(apiVersion, k8sKind, workloadInfo.Namespace, workloadInfo.Name, nil)
				c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(seed).Build()
				svc := NewService(c, nil, nil)

				resp, err := svc.GetWorkloadObject(ctx, "ignored")
				Expect(err).To(BeNil())
				Expect(resp.Code).To(Equal(http.StatusOK))
				obj := resp.Body.(openapi.Object)
				Expect(obj.Object["apiVersion"]).To(Equal(apiVersion))
				Expect(obj.Object["kind"]).To(Equal(k8sKind))
			},
			Entry("Workspace -> batch/v1 Job", "Workspace", "", "batch/v1", "Job"),
			Entry("Training -> batch/v1 Job", "Training", "", "batch/v1", "Job"),
			Entry("Distributed/PyTorch -> kubeflow.org/v1 PyTorchJob", "Distributed", "PyTorch", "kubeflow.org/v1", "PyTorchJob"),
			Entry("Distributed/MPI -> kubeflow.org/v2beta1 MPIJob", "Distributed", "MPI", "kubeflow.org/v2beta1", "MPIJob"),
			Entry("Distributed/TF -> kubeflow.org/v1 TFJob", "Distributed", "TF", "kubeflow.org/v1", "TFJob"),
			Entry("Distributed/XGBoost -> kubeflow.org/v1 XGBoostJob", "Distributed", "XGBoost", "kubeflow.org/v1", "XGBoostJob"),
			Entry("Distributed/JAX -> kubeflow.org/v1 JAXJob", "Distributed", "JAX", "kubeflow.org/v1", "JAXJob"),
			Entry("Inference -> serving.knative.dev/v1 Service", "Inference", "", "serving.knative.dev/v1", "Service"),
		)

		It("keeps the native run.ai GVK for a transitioned DistributedInference workload (not yet migrated)", func() {
			workloadInfo.Kind = "DistributedInference"
			workloadInfo.SourceApi = string(workloadsapi.SOURCEAPI_WORKLOADS_V2)
			ctx = WithWorkloadInfo(context.Background(), workloadInfo)
			ctx = auth.ComposeAuthContext(ctx, &auth.Info{RawToken: fakeBearer})

			scheme := runtime.NewScheme()
			Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
			seed := newUnstructured("run.ai/v1alpha1", "DistributedInferenceWorkload", workloadInfo.Namespace, workloadInfo.Name, nil)
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(seed).Build()
			svc := NewService(c, nil, nil)

			resp, err := svc.GetWorkloadObject(ctx, "ignored")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			obj := resp.Body.(openapi.Object)
			Expect(obj.Object["kind"]).To(Equal("DistributedInferenceWorkload"))
		})

		It("falls back to the native DistributedWorkload when framework is empty", func() {
			workloadInfo.Kind = "Distributed"
			workloadInfo.SourceApi = string(workloadsapi.SOURCEAPI_WORKLOADS_V2)
			workloadInfo.DistributedFramework = ""
			ctx = WithWorkloadInfo(context.Background(), workloadInfo)
			ctx = auth.ComposeAuthContext(ctx, &auth.Info{RawToken: fakeBearer})

			scheme := runtime.NewScheme()
			Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
			seed := newUnstructured("run.ai/v2alpha1", "DistributedWorkload", workloadInfo.Namespace, workloadInfo.Name, nil)
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(seed).Build()
			svc := NewService(c, nil, nil)

			resp, err := svc.GetWorkloadObject(ctx, "ignored")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			obj := resp.Body.(openapi.Object)
			Expect(obj.Object["kind"]).To(Equal("DistributedWorkload"))
		})

		It("returns 404 when the k8s object is missing", func() {
			scheme := runtime.NewScheme()
			Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
			c := fake.NewClientBuilder().WithScheme(scheme).Build()
			svc := NewService(c, nil, nil)

			resp, err := svc.GetWorkloadObject(ctx, "ignored")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusNotFound))
		})

		It("returns 404 when the workload namespace is empty", func() {
			workloadInfo.Namespace = ""
			ctx = WithWorkloadInfo(context.Background(), workloadInfo)
			ctx = auth.ComposeAuthContext(ctx, &auth.Info{RawToken: fakeBearer})

			scheme := runtime.NewScheme()
			Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
			c := fake.NewClientBuilder().WithScheme(scheme).Build()
			svc := NewService(c, nil, nil)

			resp, err := svc.GetWorkloadObject(ctx, "ignored")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusNotFound))
		})

		It("returns 404 when the workload name is empty", func() {
			workloadInfo.Name = ""
			ctx = WithWorkloadInfo(context.Background(), workloadInfo)
			ctx = auth.ComposeAuthContext(ctx, &auth.Info{RawToken: fakeBearer})

			scheme := runtime.NewScheme()
			Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
			c := fake.NewClientBuilder().WithScheme(scheme).Build()
			svc := NewService(c, nil, nil)

			resp, err := svc.GetWorkloadObject(ctx, "ignored")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusNotFound))
		})

		It("500s when External workload has no kubernetesRef", func() {
			workloadInfo.Kind = "Deployment" // not native; Group/Version empty
			ctx = WithWorkloadInfo(context.Background(), workloadInfo)
			ctx = auth.ComposeAuthContext(ctx, &auth.Info{RawToken: fakeBearer})

			scheme := runtime.NewScheme()
			Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
			c := fake.NewClientBuilder().WithScheme(scheme).Build()
			svc := NewService(c, nil, nil)

			resp, err := svc.GetWorkloadObject(ctx, "ignored")
			Expect(err).ToNot(BeNil())
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
		})
	})

	Context("GetPodObject", func() {
		It("returns the pod when upstream resolves it", func() {
			pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: workloadInfo.Namespace, Name: testPodA}}
			c := fake.NewClientBuilder().WithScheme(clientgoscheme.Scheme).WithObjects(pod).Build()
			fakePods := &fakePodGetter{pod: &workloadsapi.Pod{Id: "p1", Name: testPodA}}
			svc := NewService(c, nil, fakePods)

			resp, err := svc.GetPodObject(ctx, "wl-1", "p1")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(fakePods.calls).To(Equal(1), "upstream should be called exactly once (no pagination)")
		})

		It("returns 404 when upstream returns 404", func() {
			c := fake.NewClientBuilder().WithScheme(clientgoscheme.Scheme).Build()
			fakePods := &fakePodGetter{httpStatus: http.StatusNotFound, err: errors.New("not found")}
			svc := NewService(c, nil, fakePods)

			resp, err := svc.GetPodObject(ctx, "wl-1", "missing")
			Expect(err).ToNot(BeNil())
			Expect(resp.Code).To(Equal(http.StatusNotFound))
		})

		It("passes through 4xx upstream errors as-is (e.g. 400)", func() {
			c := fake.NewClientBuilder().WithScheme(clientgoscheme.Scheme).Build()
			fakePods := &fakePodGetter{httpStatus: http.StatusBadRequest, err: errors.New("bad request")}
			svc := NewService(c, nil, fakePods)

			resp, err := svc.GetPodObject(ctx, "wl-1", "p1")
			Expect(err).ToNot(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
		})

		It("passes through 5xx upstream errors as-is", func() {
			c := fake.NewClientBuilder().WithScheme(clientgoscheme.Scheme).Build()
			fakePods := &fakePodGetter{httpStatus: http.StatusInternalServerError, err: errors.New("upstream boom")}
			svc := NewService(c, nil, fakePods)

			resp, err := svc.GetPodObject(ctx, "wl-1", "p1")
			Expect(err).ToNot(BeNil())
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
		})

		It("returns 500 when upstream returns no HTTP response (transport failure)", func() {
			c := fake.NewClientBuilder().WithScheme(clientgoscheme.Scheme).Build()
			fakePods := &fakePodGetter{err: errors.New("dial tcp: connection refused")}
			svc := NewService(c, nil, fakePods)

			resp, err := svc.GetPodObject(ctx, "wl-1", "p1")
			Expect(err).ToNot(BeNil())
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
		})
	})

	Context("GetElementObject", func() {
		It("returns the element's k8s object on happy path (first-class kind)", func() {
			scheme := runtime.NewScheme()
			Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
			seed := newUnstructured("apps/v1", "Deployment", workloadInfo.Namespace, "elem-dep", nil)
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(seed).Build()

			fakeEl := &fakeElementGetter{element: makeElementResponse("elem-dep", "apps", "v1", "Deployment")}
			svc := NewService(c, fakeEl, nil)

			resp, err := svc.GetElementObject(ctx, "wl-1", "elem-1")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			obj := resp.Body.(openapi.Object)
			Expect(obj.Object["kind"]).To(Equal("Deployment"))
		})

		It("returns the element's k8s object for a non-first-class kind (RayJob)", func() {
			scheme := runtime.NewScheme()
			Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
			seed := newUnstructured("ray.io/v1", "RayJob", workloadInfo.Namespace, "elem-ray", nil)
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(seed).Build()

			fakeEl := &fakeElementGetter{element: makeElementResponse("elem-ray", "ray.io", "v1", "RayJob")}
			svc := NewService(c, fakeEl, nil)

			resp, err := svc.GetElementObject(ctx, "wl-1", "elem-1")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			obj := resp.Body.(openapi.Object)
			Expect(obj.Object["apiVersion"]).To(Equal("ray.io/v1"))
			Expect(obj.Object["kind"]).To(Equal("RayJob"))
		})

		It("returns 404 when the element has no kubernetesRef", func() {
			c := fake.NewClientBuilder().WithScheme(clientgoscheme.Scheme).Build()
			elem := workloadsapi.NewWorkloadElementResponse("elem-1", "logical-only", 1,
				*workloadsapi.NewNullableWorkloadRequestResources(nil),
				*workloadsapi.NewNullableWorkloadAllocatedResources(nil))
			fakeEl := &fakeElementGetter{element: elem}
			svc := NewService(c, fakeEl, nil)

			resp, err := svc.GetElementObject(ctx, "wl-1", "elem-1")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusNotFound))
		})

		It("returns 404 when kubernetesRef has empty group (elementGvkName invariant)", func() {
			c := fake.NewClientBuilder().WithScheme(clientgoscheme.Scheme).Build()
			fakeEl := &fakeElementGetter{element: makeElementResponse("elem-bad", "", "v1", "Deployment")}
			svc := NewService(c, fakeEl, nil)

			resp, err := svc.GetElementObject(ctx, "wl-1", "elem-1")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusNotFound))
		})

	})

	Context("sanitize", func() {
		It("strips both managedFields and last-applied-configuration", func() {
			u := newUnstructured("v1", "Pod", "ns", "p", map[string]string{
				lastAppliedAnnotation: junkAnnotation,
				keepMeAnnotation:      keepMeValue,
			})
			u.SetManagedFields([]metav1.ManagedFieldsEntry{{Manager: kubectlManager}})

			sanitize(u)

			Expect(u.GetManagedFields()).To(BeEmpty())
			Expect(u.GetAnnotations()).ToNot(HaveKey(lastAppliedAnnotation))
			Expect(u.GetAnnotations()).To(HaveKeyWithValue(keepMeAnnotation, keepMeValue))
		})

		It("is a no-op when there's nothing to strip", func() {
			u := newUnstructured("v1", "Pod", "ns", "p", map[string]string{keepMeAnnotation: keepMeValue})

			sanitize(u)

			Expect(u.GetManagedFields()).To(BeEmpty())
			Expect(u.GetAnnotations()).To(HaveKeyWithValue(keepMeAnnotation, keepMeValue))
		})

		It("strips managedFields when annotations are absent", func() {
			u := newUnstructured("v1", "Pod", "ns", "p", nil)
			u.SetManagedFields([]metav1.ManagedFieldsEntry{{Manager: kubectlManager}})

			sanitize(u)

			Expect(u.GetManagedFields()).To(BeEmpty())
			Expect(u.GetAnnotations()).To(BeEmpty())
		})

		It("strips only last-applied-configuration when managedFields is absent", func() {
			u := newUnstructured("v1", "Pod", "ns", "p", map[string]string{
				lastAppliedAnnotation: junkAnnotation,
				keepMeAnnotation:      keepMeValue,
			})

			sanitize(u)

			Expect(u.GetManagedFields()).To(BeEmpty())
			Expect(u.GetAnnotations()).ToNot(HaveKey(lastAppliedAnnotation))
			Expect(u.GetAnnotations()).To(HaveKeyWithValue(keepMeAnnotation, keepMeValue))
		})

		It("leaves annotation map intact when last-applied key is absent", func() {
			u := newUnstructured("v1", "Pod", "ns", "p", map[string]string{"only": "this"})

			sanitize(u)

			Expect(u.GetAnnotations()).To(HaveKeyWithValue("only", "this"))
			Expect(u.GetAnnotations()).To(HaveLen(1))
		})

		It("is a no-op on a nil unstructured pointer", func() {
			Expect(func() { sanitize(nil) }).ToNot(Panic())
		})
	})
})

func newUnstructured(apiVersion, kind, namespace, name string, annotations map[string]string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	gv, _ := schema.ParseGroupVersion(apiVersion)
	u.SetGroupVersionKind(gv.WithKind(kind))
	u.SetNamespace(namespace)
	u.SetName(name)
	if annotations != nil {
		u.SetAnnotations(annotations)
	}
	_ = appsv1.AddToScheme(clientgoscheme.Scheme)
	return u
}

// fakePodGetter satisfies PodGetter for tests.
//   - happy path: set `pod`
//   - upstream returned a status with error: set `err` + `httpStatus`
//   - transport failure (no HTTP response at all): set `err` only, leave
//     `httpStatus` zero
type fakePodGetter struct {
	pod        *workloadsapi.Pod
	err        error
	httpStatus int
	calls      int
}

func (f *fakePodGetter) GetWorkloadPod(_ context.Context, _, _ string) (*workloadsapi.Pod, *http.Response, error) {
	f.calls++
	if f.err != nil && f.httpStatus == 0 {
		return nil, nil, f.err
	}
	httpResp := &http.Response{StatusCode: f.httpStatus}
	if f.err != nil {
		return nil, httpResp, f.err
	}
	if httpResp.StatusCode == 0 {
		httpResp.StatusCode = http.StatusOK
	}
	return f.pod, httpResp, nil
}

type fakeElementGetter struct {
	element *workloadsapi.WorkloadElementResponse
}

func (f *fakeElementGetter) GetWorkloadElement(_ context.Context, _, _ string) (*workloadsapi.WorkloadElementResponse, *http.Response, error) {
	return f.element, &http.Response{StatusCode: http.StatusOK}, nil
}

func makeElementResponse(name, group, version, kind string) *workloadsapi.WorkloadElementResponse {
	gvk := workloadsapi.NewWorkloadElementKubernetesRefGvk()
	gvk.SetGroup(group)
	gvk.SetVersion(version)
	gvk.SetKind(kind)
	ref := workloadsapi.NewWorkloadElementKubernetesRef()
	ref.SetName(name)
	ref.SetGvk(*gvk)
	elem := workloadsapi.NewWorkloadElementResponse("elem-1", name, 1,
		*workloadsapi.NewNullableWorkloadRequestResources(nil),
		*workloadsapi.NewNullableWorkloadAllocatedResources(nil))
	elem.SetKubernetesRef(*ref)
	return elem
}
