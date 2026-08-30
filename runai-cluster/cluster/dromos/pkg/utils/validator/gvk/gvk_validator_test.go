package gvk

import (
	"context"
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAttachedResource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GVK Validation Suite")
}

var _ = Describe("GVK Validator", func() {
	var (
		ctx       context.Context
		validator *GVKValidator
	)

	BeforeEach(func() {
		ctx = context.Background()
		validator = NewGVKValidator()
	})

	Describe("Validator Name", func() {
		It("returns the correct validator name", func() {
			Expect(validator.Name()).To(Equal("GVKValidator"))
		})
	})

	Describe("AttachedResource Name Validation", func() {
		It("accepts valid attached resource name", func() {
			runner := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runner",
					Namespace: "default",
				},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Template: newPodTemplate("test-pod", "default"),
				},
			}

			_, errs := validator.Validate(ctx, nil, runner)

			Expect(errs).To(BeEmpty())
		})

		It("rejects missing attached resource name", func() {
			runner := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runner",
					Namespace: "default",
				},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Template: newPodTemplate("", "default"),
				},
			}

			_, errs := validator.Validate(ctx, nil, runner)

			Expect(errs).NotTo(BeEmpty())
			Expect(errs[0].Detail).To(Equal("template name must be set"))
		})
	})

	Describe("AttachedResource Namespace Validation", func() {
		It("accepts valid attached resource namespace", func() {
			runner := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runner",
					Namespace: "default",
				},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Template: newPodTemplate("test-pod", "default"),
				},
			}

			_, errs := validator.Validate(ctx, nil, runner)

			Expect(errs).To(BeEmpty())
		})

		It("rejects missing attached resource namespace", func() {
			runner := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runner",
					Namespace: "default",
				},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Template: newPodTemplate("test-pod", ""),
				},
			}

			_, errs := validator.Validate(ctx, nil, runner)

			Expect(errs).NotTo(BeEmpty())
			Expect(errs[0].Detail).To(Equal("template namespace must be set"))
		})
	})

	Describe("AttachedResource GVK Validation", func() {
		It("accepts valid GVK", func() {
			runner := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runner",
					Namespace: "default",
				},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Template: newPodTemplate("test-pod", "default"),
				},
			}

			_, errs := validator.Validate(ctx, nil, runner)

			Expect(errs).To(BeEmpty())
		})

		It("rejects missing kind", func() {
			runner := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runner",
					Namespace: "default",
				},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Template: newTemplateWithoutKind(),
				},
			}

			_, errs := validator.Validate(ctx, nil, runner)

			Expect(errs).NotTo(BeEmpty())
			Expect(errs[0].Detail).To(Equal("failed to parse template: template missing required apiVersion or kind fields"))
		})

		It("rejects missing apiVersion", func() {
			runner := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runner",
					Namespace: "default",
				},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Template: newTemplateWithoutAPIVersion(),
				},
			}

			_, errs := validator.Validate(ctx, nil, runner)

			Expect(errs).NotTo(BeEmpty())
			Expect(errs[0].Detail).To(Equal("failed to parse template: template missing required apiVersion or kind fields"))
		})
	})
})

func newPodTemplate(name, namespace string) runtime.RawExtension {
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test",
					Image: "nginx",
				},
			},
		},
	}

	data, _ := json.Marshal(pod)
	return runtime.RawExtension{Raw: data}
}

func newTemplateWithoutKind() runtime.RawExtension {
	template := map[string]interface{}{
		"apiVersion": "v1",
		"metadata": map[string]interface{}{
			"name":      "test",
			"namespace": "default",
		},
	}

	data, _ := json.Marshal(template)
	return runtime.RawExtension{Raw: data}
}

func newTemplateWithoutAPIVersion() runtime.RawExtension {
	template := map[string]interface{}{
		"kind": "Pod",
		"metadata": map[string]interface{}{
			"name":      "test",
			"namespace": "default",
		},
	}

	data, _ := json.Marshal(template)
	return runtime.RawExtension{Raw: data}
}
