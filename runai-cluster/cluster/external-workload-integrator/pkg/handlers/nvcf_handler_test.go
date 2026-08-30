package handlers

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	nvcfv1 "github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/nvcf/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/test"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("TestNvcfHandler", Ordered, func() {
	var (
		externalWorkload         *v1alpha1.ExternalWorkload
		expectedExternalWorkload *v1alpha1.ExternalWorkload
		cachedClient             client.WithWatch
		nvcf                     *nvcfv1.SPOTRequest
		nvcfTopOwnerHandler      *NVCFTopOwnerHandler

		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeAll(func() {
		scheme := scheme_init.Scheme()
		ns := test.BuildNamespace(test.TestNamespace, map[string]string{})
		cachedClient = fake.NewClientBuilder().WithStatusSubresource(&corev1.Pod{}).WithScheme(scheme).WithObjects(ns).Build()

		ctx, cancel = context.WithCancel(context.Background())
		DeferCleanup(func() {
			cancel()
		})

		test.InitVars()

		nvcfTopOwnerHandler = NewNVCFTopOwnerHandler(cachedClient, nil)
	})

	BeforeEach(func() {
		externalWorkload = &v1alpha1.ExternalWorkload{}
		nvcf = test.TestNvcfFunction.DeepCopy()
		expectedExternalWorkload = test.TestNvcfFunctionExpectedBasicEw.DeepCopy()

		// delete all pods from previous tests
		pods := &corev1.PodList{}
		err := cachedClient.List(ctx, pods)
		Expect(err).To(BeNil())
		for _, pod := range pods.Items {
			err = cachedClient.Delete(ctx, &pod)
			Expect(err).To(BeNil())
		}
	})

	It("Sanity", func() {
		nvcfPod := test.TestNvcfFunctionChildPod.DeepCopy()
		err := cachedClient.Create(ctx, nvcfPod)
		Expect(err).To(BeNil())

		err = nvcfTopOwnerHandler.Parse(ctx, nvcf, externalWorkload)
		Expect(err).To(BeNil())

		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)

		// validate the top-owner-id label on the pod
		err = cachedClient.Get(ctx, client.ObjectKeyFromObject(nvcfPod), nvcfPod)
		Expect(err).To(BeNil())
		Expect(nvcfPod.Labels[commonconstants.TopOwnerWorkloadIdLabel]).To(Equal(string(expectedExternalWorkload.Spec.OriginalWorkloadUID)))
		Expect(nvcfPod.Labels[commonconstants.TopOwnerUidLabel]).To(Equal(string(expectedExternalWorkload.Spec.OriginalWorkloadUID)))
		Expect(nvcfPod.Labels[commonconstants.ClusterTopOwnerUidLabel]).To(Equal(string(expectedExternalWorkload.GetUID())))
	})

	It("No pods - no ew is created", func() {
		// not creating any pods
		err := nvcfTopOwnerHandler.Parse(ctx, nvcf, externalWorkload)
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(ContainSubstring("no pods found for function version"))
	})

	It("Multiple pods", func() {
		nvcfPod1 := test.TestNvcfFunctionChildPod.DeepCopy()
		nvcfPod2 := test.TestNvcfFunctionChildPod.DeepCopy()
		nvcfPod2.Name = nvcfPod2.Name + "-2"
		nvcfPod2.UID = nvcfPod2.UID + "-2"
		err := cachedClient.Create(ctx, nvcfPod1)
		Expect(err).To(BeNil())
		err = cachedClient.Create(ctx, nvcfPod2)
		Expect(err).To(BeNil())

		expectedExternalWorkload.Spec.RequestedPods.Replicas = ptr.To(int32(2))
		expectedExternalWorkload.Spec.RequestedResources.Limits = corev1.ResourceList{
			constants.NvidiaGpuResourceName: *resource.NewQuantity(int64(10), resource.DecimalSI),
		}

		err = nvcfTopOwnerHandler.Parse(ctx, nvcf, externalWorkload)
		Expect(err).To(BeNil())

		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)
	})

	It("No environment variables in init container", func() {
		nvcfPod1 := test.TestNvcfFunctionChildPod.DeepCopy()
		nvcfPod1.Spec.InitContainers[0].Env = nil
		err := cachedClient.Create(ctx, nvcfPod1)
		Expect(err).To(BeNil())

		expectedExternalWorkload.Spec.OriginalWorkloadName = "function-version-" + nvcf.Spec.FunctionVersionID
		expectedExternalWorkload.Spec.ImageName = ""
		expectedExternalWorkload.Spec.Urls = nil
		expectedExternalWorkload.Spec.AdditionalFields = ptr.To("{\"function-id\":\"\",\"function-version-id\":\"\",\"function-version-name\":\"\"}")

		err = nvcfTopOwnerHandler.Parse(ctx, nvcf, externalWorkload)
		Expect(err).To(BeNil())

		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)
	})

	It("Status - single pod - ready -> running", func() {
		nvcfPod1 := test.TestNvcfFunctionChildPod.DeepCopy()
		nvcfPod1.Status.Phase = corev1.PodRunning
		nvcfPod1.Status.Conditions = []corev1.PodCondition{
			{
				Type:   corev1.PodReady,
				Status: corev1.ConditionTrue,
			},
		}
		err := cachedClient.Create(ctx, nvcfPod1)
		Expect(err).To(BeNil())

		expectedExternalWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWRunningPhase

		err = nvcfTopOwnerHandler.Parse(ctx, nvcf, externalWorkload)
		Expect(err).To(BeNil())

		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)
	})

	It("Status - single pod - running but not ready -> initializing", func() {
		nvcfPod1 := test.TestNvcfFunctionChildPod.DeepCopy()
		nvcfPod1.Status.Phase = corev1.PodRunning
		err := cachedClient.Create(ctx, nvcfPod1)
		Expect(err).To(BeNil())

		expectedExternalWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWInitializingPhase

		err = nvcfTopOwnerHandler.Parse(ctx, nvcf, externalWorkload)
		Expect(err).To(BeNil())

		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)
	})

	It("Status - single pod - pending -> pending", func() {
		nvcfPod1 := test.TestNvcfFunctionChildPod.DeepCopy()
		nvcfPod1.Status.Phase = corev1.PodPending
		err := cachedClient.Create(ctx, nvcfPod1)
		Expect(err).To(BeNil())

		expectedExternalWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWPendingPhase

		err = nvcfTopOwnerHandler.Parse(ctx, nvcf, externalWorkload)
		Expect(err).To(BeNil())

		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)
	})

	It("Status - single pod - failed -> failed", func() {
		nvcfPod1 := test.TestNvcfFunctionChildPod.DeepCopy()
		nvcfPod1.Status.Phase = corev1.PodFailed
		err := cachedClient.Create(ctx, nvcfPod1)
		Expect(err).To(BeNil())

		expectedExternalWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWFailedPhase

		err = nvcfTopOwnerHandler.Parse(ctx, nvcf, externalWorkload)
		Expect(err).To(BeNil())

		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)
	})

	It("Status - single pod - unknown -> unknown", func() {
		nvcfPod1 := test.TestNvcfFunctionChildPod.DeepCopy()
		nvcfPod1.Status.Phase = corev1.PodUnknown
		err := cachedClient.Create(ctx, nvcfPod1)
		Expect(err).To(BeNil())

		expectedExternalWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWUnknownPhase

		err = nvcfTopOwnerHandler.Parse(ctx, nvcf, externalWorkload)
		Expect(err).To(BeNil())

		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)
	})

	It("Status - two pods - ready and failed -> running", func() {
		nvcfPod1 := test.TestNvcfFunctionChildPod.DeepCopy()
		nvcfPod1.Status.Phase = corev1.PodRunning
		nvcfPod1.Status.Conditions = []corev1.PodCondition{
			{
				Type:   corev1.PodReady,
				Status: corev1.ConditionTrue,
			},
		}
		nvcfPod2 := test.TestNvcfFunctionChildPod.DeepCopy()
		nvcfPod2.Name = nvcfPod2.Name + "-2"
		nvcfPod2.Status.Phase = corev1.PodFailed
		err := cachedClient.Create(ctx, nvcfPod1)
		Expect(err).To(BeNil())
		err = cachedClient.Create(ctx, nvcfPod2)
		Expect(err).To(BeNil())

		err = nvcfTopOwnerHandler.Parse(ctx, nvcf, externalWorkload)
		Expect(err).To(BeNil())

		Expect(externalWorkload.Status.ObservedStatus.Phase).To(Equal(v1alpha1.ExWRunningPhase))
	})

	It("Status - two pods - running but not ready and failed -> initializing", func() {
		nvcfPod1 := test.TestNvcfFunctionChildPod.DeepCopy()
		nvcfPod1.Status.Phase = corev1.PodRunning
		nvcfPod2 := test.TestNvcfFunctionChildPod.DeepCopy()
		nvcfPod2.Name = nvcfPod2.Name + "-2"
		nvcfPod2.Status.Phase = corev1.PodFailed
		err := cachedClient.Create(ctx, nvcfPod1)
		Expect(err).To(BeNil())
		err = cachedClient.Create(ctx, nvcfPod2)
		Expect(err).To(BeNil())

		err = nvcfTopOwnerHandler.Parse(ctx, nvcf, externalWorkload)
		Expect(err).To(BeNil())

		Expect(externalWorkload.Status.ObservedStatus.Phase).To(Equal(v1alpha1.ExWInitializingPhase))
	})

	It("Status - two pods - pending and failed -> pending", func() {
		nvcfPod1 := test.TestNvcfFunctionChildPod.DeepCopy()
		nvcfPod1.Status.Phase = corev1.PodPending
		nvcfPod2 := test.TestNvcfFunctionChildPod.DeepCopy()
		nvcfPod2.Name = nvcfPod2.Name + "-2"
		nvcfPod2.Status.Phase = corev1.PodFailed
		err := cachedClient.Create(ctx, nvcfPod1)
		Expect(err).To(BeNil())
		err = cachedClient.Create(ctx, nvcfPod2)
		Expect(err).To(BeNil())

		err = nvcfTopOwnerHandler.Parse(ctx, nvcf, externalWorkload)
		Expect(err).To(BeNil())

		Expect(externalWorkload.Status.ObservedStatus.Phase).To(Equal(v1alpha1.ExWPendingPhase))
	})

	It("Status - two pods - failed and unknown -> failed", func() {
		nvcfPod1 := test.TestNvcfFunctionChildPod.DeepCopy()
		nvcfPod1.Status.Phase = corev1.PodFailed
		nvcfPod2 := test.TestNvcfFunctionChildPod.DeepCopy()
		nvcfPod2.Name = nvcfPod2.Name + "-2"
		nvcfPod2.Status.Phase = corev1.PodUnknown
		err := cachedClient.Create(ctx, nvcfPod1)
		Expect(err).To(BeNil())
		err = cachedClient.Create(ctx, nvcfPod2)
		Expect(err).To(BeNil())

		err = nvcfTopOwnerHandler.Parse(ctx, nvcf, externalWorkload)
		Expect(err).To(BeNil())

		Expect(externalWorkload.Status.ObservedStatus.Phase).To(Equal(v1alpha1.ExWFailedPhase))
	})

	It("Status - two pods - unknown and unknown -> unknown", func() {
		nvcfPod1 := test.TestNvcfFunctionChildPod.DeepCopy()
		nvcfPod1.Status.Phase = corev1.PodUnknown
		nvcfPod2 := test.TestNvcfFunctionChildPod.DeepCopy()
		nvcfPod2.Name = nvcfPod2.Name + "-2"
		nvcfPod2.Status.Phase = corev1.PodUnknown
		err := cachedClient.Create(ctx, nvcfPod1)
		Expect(err).To(BeNil())
		err = cachedClient.Create(ctx, nvcfPod2)
		Expect(err).To(BeNil())

		err = nvcfTopOwnerHandler.Parse(ctx, nvcf, externalWorkload)
		Expect(err).To(BeNil())

		Expect(externalWorkload.Status.ObservedStatus.Phase).To(Equal(v1alpha1.ExWUnknownPhase))
	})

})
