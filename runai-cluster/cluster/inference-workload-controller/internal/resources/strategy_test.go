package resources

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"strings"
)

var _ = Describe("Metadata Strategy Update", func() {
	var (
		existingPod *corev1.Pod
		newPod      *corev1.Pod
		config      *MetadataStrategyConfig
	)

	BeforeEach(func() {
		existingPod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      map[string]string{"existing": "value", "new": "old-value"},
				Annotations: map[string]string{"existing": "annotation"},
				OwnerReferences: []metav1.OwnerReference{
					{UID: "existing-uid", Name: "existing-owner"},
					{UID: "new-uid", Name: "new-owner", Controller: ptr.To(false)},
				},
				Finalizers: []string{"existing-finalizer"},
			},
		}
		newPod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Labels:          map[string]string{"new": "value"},
				Annotations:     map[string]string{"new": "annotation"},
				OwnerReferences: []metav1.OwnerReference{{UID: "new-uid", Name: "new-owner", Controller: ptr.To(true)}},
				Finalizers:      []string{"new-finalizer"},
			},
		}
	})

	Describe("handleMetadataStrategy", func() {
		Context("with Override strategy", func() {
			BeforeEach(func() {
				config = &MetadataStrategyConfig{
					Labels:      Override,
					Annotations: Override,
					Finalizers:  Override,
				}
			})

			It("should override all metadata", func() {
				deepCopy := newPod.DeepCopy()
				handleMetadataStrategy(newPod, existingPod, config)
				Expect(newPod.Labels).To(Equal(deepCopy.Labels))
				Expect(newPod.Annotations).To(Equal(deepCopy.Annotations))
				Expect(newPod.Finalizers).To(Equal(deepCopy.Finalizers))
			})
		})

		Context("with Merge strategy", func() {
			BeforeEach(func() {
				config = &MetadataStrategyConfig{
					Labels:      Merge,
					Annotations: Merge,
					Finalizers:  Merge,
				}
			})

			It("should merge all metadata", func() {
				handleMetadataStrategy(newPod, existingPod, config)

				Expect(newPod.Labels).To(HaveLen(2))
				Expect(newPod.Labels).To(HaveKeyWithValue("existing", "value"))
				Expect(newPod.Labels).To(HaveKeyWithValue("new", "value"))

				Expect(newPod.Annotations).To(HaveLen(2))
				Expect(newPod.Annotations).To(HaveKeyWithValue("existing", "annotation"))
				Expect(newPod.Annotations).To(HaveKeyWithValue("new", "annotation"))

				Expect(newPod.Finalizers).To(HaveLen(2))
			})
		})

		Context("with Ignored strategy", func() {
			BeforeEach(func() {
				config = &MetadataStrategyConfig{
					Labels:      Ignored,
					Annotations: Ignored,
					Finalizers:  Ignored,
				}
			})

			It("should not change any metadata", func() {
				originalPod := existingPod.DeepCopy()
				handleMetadataStrategy(newPod, existingPod, config)

				Expect(newPod.Labels).To(Equal(originalPod.Labels))
				Expect(newPod.Annotations).To(Equal(originalPod.Annotations))
				Expect(newPod.Finalizers).To(Equal(originalPod.Finalizers))
			})
		})
	})

	Describe("mergeLabels", func() {
		Context("when both resources have nil labels", func() {
			It("should result in nil labels", func() {
				existingPod.Labels = nil
				newPod.Labels = nil
				mergeLabels(existingPod, newPod)
				Expect(newPod.Labels).To(BeEmpty())
			})
		})

		Context("when existing resource has nil labels", func() {
			It("should use all labels from new resource", func() {
				existingPod.Labels = nil
				newPod.Labels = map[string]string{"key1": "value1", "key2": "value2"}
				mergeLabels(existingPod, newPod)
				Expect(newPod.Labels).To(HaveLen(2))
				Expect(newPod.Labels).To(HaveKeyWithValue("key1", "value1"))
				Expect(newPod.Labels).To(HaveKeyWithValue("key2", "value2"))
			})
		})

		Context("when new resource has nil labels", func() {
			It("should keep all labels from existing resource", func() {
				existingPod.Labels = map[string]string{"key1": "value1", "key2": "value2"}
				newPod.Labels = nil
				mergeLabels(existingPod, newPod)
				Expect(newPod.Labels).To(HaveLen(2))
				Expect(newPod.Labels).To(HaveKeyWithValue("key1", "value1"))
				Expect(newPod.Labels).To(HaveKeyWithValue("key2", "value2"))
			})
		})

		Context("when both resources have labels", func() {
			It("should merge labels, with new resource taking precedence for same keys", func() {
				existingPod.Labels = map[string]string{"key1": "old1", "key2": "old2", "key3": "old3"}
				newPod.Labels = map[string]string{"key1": "new1", "key2": "new2", "key4": "new4"}
				mergeLabels(existingPod, newPod)
				Expect(newPod.Labels).To(HaveLen(4))
				Expect(newPod.Labels).To(HaveKeyWithValue("key1", "new1"))
				Expect(newPod.Labels).To(HaveKeyWithValue("key2", "new2"))
				Expect(newPod.Labels).To(HaveKeyWithValue("key3", "old3"))
				Expect(newPod.Labels).To(HaveKeyWithValue("key4", "new4"))
			})
		})

		Context("when dealing with empty string values", func() {
			It("should handle empty string values correctly", func() {
				existingPod.Labels = map[string]string{"key1": "", "key2": "value2"}
				newPod.Labels = map[string]string{"key1": "new1", "key3": ""}
				mergeLabels(existingPod, newPod)
				Expect(newPod.Labels).To(HaveLen(3))
				Expect(newPod.Labels).To(HaveKeyWithValue("key1", "new1"))
				Expect(newPod.Labels).To(HaveKeyWithValue("key2", "value2"))
				Expect(newPod.Labels).To(HaveKeyWithValue("key3", ""))
			})
		})

		Context("when dealing with special characters in keys and values", func() {
			It("should preserve special characters", func() {
				existingPod.Labels = map[string]string{"key-with-dash": "value-with-dash", "key.with.dots": "value.with.dots"}
				newPod.Labels = map[string]string{"key_with_underscore": "value_with_underscore", "key@with@at": "value@with@at"}
				mergeLabels(existingPod, newPod)
				Expect(newPod.Labels).To(HaveLen(4))
				Expect(newPod.Labels).To(HaveKeyWithValue("key-with-dash", "value-with-dash"))
				Expect(newPod.Labels).To(HaveKeyWithValue("key.with.dots", "value.with.dots"))
				Expect(newPod.Labels).To(HaveKeyWithValue("key_with_underscore", "value_with_underscore"))
				Expect(newPod.Labels).To(HaveKeyWithValue("key@with@at", "value@with@at"))
			})
		})

		Context("when dealing with a large number of labels", func() {
			It("should handle merging many labels efficiently", func() {
				existingPod.Labels = make(map[string]string)
				newPod.Labels = make(map[string]string)
				for i := 0; i < 1000; i++ {
					existingPod.Labels[fmt.Sprintf("existing-key-%d", i)] = fmt.Sprintf("existing-value-%d", i)
					newPod.Labels[fmt.Sprintf("new-key-%d", i)] = fmt.Sprintf("new-value-%d", i)
				}
				mergeLabels(existingPod, newPod)
				Expect(newPod.Labels).To(HaveLen(2000))
				for i := 0; i < 1000; i++ {
					Expect(newPod.Labels).To(HaveKeyWithValue(fmt.Sprintf("existing-key-%d", i), fmt.Sprintf("existing-value-%d", i)))
					Expect(newPod.Labels).To(HaveKeyWithValue(fmt.Sprintf("new-key-%d", i), fmt.Sprintf("new-value-%d", i)))
				}
			})
		})

		Context("when dealing with label name and value length limits", func() {
			It("should preserve labels at the maximum allowed length", func() {
				maxKeyLength := strings.Repeat("a", 63)   // Kubernetes allows up to 63 characters for label keys
				maxValueLength := strings.Repeat("b", 63) // Kubernetes allows up to 63 characters for label values
				existingPod.Labels = map[string]string{maxKeyLength: maxValueLength}
				newPod.Labels = map[string]string{"short-key": "short-value"}
				mergeLabels(existingPod, newPod)
				Expect(newPod.Labels).To(HaveLen(2))
				Expect(newPod.Labels).To(HaveKeyWithValue(maxKeyLength, maxValueLength))
				Expect(newPod.Labels).To(HaveKeyWithValue("short-key", "short-value"))
			})
		})
	})

	Describe("mergeAnnotations", func() {
		Context("when both resources have nil annotations", func() {
			It("should result in nil annotations", func() {
				existingPod.Annotations = nil
				newPod.Annotations = nil
				mergeAnnotations(existingPod, newPod)
				Expect(newPod.Annotations).To(BeEmpty())
			})
		})

		Context("when existing resource has nil annotations", func() {
			It("should use all annotations from new resource", func() {
				existingPod.Annotations = nil
				newPod.Annotations = map[string]string{"anno1": "value1", "anno2": "value2"}
				mergeAnnotations(existingPod, newPod)
				Expect(newPod.Annotations).To(HaveLen(2))
				Expect(newPod.Annotations).To(HaveKeyWithValue("anno1", "value1"))
				Expect(newPod.Annotations).To(HaveKeyWithValue("anno2", "value2"))
			})
		})

		Context("when new resource has nil annotations", func() {
			It("should keep all annotations from existing resource", func() {
				existingPod.Annotations = map[string]string{"anno1": "value1", "anno2": "value2"}
				newPod.Annotations = nil
				mergeAnnotations(existingPod, newPod)
				Expect(newPod.Annotations).To(HaveLen(2))
				Expect(newPod.Annotations).To(HaveKeyWithValue("anno1", "value1"))
				Expect(newPod.Annotations).To(HaveKeyWithValue("anno2", "value2"))
			})
		})

		Context("when both resources have annotations", func() {
			It("should merge annotations, with new resource taking precedence for same keys", func() {
				existingPod.Annotations = map[string]string{"anno1": "old1", "anno2": "old2", "anno3": "old3"}
				newPod.Annotations = map[string]string{"anno1": "new1", "anno2": "new2", "anno4": "new4"}
				mergeAnnotations(existingPod, newPod)
				Expect(newPod.Annotations).To(HaveLen(4))
				Expect(newPod.Annotations).To(HaveKeyWithValue("anno1", "new1"))
				Expect(newPod.Annotations).To(HaveKeyWithValue("anno2", "new2"))
				Expect(newPod.Annotations).To(HaveKeyWithValue("anno3", "old3"))
				Expect(newPod.Annotations).To(HaveKeyWithValue("anno4", "new4"))
			})
		})

		Context("when dealing with empty string values", func() {
			It("should handle empty string values correctly", func() {
				existingPod.Annotations = map[string]string{"anno1": "", "anno2": "value2"}
				newPod.Annotations = map[string]string{"anno1": "new1", "anno3": ""}
				mergeAnnotations(existingPod, newPod)
				Expect(newPod.Annotations).To(HaveLen(3))
				Expect(newPod.Annotations).To(HaveKeyWithValue("anno1", "new1"))
				Expect(newPod.Annotations).To(HaveKeyWithValue("anno2", "value2"))
				Expect(newPod.Annotations).To(HaveKeyWithValue("anno3", ""))
			})
		})

		Context("when dealing with special characters in keys and values", func() {
			It("should preserve special characters", func() {
				existingPod.Annotations = map[string]string{
					"anno-with-dash": "value-with-dash",
					"anno.with.dots": "value.with.dots",
				}
				newPod.Annotations = map[string]string{
					"anno_with_underscore": "value_with_underscore",
					"anno@with@at":         "value@with@at",
				}
				mergeAnnotations(existingPod, newPod)
				Expect(newPod.Annotations).To(HaveLen(4))
				Expect(newPod.Annotations).To(HaveKeyWithValue("anno-with-dash", "value-with-dash"))
				Expect(newPod.Annotations).To(HaveKeyWithValue("anno.with.dots", "value.with.dots"))
				Expect(newPod.Annotations).To(HaveKeyWithValue("anno_with_underscore", "value_with_underscore"))
				Expect(newPod.Annotations).To(HaveKeyWithValue("anno@with@at", "value@with@at"))
			})
		})

		Context("when dealing with a large number of annotations", func() {
			It("should handle merging many annotations efficiently", func() {
				existingPod.Annotations = make(map[string]string)
				newPod.Annotations = make(map[string]string)
				for i := 0; i < 1000; i++ {
					existingPod.Annotations[fmt.Sprintf("existing-anno-%d", i)] = fmt.Sprintf("existing-value-%d", i)
					newPod.Annotations[fmt.Sprintf("new-anno-%d", i)] = fmt.Sprintf("new-value-%d", i)
				}
				mergeAnnotations(existingPod, newPod)
				Expect(newPod.Annotations).To(HaveLen(2000))
				for i := 0; i < 1000; i++ {
					Expect(newPod.Annotations).To(HaveKeyWithValue(fmt.Sprintf("existing-anno-%d", i), fmt.Sprintf("existing-value-%d", i)))
					Expect(newPod.Annotations).To(HaveKeyWithValue(fmt.Sprintf("new-anno-%d", i), fmt.Sprintf("new-value-%d", i)))
				}
			})
		})

		Context("when dealing with annotation name and value length limits", func() {
			It("should preserve annotations at the maximum allowed length", func() {
				// Kubernetes allows up to 253 characters for annotation keys
				maxKeyLength := strings.Repeat("a", 253)
				// Annotation values can be much longer, but let's use a reasonable large size
				maxValueLength := strings.Repeat("b", 1024)
				existingPod.Annotations = map[string]string{maxKeyLength: maxValueLength}
				newPod.Annotations = map[string]string{"short-anno": "short-value"}
				mergeAnnotations(existingPod, newPod)
				Expect(newPod.Annotations).To(HaveLen(2))
				Expect(newPod.Annotations).To(HaveKeyWithValue(maxKeyLength, maxValueLength))
				Expect(newPod.Annotations).To(HaveKeyWithValue("short-anno", "short-value"))
			})
		})

		Context("when dealing with Kubernetes-specific annotations", func() {
			It("should handle Kubernetes-specific annotations correctly", func() {
				existingPod.Annotations = map[string]string{
					"kubectl.kubernetes.io/last-applied-configuration": `{"apiVersion":"v1","kind":"Pod","metadata":{"name":"example"}}`,
				}
				newPod.Annotations = map[string]string{
					"kubernetes.io/change-cause": "Update pod configuration",
				}
				mergeAnnotations(existingPod, newPod)
				Expect(newPod.Annotations).To(HaveLen(2))
				Expect(newPod.Annotations).To(HaveKey("kubectl.kubernetes.io/last-applied-configuration"))
				Expect(newPod.Annotations).To(HaveKeyWithValue("kubernetes.io/change-cause", "Update pod configuration"))
			})
		})

		Context("when dealing with JSON-encoded annotation values", func() {
			It("should preserve JSON-encoded values", func() {
				jsonValue := `{"key1":"value1","key2":{"nestedKey":"nestedValue"}}`
				existingPod.Annotations = map[string]string{"json-anno": jsonValue}
				newPod.Annotations = map[string]string{"other-anno": "other-value"}
				mergeAnnotations(existingPod, newPod)
				Expect(newPod.Annotations).To(HaveLen(2))
				Expect(newPod.Annotations).To(HaveKeyWithValue("json-anno", jsonValue))
				Expect(newPod.Annotations).To(HaveKeyWithValue("other-anno", "other-value"))
			})
		})
	})

	Describe("mergeFinalizers", func() {
		Context("when both resources have nil finalizers", func() {
			It("should result in nil finalizers", func() {
				existingPod.Finalizers = nil
				newPod.Finalizers = nil
				mergeFinalizers(existingPod, newPod)
				Expect(newPod.Finalizers).To(BeNil())
			})
		})

		Context("when existing resource has nil finalizers", func() {
			It("should use all finalizers from new resource", func() {
				existingPod.Finalizers = nil
				newPod.Finalizers = []string{"finalizer1", "finalizer2"}
				mergeFinalizers(existingPod, newPod)
				Expect(newPod.Finalizers).To(Equal([]string{"finalizer1", "finalizer2"}))
			})
		})

		Context("when new resource has nil finalizers", func() {
			It("should keep all finalizers from existing resource", func() {
				existingPod.Finalizers = []string{"finalizer1", "finalizer2"}
				newPod.Finalizers = nil
				mergeFinalizers(existingPod, newPod)
				Expect(newPod.Finalizers).To(Equal([]string{"finalizer1", "finalizer2"}))
			})
		})

		Context("when new resource has empty finalizers slice", func() {
			It("should keep all finalizers from existing resource", func() {
				existingPod.Finalizers = []string{"finalizer1", "finalizer2"}
				newPod.Finalizers = []string{}
				mergeFinalizers(existingPod, newPod)
				Expect(newPod.Finalizers).To(Equal([]string{"finalizer1", "finalizer2"}))
			})
		})

		Context("when both resources have finalizers", func() {
			It("should merge finalizers, preserving order and removing duplicates", func() {
				existingPod.Finalizers = []string{"finalizer1", "finalizer2", "finalizer3"}
				newPod.Finalizers = []string{"finalizer2", "finalizer4", "finalizer1"}
				mergeFinalizers(existingPod, newPod)
				Expect(newPod.Finalizers).To(Equal([]string{"finalizer1", "finalizer2", "finalizer3", "finalizer4"}))
			})
		})

		Context("when new finalizers are a subset of existing finalizers", func() {
			It("should preserve all existing finalizers", func() {
				existingPod.Finalizers = []string{"finalizer1", "finalizer2", "finalizer3"}
				newPod.Finalizers = []string{"finalizer2"}
				mergeFinalizers(existingPod, newPod)
				Expect(newPod.Finalizers).To(Equal([]string{"finalizer1", "finalizer2", "finalizer3"}))
			})
		})

		Context("when existing finalizers are a subset of new finalizers", func() {
			It("should add new finalizers while preserving existing ones", func() {
				existingPod.Finalizers = []string{"finalizer2"}
				newPod.Finalizers = []string{"finalizer1", "finalizer2", "finalizer3"}
				mergeFinalizers(existingPod, newPod)
				Expect(newPod.Finalizers).To(Equal([]string{"finalizer1", "finalizer2", "finalizer3"}))
			})
		})

		Context("when dealing with case-sensitive finalizers", func() {
			It("should treat different cases as distinct finalizers", func() {
				existingPod.Finalizers = []string{"Finalizer1", "finalizer2"}
				newPod.Finalizers = []string{"finalizer1", "Finalizer2"}
				mergeFinalizers(existingPod, newPod)
				Expect(newPod.Finalizers).To(Equal([]string{"Finalizer1", "Finalizer2", "finalizer1", "finalizer2"}))
			})
		})

		Context("when dealing with very long finalizer names", func() {
			It("should handle long finalizer names correctly", func() {
				longFinalizer := strings.Repeat("a", 253) // Max length for DNS label
				existingPod.Finalizers = []string{longFinalizer}
				newPod.Finalizers = []string{"short-finalizer"}
				mergeFinalizers(existingPod, newPod)
				Expect(newPod.Finalizers).To(Equal([]string{longFinalizer, "short-finalizer"}))
			})
		})
	})
})
