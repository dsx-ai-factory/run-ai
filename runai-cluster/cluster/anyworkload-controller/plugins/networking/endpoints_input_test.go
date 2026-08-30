package networking

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	endpoints "github.com/run-ai/runai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("classifyInput", func() {
	newAnnotatedRunner := func(rawAnnotation string) *dromosv1alpha1.WorkloadRunner {
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testWorkloadName,
				Namespace: testNamespace,
				Annotations: map[string]string{
					endpoints.Annotation: rawAnnotation,
				},
			},
		}
		return runner
	}

	validPayload := func() string {
		raw := `{"schemaVersion":"v1","endpoints":[{"services":[{"create":{"metadata":{"name":"jupyter"},"spec":{"ports":[{"port":8888}]}}}]}]}`
		return raw
	}

	It("returns inputNone when neither field nor annotation is set", func() {
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{Name: testWorkloadName, Namespace: testNamespace},
		}
		shape, cfg, err := classifyInput(runner)
		Expect(err).ToNot(HaveOccurred())
		Expect(shape).To(Equal(inputNone))
		Expect(cfg).To(BeNil())
	})

	It("returns inputLegacy when only spec.instructions.networking is set", func() {
		runner := newTestRunner(testWorkloadName, testNamespace, &dromosv1alpha1.NetworkingConfig{
			Services: []dromosv1alpha1.NetworkingServiceItem{{
				Create: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc"}},
			}},
		})
		shape, cfg, err := classifyInput(runner)
		Expect(err).ToNot(HaveOccurred())
		Expect(shape).To(Equal(inputLegacy))
		Expect(cfg).To(BeNil())
	})

	It("returns inputAnnotation when only run.ai/endpoints is set with a well-formed v1 payload", func() {
		runner := newAnnotatedRunner(validPayload())
		shape, cfg, err := classifyInput(runner)
		Expect(err).ToNot(HaveOccurred())
		Expect(shape).To(Equal(inputAnnotation))
		Expect(cfg).ToNot(BeNil())
		Expect(cfg.SchemaVersion).To(Equal(endpoints.SchemaVersionV1))
		Expect(cfg.Endpoints).To(HaveLen(1))
	})

	It("returns an error and inputNone when the annotation JSON is malformed", func() {
		runner := newAnnotatedRunner(`{"schemaVersion":"v1","endpoints":[`)
		shape, cfg, err := classifyInput(runner)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("malformed"))
		Expect(shape).To(Equal(inputNone))
		Expect(cfg).To(BeNil())
	})

	It("returns an error and inputNone when the annotation schemaVersion is unsupported", func() {
		runner := newAnnotatedRunner(`{"schemaVersion":"v9","endpoints":[]}`)
		shape, cfg, err := classifyInput(runner)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unsupported"))
		Expect(shape).To(Equal(inputNone))
		Expect(cfg).To(BeNil())
	})

	It("returns an error and inputNone when both the legacy field and the annotation are set", func() {
		runner := newTestRunner(testWorkloadName, testNamespace, &dromosv1alpha1.NetworkingConfig{
			Services: []dromosv1alpha1.NetworkingServiceItem{{
				Create: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc"}},
			}},
		})
		runner.Annotations = map[string]string{endpoints.Annotation: validPayload()}

		shape, cfg, err := classifyInput(runner)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("mutually exclusive"))
		Expect(shape).To(Equal(inputNone))
		Expect(cfg).ToNot(BeNil(), "parsed config is surfaced alongside the mutual-exclusivity error for caller convenience")
	})
})
