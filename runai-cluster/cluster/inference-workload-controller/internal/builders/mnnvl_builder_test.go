package builders

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("MNNVL node pool support", func() {
	Describe("nodePoolsSupportMnnvl", func() {
		// On the KAI CRD the detection result is an annotation, so anything other than a
		// value that parses as true has to read as not detected — defaulting the other way
		// would place workloads on nodes without the interconnect.
		kaiPool := func(name string, annotations map[string]string) kaiv1alpha1.NodePool {
			return kaiv1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: name, Annotations: annotations},
			}
		}
		detection := func(value string) map[string]string {
			return map[string]string{kaiv1alpha1.AnnotationGPUNetworkAccelerationDetected: value}
		}

		DescribeTable("reads the detection annotation",
			func(annotations map[string]string, expected bool) {
				pools := []kaiv1alpha1.NodePool{kaiPool("np1", annotations)}

				Expect(nodePoolsSupportMnnvl(pools)).To(Equal(expected))
			},
			Entry("detected", detection("true"), true),
			Entry("not detected", detection("false"), false),
			Entry("empty value", detection(""), false),
			Entry("unparseable value", detection("yes-please"), false),
			Entry("annotation absent", map[string]string{}, false),
			Entry("no annotations at all", nil, false),
		)

		It("supports the workload only when every pool is detected", func() {
			pools := []kaiv1alpha1.NodePool{
				kaiPool("np1", detection("true")),
				kaiPool("np2", detection("true")),
			}

			Expect(nodePoolsSupportMnnvl(pools)).To(BeTrue())
		})

		It("rejects the workload when any single pool is not detected", func() {
			pools := []kaiv1alpha1.NodePool{
				kaiPool("np1", detection("true")),
				kaiPool("np2", detection("false")),
				kaiPool("np3", detection("true")),
			}

			Expect(nodePoolsSupportMnnvl(pools)).To(BeFalse())
		})

		// An empty list means the workload named no pools that had to be checked, which is
		// not the same as a pool failing the check.
		It("supports the workload when there are no pools to check", func() {
			Expect(nodePoolsSupportMnnvl(nil)).To(BeTrue())
		})
	})

})
