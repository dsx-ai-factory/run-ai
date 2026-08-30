package runaimetrics

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRunaiMetrics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Runai metrics suite")
}

var _ = Describe("vLLM metric names", func() {
	Context("per-workload and per-pod wrappers", func() {
		It("uses the recording-rule naming convention", func() {
			names := []VLLMMetricName{
				VLLMNumRequestsRunning,
				VLLMNumRequestsWaiting,
				VLLMRequestSuccess,
				VLLMRequestSuccessRate,
				VLLMKVCacheUsagePerc,
				VLLMPromptTokensPerSecond,
				VLLMGenerationTokensPerSecond,
				VLLMTimeToFirstTokenSecondsBucket,
				VLLME2ERequestLatencySecondsBucket,
				VLLMTimeToFirstTokenSecondsPercentile,
				VLLME2ERequestLatencySecondsPercentile,
				VLLMRequestQueueTimeSecondsPercentile,
				VLLMRequestPrefillTimeSecondsPercentile,
				VLLMRequestDecodeTimeSecondsPercentile,
				VLLMInterTokenLatencySecondsPercentile,
			}

			for _, name := range names {
				workloadName := RunaiVLLMPerWorkloadMetric{Name: name}.GetName()
				podName := RunaiVLLMPerPodMetric{Name: name}.GetName()
				Expect(workloadName).To(Equal("runai_vllm_" + string(name) + "_per_workload"))
				Expect(podName).To(Equal("runai_vllm_" + string(name) + "_per_pod"))
			}
		})
	})
})
